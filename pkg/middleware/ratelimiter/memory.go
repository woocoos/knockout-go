package ratelimiter

import (
	"hash/fnv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// rateLimitEntry 跟踪每个限流 key 的状态（key 可以是用户、IP、租户等）
type rateLimitEntry struct {
	ts     int64
	tokens uint
}

// shardMap 是分片并发 map，用于减少锁竞争。
// 不使用单一全局锁，而是使用多个分片，使不同 key 可以并发处理，最小化竞争。
type shardMap struct {
	shards [16]struct {
		mu   sync.Mutex
		data map[string]rateLimitEntry
	}
}

func (sm *shardMap) getShard(key string) *struct {
	mu   sync.Mutex
	data map[string]rateLimitEntry
} {
	// 使用 FNV-32a 哈希以获得更好的分片分布。
	// 简单的首字节哈希 (key[0] % 16) 在 key 有相同前缀时会导致冲突
	// （例如用户 ID 都以 "1" 开头，IP 地址都以 "192" 开头）。
	// 注意：空 key 在实际中不会出现（KeySkip 对空 key 返回 true），
	// 但这里防御性地使用 shard 0
	if len(key) == 0 {
		return &sm.shards[0]
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	idx := h.Sum32() % 16
	return &sm.shards[idx]
}

func newShardMap() *shardMap {
	sm := &shardMap{}
	for i := range sm.shards {
		sm.shards[i].data = make(map[string]rateLimitEntry)
	}
	return sm
}

// SafeInMemoryStore 是完全线程安全的内存存储实现，
// 使用分片锁减少高并发下的竞争，
// 比单一全局锁提供更好的性能。
type SafeInMemoryStore struct {
	rate  int64 // 速率窗口（秒）
	limit uint
	data  *shardMap // 分片 map，用于并发访问
	skip  func(*gin.Context) bool
}

// NewSafeInMemoryStore 创建一个新的线程安全的内存限流存储
func NewSafeInMemoryStore(options *InMemoryOptions) *SafeInMemoryStore {
	store := &SafeInMemoryStore{
		rate:  int64(options.Rate.Seconds()),
		limit: options.Limit,
		skip:  options.Skip,
		data:  newShardMap(),
	}
	// 启动后台清理 goroutine
	go store.clearInBackground()
	return store
}

// Limit 实现 Store 接口，带有正确的并发控制。
// 使用分片锁最小化竞争：不同 key 大概率落在不同分片，允许并行处理。
func (s *SafeInMemoryStore) Limit(key string, c *gin.Context) Info {
	shard := s.data.getShard(key)
	shard.mu.Lock()
	defer shard.mu.Unlock()

	u, ok := shard.data[key]
	if !ok {
		u = rateLimitEntry{ts: time.Now().Unix(), tokens: s.limit}
	}

	// 如果速率窗口已过期，重置令牌
	if u.ts+s.rate <= time.Now().Unix() {
		u.ts = time.Now().Unix()
		u.tokens = s.limit
	}

	// 检查跳过条件
	if s.skip != nil && s.skip(c) {
		return Info{
			Limit:         s.limit,
			RateLimited:   false,
			ResetTime:     time.Now().Add(time.Duration((s.rate - (time.Now().Unix() - u.ts)) * int64(time.Second))),
			RemainingHits: u.tokens,
		}
	}

	// 检查是否已被限流
	if u.tokens <= 0 {
		return Info{
			Limit:         s.limit,
			RateLimited:   true,
			ResetTime:     time.Now().Add(time.Duration((s.rate - (time.Now().Unix() - u.ts)) * int64(time.Second))),
			RemainingHits: 0,
		}
	}

	// 消耗一个令牌
	u.tokens--
	shard.data[key] = u

	return Info{
		Limit:         s.limit,
		RateLimited:   false,
		ResetTime:     time.Now().Add(time.Duration((s.rate - (time.Now().Unix() - u.ts)) * int64(time.Second))),
		RemainingHits: u.tokens,
	}
}

func (s *SafeInMemoryStore) clearInBackground() {
	for {
		time.Sleep(time.Minute)
		now := time.Now().Unix()
		for i := range s.data.shards {
			s.data.shards[i].mu.Lock()
			for k, u := range s.data.shards[i].data {
				if u.ts+s.rate <= now {
					delete(s.data.shards[i].data, k)
				}
			}
			s.data.shards[i].mu.Unlock()
		}
	}
}

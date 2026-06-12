package ratelimiter

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// redisRateLimitScript 是原子性检查和更新限流状态的 Lua 脚本。
// 修复了 gin-rate-limit RedisStore 的两个问题：
// 1. 竞态条件：Pipeline (Get→Check→Set) 非原子操作；Lua 脚本在 Redis 中是原子的。
// 2. 窗口滑动：原实现每次请求都重置 ts；本实现仅在窗口过期时重置 ts。
var redisRateLimitScript = redis.NewScript(`
local key = KEYS[1]
local tsKey = KEYS[2]
local rate = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local ttl = rate * 2

local ts = redis.call('GET', tsKey)
local hits = redis.call('GET', key)
local windowExpired = false

if not ts or (tonumber(ts) + rate) <= now then
    ts = now
    windowExpired = true
    redis.call('SET', tsKey, ts, 'EX', ttl)
    hits = 0
else
    ts = tonumber(ts)
    hits = tonumber(hits or '0')
end

if hits >= limit then
    return {1, limit, 0, ts}
end

hits = hits + 1
if windowExpired then
    redis.call('SET', key, hits, 'EX', ttl)
else
    redis.call('INCR', key)
    redis.call('EXPIRE', key, ttl)
end

local remaining = limit - hits
if remaining < 0 then remaining = 0 end
return {0, limit, remaining, ts}
`)

// SafeRedisStore 是使用 Lua 脚本的线程安全的 Redis 存储实现。
// 使用 Lua 脚本确保原子性，修复了：
// 1. 竞态条件：Lua 脚本确保 Redis 中原子读取-检查-写入。
// 2. 窗口滑动：ts 仅在新窗口开始时设置，而非每次请求。
type SafeRedisStore struct {
	client *redis.Client
	rate   int64
	limit  uint
	skip   func(*gin.Context) bool
}

// NewSafeRedisStore 创建一个新的线程安全的 Redis 限流存储
func NewSafeRedisStore(options *RedisOptions) *SafeRedisStore {
	return &SafeRedisStore{
		client: options.RedisClient,
		rate:   int64(options.Rate.Seconds()),
		limit:  options.Limit,
		skip:   options.Skip,
	}
}

// Limit 实现 Store 接口，使用原子 Lua 脚本（无需互斥锁）
func (s *SafeRedisStore) Limit(key string, c *gin.Context) Info {
	now := time.Now().Unix()

	if s.skip != nil && s.skip(c) {
		return Info{
			Limit:         s.limit,
			RateLimited:   false,
			ResetTime:     time.Now().Add(time.Duration(s.rate) * time.Second),
			RemainingHits: s.limit,
		}
	}

	result, err := redisRateLimitScript.Run(c.Request.Context(), s.client,
		[]string{key, key + ":ts"},
		s.rate, s.limit, now,
	).Slice()
	if err != nil {
		return Info{
			Limit:         s.limit,
			RateLimited:   false,
			ResetTime:     time.Now().Add(time.Duration(s.rate) * time.Second),
			RemainingHits: s.limit,
		}
	}

	rateLimited := toInt64(result[0]) == 1
	limit := uint(toInt64(result[1]))
	remaining := uint(toInt64(result[2]))
	ts := toInt64(result[3])

	return Info{
		Limit:         limit,
		RateLimited:   rateLimited,
		ResetTime:     time.Unix(ts, 0).Add(time.Duration(s.rate) * time.Second),
		RemainingHits: remaining,
	}
}

// toInt64 将 Redis Lua 脚本的返回值转换为 int64。
// Redis 根据数值范围返回 int64 或 string 类型。
func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		n, _ := strconv.ParseInt(val, 10, 64)
		return n
	default:
		return 0
	}
}

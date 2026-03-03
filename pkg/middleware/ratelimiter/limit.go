// Package ratelimiter Web API 限流器.
// 限流方式:
// - user: 基于用户名
// - tenant: 基于租户
// - ip: 基于客户端IP
// 同时支持指定的key,应用于或排除出限流策略,以方便针对指定用户才限流,或某用户不限流.
// 配置:
// rateLimiter:
//
//	inMemoryOptions:
//	  rate: 1s
//	  limit: 100
//	redisOptions:
//	  rate: 1s
//	  limit: 100
//	  panicOnError: false
//	exclude:       # 排除路径，这些路径不限流
//	  - /health
//	  - /metrics
//	options:
//	  storeKey: default  # if use cache manager.
//	  keyFunc: user  # user, tenant, ip
//	  includeKeys:   # 只对这些 key 应用限流（如果设置，其他 key 不限流）
//	    - user1
//	    - user2
//	  excludeKeys:   # 这些 key 不限流（优先级高于 includeKeys）
//	    - admin
//	    - system
package ratelimiter

import (
	"errors"
	"fmt"
	"strconv"

	ratelimit "github.com/JGLTechnologies/gin-rate-limit"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/tsingsun/woocoo/pkg/cache"
	"github.com/tsingsun/woocoo/pkg/cache/redisc"
	"github.com/tsingsun/woocoo/pkg/conf"
	"github.com/tsingsun/woocoo/pkg/security"
	"github.com/tsingsun/woocoo/web"
	"github.com/tsingsun/woocoo/web/handler"
	"github.com/woocoos/knockout-go/pkg/identity"
)

// ErrRedisClientNotFound is returned when redis client is not found in cache manager
var ErrRedisClientNotFound = errors.New("redis client not found for rate limiter")

// Config 限流配置
type Config struct {
	// 不需要限流的接口
	Exclude []string `json:"exclude"`
	// 根据Exclude产生的Path.PathSkipper
	Skipper handler.Skipper
	// RedisOptions store为redis时的配置
	RedisOptions *ratelimit.RedisOptions `json:"redisOptions"`
	// InMemoryOptions store为memory时的配置
	InMemoryOptions *ratelimit.InMemoryOptions `json:"inMemoryOptions"`
	// StoreKey the key of cache manager.
	StoreKey string `json:"storeKey"`
	// KeyFunc
	KeyFunc string `json:"keyFunc"`
	// IncludeKeys 这些 key 应用限流（如果设置，其他 key 不限流）
	IncludeKeys []string `json:"includeKeys"`
	// ExcludeKeys 这些 key 不限流（优先级高于 includeKeys）
	ExcludeKeys            []string `json:"excludeKeys"`
	includeMap, excludeMap map[string]struct{}
}

func (mid *Config) Name() string {
	return "rateLimiter"
}

func (mid *Config) ApplyFunc(cfg *conf.Configuration) gin.HandlerFunc {
	if err := cfg.Unmarshal(mid); err != nil {
		panic(err)
	}
	if mid.Skipper == nil {
		mid.Skipper = handler.PathSkipper(mid.Exclude)
	}
	// Convert to map for O(1) lookup
	mid.includeMap = make(map[string]struct{})
	for _, k := range mid.IncludeKeys {
		mid.includeMap[k] = struct{}{}
	}
	mid.excludeMap = make(map[string]struct{})
	for _, k := range mid.ExcludeKeys {
		mid.excludeMap[k] = struct{}{}
	}
	var store ratelimit.Store
	switch {
	case mid.RedisOptions != nil:
		// Get Redis client from cache manager
		client, err := getRedisClientFromCache(mid.StoreKey)
		if err != nil {
			panic(fmt.Sprintf("redis client not found for rate limiter: %s,%v", mid.StoreKey, err))
		}
		mid.RedisOptions.RedisClient = client
		store = ratelimit.RedisStore(mid.RedisOptions)
	default:
		// Default to memory store
		store = ratelimit.InMemoryStore(mid.InMemoryOptions)
	}

	opts := &ratelimit.Options{}
	opts.KeyFunc = getKeyFunc(mid.KeyFunc)

	handlerFunc := mid.RateLimiter(store, opts)
	return func(c *gin.Context) {
		if mid.Skipper(c) {
			return
		}
		handlerFunc(c)
	}
}

// KeySkip 检查key是否应该被忽略
func (mid *Config) KeySkip(key string) bool {
	if key == "" {
		// If key is empty, skip rate limiting
		return true
	}

	// Check exclude first (higher priority)
	if _, excluded := mid.excludeMap[key]; excluded {
		return true // Skip rate limiting for excluded keys
	}

	// Check include list
	if len(mid.includeMap) > 0 {
		if _, included := mid.includeMap[key]; !included {
			return true // Skip rate limiting for keys not in include list
		}
	}
	return false // Apply rate limiting
}

// RateLimiter is a function to get gin.HandlerFunc.
// Base Logic is copied from ratelimit.RateLimiter
func (mid *Config) RateLimiter(s ratelimit.Store, options *ratelimit.Options) gin.HandlerFunc {
	if options == nil {
		options = &ratelimit.Options{}
	}
	if options.ErrorHandler == nil {
		options.ErrorHandler = func(c *gin.Context, info ratelimit.Info) {
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", info.Limit))
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
			c.String(429, "Too many requests")
		}
	}
	if options.BeforeResponse == nil {
		options.BeforeResponse = func(c *gin.Context, info ratelimit.Info) {
			c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%d", info.Limit))
			c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%v", info.RemainingHits))
			c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", info.ResetTime.Unix()))
		}
	}
	if options.KeyFunc == nil {
		options.KeyFunc = func(c *gin.Context) string {
			return c.ClientIP() + c.FullPath()
		}
	}
	return func(c *gin.Context) {
		key := options.KeyFunc(c)
		if mid.KeySkip(key) {
			return
		}
		info := s.Limit(key, c)
		options.BeforeResponse(c, info)
		if c.IsAborted() {
			return
		}
		if info.RateLimited {
			options.ErrorHandler(c, info)
			c.Abort()
		} else {
			c.Next()
		}
	}
}

// keyFuncUser returns a key function that uses the user name from security context
func keyFuncUser() func(*gin.Context) string {
	return func(c *gin.Context) string {
		p, ok := security.FromContext(c)
		if !ok {
			return ""
		}
		return p.Identity().Name()
	}
}

// keyFuncTenant returns a key function that uses the tenant ID from context
func keyFuncTenant() func(*gin.Context) string {
	return func(c *gin.Context) string {
		tid, err := identity.TenantIDFromContext(c)
		if err != nil {
			return ""
		}
		return strconv.Itoa(tid)
	}
}

// getKeyFunc returns the key function based on the name
func getKeyFunc(name string) func(*gin.Context) string {
	switch name {
	case "user":
		return keyFuncUser()
	case "tenant":
		return keyFuncTenant()
	default:
		return nil
	}
}

// getRedisClientFromCache tries to get redis client from cache manager
func getRedisClientFromCache(storeKey string) (*redis.Client, error) {
	cacheInst, err := cache.GetCache(storeKey)
	if err != nil {
		return nil, err
	}
	if rc, ok := cacheInst.(*redisc.Redisc); ok {
		if client, ok := rc.RedisClient().(*redis.Client); ok {
			return client, nil
		}
	}
	return nil, ErrRedisClientNotFound
}

// RegisterMiddleware register rate limit middleware
func RegisterMiddleware() web.Option {
	return web.WithMiddlewareNewFunc("rateLimit", func() handler.Middleware {
		return &Config{}
	})
}

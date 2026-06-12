package ratelimiter

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// Info 限流结果信息
type Info struct {
	// Limit 限流窗口内允许的最大请求数
	Limit uint
	// RateLimited 是否被限流
	RateLimited bool
	// ResetTime 限流窗口重置时间
	ResetTime time.Time
	// RemainingHits 剩余可用请求数
	RemainingHits uint
}

// Store 限流存储接口
type Store interface {
	Limit(key string, c *gin.Context) Info
}

// Options 限流器选项
type Options struct {
	// KeyFunc 从请求中提取限流 key 的函数
	KeyFunc func(*gin.Context) string
	// ErrorHandler 被限流时的响应处理函数
	ErrorHandler func(*gin.Context, Info)
	// BeforeResponse 每次请求的响应前处理函数（设置 header 等）
	BeforeResponse func(*gin.Context, Info)
}

// InMemoryOptions 内存存储配置
type InMemoryOptions struct {
	// Rate 限流窗口时长
	Rate time.Duration
	// Limit 窗口内允许的最大请求数
	Limit uint
	// Skip 跳过限流的条件判断函数
	Skip func(*gin.Context) bool
}

// RedisOptions Redis 存储配置
type RedisOptions struct {
	// RedisClient Redis 客户端实例
	RedisClient *redis.Client
	// Rate 限流窗口时长
	Rate time.Duration
	// Limit 窗口内允许的最大请求数
	Limit uint
	// Skip 跳过限流的条件判断函数
	Skip func(*gin.Context) bool
}

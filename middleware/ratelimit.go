// package middleware 提供 Gin 中间件。
//
// 本文件实现接口限流器，支持两种后端：
//   - memory：单实例内存固定窗口限流，适合开发或单节点部署
//   - redis：基于 Redis + Lua 脚本的原子固定窗口限流，适合多实例分布式部署
//
// 限流维度：客户端 IP。触发限流时返回 429 Too Many Requests。
package middleware

import (
	"blog/config"
	"blog/database"
	"blog/utils"
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor 记录某个 key 在当前窗口内的请求计数与窗口过期时间（内存模式）。
type visitor struct {
	count   int
	expires time.Time
}

// MemoryRateLimiter 是内存固定窗口限流器。
type MemoryRateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	requests int
	window   time.Duration
}

// NewMemoryRateLimiter 创建一个内存固定窗口限流器。
func NewMemoryRateLimiter(requests int, window time.Duration) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		visitors: make(map[string]*visitor),
		requests: requests,
		window:   window,
	}
}

// Allow 判断指定 key 是否允许通过。
func (rl *MemoryRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]

	// 窗口已过期，重置计数
	if !exists || now.After(v.expires) {
		rl.visitors[key] = &visitor{
			count:   1,
			expires: now.Add(rl.window),
		}
		return true
	}

	// 窗口内计数 +1
	v.count++
	return v.count <= rl.requests
}

// cleanup 清理过期的 visitor 记录，防止内存无限增长。
func (rl *MemoryRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, v := range rl.visitors {
		if now.After(v.expires) {
			delete(rl.visitors, key)
		}
	}
}

// redisRateLimitScript 是 Redis 限流 Lua 脚本。
// 原子化执行：INCR key -> 如果是首次设置 EXPIRE -> 判断是否超过阈值。
// 返回 1 表示允许通过，0 表示触发限流。
const redisRateLimitScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end
if current > limit then
    return 0
end
return 1
`

// RedisRateLimiter 是 Redis 分布式固定窗口限流器。
type RedisRateLimiter struct {
	requests int
	window   time.Duration
}

// NewRedisRateLimiter 创建一个 Redis 分布式固定窗口限流器。
func NewRedisRateLimiter(requests int, window time.Duration) *RedisRateLimiter {
	return &RedisRateLimiter{
		requests: requests,
		window:   window,
	}
}

// Allow 判断指定 key 是否允许通过。
// Redis 不可用时降级放行，避免影响正常业务。
func (rl *RedisRateLimiter) Allow(key string) bool {
	if database.Redis == nil {
		return true
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	fullKey := "blog:ratelimit:" + key
	result, err := database.Redis.Eval(ctx, redisRateLimitScript, []string{fullKey}, rl.requests, int(rl.window.Seconds())).Result()
	if err != nil {
		utils.Logger.Errorf("Redis 限流执行失败: %v", err)
		return true // 降级放行
	}

	allowed, ok := result.(int64)
	if !ok {
		return true
	}
	return allowed == 1
}

// RateLimiter 是限流器接口。
type RateLimiter interface {
	Allow(key string) bool
}

// RateLimit 返回 Gin 限流中间件。
//
// 限流规则从 config.C.RateLimit 读取：
//   - Enabled: 是否启用
//   - Mode: "memory" 或 "redis"，配置为 redis 但 Redis 不可用时自动降级为 memory
//   - Requests: 每个窗口最大请求数
//   - WindowSec: 窗口长度（秒）
//
// 触发限流时返回 429，并在响应头中携带 Retry-After。
func RateLimit() gin.HandlerFunc {
	cfg := config.C.RateLimit
	if !cfg.Enabled || cfg.Requests <= 0 || cfg.WindowSec <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	window := time.Duration(cfg.WindowSec) * time.Second
	var limiter RateLimiter

	if cfg.Mode == "redis" && database.Redis != nil {
		limiter = NewRedisRateLimiter(cfg.Requests, window)
	} else {
		memLimiter := NewMemoryRateLimiter(cfg.Requests, window)
		limiter = memLimiter

		// 启动后台 goroutine 定期清理过期记录
		go func() {
			ticker := time.NewTicker(window)
			defer ticker.Stop()
			for range ticker.C {
				memLimiter.cleanup()
			}
		}()
	}

	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			key = "unknown"
		}

		if !limiter.Allow(key) {
			c.Header("Retry-After", strconv.Itoa(cfg.WindowSec))
			utils.Error(c, utils.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

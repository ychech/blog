// package middleware 提供 Gin 中间件。
//
// 本文件实现基于固定窗口的接口限流器：
//   - 以客户端 IP 为维度统计请求次数
//   - 在配置的时间窗口内超过阈值时返回 429 Too Many Requests
//   - 仅适用于单实例部署；多实例部署需改用 Redis 等分布式限流方案
package middleware

import (
	"blog/config"
	"blog/utils"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// visitor 记录某个 IP 在当前窗口内的请求计数与窗口过期时间。
type visitor struct {
	count   int
	expires time.Time
}

// RateLimiter 是固定窗口限流器。
// 使用 map 在内存中维护每个客户端 IP 的请求计数。
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	requests int
	window   time.Duration
}

// NewRateLimiter 创建一个固定窗口限流器。
//
// 参数 requests 是单个窗口内允许的最大请求数；
// 参数 window 是窗口时长。
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*visitor),
		requests: requests,
		window:   window,
	}
}

// Allow 判断指定 key 是否允许通过。
// 返回 true 表示未触发限流；返回 false 表示已触发限流。
func (rl *RateLimiter) Allow(key string) bool {
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
	if v.count > rl.requests {
		return false
	}
	return true
}

// cleanup 清理过期的 visitor 记录，防止内存无限增长。
// 可由后台 goroutine 定期调用，也可以在 Allow 中按需清理。
func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, v := range rl.visitors {
		if now.After(v.expires) {
			delete(rl.visitors, key)
		}
	}
}

// RateLimit 返回 Gin 限流中间件。
//
// 限流规则从 config.C.RateLimit 读取：
//   - Enabled: 是否启用
//   - Requests: 每个窗口最大请求数
//   - WindowSec: 窗口长度（秒）
//
// 触发限流时返回 429，并在响应头中携带 Retry-After 提示客户端多久后重试。
func RateLimit() gin.HandlerFunc {
	cfg := config.C.RateLimit
	if !cfg.Enabled || cfg.Requests <= 0 || cfg.WindowSec <= 0 {
		// 配置关闭或非法时，直接放行，不影响请求链路
		return func(c *gin.Context) {
			c.Next()
		}
	}

	rl := NewRateLimiter(cfg.Requests, time.Duration(cfg.WindowSec)*time.Second)

	// 启动后台 goroutine 定期清理过期记录
	window := time.Duration(cfg.WindowSec) * time.Second
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return func(c *gin.Context) {
		key := c.ClientIP()
		if key == "" {
			key = "unknown"
		}

		if !rl.Allow(key) {
			c.Header("Retry-After", strconv.Itoa(cfg.WindowSec))
			utils.Error(c, utils.CodeTooManyRequests, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

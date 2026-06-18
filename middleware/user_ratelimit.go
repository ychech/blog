// package middleware 提供 Gin 中间件。
//
// 本文件实现按用户维度的接口限流，登录用户使用 userID，未登录用户回退到 IP。
package middleware

import (
	"blog/config"
	"blog/database"
	"blog/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserRateLimit 返回按用户维度限流的 Gin 中间件。
// 应挂载在需要登录的接口组上，或放在 JWTAuth 之后；这样能从 context 读取 userID。
// 未登录时回退到客户端 IP。
func UserRateLimit() gin.HandlerFunc {
	cfg := config.C.UserRateLimit
	if !cfg.Enabled || cfg.Requests <= 0 || cfg.WindowSec <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	window := time.Duration(cfg.WindowSec) * time.Second
	var limiter RateLimiter

	if database.Redis != nil {
		limiter = NewRedisRateLimiter(cfg.Requests, window)
	} else {
		memLimiter := NewMemoryRateLimiter(cfg.Requests, window)
		limiter = memLimiter

		go func() {
			ticker := time.NewTicker(window)
			defer ticker.Stop()
			for range ticker.C {
				memLimiter.cleanup()
			}
		}()
	}

	return func(c *gin.Context) {
		key := "user_"
		if userID, ok := GetCurrentUserID(c); ok {
			key += strconv.FormatUint(uint64(userID), 10)
		} else {
			ip := c.ClientIP()
			if ip == "" {
				ip = "unknown"
			}
			key += "ip_" + ip
		}

		if !limiter.Allow(key) {
			c.Header("Retry-After", strconv.Itoa(cfg.WindowSec))
			utils.Error(c, utils.CodeTooManyRequests, utils.T(utils.GetLocale(c), "request_too_frequent"))
			c.Abort()
			return
		}
		c.Next()
	}
}

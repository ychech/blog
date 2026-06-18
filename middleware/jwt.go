// package middleware 提供 Gin 中间件：JWT 认证、请求日志、panic 恢复、跨域处理。
package middleware

import (
	"blog/model"
	"blog/service"
	"blog/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件。
// 从请求头 Authorization: Bearer <token> 中提取 token 并校验；
// 校验通过后将 userID 与 username 写入 gin.Context，供后续 handler 使用。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "缺少 Authorization 请求头")
			c.Abort()
			return
		}

		// 支持 "Bearer token" 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "Authorization 格式错误，应为 Bearer <token>")
			c.Abort()
			return
		}

		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			utils.Unauthorized(c, "Token 无效或已过期")
			c.Abort()
			return
		}

		// 校验 Token 是否已被拉黑
		blacklisted, err := service.IsTokenBlacklisted(claims.ID)
		if err != nil {
			utils.Unauthorized(c, "Token 校验失败")
			c.Abort()
			return
		}
		if blacklisted {
			utils.Unauthorized(c, "Token 已失效")
			c.Abort()
			return
		}

		// 将用户信息存入 context，后续 handler 可以通过 c.GetUint("userID") 获取
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("jti", claims.ID)
		c.Next()
	}
}

// GetCurrentUserRole 从 gin context 获取当前用户角色
func GetCurrentUserRole(c *gin.Context) (model.UserRole, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	r, ok := role.(model.UserRole)
	return r, ok
}

// GetCurrentUserID 从 gin context 获取当前用户 ID
func GetCurrentUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("userID")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// package middleware 提供 Gin 中间件：JWT 认证、请求日志、panic 恢复、跨域处理。
package middleware

import (
	"blog/model"
	"blog/utils"

	"github.com/gin-gonic/gin"
)

// AdminAuth 管理员权限中间件。
// 必须在 JWTAuth 之后使用；从 JWT claims 中读取角色，避免每次请求都查数据库。
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetCurrentUserRole(c)
		if !ok {
			utils.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		if role != model.UserRoleAdmin {
			utils.Error(c, utils.CodeForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		c.Set("isAdmin", true)
		c.Next()
	}
}

// IsAdmin 从 gin context 获取当前用户是否为管理员
func IsAdmin(c *gin.Context) bool {
	// 优先读取 AdminAuth 设置的标记
	isAdmin, exists := c.Get("isAdmin")
	if exists {
		admin, ok := isAdmin.(bool)
		if ok && admin {
			return true
		}
	}

	// 兜底：根据角色判断
	role, ok := GetCurrentUserRole(c)
	return ok && role == model.UserRoleAdmin
}

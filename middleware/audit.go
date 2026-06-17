// package middleware 提供 Gin 中间件。
//
// 本文件实现操作审计日志中间件，记录管理员的关键写操作。
package middleware

import (
	"blog/model"
	"blog/service"
	"blog/utils"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuditLog 返回审计日志中间件。
// 仅记录管理员用户的 POST/PUT/DELETE 请求，便于合规审计。
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只记录写操作
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "DELETE" && method != "PATCH" {
			return
		}

		// 只记录管理员操作
		role, exists := c.Get("role")
		if !exists || role != model.UserRoleAdmin {
			return
		}

		userID, _ := GetCurrentUserID(c)
		username, _ := c.Get("username")
		usernameStr, _ := username.(string)
		if usernameStr == "" {
			usernameStr = "unknown"
		}

		action := actionFromMethod(method)
		resource, resourceID := parseResource(c.Request.URL.Path)
		details := c.Request.URL.Path

		go func() {
			if err := service.CreateAuditLog(userID, usernameStr, action, resource, resourceID, details, c.ClientIP()); err != nil {
				utils.Logger.Errorf("写入审计日志失败: %v", err)
			}
		}()
	}
}

func actionFromMethod(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return method
	}
}

// parseResource 从请求路径中解析资源类型和 ID。
// 例如：/api/posts/123 -> ("post", 123)；/api/categories -> ("category", 0)
func parseResource(path string) (string, uint) {
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return "", 0
	}

	resource := parts[0]
	// 去掉末尾的 s，简单单数化
	if strings.HasSuffix(resource, "s") && len(resource) > 1 {
		resource = resource[:len(resource)-1]
	}

	if len(parts) >= 2 {
		if id, err := strconv.ParseUint(parts[1], 10, 64); err == nil {
			return resource, uint(id)
		}
	}
	return resource, 0
}

// package middleware 提供 Gin 中间件。
//
// 本文件实现语言偏好解析中间件。
package middleware

import (
	"blog/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// Locale 解析请求中的语言偏好。
// 优先读取 query 参数 lang，其次读取 Accept-Language 请求头。
func Locale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.Query("lang")
		if locale == "" {
			locale = c.GetHeader("Accept-Language")
		}
		locale = normalizeLocale(locale)
		c.Set("locale", locale)
		c.Next()
	}
}

func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if strings.HasPrefix(locale, "zh") {
		return "zh"
	}
	if strings.HasPrefix(locale, "en") {
		return "en"
	}
	return utils.DefaultLocale
}

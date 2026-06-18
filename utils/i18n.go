// package utils 提供项目通用工具。
//
// 本文件实现基础国际化（i18n）支持：根据请求中的语言标识返回对应文案。
package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// DefaultLocale 默认语言。
const DefaultLocale = "zh"

var translations = map[string]map[string]string{
	"zh": {
		"request_too_frequent": "请求过于频繁，请稍后再试",
		"route_not_found":      "接口不存在",
		"unauthorized":         "请先登录",
		"forbidden":            "无权访问",
		"server_error":         "服务器内部错误",
	},
	"en": {
		"request_too_frequent": "Too many requests, please try again later",
		"route_not_found":      "API not found",
		"unauthorized":         "Please login first",
		"forbidden":            "Forbidden",
		"server_error":         "Internal server error",
	},
}

// T 根据语言和 key 返回翻译文案。
// 如果找不到对应语言或 key，则返回 key 本身。
func T(locale, key string) string {
	locale = normalizeLocale(locale)
	if msgs, ok := translations[locale]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return key
}

// GetLocale 从 gin context 获取当前语言。
func GetLocale(c *gin.Context) string {
	if locale, exists := c.Get("locale"); exists {
		if s, ok := locale.(string); ok && s != "" {
			return s
		}
	}
	return DefaultLocale
}

// normalizeLocale 规范化语言标识，目前只支持 zh/en。
func normalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if strings.HasPrefix(locale, "zh") {
		return "zh"
	}
	if strings.HasPrefix(locale, "en") {
		return "en"
	}
	return DefaultLocale
}

package handler

import (
	"blog/config"
	"blog/service"
	"blog/utils"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkWebsocketOrigin,
}

// checkWebsocketOrigin 校验 WebSocket 跨域来源。
// 未配置 BaseURL 或开发环境（localhost）允许所有来源；生产环境仅允许 BaseURL 对应域名。
func checkWebsocketOrigin(r *http.Request) bool {
	if config.C == nil || config.C.App.BaseURL == "" {
		return true
	}

	base, err := url.Parse(config.C.App.BaseURL)
	if err != nil {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}

	// 开发环境放宽限制
	if strings.Contains(base.Host, "localhost") || strings.Contains(base.Host, "127.0.0.1") {
		return true
	}

	o, err := url.Parse(origin)
	if err != nil {
		return false
	}
	return o.Host == base.Host
}

// NotificationWebSocket 建立实时通知 WebSocket 连接。
// 客户端需在 query 中携带 token：/ws/notifications?token=<jwt>
func NotificationWebSocket(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		utils.BadRequest(c, "缺少 token 参数")
		return
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		utils.Unauthorized(c, "Token 无效或已过期")
		return
	}

	blacklisted, err := service.IsTokenBlacklisted(claims.ID)
	if err != nil {
		utils.Unauthorized(c, "Token 校验失败")
		return
	}
	if blacklisted {
		utils.Unauthorized(c, "Token 已失效")
		return
	}

	isActive, err := service.CheckUserActive(claims.UserID)
	if err != nil || !isActive {
		utils.Unauthorized(c, "账号已被禁用")
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		utils.Logger.Errorf("WebSocket 升级失败: %v", err)
		return
	}

	service.RegisterWSClient(claims.UserID, conn)
}

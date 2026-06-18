package handler

import (
	"blog/service"
	"blog/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 生产环境建议根据前端域名做白名单校验；开发环境允许跨域。
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
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

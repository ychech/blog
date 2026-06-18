// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现站内信（私信）接口。
package handler

import (
	"blog/middleware"
	"blog/model"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// MessageHandler 私信处理器。
type MessageHandler struct{}

// NewMessageHandler 创建私信处理器。
func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

// Send 发送私信（需要登录）。
// @Summary 发送私信
// @Tags 私信
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.SendMessageRequest true "接收者 ID 与内容"
// @Success 200 {object} utils.Response{data=model.Message}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /messages [post]
func (h *MessageHandler) Send(c *gin.Context) {
	senderID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var req model.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	msg, err := service.SendMessage(senderID, req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, msg)
}

// ListConversations 获取当前用户的会话列表（需要登录）。
// @Summary 获取会话列表
// @Tags 私信
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Router /messages/conversations [get]
func (h *MessageHandler) ListConversations(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListConversations(userID, page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// ListMessages 获取与指定用户的私信记录（需要登录）。
// @Summary 获取私信记录
// @Tags 私信
// @Produce json
// @Security BearerAuth
// @Param user_id path int true "对方用户 ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Router /messages/{user_id} [get]
func (h *MessageHandler) ListMessages(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	otherID, err := strconv.Atoi(c.Param("user_id"))
	if err != nil || otherID <= 0 {
		utils.BadRequest(c, "用户 ID 无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListMessages(userID, uint(otherID), page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// CountUnread 获取当前用户未读私信数（需要登录）。
// @Summary 获取未读私信数
// @Tags 私信
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=map[string]int64}
// @Failure 401 {object} utils.Response
// @Router /messages/unread-count [get]
func (h *MessageHandler) CountUnread(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	count, err := service.CountUnreadMessages(userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, gin.H{"count": count})
}

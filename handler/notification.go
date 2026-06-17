// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现通知相关的 HTTP 接口：列表查询、标记已读、未读数统计。
package handler

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NotificationHandler 通知处理器。
type NotificationHandler struct{}

// NewNotificationHandler 创建通知处理器。
func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

// List 获取当前登录用户的通知列表。
// @Summary 获取通知列表
// @Tags 通知
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Router /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	resp, err := service.ListNotifications(userID, page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// MarkAsRead 将指定通知标记为已读。
// @Summary 标记通知为已读
// @Tags 通知
// @Produce json
// @Security BearerAuth
// @Param id path int true "通知 ID"
// @Success 200 {object} utils.Response
// @Router /notifications/{id}/read [put]
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "通知 ID 无效")
		return
	}

	if err := service.MarkNotificationAsRead(userID, uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "已标记为已读"})
}

// MarkAllAsRead 将当前登录用户的所有通知标记为已读。
// @Summary 标记所有通知为已读
// @Tags 通知
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=map[string]int64}
// @Router /notifications/read-all [put]
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	affected, err := service.MarkAllNotificationsAsRead(userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, gin.H{"affected": affected})
}

// Delete 删除指定通知。
// @Summary 删除通知
// @Tags 通知
// @Produce json
// @Security BearerAuth
// @Param id path int true "通知 ID"
// @Success 200 {object} utils.Response
// @Router /notifications/{id} [delete]
func (h *NotificationHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "通知 ID 无效")
		return
	}

	if err := service.DeleteNotification(userID, uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, nil)
}

// CountUnread 获取当前登录用户的未读通知数。
// @Summary 获取未读通知数
// @Tags 通知
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=map[string]int64}
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) CountUnread(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	count, err := service.CountUnreadNotifications(userID)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, gin.H{"count": count})
}

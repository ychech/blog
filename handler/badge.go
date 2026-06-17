// package handler 负责处理 HTTP 请求：解析参数、调用 service、返回统一响应。
// 不直接操作数据库，所有业务逻辑委托给 service 层。
package handler

import (
	"blog/middleware"
	"blog/model"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BadgeHandler 勋章处理器
type BadgeHandler struct {
	service *service.BadgeService
}

// NewBadgeHandler 创建勋章处理器
func NewBadgeHandler() *BadgeHandler {
	return &BadgeHandler{service: service.NewBadgeService()}
}

// Create 创建勋章（管理员）
// @Summary 创建勋章（管理员）
// @Tags 勋章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreateBadgeRequest true "勋章信息"
// @Success 201 {object} utils.Response{data=model.Badge}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /badges [post]
func (h *BadgeHandler) Create(c *gin.Context) {
	var req model.CreateBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	badge, err := h.service.Create(req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, badge)
}

// List 获取所有勋章
// @Summary 获取所有勋章
// @Tags 勋章
// @Produce json
// @Success 200 {object} utils.Response{data=[]model.Badge}
// @Router /badges [get]
func (h *BadgeHandler) List(c *gin.Context) {
	badges, err := h.service.List()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, badges)
}

// Get 获取勋章详情
// @Summary 获取勋章详情
// @Tags 勋章
// @Produce json
// @Param id path int true "勋章 ID"
// @Success 200 {object} utils.Response{data=model.Badge}
// @Failure 404 {object} utils.Response
// @Router /badges/{id} [get]
func (h *BadgeHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "勋章 ID 格式错误")
		return
	}

	badge, err := h.service.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "勋章不存在")
		return
	}

	utils.Success(c, badge)
}

// Update 更新勋章（管理员）
// @Summary 更新勋章（管理员）
// @Tags 勋章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "勋章 ID"
// @Param body body model.UpdateBadgeRequest true "勋章信息"
// @Success 200 {object} utils.Response{data=model.Badge}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /badges/{id} [put]
func (h *BadgeHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "勋章 ID 格式错误")
		return
	}

	var req model.UpdateBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	badge, err := h.service.Update(uint(id), req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, badge)
}

// Delete 删除勋章（管理员）
// @Summary 删除勋章（管理员）
// @Tags 勋章
// @Security BearerAuth
// @Param id path int true "勋章 ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /badges/{id} [delete]
func (h *BadgeHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "勋章 ID 格式错误")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, nil)
}

// Award 颁发勋章给用户（管理员）
// @Summary 颁发勋章给用户（管理员）
// @Tags 勋章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.AwardBadgeRequest true "用户 ID、勋章 ID 与原因"
// @Success 200 {object} utils.Response{data=model.UserBadge}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /badges/award [post]
func (h *BadgeHandler) Award(c *gin.Context) {
	var req model.AwardBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	userBadge, err := h.service.Award(req.UserID, req.BadgeID, req.Reason)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, userBadge)
}

// GetUserBadges 获取指定用户的勋章列表
// @Summary 获取指定用户的勋章列表
// @Tags 勋章
// @Produce json
// @Param id path int true "用户 ID"
// @Success 200 {object} utils.Response{data=[]model.UserBadge}
// @Router /users/{id}/badges [get]
func (h *BadgeHandler) GetUserBadges(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "用户 ID 格式错误")
		return
	}

	userBadges, err := h.service.GetUserBadges(uint(id))
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, userBadges)
}

// GetMyBadges 获取当前登录用户的勋章列表
// @Summary 获取当前登录用户的勋章列表
// @Tags 勋章
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=[]model.UserBadge}
// @Failure 401 {object} utils.Response
// @Router /auth/badges [get]
func (h *BadgeHandler) GetMyBadges(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	userBadges, err := h.service.GetUserBadges(userID)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, userBadges)
}

// Revoke 收回用户勋章（管理员）
// @Summary 收回用户勋章（管理员）
// @Tags 勋章
// @Security BearerAuth
// @Param id path int true "用户勋章 ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /user-badges/{id} [delete]
func (h *BadgeHandler) Revoke(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "用户勋章 ID 格式错误")
		return
	}

	if err := h.service.Revoke(uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, nil)
}

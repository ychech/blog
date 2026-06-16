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
func (h *BadgeHandler) List(c *gin.Context) {
	badges, err := h.service.List()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, badges)
}

// Get 获取勋章详情
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

// package handler 负责处理 HTTP 请求：解析参数、调用 service、返回统一响应。
// 不直接操作数据库，所有业务逻辑委托给 service 层。
package handler

import (
	"blog/middleware"
	"blog/model"
	"blog/service"
	"blog/utils"

	"github.com/gin-gonic/gin"
)

// AuthHandler 用户认证处理器，处理注册、登录、获取当前用户等接口。
type AuthHandler struct {
	userService *service.UserService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{userService: service.NewUserService()}
}

// Register 用户注册
// @Summary 用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body model.RegisterRequest true "注册信息"
// @Success 200 {object} utils.Response{data=model.LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 1001 {object} utils.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.userService.Register(req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// Login 用户登录
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body model.LoginRequest true "登录凭证"
// @Success 200 {object} utils.Response{data=model.LoginResponse}
// @Failure 400 {object} utils.Response
// @Failure 1001 {object} utils.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	resp, err := h.userService.Login(req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// Me 获取当前登录用户信息
// @Summary 获取当前登录用户信息
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=model.User}
// @Failure 401 {object} utils.Response
// @Router /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "未登录")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// UpdateProfile 更新当前登录用户资料
// @Summary 更新当前登录用户资料
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.UpdateProfileRequest true "用户资料"
// @Success 200 {object} utils.Response{data=model.User}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/me [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req model.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.userService.UpdateProfile(userID, req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, user)
}

// AdminListUsers 管理员获取用户列表
// @Summary 管理员获取用户列表
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /auth/users [get]
func (h *AuthHandler) AdminListUsers(c *gin.Context) {
	var query struct {
		Page     int `form:"page"`
		PageSize int `form:"page_size"`
	}
	c.ShouldBindQuery(&query)

	resp, err := h.userService.List(query.Page, query.PageSize)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, resp)
}

// AdminGetStats 管理员获取站点统计
// @Summary 管理员获取站点统计
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=object{user_count=int,post_count=int,comment_count=int,category_count=int,tag_count=int,badge_count=int}}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /auth/stats [get]
func (h *AuthHandler) AdminGetStats(c *gin.Context) {
	stats, err := h.userService.GetStats()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, stats)
}

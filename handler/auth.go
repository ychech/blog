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

// SendVerificationEmail 发送邮箱验证码（需要登录）
// @Summary 发送邮箱验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/send-verification-email [post]
func (h *AuthHandler) SendVerificationEmail(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	if user.Email == "" {
		utils.BadRequest(c, "请先绑定邮箱")
		return
	}

	if err := service.SendVerificationEmail(user.ID, user.Email); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "验证邮件已发送，请查收"})
}

// VerifyEmail 验证邮箱验证码（需要登录）
// @Summary 验证邮箱验证码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.VerifyEmailRequest true "验证码"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var req model.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := service.VerifyEmail(userID, req.Code); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "邮箱验证成功"})
}

// ChangePassword 修改当前登录用户密码
// @Summary 修改当前登录用户密码
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.ChangePasswordRequest true "原密码与新密码"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /auth/change-password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "未登录")
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.userService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "密码修改成功"})
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

// AdminGetUser 管理员获取用户详情
// @Summary 管理员获取用户详情
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户 ID"
// @Success 200 {object} utils.Response{data=model.User}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Failure 404 {object} utils.Response
// @Router /auth/users/{id} [get]
func (h *AuthHandler) AdminGetUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "用户 ID 格式错误")
		return
	}

	user, err := h.userService.GetUserDetail(uint(id))
	if err != nil {
		utils.NotFound(c, "用户不存在")
		return
	}

	utils.Success(c, user)
}

// AdminUpdateUserRole 管理员更新用户角色
// @Summary 管理员更新用户角色
// @Tags 认证
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户 ID"
// @Param body body model.UpdateUserRoleRequest true "角色"
// @Success 200 {object} utils.Response{data=model.User}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /auth/users/{id}/role [put]
func (h *AuthHandler) AdminUpdateUserRole(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "用户 ID 格式错误")
		return
	}

	var req model.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	user, err := h.userService.UpdateRole(uint(id), req.Role)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, user)
}

// AdminDeleteUser 管理员删除用户
// @Summary 管理员删除用户
// @Tags 认证
// @Security BearerAuth
// @Param id path int true "用户 ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /auth/users/{id} [delete]
func (h *AuthHandler) AdminDeleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "用户 ID 格式错误")
		return
	}

	if err := h.userService.DeleteUser(uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, nil)
}

// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现用户关注/粉丝接口。
package handler

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UserFollowHandler 用户关注处理器。
type UserFollowHandler struct{}

// NewUserFollowHandler 创建用户关注处理器。
func NewUserFollowHandler() *UserFollowHandler {
	return &UserFollowHandler{}
}

// Follow 关注用户（需要登录）。
// @Summary 关注用户
// @Tags 用户关注
// @Security BearerAuth
// @Param id path int true "目标用户 ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /users/{id}/follow [post]
func (h *UserFollowHandler) Follow(c *gin.Context) {
	followerID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "用户 ID 无效")
		return
	}

	if err := service.FollowUser(followerID, uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "关注成功"})
}

// Unfollow 取消关注（需要登录）。
// @Summary 取消关注
// @Tags 用户关注
// @Security BearerAuth
// @Param id path int true "目标用户 ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /users/{id}/follow [delete]
func (h *UserFollowHandler) Unfollow(c *gin.Context) {
	followerID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "用户 ID 无效")
		return
	}

	if err := service.UnfollowUser(followerID, uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "已取消关注"})
}

// Followers 获取用户粉丝列表。
// @Summary 粉丝列表
// @Tags 用户关注
// @Produce json
// @Param id path int true "用户 ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Router /users/{id}/followers [get]
func (h *UserFollowHandler) Followers(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "用户 ID 无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListFollowers(uint(id), page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// Following 获取用户关注列表。
// @Summary 关注列表
// @Tags 用户关注
// @Produce json
// @Param id path int true "用户 ID"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Router /users/{id}/following [get]
func (h *UserFollowHandler) Following(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "用户 ID 无效")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListFollowing(uint(id), page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

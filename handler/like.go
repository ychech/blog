// package handler 负责处理 HTTP 请求：解析参数、调用 service、返回统一响应。
// 不直接操作数据库，所有业务逻辑委托给 service 层。
package handler

import (
	"blog/middleware"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// LikeHandler 点赞处理器
type LikeHandler struct {
	service *service.LikeService
}

// NewLikeHandler 创建点赞处理器
func NewLikeHandler() *LikeHandler {
	return &LikeHandler{service: service.NewLikeService()}
}

// Toggle 切换点赞（需要登录）
// @Summary 切换文章点赞状态
// @Tags 点赞
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response{data=object{liked=bool,count=int}}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts/{id}/like [post]
func (h *LikeHandler) Toggle(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "文章 ID 格式错误")
		return
	}

	liked, err := h.service.Toggle(uint(id), userID)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	count := h.service.GetLikeCount(uint(id))
	utils.Success(c, gin.H{
		"liked": liked,
		"count": count,
	})
}

// Status 获取当前用户对某篇文章的点赞状态（可选登录）
// @Summary 获取文章点赞状态
// @Tags 点赞
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response{data=object{liked=bool,count=int}}
// @Failure 400 {object} utils.Response
// @Router /posts/{id}/like [get]
func (h *LikeHandler) Status(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "文章 ID 格式错误")
		return
	}

	count := h.service.GetLikeCount(uint(id))
	resp := gin.H{
		"count": count,
		"liked": false,
	}

	if userID, ok := middleware.GetCurrentUserID(c); ok {
		resp["liked"] = h.service.IsLiked(uint(id), userID)
	}

	utils.Success(c, resp)
}

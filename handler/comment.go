package handler

import (
	"blog/middleware"
	"blog/model"
	"blog/service"
	"blog/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CommentHandler 评论处理器，支持一级评论与嵌套回复的创建、查询和删除。
type CommentHandler struct {
	service *service.CommentService
}

// NewCommentHandler 创建评论处理器
func NewCommentHandler() *CommentHandler {
	return &CommentHandler{service: service.NewCommentService()}
}

// Create 创建评论（需要登录）
// @Summary 创建评论
// @Tags 评论
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreateCommentRequest true "评论内容"
// @Success 201 {object} utils.Response{data=model.Comment}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /comments [post]
func (h *CommentHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	username, _ := c.Get("username")
	authorName, _ := username.(string)

	var req model.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	comment, err := h.service.Create(userID, authorName, req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.SuccessWithStatus(c, http.StatusCreated, comment)
}

// ListByPost 获取文章评论列表
// @Summary 获取文章评论列表
// @Tags 评论
// @Produce json
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response{data=[]model.Comment}
// @Failure 400 {object} utils.Response
// @Router /posts/{id}/comments [get]
func (h *CommentHandler) ListByPost(c *gin.Context) {
	postID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "文章 ID 格式错误")
		return
	}

	comments, err := h.service.ListByPost(uint(postID))
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, comments)
}

// Update 更新评论（需要登录；管理员可编辑任意评论）
// @Summary 更新评论
// @Tags 评论
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论 ID"
// @Param body body model.UpdateCommentRequest true "评论内容"
// @Success 200 {object} utils.Response{data=model.Comment}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /comments/{id} [put]
func (h *CommentHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "评论 ID 格式错误")
		return
	}

	var req model.UpdateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	comment, err := h.service.Update(uint(id), userID, middleware.IsAdmin(c), req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, comment)
}

// PinComment 置顶/取消置顶评论（管理员）。
// @Summary 置顶/取消置顶评论
// @Tags 评论
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论 ID"
// @Param pinned query bool true "是否置顶"
// @Success 200 {object} utils.Response{data=model.Comment}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/comments/{id}/pin [put]
func (h *CommentHandler) PinComment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "评论 ID 无效")
		return
	}

	pinned, _ := strconv.ParseBool(c.DefaultQuery("pinned", "true"))
	comment, err := h.service.PinComment(uint(id), pinned)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, comment)
}

// EssenceComment 加精/取消加精评论（管理员）。
// @Summary 加精/取消加精评论
// @Tags 评论
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论 ID"
// @Param essence query bool true "是否精华"
// @Success 200 {object} utils.Response{data=model.Comment}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/comments/{id}/essence [put]
func (h *CommentHandler) EssenceComment(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "评论 ID 无效")
		return
	}

	essence, _ := strconv.ParseBool(c.DefaultQuery("essence", "true"))
	comment, err := h.service.EssenceComment(uint(id), essence)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, comment)
}

// Delete 删除评论（需要登录；管理员可删除任意评论）
// @Summary 删除评论
// @Tags 评论
// @Security BearerAuth
// @Param id path int true "评论 ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /comments/{id} [delete]
func (h *CommentHandler) Delete(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "评论 ID 格式错误")
		return
	}

	if err := h.service.Delete(uint(id), userID, middleware.IsAdmin(c)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, nil)
}

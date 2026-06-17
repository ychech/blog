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

// PostHandler 文章处理器，负责文章相关的 HTTP 接口。
type PostHandler struct {
	service *service.PostService
}

// NewPostHandler 创建文章处理器
func NewPostHandler() *PostHandler {
	return &PostHandler{service: service.NewPostService()}
}

// Create 创建文章（需要登录）
// @Summary 创建文章
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreatePostRequest true "文章信息"
// @Success 201 {object} utils.Response{data=model.Post}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var req model.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	post, err := h.service.Create(userID, req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.SuccessWithStatus(c, http.StatusCreated, post)
}

// List 获取文章列表
// @Summary 获取文章列表
// @Tags 文章
// @Produce json
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页数量，默认 10"
// @Param keyword query string false "关键词搜索"
// @Param category_id query int false "分类 ID"
// @Param tag_id query int false "标签 ID"
// @Param status query string false "状态：draft/published"
// @Param order_by query string false "排序：created_at/view_count"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Router /posts [get]
func (h *PostHandler) List(c *gin.Context) {
	var query model.PostQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, "查询参数错误: "+err.Error())
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)
	resp, err := h.service.List(query, userID, middleware.IsAdmin(c))
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, resp)
}

// Get 获取文章详情
// @Summary 获取文章详情
// @Tags 文章
// @Produce json
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response{data=model.Post}
// @Failure 404 {object} utils.Response
// @Router /posts/{id} [get]
func (h *PostHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "文章 ID 格式错误")
		return
	}

	userID, _ := middleware.GetCurrentUserID(c)
	isAdmin := middleware.IsAdmin(c)

	// 先尝试缓存（缓存只存已发布且无需鉴权的文章）
	if post, ok := service.GetPostCache(uint(id)); ok {
		// 即使命中缓存，也要异步增加浏览量，并刷新缓存中的 view_count
		go h.refreshViewCountAndCache(uint(id), post)
		utils.Success(c, post)
		return
	}

	post, err := h.service.GetByID(uint(id), userID, isAdmin)
	if err != nil {
		utils.NotFound(c, "文章不存在")
		return
	}

	// 异步增加浏览量
	go func() {
		if err := h.service.IncrementViewCount(uint(id)); err != nil {
			utils.Logger.Errorf("增加文章浏览量失败: %v", err)
		}
	}()

	// 仅已发布文章写入缓存
	if post.Status == model.PostStatusPublished {
		service.SetPostCache(post)
	}

	utils.Success(c, post)
}

// refreshViewCountAndCache 增加浏览量并刷新缓存中的 view_count。
// 用于缓存命中时避免返回过时的浏览量。
func (h *PostHandler) refreshViewCountAndCache(postID uint, post *model.Post) {
	if err := h.service.IncrementViewCount(postID); err != nil {
		utils.Logger.Errorf("增加文章浏览量失败: %v", err)
	}
	viewCount, err := service.GetPostViewCount(postID)
	if err == nil {
		post.ViewCount = viewCount
		service.SetPostCache(post)
	}
}

// Update 更新文章（需要登录；管理员可修改任意文章，普通用户只能修改自己的文章）
// @Summary 更新文章
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Param body body model.UpdatePostRequest true "文章更新信息"
// @Success 200 {object} utils.Response{data=model.Post}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts/{id} [put]
func (h *PostHandler) Update(c *gin.Context) {
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

	var req model.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	post, err := h.service.Update(uint(id), userID, middleware.IsAdmin(c), req)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, post)
}

// Delete 删除文章（需要登录；管理员可删除任意文章，普通用户只能删除自己的文章）
// @Summary 删除文章
// @Tags 文章
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts/{id} [delete]
func (h *PostHandler) Delete(c *gin.Context) {
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

	if err := h.service.Delete(uint(id), userID, middleware.IsAdmin(c)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, nil)
}

// Hot 获取热门文章
// @Summary 获取热门文章
// @Tags 文章
// @Produce json
// @Param limit query int false "数量限制，默认 10"
// @Success 200 {object} utils.Response{data=[]model.Post}
// @Router /posts/hot [get]
func (h *PostHandler) Hot(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	posts, err := h.service.GetHotPosts(limit)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, posts)
}

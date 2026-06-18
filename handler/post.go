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

	// 创建成功后清除草稿
	_ = service.ClearPostDraft(userID)

	utils.SuccessWithStatus(c, http.StatusCreated, post)
}

// SaveDraft 自动保存文章草稿（需要登录）。
// @Summary 保存文章草稿
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreatePostRequest true "草稿内容"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts/drafts [post]
func (h *PostHandler) SaveDraft(c *gin.Context) {
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

	if err := service.SavePostDraft(userID, &req); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "草稿保存成功"})
}

// GetDraft 获取当前登录用户的文章草稿（需要登录）。
// @Summary 获取文章草稿
// @Tags 文章
// @Produce json
// @Security BearerAuth
// @Success 200 {object} utils.Response{data=model.CreatePostRequest}
// @Failure 401 {object} utils.Response
// @Router /posts/drafts [get]
func (h *PostHandler) GetDraft(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	draft, err := service.GetPostDraft(userID)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, "暂无草稿或读取失败")
		return
	}

	utils.Success(c, draft)
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
// @Param date_from query string false "开始日期 2006-01-02"
// @Param date_to query string false "结束日期 2006-01-02"
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

	// 登录用户记录阅读历史
	if userID != 0 {
		go service.RecordReadHistory(userID, post.ID)
	}

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

// AddFavorite 收藏文章（需要登录）。
// @Summary 收藏文章
// @Tags 文章
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts/{id}/favorite [post]
func (h *PostHandler) AddFavorite(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "文章 ID 无效")
		return
	}

	if err := service.AddFavorite(userID, uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "收藏成功"})
}

// RemoveFavorite 取消收藏（需要登录）。
// @Summary 取消收藏
// @Tags 文章
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /posts/{id}/favorite [delete]
func (h *PostHandler) RemoveFavorite(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "文章 ID 无效")
		return
	}

	if err := service.RemoveFavorite(userID, uint(id)); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "已取消收藏"})
}

// ListFavorites 获取当前用户收藏列表（需要登录）。
// @Summary 我的收藏
// @Tags 文章
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Router /auth/favorites [get]
func (h *PostHandler) ListFavorites(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListUserFavorites(userID, page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// Feed 获取当前用户的动态 Feed（自己 + 关注用户的文章，需要登录）。
// @Summary 用户动态 Feed
// @Tags 文章
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Router /feed [get]
func (h *PostHandler) Feed(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.GetUserFeed(userID, page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// BatchDelete 批量删除文章（管理员）。
// @Summary 批量删除文章
// @Tags 管理后台
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.BatchDeleteRequest true "文章 ID 列表"
// @Success 200 {object} utils.Response
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/posts/batch-delete [post]
func (h *PostHandler) BatchDelete(c *gin.Context) {
	var req model.BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	if err := h.service.BatchDelete(req.IDs); err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, gin.H{"message": "批量删除成功"})
}

// ListReadHistory 获取当前用户阅读历史（需要登录）。
// @Summary 阅读历史
// @Tags 文章
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Router /auth/read-history [get]
func (h *PostHandler) ListReadHistory(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListReadHistory(userID, page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

package handler

import (
	"blog/model"
	"blog/service"
	"blog/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 文章分类处理器，提供分类的增删改查接口。
type CategoryHandler struct {
	service *service.CategoryService
}

// NewCategoryHandler 创建分类处理器
func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{service: service.NewCategoryService()}
}

// Create 创建分类
// @Summary 创建分类（管理员）
// @Tags 分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreateCategoryRequest true "分类名称"
// @Success 201 {object} utils.Response{data=model.Category}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	category, err := h.service.Create(req.Name)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.SuccessWithStatus(c, http.StatusCreated, category)
}

// List 获取分类列表
// @Summary 获取分类列表
// @Tags 分类
// @Produce json
// @Success 200 {object} utils.Response{data=[]model.Category}
// @Router /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.service.List()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, categories)
}

// Update 更新分类
// @Summary 更新分类（管理员）
// @Tags 分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "分类 ID"
// @Param body body model.CreateCategoryRequest true "分类名称"
// @Success 200 {object} utils.Response{data=model.Category}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "分类 ID 格式错误")
		return
	}

	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	category, err := h.service.Update(uint(id), req.Name)
	if err != nil {
		utils.NotFound(c, "分类不存在")
		return
	}

	utils.Success(c, category)
}

// Delete 删除分类
// @Summary 删除分类（管理员）
// @Tags 分类
// @Security BearerAuth
// @Param id path int true "分类 ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "分类 ID 格式错误")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

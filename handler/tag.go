package handler

import (
	"blog/model"
	"blog/service"
	"blog/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TagHandler 文章标签处理器，提供标签的创建、列表和删除接口。
type TagHandler struct {
	service *service.TagService
}

// NewTagHandler 创建标签处理器
func NewTagHandler() *TagHandler {
	return &TagHandler{service: service.NewTagService()}
}

// Create 创建标签
// @Summary 创建标签（管理员）
// @Tags 标签
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreateTagRequest true "标签名称"
// @Success 201 {object} utils.Response{data=model.Tag}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /tags [post]
func (h *TagHandler) Create(c *gin.Context) {
	var req model.CreateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tag, err := h.service.Create(req.Name)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.SuccessWithStatus(c, http.StatusCreated, tag)
}

// List 获取标签列表
// @Summary 获取标签列表
// @Tags 标签
// @Produce json
// @Success 200 {object} utils.Response{data=[]model.Tag}
// @Router /tags [get]
func (h *TagHandler) List(c *gin.Context) {
	tags, err := h.service.List()
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, tags)
}

// Update 更新标签
// @Summary 更新标签（管理员）
// @Tags 标签
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "标签 ID"
// @Param body body model.UpdateTagRequest true "标签名称"
// @Success 200 {object} utils.Response{data=model.Tag}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /tags/{id} [put]
func (h *TagHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "标签 ID 格式错误")
		return
	}

	var req model.UpdateTagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	tag, err := h.service.Update(uint(id), req.Name)
	if err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, tag)
}

// Delete 删除标签
// @Summary 删除标签（管理员）
// @Tags 标签
// @Security BearerAuth
// @Param id path int true "标签 ID"
// @Success 200 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /tags/{id} [delete]
func (h *TagHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "标签 ID 格式错误")
		return
	}

	if err := h.service.Delete(uint(id)); err != nil {
		utils.InternalError(c, err.Error())
		return
	}

	utils.Success(c, nil)
}

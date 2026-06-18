// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现评论举报与后台审核接口。
package handler

import (
	"blog/middleware"
	"blog/model"
	"blog/service"
	"blog/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

// CommentReportHandler 评论举报处理器。
type CommentReportHandler struct{}

// NewCommentReportHandler 创建评论举报处理器。
func NewCommentReportHandler() *CommentReportHandler {
	return &CommentReportHandler{}
}

// Create 创建评论举报（已登录用户）。
// @Summary 举报评论
// @Tags 评论举报
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "评论 ID"
// @Param body body model.CreateCommentReportRequest true "举报原因"
// @Success 200 {object} utils.Response{data=model.CommentReport}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Router /comments/{id}/reports [post]
func (h *CommentReportHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetCurrentUserID(c)
	if !ok {
		utils.Unauthorized(c, "请先登录")
		return
	}

	commentID, err := strconv.Atoi(c.Param("id"))
	if err != nil || commentID <= 0 {
		utils.BadRequest(c, "评论 ID 无效")
		return
	}

	var req model.CreateCommentReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	report, err := service.CreateCommentReport(uint(commentID), userID, req.Reason)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, report)
}

// List 查询评论举报列表（管理员）。
// @Summary 查询评论举报列表
// @Tags 评论举报
// @Produce json
// @Security BearerAuth
// @Param status query string false "状态：pending/approved/rejected"
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/comment-reports [get]
func (h *CommentReportHandler) List(c *gin.Context) {
	status := model.CommentReportStatus(c.Query("status"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListCommentReports(status, page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

// UpdateStatus 更新举报状态（管理员）。
// @Summary 审核评论举报
// @Tags 评论举报
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "举报 ID"
// @Param body body model.UpdateCommentReportStatusRequest true "状态"
// @Success 200 {object} utils.Response{data=model.CommentReport}
// @Failure 400 {object} utils.Response
// @Failure 401 {object} utils.Response
// @Failure 403 {object} utils.Response
// @Router /admin/comment-reports/{id}/status [put]
func (h *CommentReportHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		utils.BadRequest(c, "举报 ID 无效")
		return
	}

	var req model.UpdateCommentReportStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	report, err := service.UpdateCommentReportStatus(uint(id), req.Status)
	if err != nil {
		utils.Error(c, utils.CodeBusinessError, err.Error())
		return
	}

	utils.Success(c, report)
}

// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现管理员后台相关接口：审计日志查询等。
package handler

import (
	"blog/model"
	"blog/service"
	"blog/utils"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理后台处理器。
type AdminHandler struct{}

// NewAdminHandler 创建管理后台处理器。
func NewAdminHandler() *AdminHandler {
	return &AdminHandler{}
}

// ListAuditLogs 查询审计日志列表（管理员）。
// @Summary 查询审计日志
// @Tags 管理后台
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param action query string false "动作，如 CREATE/UPDATE/DELETE"
// @Param resource query string false "资源，如 post/user/category"
// @Param user_id query int false "操作人用户 ID"
// @Param start_time query string false "开始时间，RFC3339 或 2006-01-02"
// @Param end_time query string false "结束时间，RFC3339 或 2006-01-02"
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Router /api/audit-logs [get]
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	var query model.AuditLogQuery
	_ = c.ShouldBindQuery(&query)

	query.StartTime = parseTimeParam(c.Query("start_time"))
	query.EndTime = parseTimeParam(c.Query("end_time"))

	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	resp, err := service.ListAuditLogs(query)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

func parseTimeParam(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	// 支持日期简写，解析为当天开始
	if t, err := time.Parse("2006-01-02", value); err == nil {
		return t
	}
	return time.Time{}
}

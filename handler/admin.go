// package handler 提供 HTTP 请求处理函数。
//
// 本文件实现管理员后台相关接口：审计日志查询等。
package handler

import (
	"blog/service"
	"blog/utils"
	"strconv"

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
// @Success 200 {object} utils.Response{data=model.ListResponse}
// @Router /api/audit-logs [get]
func (h *AdminHandler) ListAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	resp, err := service.ListAuditLogs(page, pageSize)
	if err != nil {
		utils.Error(c, utils.CodeInternalError, err.Error())
		return
	}

	utils.Success(c, resp)
}

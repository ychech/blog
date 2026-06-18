package service

import (
	"blog/database"
	"blog/model"
	"fmt"
	"strings"
)

// CreateCommentReport 创建评论举报。
func CreateCommentReport(commentID, userID uint, reason string) (*model.CommentReport, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, fmt.Errorf("举报原因不能为空")
	}

	// 确认评论存在
	var comment model.Comment
	if err := database.DB.First(&comment, commentID).Error; err != nil {
		return nil, fmt.Errorf("评论不存在")
	}

	report := model.CommentReport{
		CommentID: commentID,
		UserID:    userID,
		Reason:    reason,
		Status:    model.CommentReportStatusPending,
	}
	if err := database.DB.Create(&report).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

// ListCommentReports 查询评论举报列表（管理员）。
func ListCommentReports(status model.CommentReportStatus, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.CommentReport{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var reports []model.CommentReport
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&reports).Error; err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  reports,
	}, nil
}

// UpdateCommentReportStatus 更新举报状态（管理员）。
// 状态更新为 approved 时，会软删除对应评论。
func UpdateCommentReportStatus(reportID uint, status model.CommentReportStatus) (*model.CommentReport, error) {
	if status != model.CommentReportStatusPending &&
		status != model.CommentReportStatusApproved &&
		status != model.CommentReportStatusRejected {
		return nil, fmt.Errorf("无效的举报状态")
	}

	var report model.CommentReport
	if err := database.DB.First(&report, reportID).Error; err != nil {
		return nil, err
	}

	if err := database.DB.Model(&report).Update("status", status).Error; err != nil {
		return nil, err
	}

	if status == model.CommentReportStatusApproved {
		_ = database.DB.Delete(&model.Comment{}, report.CommentID).Error
	}

	return &report, nil
}

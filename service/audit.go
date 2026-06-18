// package service 实现操作审计日志业务逻辑。
package service

import (
	"blog/database"
	"blog/model"
)

// CreateAuditLog 创建一条审计日志。
func CreateAuditLog(userID uint, username, action, resource string, resourceID uint, details, ip string) error {
	log := model.AuditLog{
		UserID:     userID,
		Username:   username,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		Details:    details,
		IP:         ip,
	}
	return database.DB.Create(&log).Error
}

// ListAuditLogs 查询审计日志列表（仅管理员使用），支持按动作、资源、用户、时间范围过滤。
func ListAuditLogs(q model.AuditLogQuery) (*model.ListResponse, error) {
	page, pageSize := q.Page, q.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.AuditLog{})
	if q.Action != "" {
		query = query.Where("action = ?", q.Action)
	}
	if q.Resource != "" {
		query = query.Where("resource = ?", q.Resource)
	}
	if q.UserID > 0 {
		query = query.Where("user_id = ?", q.UserID)
	}
	if !q.StartTime.IsZero() {
		query = query.Where("created_at >= ?", q.StartTime)
	}
	if !q.EndTime.IsZero() {
		query = query.Where("created_at <= ?", q.EndTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var logs []model.AuditLog
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  logs,
	}, nil
}

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

// ListAuditLogs 查询审计日志列表（仅管理员使用）。
func ListAuditLogs(page, pageSize int) (*model.ListResponse, error) {
	var total int64
	query := database.DB.Model(&model.AuditLog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
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

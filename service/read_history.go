package service

import (
	"blog/database"
	"blog/model"
	"time"
)

// RecordReadHistory 记录用户阅读历史。
// 如果同一用户重复阅读同一篇文章，只更新阅读时间。
func RecordReadHistory(userID, postID uint) {
	if userID == 0 || postID == 0 {
		return
	}

	var history model.ReadHistory
	err := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).First(&history).Error
	if err == nil {
		history.ReadAt = time.Now()
		_ = database.DB.Save(&history).Error
		return
	}

	history = model.ReadHistory{
		UserID: userID,
		PostID: postID,
		ReadAt: time.Now(),
	}
	_ = database.DB.Create(&history).Error
}

// ListReadHistory 查询用户阅读历史。
func ListReadHistory(userID uint, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.ReadHistory{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var histories []model.ReadHistory
	if err := query.Order("read_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&histories).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if len(histories) > 0 {
		ids := make([]uint, 0, len(histories))
		for _, h := range histories {
			ids = append(ids, h.PostID)
		}
		database.DB.Where("id IN ?", ids).Find(&posts)
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  posts,
	}, nil
}

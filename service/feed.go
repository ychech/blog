package service

import (
	"blog/database"
	"blog/model"
)

// GetUserFeed 获取当前用户的动态 Feed（自己 + 关注用户的已发布文章）。
func GetUserFeed(userID uint, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 查询关注用户 ID
	var followingIDs []uint
	database.DB.Model(&model.UserFollow{}).
		Where("follower_id = ?", userID).
		Pluck("following_id", &followingIDs)

	authorIDs := append(followingIDs, userID)

	query := database.DB.Model(&model.Post{}).
		Where("author_id IN ? AND status = ?", authorIDs, model.PostStatusPublished)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if err := query.Order("created_at DESC").
		Preload("Author").Preload("Category").Preload("Tags").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  posts,
	}, nil
}

package service

import (
	"blog/database"
	"blog/model"
	"fmt"
)

// AddFavorite 收藏文章。
func AddFavorite(userID, postID uint) error {
	// 确认文章存在且已发布
	var post model.Post
	if err := database.DB.Where("id = ? AND status = ?", postID, model.PostStatusPublished).First(&post).Error; err != nil {
		return fmt.Errorf("文章不存在或未发布")
	}

	favorite := model.Favorite{UserID: userID, PostID: postID}
	if err := database.DB.Create(&favorite).Error; err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("已经收藏过该文章")
		}
		return err
	}
	return nil
}

// RemoveFavorite 取消收藏。
func RemoveFavorite(userID, postID uint) error {
	result := database.DB.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Favorite{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("未收藏该文章")
	}
	return nil
}

// ListUserFavorites 查询用户收藏的文章列表。
func ListUserFavorites(userID uint, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.Favorite{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var favorites []model.Favorite
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&favorites).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if len(favorites) > 0 {
		ids := make([]uint, 0, len(favorites))
		for _, f := range favorites {
			ids = append(ids, f.PostID)
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

// IsFavorite 查询用户是否收藏了某篇文章。
func IsFavorite(userID, postID uint) bool {
	var count int64
	database.DB.Model(&model.Favorite{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count)
	return count > 0
}

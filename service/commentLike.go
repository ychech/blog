// package service 实现业务逻辑，协调数据库访问、缓存、密码加密与 JWT 生成。
// handler 层不直接访问数据库，所有业务规则都在此层处理。
package service

import (
	"blog/database"
	"blog/model"
	"errors"

	"gorm.io/gorm"
)

// CommentLikeService 评论点赞服务
type CommentLikeService struct{}

// NewCommentLikeService 创建评论点赞服务
func NewCommentLikeService() *CommentLikeService {
	return &CommentLikeService{}
}

// Toggle 切换评论点赞状态：已点赞则取消，未点赞则点赞
func (s *CommentLikeService) Toggle(commentID, userID uint) (bool, error) {
	var like model.CommentLike
	err := database.DB.Where("comment_id = ? AND user_id = ?", commentID, userID).First(&like).Error

	// 已存在则取消点赞（软删除）
	if err == nil {
		if err := database.DB.Delete(&like).Error; err != nil {
			return false, err
		}
		return false, nil
	}

	// 不存在则新增点赞
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	like = model.CommentLike{CommentID: commentID, UserID: userID}
	if err := database.DB.Create(&like).Error; err != nil {
		return false, err
	}
	return true, nil
}

// IsLiked 检查用户是否已点赞某条评论
func (s *CommentLikeService) IsLiked(commentID, userID uint) bool {
	var count int64
	database.DB.Model(&model.CommentLike{}).Where("comment_id = ? AND user_id = ?", commentID, userID).Count(&count)
	return count > 0
}

// GetLikeCount 获取评论点赞数
func (s *CommentLikeService) GetLikeCount(commentID uint) int64 {
	var count int64
	database.DB.Model(&model.CommentLike{}).Where("comment_id = ?", commentID).Count(&count)
	return count
}

// BatchGetLikeCounts 批量获取多条评论的点赞数，避免 N+1 查询。
func (s *CommentLikeService) BatchGetLikeCounts(commentIDs []uint) map[uint]int64 {
	result := make(map[uint]int64, len(commentIDs))
	if len(commentIDs) == 0 {
		return result
	}

	var rows []struct {
		CommentID uint  `gorm:"column:comment_id"`
		Count     int64 `gorm:"column:count"`
	}
	database.DB.Model(&model.CommentLike{}).
		Select("comment_id, COUNT(*) as count").
		Where("comment_id IN ?", commentIDs).
		Group("comment_id").
		Find(&rows)

	for _, row := range rows {
		result[row.CommentID] = row.Count
	}
	return result
}

// BatchIsLiked 批量检查用户对多条评论是否已点赞。
func (s *CommentLikeService) BatchIsLiked(commentIDs []uint, userID uint) map[uint]bool {
	result := make(map[uint]bool, len(commentIDs))
	if len(commentIDs) == 0 || userID == 0 {
		return result
	}

	var rows []struct {
		CommentID uint `gorm:"column:comment_id"`
	}
	database.DB.Model(&model.CommentLike{}).
		Select("comment_id").
		Where("comment_id IN ? AND user_id = ?", commentIDs, userID).
		Find(&rows)

	for _, row := range rows {
		result[row.CommentID] = true
	}
	return result
}

// package service 实现业务逻辑，协调数据库访问、缓存、密码加密与 JWT 生成。
// handler 层不直接访问数据库，所有业务规则都在此层处理。
package service

import (
	"blog/database"
	"blog/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LikeService 点赞服务
type LikeService struct{}

// NewLikeService 创建点赞服务
func NewLikeService() *LikeService {
	return &LikeService{}
}

// Toggle 切换点赞状态：已点赞则取消，未点赞则点赞。
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 原子切换，避免并发重复点赞。
// 新增点赞时会异步通知文章作者。
func (s *LikeService) Toggle(postID, userID uint) (bool, error) {
	var liked bool
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		like := model.Like{PostID: postID, UserID: userID}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "post_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"deleted_at": gorm.Expr("CASE WHEN deleted_at IS NULL THEN CURRENT_TIMESTAMP ELSE NULL END"),
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}),
		}).Create(&like).Error; err != nil {
			return err
		}

		var final model.Like
		if err := tx.Unscoped().Where("post_id = ? AND user_id = ?", postID, userID).First(&final).Error; err != nil {
			return err
		}
		liked = !final.DeletedAt.Valid
		return nil
	})
	if err != nil {
		return false, err
	}

	if liked {
		s.notifyPostAuthor(postID, userID)
	}
	return liked, nil
}

func (s *LikeService) notifyPostAuthor(postID, userID uint) {
	var post model.Post
	if err := database.DB.Select("id, title, author_id").First(&post, postID).Error; err != nil {
		return
	}
	if post.AuthorID == userID {
		return
	}

	var user model.User
	if err := database.DB.Select("id, nickname, username").First(&user, userID).Error; err != nil {
		return
	}

	nickname := user.Nickname
	if nickname == "" {
		nickname = user.Username
	}
	notifyAsync(func() error {
		return CreatePostLikeNotification(post.AuthorID, post.ID, nickname, post.Title)
	})
}

// IsLiked 检查用户是否已点赞
func (s *LikeService) IsLiked(postID, userID uint) bool {
	var count int64
	database.DB.Model(&model.Like{}).Where("post_id = ? AND user_id = ?", postID, userID).Count(&count)
	return count > 0
}

// GetLikeCount 获取文章点赞数
func (s *LikeService) GetLikeCount(postID uint) int64 {
	var count int64
	database.DB.Model(&model.Like{}).Where("post_id = ?", postID).Count(&count)
	return count
}

// BatchGetLikeCounts 批量获取多篇文章的点赞数，避免 N+1 查询。
// 返回 map[postID]count
func (s *LikeService) BatchGetLikeCounts(postIDs []uint) map[uint]int64 {
	result := make(map[uint]int64, len(postIDs))
	if len(postIDs) == 0 {
		return result
	}

	var rows []struct {
		PostID uint  `gorm:"column:post_id"`
		Count  int64 `gorm:"column:count"`
	}
	database.DB.Model(&model.Like{}).
		Select("post_id, COUNT(*) as count").
		Where("post_id IN ?", postIDs).
		Group("post_id").
		Find(&rows)

	for _, row := range rows {
		result[row.PostID] = row.Count
	}
	return result
}

// BatchIsLiked 批量检查用户对多篇文章是否已点赞。
// 返回 map[postID]liked
func (s *LikeService) BatchIsLiked(postIDs []uint, userID uint) map[uint]bool {
	result := make(map[uint]bool, len(postIDs))
	if len(postIDs) == 0 || userID == 0 {
		return result
	}

	var rows []struct {
		PostID uint `gorm:"column:post_id"`
	}
	database.DB.Model(&model.Like{}).
		Select("post_id").
		Where("post_id IN ? AND user_id = ?", postIDs, userID).
		Find(&rows)

	for _, row := range rows {
		result[row.PostID] = true
	}
	return result
}

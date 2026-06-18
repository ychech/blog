// package service 实现业务逻辑，协调数据库访问、缓存、密码加密与 JWT 生成。
// handler 层不直接访问数据库，所有业务规则都在此层处理。
package service

import (
	"blog/database"
	"blog/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// CommentLikeService 评论点赞服务
type CommentLikeService struct{}

// NewCommentLikeService 创建评论点赞服务
func NewCommentLikeService() *CommentLikeService {
	return &CommentLikeService{}
}

// Toggle 切换评论点赞状态：已点赞则取消，未点赞则点赞。
// 使用 INSERT ... ON DUPLICATE KEY UPDATE 原子切换，避免并发重复点赞。
// 新增点赞时会异步通知评论作者。
func (s *CommentLikeService) Toggle(commentID, userID uint) (bool, error) {
	var liked bool
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		like := model.CommentLike{CommentID: commentID, UserID: userID}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "comment_id"}, {Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"deleted_at": gorm.Expr("CASE WHEN deleted_at IS NULL THEN CURRENT_TIMESTAMP ELSE NULL END"),
				"updated_at": gorm.Expr("CURRENT_TIMESTAMP"),
			}),
		}).Create(&like).Error; err != nil {
			return err
		}

		var final model.CommentLike
		if err := tx.Unscoped().Where("comment_id = ? AND user_id = ?", commentID, userID).First(&final).Error; err != nil {
			return err
		}
		liked = !final.DeletedAt.Valid
		return nil
	})
	if err != nil {
		return false, err
	}

	if liked {
		s.notifyCommentAuthor(commentID, userID)
	}
	return liked, nil
}

func (s *CommentLikeService) notifyCommentAuthor(commentID, userID uint) {
	var comment model.Comment
	if err := database.DB.Select("id, author_id").First(&comment, commentID).Error; err != nil {
		return
	}
	if comment.AuthorID == userID {
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
		return CreateCommentLikeNotification(comment.AuthorID, comment.ID, nickname)
	})
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

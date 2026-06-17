package service

import (
	"blog/database"
	"blog/model"
	"fmt"
	"strings"
)

// CommentService 评论服务，处理评论创建、文章评论列表查询与删除。
type CommentService struct{}

// NewCommentService 创建评论服务
func NewCommentService() *CommentService {
	return &CommentService{}
}

// Create 创建评论
func (s *CommentService) Create(authorID uint, authorName string, req model.CreateCommentRequest) (*model.Comment, error) {
	// 检查文章是否存在
	var post model.Post
	if err := database.DB.First(&post, req.PostID).Error; err != nil {
		return nil, fmt.Errorf("文章不存在")
	}

	// 如果回复某条评论，检查父评论是否存在
	if req.ParentID != nil {
		var parent model.Comment
		if err := database.DB.First(&parent, *req.ParentID).Error; err != nil {
			return nil, fmt.Errorf("父评论不存在")
		}
		if parent.PostID != req.PostID {
			return nil, fmt.Errorf("父评论不属于该文章")
		}
	}

	content := strings.TrimSpace(req.Content)
	if err := validateNonEmptyTrimmed(content, "评论内容"); err != nil {
		return nil, err
	}
	if err := validateMaxLength(content, "评论内容", 5000); err != nil {
		return nil, err
	}

	comment := model.Comment{
		PostID:     req.PostID,
		ParentID:   req.ParentID,
		AuthorID:   authorID,
		AuthorName: authorName,
		Content:    content,
	}

	if err := database.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	// 如果回复了他人评论，异步发送通知给父评论作者
	if req.ParentID != nil {
		var parent model.Comment
		if err := database.DB.First(&parent, *req.ParentID).Error; err == nil {
			if parent.AuthorID != authorID {
				go CreateCommentReplyNotification(parent.AuthorID, comment.ID, authorName, post.Title)
			}
		}
	}

	return &comment, nil
}

// ListByPost 获取文章的评论列表（包含回复），并填充每条评论的点赞数。
func (s *CommentService) ListByPost(postID uint) ([]model.Comment, error) {
	var comments []model.Comment
	if err := database.DB.
		Where("post_id = ?", postID).
		Order("created_at ASC").
		Find(&comments).Error; err != nil {
		return nil, err
	}

	// 批量填充评论点赞数（单次查询，避免 N+1）
	commentIDs := make([]uint, len(comments))
	for i := range comments {
		commentIDs[i] = comments[i].ID
	}
	likeCounts := NewCommentLikeService().BatchGetLikeCounts(commentIDs)
	for i := range comments {
		comments[i].LikeCount = likeCounts[comments[i].ID]
	}
	return comments, nil
}

// Delete 删除评论。管理员可删除任意评论，普通用户只能删除自己的评论。
func (s *CommentService) Delete(id, currentUserID uint, isAdmin bool) error {
	var comment model.Comment
	if err := database.DB.First(&comment, id).Error; err != nil {
		return err
	}
	if !isAdmin && comment.AuthorID != currentUserID {
		return fmt.Errorf("无权删除该评论")
	}
	return database.DB.Delete(&comment).Error
}

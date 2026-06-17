// package service 实现通知相关业务逻辑。
//
// 当前支持：
//   - 评论回复通知：当用户 B 回复用户 A 的评论时，向用户 A 发送通知
package service

import (
	"blog/database"
	"blog/model"
	"fmt"
)

// CreateCommentReplyNotification 创建评论回复通知。
// parentUserID: 被回复评论的作者 ID
// replyCommentID: 回复评论的 ID
// replierNickname: 回复者昵称
// postTitle: 文章标题（用于通知内容）
func CreateCommentReplyNotification(parentUserID, replyCommentID uint, replierNickname, postTitle string) error {
	if parentUserID == 0 {
		return nil
	}

	notification := model.Notification{
		UserID:    parentUserID,
		Type:      model.NotificationTypeCommentReply,
		Title:     "有人回复了你的评论",
		Content:   fmt.Sprintf("%s 回复了你在《%s》中的评论", replierNickname, postTitle),
		RelatedID: replyCommentID,
		IsRead:    false,
	}

	return database.DB.Create(&notification).Error
}

// ListNotifications 查询指定用户的通知列表。
func ListNotifications(userID uint, page, pageSize int) (*model.ListResponse, error) {
	var total int64
	query := database.DB.Model(&model.Notification{}).Where("user_id = ?", userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 10
	}

	var notifications []model.Notification
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&notifications).Error; err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  notifications,
	}, nil
}

// MarkNotificationAsRead 将指定通知标记为已读。
// 仅允许通知接收者自己标记。
func MarkNotificationAsRead(userID, notificationID uint) error {
	result := database.DB.Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("通知不存在或无权限")
	}
	return nil
}

// CountUnreadNotifications 统计用户未读通知数量。
func CountUnreadNotifications(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

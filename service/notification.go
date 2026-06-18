// package service 实现通知相关业务逻辑。
//
// 当前支持：
//   - 评论回复通知
//   - 文章/评论点赞通知
//   - 被关注通知
//   - 私信通知
//   - 勋章颁发通知
package service

import (
	"blog/database"
	"blog/model"
	"blog/utils"
	"fmt"
)

// CreateNotification 创建一条通用通知，并在写入成功后尝试实时推送给在线用户。
func CreateNotification(userID uint, nType model.NotificationType, title, content string, relatedID uint) error {
	if userID == 0 {
		return nil
	}
	notification := &model.Notification{
		UserID:    userID,
		Type:      nType,
		Title:     title,
		Content:   content,
		RelatedID: relatedID,
		IsRead:    false,
	}
	if err := database.DB.Create(notification).Error; err != nil {
		return err
	}
	NotifyUserRealtime(userID, notification)
	return nil
}

// CreateCommentReplyNotification 创建评论回复通知。
// parentUserID: 被回复评论的作者 ID
// replyCommentID: 回复评论的 ID
// replierNickname: 回复者昵称
// postTitle: 文章标题（用于通知内容）
func CreateCommentReplyNotification(parentUserID, replyCommentID uint, replierNickname, postTitle string) error {
	return CreateNotification(parentUserID, model.NotificationTypeCommentReply,
		"有人回复了你的评论",
		fmt.Sprintf("%s 回复了你在《%s》中的评论", replierNickname, postTitle),
		replyCommentID)
}

// CreatePostLikeNotification 创建文章点赞通知。
func CreatePostLikeNotification(authorID, postID uint, likerNickname, postTitle string) error {
	return CreateNotification(authorID, model.NotificationTypePostLike,
		"有人赞了你的文章",
		fmt.Sprintf("%s 赞了你的文章《%s》", likerNickname, postTitle),
		postID)
}

// CreateCommentLikeNotification 创建评论点赞通知。
func CreateCommentLikeNotification(authorID, commentID uint, likerNickname string) error {
	return CreateNotification(authorID, model.NotificationTypeCommentLike,
		"有人赞了你的评论",
		fmt.Sprintf("%s 赞了你的评论", likerNickname),
		commentID)
}

// CreateFollowNotification 创建被关注通知。
func CreateFollowNotification(followingID, followerID uint, followerNickname string) error {
	return CreateNotification(followingID, model.NotificationTypeFollow,
		"新增关注",
		fmt.Sprintf("%s 关注了你", followerNickname),
		followerID)
}

// CreateMessageNotification 创建私信通知。
func CreateMessageNotification(receiverID, messageID uint, senderNickname string) error {
	return CreateNotification(receiverID, model.NotificationTypeMessage,
		"收到一条新私信",
		fmt.Sprintf("%s 给你发了一条私信", senderNickname),
		messageID)
}

// CreateBadgeAwardNotification 创建勋章颁发通知。
func CreateBadgeAwardNotification(userID, userBadgeID uint, badgeName string) error {
	return CreateNotification(userID, model.NotificationTypeBadgeAward,
		"获得新勋章",
		fmt.Sprintf("恭喜你获得「%s」勋章", badgeName),
		userBadgeID)
}

// notifyAsync 异步创建通知，错误仅记录日志，不阻塞主流程。
func notifyAsync(fn func() error) {
	go func() {
		if err := fn(); err != nil {
			utils.Logger.Errorf("创建通知失败: %v", err)
		}
	}()
}

// ListNotifications 查询指定用户的通知列表，支持按类型过滤。
// notificationType 为空字符串时表示全部类型。
func ListNotifications(userID uint, notificationType model.NotificationType, page, pageSize int) (*model.ListResponse, error) {
	var total int64
	query := database.DB.Model(&model.Notification{}).Where("user_id = ?", userID)
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}
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

// MarkAllNotificationsAsRead 将用户所有未读通知标记为已读。
func MarkAllNotificationsAsRead(userID uint) (int64, error) {
	result := database.DB.Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true)
	return result.RowsAffected, result.Error
}

// DeleteNotification 删除指定通知。
// 仅允许通知接收者自己删除。
func DeleteNotification(userID, notificationID uint) error {
	result := database.DB.Where("id = ? AND user_id = ?", notificationID, userID).Delete(&model.Notification{})
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

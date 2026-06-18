package service

import (
	"blog/database"
	"blog/model"
	"fmt"
	"strings"
)

// SendMessage 发送站内私信。
func SendMessage(senderID uint, req model.SendMessageRequest) (*model.Message, error) {
	if senderID == req.ReceiverID {
		return nil, fmt.Errorf("不能给自己发送私信")
	}

	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("私信内容不能为空")
	}
	if len([]rune(content)) > 2000 {
		return nil, fmt.Errorf("私信内容不能超过 2000 个字符")
	}

	// 确认接收者存在
	var receiver model.User
	if err := database.DB.First(&receiver, req.ReceiverID).Error; err != nil {
		return nil, fmt.Errorf("接收者不存在")
	}

	msg := model.Message{
		SenderID:   senderID,
		ReceiverID: req.ReceiverID,
		Content:    content,
		IsRead:     false,
	}
	if err := database.DB.Create(&msg).Error; err != nil {
		return nil, err
	}

	notifyMessageNotification(senderID, receiver.ID, msg.ID)
	return &msg, nil
}

func notifyMessageNotification(senderID, receiverID, messageID uint) {
	var sender model.User
	if err := database.DB.Select("id, nickname, username").First(&sender, senderID).Error; err != nil {
		return
	}
	nickname := sender.Nickname
	if nickname == "" {
		nickname = sender.Username
	}
	notifyAsync(func() error {
		return CreateMessageNotification(receiverID, messageID, nickname)
	})
}

// ListConversations 查询当前用户的会话列表。
func ListConversations(userID uint, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	// 子查询：每个会话对方的最新一条消息 ID
	subQuery := `
		SELECT MAX(id) AS id FROM messages
		WHERE sender_id = ? OR receiver_id = ?
		GROUP BY CASE
			WHEN sender_id = ? THEN receiver_id
			ELSE sender_id
		END
	`
	var lastMsgIDs []uint
	if err := database.DB.Raw(subQuery, userID, userID, userID).Scan(&lastMsgIDs).Error; err != nil {
		return nil, err
	}

	var total int64 = int64(len(lastMsgIDs))
	var conversations []model.Conversation
	if len(lastMsgIDs) > 0 {
		var lastMessages []model.Message
		if err := database.DB.Where("id IN ?", lastMsgIDs).Order("created_at DESC").
			Offset((page - 1) * pageSize).Limit(pageSize).Find(&lastMessages).Error; err != nil {
			return nil, err
		}

		for _, msg := range lastMessages {
			otherID := msg.SenderID
			if otherID == userID {
				otherID = msg.ReceiverID
			}

			var user model.User
			database.DB.Select("id, username, nickname, avatar").First(&user, otherID)

			var unread int64
			database.DB.Model(&model.Message{}).
				Where("sender_id = ? AND receiver_id = ? AND is_read = ?", otherID, userID, false).
				Count(&unread)

			conversations = append(conversations, model.Conversation{
				UserID:        user.ID,
				Username:      user.Username,
				Nickname:      user.Nickname,
				Avatar:        user.Avatar,
				LastContent:   msg.Content,
				LastMessageAt: msg.CreatedAt,
				UnreadCount:   unread,
			})
		}
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  conversations,
	}, nil
}

// ListMessages 查询与指定用户的私信记录，并将对方发送的未读消息标记为已读。
func ListMessages(userID, otherUserID uint, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query := database.DB.Model(&model.Message{}).
		Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?)",
			userID, otherUserID, otherUserID, userID)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var messages []model.Message
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&messages).Error; err != nil {
		return nil, err
	}

	// 将对方发送的未读消息标记为已读
	database.DB.Model(&model.Message{}).
		Where("sender_id = ? AND receiver_id = ? AND is_read = ?", otherUserID, userID, false).
		Update("is_read", true)

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  messages,
	}, nil
}

// CountUnreadMessages 统计用户未读私信数。
func CountUnreadMessages(userID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.Message{}).
		Where("receiver_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

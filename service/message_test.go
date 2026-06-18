package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestMessageService(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	userA := model.User{Username: "userA", Password: "hash"}
	userB := model.User{Username: "userB", Password: "hash"}
	if err := database.DB.Create(&userA).Error; err != nil {
		t.Fatalf("创建用户 A 失败: %v", err)
	}
	if err := database.DB.Create(&userB).Error; err != nil {
		t.Fatalf("创建用户 B 失败: %v", err)
	}

	// 发送私信
	msg, err := SendMessage(userA.ID, model.SendMessageRequest{
		ReceiverID: userB.ID,
		Content:    "hello",
	})
	if err != nil {
		t.Fatalf("发送私信失败: %v", err)
	}
	if msg.Content != "hello" {
		t.Errorf("私信内容错误: %s", msg.Content)
	}

	// 不能给自己发
	_, err = SendMessage(userA.ID, model.SendMessageRequest{
		ReceiverID: userA.ID,
		Content:    "self",
	})
	if err == nil {
		t.Error("不能给自己发私信")
	}

	// 未读数
	count, err := CountUnreadMessages(userB.ID)
	if err != nil {
		t.Fatalf("统计未读失败: %v", err)
	}
	if count != 1 {
		t.Errorf("未读数期望 1，得到 %d", count)
	}

	// 会话列表
	convResp, err := ListConversations(userB.ID, 1, 10)
	if err != nil {
		t.Fatalf("会话列表失败: %v", err)
	}
	if convResp.Total != 1 {
		t.Errorf("会话数期望 1，得到 %d", convResp.Total)
	}

	// 读取私信
	msgResp, err := ListMessages(userB.ID, userA.ID, 1, 10)
	if err != nil {
		t.Fatalf("读取私信失败: %v", err)
	}
	if msgResp.Total != 1 {
		t.Errorf("私信记录数期望 1，得到 %d", msgResp.Total)
	}

	// 读取后未读应为 0
	count, _ = CountUnreadMessages(userB.ID)
	if count != 0 {
		t.Errorf("读取后未读数期望 0，得到 %d", count)
	}
}

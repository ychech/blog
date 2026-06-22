package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"errors"
	"testing"
	"time"
)

func TestSendNotificationEmail_SkipsWhenDisabled(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = &config.Config{
		Email: config.EmailConfig{NotificationEmailEnabled: false},
	}

	user := model.User{Username: "emailuser1", Password: "hash", Email: "a@example.com", EmailVerified: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	called := false
	sendEmailFunc = func(_, _, _ string) error { called = true; return nil }
	defer func() { sendEmailFunc = defaultSendEmailFunc }()

	SendNotificationEmail(user.ID, &model.Notification{ID: 1, UserID: user.ID, Type: model.NotificationTypeFollow, Title: "t", Content: "c", CreatedAt: time.Now()})
	time.Sleep(100 * time.Millisecond)

	if called {
		t.Error("通知邮件开关关闭时不应发送邮件")
	}
}

func TestSendNotificationEmail_SkipsWhenNoEmail(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = &config.Config{
		Email: config.EmailConfig{
			NotificationEmailEnabled: true,
			Host:                     "smtp.example.com",
			Username:                 "user",
			Password:                 "pass",
		},
	}

	user := model.User{Username: "emailuser2", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	called := false
	sendEmailFunc = func(_, _, _ string) error { called = true; return nil }
	defer func() { sendEmailFunc = defaultSendEmailFunc }()

	SendNotificationEmail(user.ID, &model.Notification{ID: 2, UserID: user.ID, Type: model.NotificationTypeFollow, Title: "t", Content: "c", CreatedAt: time.Now()})
	time.Sleep(100 * time.Millisecond)

	if called {
		t.Error("用户未绑定邮箱时不应发送邮件")
	}
}

func TestSendNotificationEmail_SkipsWhenUnverified(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = &config.Config{
		Email: config.EmailConfig{
			NotificationEmailEnabled: true,
			Host:                     "smtp.example.com",
			Username:                 "user",
			Password:                 "pass",
		},
	}

	user := model.User{Username: "emailuser3", Password: "hash", Email: "b@example.com", EmailVerified: false}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	called := false
	sendEmailFunc = func(_, _, _ string) error { called = true; return nil }
	defer func() { sendEmailFunc = defaultSendEmailFunc }()

	SendNotificationEmail(user.ID, &model.Notification{ID: 3, UserID: user.ID, Type: model.NotificationTypeFollow, Title: "t", Content: "c", CreatedAt: time.Now()})
	time.Sleep(100 * time.Millisecond)

	if called {
		t.Error("邮箱未验证时不应发送邮件")
	}
}

func TestSendNotificationEmail_SendsWhenConfigured(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = &config.Config{
		Email: config.EmailConfig{
			NotificationEmailEnabled: true,
			Host:                     "smtp.example.com",
			Username:                 "user",
			Password:                 "pass",
		},
	}

	user := model.User{Username: "emailuser4", Password: "hash", Email: "c@example.com", EmailVerified: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	called := false
	sendEmailFunc = func(to, subject, body string) error {
		called = true
		if to != user.Email {
			t.Errorf("收件人错误: %s", to)
		}
		if subject == "" || body == "" {
			t.Error("邮件标题或正文不能为空")
		}
		return nil
	}
	defer func() { sendEmailFunc = defaultSendEmailFunc }()

	SendNotificationEmail(user.ID, &model.Notification{
		ID:        4,
		UserID:    user.ID,
		Type:      model.NotificationTypeCommentReply,
		Title:     "有人回复了你",
		Content:   "reply content",
		CreatedAt: time.Now(),
	})
	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("配置完整且邮箱已验证时应发送邮件")
	}
}

func TestSendNotificationEmail_LogsError(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = &config.Config{
		Email: config.EmailConfig{
			NotificationEmailEnabled: true,
			Host:                     "smtp.example.com",
			Username:                 "user",
			Password:                 "pass",
		},
	}

	user := model.User{Username: "emailuser5", Password: "hash", Email: "d@example.com", EmailVerified: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	sendEmailFunc = func(_, _, _ string) error { return errors.New("mock error") }
	defer func() { sendEmailFunc = defaultSendEmailFunc }()

	// 错误仅记录日志，不应 panic
	SendNotificationEmail(user.ID, &model.Notification{ID: 5, UserID: user.ID, Type: model.NotificationTypePostLike, Title: "t", Content: "c", CreatedAt: time.Now()})
	time.Sleep(100 * time.Millisecond)
}

// defaultSendEmailFunc 保存原始的邮件发送函数，用于测试恢复。
var defaultSendEmailFunc = sendEmailFunc

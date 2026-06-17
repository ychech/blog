package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestNotificationService(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "notifyuser", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建两条未读通知
	for i := 0; i < 2; i++ {
		if err := CreateCommentReplyNotification(user.ID, uint(i+1), "replier", "post"); err != nil {
			t.Fatalf("创建通知失败: %v", err)
		}
	}

	count, err := CountUnreadNotifications(user.ID)
	if err != nil {
		t.Fatalf("统计未读失败: %v", err)
	}
	if count != 2 {
		t.Errorf("未读数期望 2，得到 %d", count)
	}

	// 标记单条已读
	if err := MarkNotificationAsRead(user.ID, 1); err != nil {
		t.Fatalf("标记已读失败: %v", err)
	}

	count, _ = CountUnreadNotifications(user.ID)
	if count != 1 {
		t.Errorf("未读数期望 1，得到 %d", count)
	}

	// 标记全部已读
	affected, err := MarkAllNotificationsAsRead(user.ID)
	if err != nil {
		t.Fatalf("标记全部已读失败: %v", err)
	}
	if affected != 1 {
		t.Errorf("影响行数期望 1，得到 %d", affected)
	}

	count, _ = CountUnreadNotifications(user.ID)
	if count != 0 {
		t.Errorf("未读数期望 0，得到 %d", count)
	}

	// 删除通知
	if err := DeleteNotification(user.ID, 1); err != nil {
		t.Fatalf("删除通知失败: %v", err)
	}
}

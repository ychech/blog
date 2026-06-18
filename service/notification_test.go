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

func TestNotificationService_ListWithTypeFilter(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "notifyuser2", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	if err := CreatePostLikeNotification(user.ID, 1, "liker", "post"); err != nil {
		t.Fatalf("创建点赞通知失败: %v", err)
	}
	if err := CreateFollowNotification(user.ID, 2, "follower"); err != nil {
		t.Fatalf("创建关注通知失败: %v", err)
	}
	if err := CreateMessageNotification(user.ID, 3, "sender"); err != nil {
		t.Fatalf("创建私信通知失败: %v", err)
	}

	all, err := ListNotifications(user.ID, "", 1, 10)
	if err != nil {
		t.Fatalf("查询全部通知失败: %v", err)
	}
	if all.Total != 3 {
		t.Errorf("全部通知数期望 3，得到 %d", all.Total)
	}

	follows, err := ListNotifications(user.ID, model.NotificationTypeFollow, 1, 10)
	if err != nil {
		t.Fatalf("按类型查询通知失败: %v", err)
	}
	if follows.Total != 1 {
		t.Errorf("关注通知数期望 1，得到 %d", follows.Total)
	}
	notifs, ok := follows.Data.([]model.Notification)
	if !ok || len(notifs) == 0 || notifs[0].Type != model.NotificationTypeFollow {
		t.Errorf("过滤结果类型不匹配")
	}
}

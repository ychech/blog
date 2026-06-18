package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestUserFollowService(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	userA := model.User{Username: "followerA", Password: "hash"}
	userB := model.User{Username: "followingB", Password: "hash"}
	if err := database.DB.Create(&userA).Error; err != nil {
		t.Fatalf("创建用户 A 失败: %v", err)
	}
	if err := database.DB.Create(&userB).Error; err != nil {
		t.Fatalf("创建用户 B 失败: %v", err)
	}

	// 关注
	if err := FollowUser(userA.ID, userB.ID); err != nil {
		t.Fatalf("关注失败: %v", err)
	}
	if !IsFollowing(userA.ID, userB.ID) {
		t.Error("应已关注")
	}

	// 重复关注应失败
	if err := FollowUser(userA.ID, userB.ID); err == nil {
		t.Error("重复关注应失败")
	}

	// 粉丝列表
	resp, err := ListFollowers(userB.ID, 1, 10)
	if err != nil {
		t.Fatalf("粉丝列表失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("粉丝数期望 1，得到 %d", resp.Total)
	}

	// 关注列表
	resp, err = ListFollowing(userA.ID, 1, 10)
	if err != nil {
		t.Fatalf("关注列表失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("关注数期望 1，得到 %d", resp.Total)
	}

	// 取消关注
	if err := UnfollowUser(userA.ID, userB.ID); err != nil {
		t.Fatalf("取消关注失败: %v", err)
	}
	if IsFollowing(userA.ID, userB.ID) {
		t.Error("取消关注后应未关注")
	}
}

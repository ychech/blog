package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestGetUserFeed(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	userA := model.User{Username: "feedA", Password: "hash"}
	userB := model.User{Username: "feedB", Password: "hash"}
	if err := database.DB.Create(&userA).Error; err != nil {
		t.Fatalf("创建用户 A 失败: %v", err)
	}
	if err := database.DB.Create(&userB).Error; err != nil {
		t.Fatalf("创建用户 B 失败: %v", err)
	}
	category := model.Category{Name: "feedcat"}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	post := model.Post{
		Title:      "feedpost",
		Content:    "content",
		AuthorID:   userB.ID,
		CategoryID: &category.ID,
		Status:     model.PostStatusPublished,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	// userA 关注 userB
	if err := FollowUser(userA.ID, userB.ID); err != nil {
		t.Fatalf("关注失败: %v", err)
	}

	resp, err := GetUserFeed(userA.ID, 1, 10)
	if err != nil {
		t.Fatalf("获取 Feed 失败: %v", err)
	}
	if resp.Total != 1 { // 关注用户 userB 的一篇文章
		t.Errorf("Feed 数量期望 1，得到 %d", resp.Total)
	}
}

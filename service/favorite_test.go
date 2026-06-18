package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestFavoriteService(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "favuser", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	category := model.Category{Name: "favcat"}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	post := model.Post{
		Title:      "favpost",
		Content:    "content",
		AuthorID:   user.ID,
		CategoryID: &category.ID,
		Status:     model.PostStatusPublished,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	// 收藏
	if err := AddFavorite(user.ID, post.ID); err != nil {
		t.Fatalf("收藏失败: %v", err)
	}
	if !IsFavorite(user.ID, post.ID) {
		t.Error("应已收藏")
	}

	// 重复收藏应失败
	if err := AddFavorite(user.ID, post.ID); err == nil {
		t.Error("重复收藏应失败")
	}

	// 查询收藏列表
	resp, err := ListUserFavorites(user.ID, 1, 10)
	if err != nil {
		t.Fatalf("查询收藏失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("收藏数期望 1，得到 %d", resp.Total)
	}

	// 取消收藏
	if err := RemoveFavorite(user.ID, post.ID); err != nil {
		t.Fatalf("取消收藏失败: %v", err)
	}
	if IsFavorite(user.ID, post.ID) {
		t.Error("取消收藏后应未收藏")
	}
}

func TestReadHistoryService(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "histuser", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	category := model.Category{Name: "histcat"}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	post := model.Post{
		Title:      "histpost",
		Content:    "content",
		AuthorID:   user.ID,
		CategoryID: &category.ID,
		Status:     model.PostStatusPublished,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	RecordReadHistory(user.ID, post.ID)
	RecordReadHistory(user.ID, post.ID) // 重复记录应更新时间

	resp, err := ListReadHistory(user.ID, 1, 10)
	if err != nil {
		t.Fatalf("查询阅读历史失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("历史记录数期望 1，得到 %d", resp.Total)
	}
}

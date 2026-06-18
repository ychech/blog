package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestGetAdminStats(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "statuser", Password: "hash", IsActive: true}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	category := model.Category{Name: "statcat"}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	post := model.Post{
		Title:      "statpost",
		Content:    "content",
		AuthorID:   user.ID,
		CategoryID: &category.ID,
		Status:     model.PostStatusPublished,
		ViewCount:  100,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}
	comment := model.Comment{
		PostID:   post.ID,
		AuthorID: user.ID,
		Content:  "statcomment",
	}
	if err := database.DB.Create(&comment).Error; err != nil {
		t.Fatalf("创建评论失败: %v", err)
	}

	stats, err := GetAdminStats()
	if err != nil {
		t.Fatalf("获取仪表盘统计失败: %v", err)
	}

	if stats.Counts.Users != 1 {
		t.Errorf("用户数期望 1，得到 %d", stats.Counts.Users)
	}
	if stats.Counts.Posts != 1 {
		t.Errorf("文章数期望 1，得到 %d", stats.Counts.Posts)
	}
	if stats.Counts.Comments != 1 {
		t.Errorf("评论数期望 1，得到 %d", stats.Counts.Comments)
	}
	if len(stats.HotPosts) != 1 {
		t.Errorf("热门文章期望 1，得到 %d", len(stats.HotPosts))
	}
	if len(stats.Trends.Users) != 7 {
		t.Errorf("趋势数据长度期望 7，得到 %d", len(stats.Trends.Users))
	}
}

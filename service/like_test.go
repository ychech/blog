package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestLikeService_Toggle(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// 创建测试用户和文章
	user := model.User{Username: "likeuser", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	category := model.Category{Name: "c"}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}
	post := model.Post{
		Title:      "post",
		Content:    "content",
		AuthorID:   user.ID,
		CategoryID: &category.ID,
		Status:     model.PostStatusPublished,
	}
	if err := database.DB.Create(&post).Error; err != nil {
		t.Fatalf("创建文章失败: %v", err)
	}

	svc := NewLikeService()

	// 第一次点赞
	liked, err := svc.Toggle(post.ID, user.ID)
	if err != nil {
		t.Fatalf("点赞失败: %v", err)
	}
	if !liked {
		t.Error("第一次点赞应返回 true")
	}

	// 再次调用取消点赞
	liked, err = svc.Toggle(post.ID, user.ID)
	if err != nil {
		t.Fatalf("取消点赞失败: %v", err)
	}
	if liked {
		t.Error("再次调用应返回 false")
	}

	// 第三次恢复点赞
	liked, err = svc.Toggle(post.ID, user.ID)
	if err != nil {
		t.Fatalf("恢复点赞失败: %v", err)
	}
	if !liked {
		t.Error("恢复点赞应返回 true")
	}

	if count := svc.GetLikeCount(post.ID); count != 1 {
		t.Errorf("点赞数期望 1，得到 %d", count)
	}
}

package service

import (
	"blog/database"
	"blog/model"
	"strings"
	"testing"
)

func TestPostService_Create_Validation(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "postuser", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}
	category := model.Category{Name: "cat"}
	if err := database.DB.Create(&category).Error; err != nil {
		t.Fatalf("创建分类失败: %v", err)
	}

	svc := NewPostService()

	// 标题为空
	_, err := svc.Create(user.ID, model.CreatePostRequest{
		Title:      "   ",
		Content:    "content",
		CategoryID: category.ID,
	})
	if err == nil {
		t.Error("空标题应校验失败")
	}

	// 内容为空
	_, err = svc.Create(user.ID, model.CreatePostRequest{
		Title:      "title",
		Content:    "   ",
		CategoryID: category.ID,
	})
	if err == nil {
		t.Error("空内容应校验失败")
	}

	// 标题超长
	_, err = svc.Create(user.ID, model.CreatePostRequest{
		Title:      strings.Repeat("a", 256),
		Content:    "content",
		CategoryID: category.ID,
	})
	if err == nil {
		t.Error("标题超过 255 字符应校验失败")
	}

	// 正常创建
	post, err := svc.Create(user.ID, model.CreatePostRequest{
		Title:      "title",
		Content:    "content",
		CategoryID: category.ID,
	})
	if err != nil {
		t.Fatalf("正常创建失败: %v", err)
	}
	if post.Title != "title" {
		t.Errorf("标题被错误修改: %s", post.Title)
	}
}

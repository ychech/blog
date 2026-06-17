package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestCommentService_UpdateAndDelete(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "commentuser", Password: "hash"}
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

	svc := NewCommentService()

	// 创建父评论
	parent, err := svc.Create(user.ID, "author", model.CreateCommentRequest{
		PostID:  post.ID,
		Content: "parent",
	})
	if err != nil {
		t.Fatalf("创建评论失败: %v", err)
	}

	// 创建子评论
	child, err := svc.Create(user.ID, "author", model.CreateCommentRequest{
		PostID:   post.ID,
		ParentID: &parent.ID,
		Content:  "child",
	})
	if err != nil {
		t.Fatalf("创建子评论失败: %v", err)
	}

	// 编辑父评论
	updated, err := svc.Update(parent.ID, user.ID, false, model.UpdateCommentRequest{Content: "updated"})
	if err != nil {
		t.Fatalf("更新评论失败: %v", err)
	}
	if updated.Content != "updated" {
		t.Errorf("评论内容未更新: %s", updated.Content)
	}

	// 删除父评论，子评论应级联删除
	if err := svc.Delete(parent.ID, user.ID, false); err != nil {
		t.Fatalf("删除评论失败: %v", err)
	}

	var remaining int64
	if err := database.DB.Unscoped().Model(&model.Comment{}).Where("post_id = ?", post.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("统计评论失败: %v", err)
	}
	if remaining != 2 {
		t.Errorf("期望软删除 2 条评论，实际 %d", remaining)
	}

	var deletedCount int64
	if err := database.DB.Model(&model.Comment{}).Where("post_id = ?", post.ID).Count(&deletedCount).Error; err != nil {
		t.Fatalf("统计未删除评论失败: %v", err)
	}
	if deletedCount != 0 {
		t.Errorf("期望可见评论 0，实际 %d", deletedCount)
	}

	_ = child
}

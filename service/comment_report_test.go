package service

import (
	"blog/database"
	"blog/model"
	"testing"
)

func TestCommentReportService(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "reporter", Password: "hash"}
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
	comment := model.Comment{
		PostID:   post.ID,
		AuthorID: user.ID,
		Content:  "bad comment",
	}
	if err := database.DB.Create(&comment).Error; err != nil {
		t.Fatalf("创建评论失败: %v", err)
	}

	// 创建举报
	report, err := CreateCommentReport(comment.ID, user.ID, "spam")
	if err != nil {
		t.Fatalf("创建举报失败: %v", err)
	}
	if report.Status != model.CommentReportStatusPending {
		t.Errorf("默认状态应为 pending: %s", report.Status)
	}

	// 查询举报
	resp, err := ListCommentReports(model.CommentReportStatusPending, 1, 10)
	if err != nil {
		t.Fatalf("查询举报失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("pending 举报数期望 1，得到 %d", resp.Total)
	}

	// 审核通过，应软删除评论
	updated, err := UpdateCommentReportStatus(report.ID, model.CommentReportStatusApproved)
	if err != nil {
		t.Fatalf("更新举报状态失败: %v", err)
	}
	if updated.Status != model.CommentReportStatusApproved {
		t.Errorf("状态未更新为 approved: %s", updated.Status)
	}

	var remaining int64
	if err := database.DB.Model(&model.Comment{}).Where("post_id = ?", post.ID).Count(&remaining).Error; err != nil {
		t.Fatalf("统计评论失败: %v", err)
	}
	if remaining != 0 {
		t.Errorf("审核通过后评论应被软删除，剩余 %d", remaining)
	}
}

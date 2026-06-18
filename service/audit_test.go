package service

import (
	"blog/database"
	"blog/model"
	"testing"
	"time"
)

func TestListAuditLogs(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := model.User{Username: "audituser", Password: "hash"}
	if err := database.DB.Create(&user).Error; err != nil {
		t.Fatalf("创建用户失败: %v", err)
	}

	// 创建两条审计日志
	for i := 0; i < 2; i++ {
		action := "CREATE"
		resource := "post"
		if i == 1 {
			action = "DELETE"
			resource = "user"
		}
		if err := CreateAuditLog(user.ID, user.Username, action, resource, uint(i+1), "detail", "127.0.0.1"); err != nil {
			t.Fatalf("创建审计日志失败: %v", err)
		}
	}

	// 按 action 过滤
	resp, err := ListAuditLogs(model.AuditLogQuery{Action: "CREATE", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("查询审计日志失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("按 action 过滤总数期望 1，得到 %d", resp.Total)
	}

	// 按 resource 过滤
	resp, err = ListAuditLogs(model.AuditLogQuery{Resource: "user", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("查询审计日志失败: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("按 resource 过滤总数期望 1，得到 %d", resp.Total)
	}

	// 按时间范围过滤（包含所有）
	resp, err = ListAuditLogs(model.AuditLogQuery{
		Page:      1,
		PageSize:  10,
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("按时间范围查询失败: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("时间范围过滤总数期望 2，得到 %d", resp.Total)
	}
}

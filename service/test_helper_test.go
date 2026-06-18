// package service 提供业务逻辑实现。
//
// 本文件包含测试辅助函数，用于在单元测试中创建独立的内存数据库。
package service

import (
	"blog/database"
	"blog/model"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建一个 SQLite 内存数据库，并自动迁移模型，用于单元测试。
// 该函数会替换全局 database.DB，测试结束后通过返回的 cleanup 函数恢复。
func setupTestDB(t *testing.T) (cleanup func()) {
	t.Helper()

	originalDB := database.DB

	// 每个测试使用独立的内存数据库，避免并行/顺序执行时数据互相污染
	dbName := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("创建测试数据库失败: %v", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.CommentLike{},
		&model.Badge{},
		&model.UserBadge{},
		&model.Notification{},
		&model.AuditLog{},
		&model.CommentReport{},
		&model.Message{},
	); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}

	database.DB = db

	return func() {
		database.DB = originalDB
	}
}

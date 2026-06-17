// package service 提供业务逻辑实现。
//
// 本文件包含测试辅助函数，用于在单元测试中创建独立的内存数据库。
package service

import (
	"blog/database"
	"blog/model"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建一个 SQLite 内存数据库，并自动迁移模型，用于单元测试。
// 该函数会替换全局 database.DB，测试结束后通过返回的 cleanup 函数恢复。
func setupTestDB(t *testing.T) (cleanup func()) {
	t.Helper()

	originalDB := database.DB

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
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
	); err != nil {
		t.Fatalf("迁移测试数据库失败: %v", err)
	}

	database.DB = db

	return func() {
		database.DB = originalDB
	}
}

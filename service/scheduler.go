package service

import (
	"blog/database"
	"blog/model"
	"blog/utils"
	"context"
	"time"

	"gorm.io/gorm"
)

// StartScheduler 启动后台调度任务。
// 当前主要任务：每 60 秒扫描一次状态为 scheduled 且 publish_at 已到的文章，
// 将其状态更新为 published。
func StartScheduler(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// 启动时立即执行一次
	publishScheduledPosts()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			publishScheduledPosts()
		}
	}
}

func publishScheduledPosts() {
	now := time.Now()
	result := database.DB.Model(&model.Post{}).
		Where("status = ? AND publish_at <= ?", model.PostStatusScheduled, now).
		Updates(map[string]interface{}{
			"status":     model.PostStatusPublished,
			"publish_at": gorm.Expr("NULL"),
		})

	if result.Error != nil {
		utils.Logger.Errorf("定时发布任务失败: %v", result.Error)
		return
	}
	if result.RowsAffected > 0 {
		utils.Logger.Infof("定时发布文章 %d 篇", result.RowsAffected)
	}
}

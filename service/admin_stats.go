package service

import (
	"blog/database"
	"blog/model"
	"fmt"
	"time"
)

// GetAdminStats 获取后台仪表盘统计信息。
func GetAdminStats() (*model.AdminStats, error) {
	stats := &model.AdminStats{}

	// 总数统计
	counts := &stats.Counts
	database.DB.Model(&model.User{}).Count(&counts.Users)
	database.DB.Model(&model.Post{}).Count(&counts.Posts)
	database.DB.Model(&model.Comment{}).Count(&counts.Comments)
	database.DB.Model(&model.Category{}).Count(&counts.Categories)
	database.DB.Model(&model.Tag{}).Count(&counts.Tags)
	database.DB.Model(&model.Badge{}).Count(&counts.Badges)

	// 近 7 天趋势
	stats.Trends.Users = countTrendByDate(&model.User{}, "created_at", 7)
	stats.Trends.Posts = countTrendByDate(&model.Post{}, "created_at", 7)
	stats.Trends.Comments = countTrendByDate(&model.Comment{}, "created_at", 7)

	// 热门文章 Top 5
	database.DB.Model(&model.Post{}).Where("status = ?", model.PostStatusPublished).
		Order("view_count DESC").Limit(5).Find(&stats.HotPosts)

	// 热门标签 Top 5（按关联文章数）
	var topTags []model.Tag
	if err := database.DB.Raw(`
		SELECT t.*, COUNT(pt.post_id) AS post_count
		FROM tags t
		LEFT JOIN post_tags pt ON t.id = pt.tag_id
		GROUP BY t.id
		ORDER BY post_count DESC
		LIMIT 5
	`).Scan(&topTags).Error; err != nil {
		return nil, fmt.Errorf("查询热门标签失败: %w", err)
	}
	stats.TopTags = topTags

	return stats, nil
}

// countTrendByDate 统计指定表过去 n 天每天的新增数量。
func countTrendByDate(table interface{}, timeColumn string, days int) []model.TrendData {
	result := make([]model.TrendData, 0, days)
	now := time.Now()
	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		end := start.AddDate(0, 0, 1)

		var count int64
		database.DB.Model(table).
			Where(fmt.Sprintf("%s >= ? AND %s < ?", timeColumn, timeColumn), start, end).
			Count(&count)

		result = append(result, model.TrendData{
			Date:  start.Format("2006-01-02"),
			Count: count,
		})
	}
	return result
}

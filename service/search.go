// package service 实现 Meilisearch 搜索引擎集成。
//
// 当 Meilisearch 启用时，文章创建/更新/删除会同步到搜索索引，
// 搜索接口优先使用 Meilisearch 返回结果，失败时降级为 MySQL 模糊搜索。
package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"encoding/json"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

// SearchClient 全局 Meilisearch 客户端。
var SearchClient meilisearch.ServiceManager

// InitSearch 初始化 Meilisearch 客户端和索引。
func InitSearch(cfg config.MeilisearchConfig) error {
	if !cfg.Enabled {
		return nil
	}

	client := meilisearch.New(cfg.Host, meilisearch.WithAPIKey(cfg.APIKey))

	// 测试连接
	if _, err := client.Health(); err != nil {
		return fmt.Errorf("Meilisearch 连接失败: %w", err)
	}

	SearchClient = client

	// 创建索引（如果不存在）
	if _, err := client.GetIndex(cfg.Index); err != nil {
		if _, err := client.CreateIndex(&meilisearch.IndexConfig{
			Uid:        cfg.Index,
			PrimaryKey: "id",
		}); err != nil {
			return fmt.Errorf("创建 Meilisearch 索引失败: %w", err)
		}
	}

	index := client.Index(cfg.Index)

	// 配置可搜索字段
	_, _ = index.UpdateSearchableAttributes(&[]string{"title", "summary", "content"})

	// 配置过滤字段
	filterAttrs := []interface{}{"status"}
	_, _ = index.UpdateFilterableAttributes(&filterAttrs)

	return nil
}

// IndexPost 将文章索引到 Meilisearch。
func IndexPost(post *model.Post) error {
	if SearchClient == nil {
		return nil
	}

	cfg := config.C.Meilisearch
	doc := map[string]interface{}{
		"id":         post.ID,
		"title":      post.Title,
		"summary":    post.Summary,
		"content":    stripHTML(post.Content),
		"status":     post.Status,
		"created_at": post.CreatedAt,
	}

	_, err := SearchClient.Index(cfg.Index).AddDocuments(doc, nil)
	return err
}

// DeletePostIndex 从 Meilisearch 删除文章索引。
func DeletePostIndex(postID uint) error {
	if SearchClient == nil {
		return nil
	}

	cfg := config.C.Meilisearch
	_, err := SearchClient.Index(cfg.Index).DeleteDocument(fmt.Sprintf("%d", postID), nil)
	return err
}

// SearchPosts 使用 Meilisearch 搜索文章。
// 如果 Meilisearch 不可用，降级为 MySQL LIKE 搜索。
func SearchPosts(keyword string, page, pageSize int) ([]model.Post, int64, error) {
	if SearchClient != nil {
		posts, total, err := searchWithMeilisearch(keyword, page, pageSize)
		if err == nil {
			return posts, total, nil
		}
	}

	return searchWithMySQL(keyword, page, pageSize)
}

func searchWithMeilisearch(keyword string, page, pageSize int) ([]model.Post, int64, error) {
	cfg := config.C.Meilisearch
	resp, err := SearchClient.Index(cfg.Index).Search(keyword, &meilisearch.SearchRequest{
		Limit:  int64(pageSize),
		Offset: int64((page - 1) * pageSize),
		Filter: []string{"status = published"},
	})
	if err != nil {
		return nil, 0, err
	}

	var ids []uint
	for _, hit := range resp.Hits {
		raw, ok := hit["id"]
		if !ok {
			continue
		}
		var id uint
		if err := json.Unmarshal(raw, &id); err != nil {
			continue
		}
		ids = append(ids, id)
	}

	var posts []model.Post
	if len(ids) > 0 {
		if err := database.DB.Where("id IN ?", ids).Find(&posts).Error; err != nil {
			return nil, 0, err
		}
	}

	return posts, int64(resp.EstimatedTotalHits), nil
}

func searchWithMySQL(keyword string, page, pageSize int) ([]model.Post, int64, error) {
	var total int64
	query := database.DB.Model(&model.Post{}).Where("status = ?", model.PostStatusPublished).
		Where("title LIKE ? OR summary LIKE ? OR content LIKE ?", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var posts []model.Post
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

// stripHTML 去除 HTML 标签，提取纯文本。
func stripHTML(html string) string {
	// 简单实现：只去除 < > 标签
	var result []rune
	inTag := false
	for _, r := range html {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result = append(result, r)
		}
	}
	return string(result)
}

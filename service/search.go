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
	"time"

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
	filterAttrs := []interface{}{"status", "category_id", "tag_ids", "published_at"}
	_, _ = index.UpdateFilterableAttributes(&filterAttrs)

	return nil
}

// IndexPost 将文章索引到 Meilisearch。
func IndexPost(post *model.Post) error {
	if SearchClient == nil {
		return nil
	}

	cfg := config.C.Meilisearch

	tagIDs := make([]uint, 0, len(post.Tags))
	for _, tag := range post.Tags {
		tagIDs = append(tagIDs, tag.ID)
	}

	var categoryID uint
	if post.CategoryID != nil {
		categoryID = *post.CategoryID
	}

	doc := map[string]interface{}{
		"id":           post.ID,
		"title":        post.Title,
		"summary":      post.Summary,
		"content":      stripHTML(post.Content),
		"status":       post.Status,
		"category_id":  categoryID,
		"tag_ids":      tagIDs,
		"published_at": post.CreatedAt.Unix(),
		"created_at":   post.CreatedAt,
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

// SearchPostIDs 使用 Meilisearch 搜索文章，返回匹配的文章 ID 列表与估算总数。
// 支持按状态、分类、标签、发布时间范围过滤。
// 如果 Meilisearch 不可用，返回错误，由调用方降级到 MySQL LIKE 搜索。
func SearchPostIDs(keyword string, query model.PostQuery, limit int64) ([]uint, int64, error) {
	if SearchClient == nil {
		return nil, 0, fmt.Errorf("Meilisearch 未启用")
	}

	filters := buildMeilisearchFilters(query)

	cfg := config.C.Meilisearch
	resp, err := SearchClient.Index(cfg.Index).Search(keyword, &meilisearch.SearchRequest{
		Limit:  limit,
		Offset: 0,
		Filter: filters,
	})
	if err != nil {
		return nil, 0, err
	}

	ids := make([]uint, 0, len(resp.Hits))
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

	return ids, int64(resp.EstimatedTotalHits), nil
}

// buildMeilisearchFilters 根据 PostQuery 构造 Meilisearch 过滤条件。
func buildMeilisearchFilters(query model.PostQuery) []string {
	filters := []string{"status = published"}
	if query.Status != "" {
		filters = []string{fmt.Sprintf("status = %s", query.Status)}
	}
	if query.CategoryID > 0 {
		filters = append(filters, fmt.Sprintf("category_id = %d", query.CategoryID))
	}
	if query.TagID > 0 {
		filters = append(filters, fmt.Sprintf("tag_ids = %d", query.TagID))
	}
	if query.DateFrom != "" {
		if t, err := time.Parse("2006-01-02", query.DateFrom); err == nil {
			filters = append(filters, fmt.Sprintf("published_at >= %d", t.Unix()))
		}
	}
	if query.DateTo != "" {
		if t, err := time.Parse("2006-01-02", query.DateTo); err == nil {
			filters = append(filters, fmt.Sprintf("published_at <= %d", t.Add(24*time.Hour).Unix()-1))
		}
	}
	return filters
}

// SearchPosts 使用 Meilisearch 搜索文章。
// 如果 Meilisearch 不可用，降级为 MySQL LIKE 搜索。
// 注意：返回的结果未预加载 Author/Category/Tags，也未填充 LikeCount，
// 建议优先使用 SearchPostIDs 结合数据库查询获取完整数据。
func SearchPosts(keyword string, query model.PostQuery, page, pageSize int) ([]model.Post, int64, error) {
	ids, total, err := SearchPostIDs(keyword, query, int64(pageSize))
	if err == nil {
		var posts []model.Post
		if len(ids) > 0 {
			if err := database.DB.Where("id IN ?", ids).Find(&posts).Error; err != nil {
				return nil, 0, err
			}
		}
		return posts, total, nil
	}

	return searchWithMySQL(keyword, page, pageSize)
}

func searchWithMeilisearch(keyword string, query model.PostQuery, page, pageSize int) ([]model.Post, int64, error) {
	posts, total, err := SearchPosts(keyword, query, page, pageSize)
	return posts, total, err
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

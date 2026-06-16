package service

import (
	"blog/database"
	"blog/model"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Redis key 常量
const (
	CacheKeyCategories = "blog:categories"
	CacheKeyTags       = "blog:tags"
	CacheKeyHotPosts   = "blog:hot_posts"
	CacheKeyPost       = "blog:post:%d"
	CacheExpireShort   = 5 * time.Minute
	CacheExpireLong    = 30 * time.Minute
)

// isRedisAvailable Redis 是否可用
func isRedisAvailable() bool {
	return database.Redis != nil
}

// GetCategoryCache 获取分类缓存
func GetCategoryCache() ([]model.Category, bool) {
	if !isRedisAvailable() {
		return nil, false
	}
	data, err := database.Redis.Get(context.Background(), CacheKeyCategories).Bytes()
	if err != nil {
		return nil, false
	}
	var categories []model.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, false
	}
	return categories, true
}

// SetCategoryCache 设置分类缓存
func SetCategoryCache(categories []model.Category) {
	if !isRedisAvailable() {
		return
	}
	data, _ := json.Marshal(categories)
	database.Redis.Set(context.Background(), CacheKeyCategories, data, CacheExpireLong)
}

// ClearCategoryCache 清除分类缓存
func ClearCategoryCache() {
	if !isRedisAvailable() {
		return
	}
	database.Redis.Del(context.Background(), CacheKeyCategories)
}

// GetTagCache 获取标签缓存
func GetTagCache() ([]model.Tag, bool) {
	if !isRedisAvailable() {
		return nil, false
	}
	data, err := database.Redis.Get(context.Background(), CacheKeyTags).Bytes()
	if err != nil {
		return nil, false
	}
	var tags []model.Tag
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, false
	}
	return tags, true
}

// SetTagCache 设置标签缓存
func SetTagCache(tags []model.Tag) {
	if !isRedisAvailable() {
		return
	}
	data, _ := json.Marshal(tags)
	database.Redis.Set(context.Background(), CacheKeyTags, data, CacheExpireLong)
}

// ClearTagCache 清除标签缓存
func ClearTagCache() {
	if !isRedisAvailable() {
		return
	}
	database.Redis.Del(context.Background(), CacheKeyTags)
}

// GetHotPostsCache 获取热门文章缓存
func GetHotPostsCache() ([]model.Post, bool) {
	if !isRedisAvailable() {
		return nil, false
	}
	data, err := database.Redis.Get(context.Background(), CacheKeyHotPosts).Bytes()
	if err != nil {
		return nil, false
	}
	var posts []model.Post
	if err := json.Unmarshal(data, &posts); err != nil {
		return nil, false
	}
	return posts, true
}

// SetHotPostsCache 设置热门文章缓存
func SetHotPostsCache(posts []model.Post) {
	if !isRedisAvailable() {
		return
	}
	data, _ := json.Marshal(posts)
	database.Redis.Set(context.Background(), CacheKeyHotPosts, data, CacheExpireShort)
}

// ClearHotPostsCache 清除热门文章缓存
func ClearHotPostsCache() {
	if !isRedisAvailable() {
		return
	}
	database.Redis.Del(context.Background(), CacheKeyHotPosts)
}

// GetPostCache 获取文章详情缓存
func GetPostCache(id uint) (*model.Post, bool) {
	if !isRedisAvailable() {
		return nil, false
	}
	key := formatPostCacheKey(id)
	data, err := database.Redis.Get(context.Background(), key).Bytes()
	if err != nil {
		return nil, false
	}
	var post model.Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, false
	}
	return &post, true
}

// SetPostCache 设置文章详情缓存
func SetPostCache(post *model.Post) {
	if !isRedisAvailable() {
		return
	}
	key := formatPostCacheKey(post.ID)
	data, _ := json.Marshal(post)
	database.Redis.Set(context.Background(), key, data, CacheExpireShort)
}

// ClearPostCache 清除文章详情缓存
func ClearPostCache(id uint) {
	if !isRedisAvailable() {
		return
	}
	database.Redis.Del(context.Background(), formatPostCacheKey(id))
}

func formatPostCacheKey(id uint) string {
	return fmt.Sprintf("blog:post:%d", id)
}

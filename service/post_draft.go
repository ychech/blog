package service

import (
	"blog/database"
	"blog/model"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

const (
	postDraftPrefix = "blog:post_draft:"
	postDraftTTL    = 7 * 24 * time.Hour
)

func postDraftKey(userID uint) string {
	return fmt.Sprintf("%s%d", postDraftPrefix, userID)
}

// SavePostDraft 保存用户文章草稿到 Redis。
func SavePostDraft(userID uint, draft *model.CreatePostRequest) error {
	if !isRedisAvailable() {
		return fmt.Errorf("Redis 不可用，无法保存草稿")
	}

	data, err := json.Marshal(draft)
	if err != nil {
		return fmt.Errorf("序列化草稿失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return database.Redis.Set(ctx, postDraftKey(userID), data, postDraftTTL).Err()
}

// GetPostDraft 从 Redis 读取用户文章草稿。
func GetPostDraft(userID uint) (*model.CreatePostRequest, error) {
	if !isRedisAvailable() {
		return nil, fmt.Errorf("Redis 不可用，无法读取草稿")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	data, err := database.Redis.Get(ctx, postDraftKey(userID)).Bytes()
	if err != nil {
		return nil, err
	}

	var draft model.CreatePostRequest
	if err := json.Unmarshal(data, &draft); err != nil {
		return nil, fmt.Errorf("反序列化草稿失败: %w", err)
	}
	return &draft, nil
}

// ClearPostDraft 清除用户文章草稿。
func ClearPostDraft(userID uint) error {
	if !isRedisAvailable() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return database.Redis.Del(ctx, postDraftKey(userID)).Err()
}

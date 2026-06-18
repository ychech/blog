package service

import (
	"blog/database"
	"blog/utils"
	"context"
	"time"
)

const tokenBlacklistPrefix = "blog:token:blacklist:"

// BlacklistToken 将指定 Token 加入黑名单，TTL 与 Token 剩余有效期一致。
// Redis 不可用时直接返回 nil，不会阻断业务流程。
func BlacklistToken(claims *utils.JWTClaims) error {
	if !isRedisAvailable() {
		return nil
	}

	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil
	}

	key := tokenBlacklistPrefix + claims.ID
	return database.Redis.Set(context.Background(), key, "1", ttl).Err()
}

// IsTokenBlacklisted 判断指定 JTI 是否已被拉黑。
func IsTokenBlacklisted(jti string) (bool, error) {
	if !isRedisAvailable() {
		return false, nil
	}

	val, err := database.Redis.Get(context.Background(), tokenBlacklistPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	return val == "1", nil
}

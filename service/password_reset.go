package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"blog/utils"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	PasswordResetPrefix = "blog:password_reset:"
	ResetTokenTTL       = 1 * time.Hour
)

// generateResetToken 生成 32 字节随机十六进制字符串。
func generateResetToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func passwordResetKey(token string) string {
	return PasswordResetPrefix + token
}

// CreatePasswordResetToken 为指定邮箱生成密码重置令牌，并将令牌存入 Redis。
// 无论邮箱是否存在都返回 nil，避免泄露用户注册信息。
func CreatePasswordResetToken(email string) (string, *model.User, error) {
	if !isRedisAvailable() {
		return "", nil, fmt.Errorf("Redis 不可用，无法重置密码")
	}

	email = utils.NormalizeEmail(email)
	if email == "" {
		return "", nil, fmt.Errorf("邮箱不能为空")
	}

	var user model.User
	if err := database.DB.Where("email = ?", email).First(&user).Error; err != nil {
		// 邮箱不存在也返回空错误，由调用方统一提示“如果邮箱存在则发送”
		return "", nil, nil
	}

	token := generateResetToken()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := database.Redis.Set(ctx, passwordResetKey(token), user.ID, ResetTokenTTL).Err(); err != nil {
		return "", nil, fmt.Errorf("保存重置令牌失败: %w", err)
	}

	return token, &user, nil
}

// SendPasswordResetEmail 发送密码重置邮件。
func SendPasswordResetEmail(email string, token string) error {
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}

	subject := "密码重置"
	body := fmt.Sprintf(
		"<p>您好，您正在重置博客账户密码。</p>"+
			"<p>请在 1 小时内点击下方链接完成重置：</p>"+
			"<p><a href=\"%s/reset-password?token=%s\">重置密码</a></p>"+
			"<p>如非本人操作，请忽略此邮件。</p>",
		config.C.App.BaseURL, token,
	)
	return utils.SendEmail(email, subject, body)
}

// ResetPassword 使用令牌重置用户密码。
func ResetPassword(token, newPassword string) error {
	if !isRedisAvailable() {
		return fmt.Errorf("Redis 不可用，无法重置密码")
	}
	if token == "" {
		return fmt.Errorf("重置令牌不能为空")
	}
	if len(newPassword) < 6 {
		return fmt.Errorf("新密码不能少于 6 位")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := passwordResetKey(token)
	userIDStr, err := database.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("重置令牌已过期或无效")
	}
	if err != nil {
		return fmt.Errorf("读取重置令牌失败: %w", err)
	}

	var userID uint
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return fmt.Errorf("重置令牌无效")
	}

	hashedPwd, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPwd).Error; err != nil {
		return fmt.Errorf("重置密码失败: %w", err)
	}

	_, _ = database.Redis.Del(ctx, key).Result()
	return nil
}

// package service 实现邮箱验证相关业务逻辑。
//
// 验证码生成后存入 Redis 并设置 TTL；用户提交正确验证码后，
// 将 users.email_verified 更新为 true，并删除 Redis 中的验证码。
package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"blog/utils"
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// EmailVerificationCodePrefix 是邮箱验证码在 Redis 中的 key 前缀。
	EmailVerificationCodePrefix = "blog:email:code:"
)

// emailVerificationKey 根据用户 ID 生成验证码 Redis key。
func emailVerificationKey(userID uint) string {
	return fmt.Sprintf("%s%d", EmailVerificationCodePrefix, userID)
}

// generateVerificationCode 生成 6 位数字验证码。
func generateVerificationCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%06d", r.Intn(1000000))
}

// SendVerificationEmail 向指定用户发送邮箱验证码。
//
// 前置条件：
//   - 邮箱验证功能已启用（config.C.EmailVerification.Enabled）
//   - Redis 可用（用于存储验证码）
//   - 用户邮箱非空
func SendVerificationEmail(userID uint, email string) error {
	cfg := config.C.EmailVerification
	if !cfg.Enabled {
		return fmt.Errorf("邮箱验证未启用")
	}
	if email == "" {
		return fmt.Errorf("邮箱不能为空")
	}
	if database.Redis == nil {
		return fmt.Errorf("Redis 不可用，无法发送验证码")
	}

	code := generateVerificationCode()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	ttl := time.Duration(cfg.CodeTTLMin) * time.Minute
	if err := database.Redis.Set(ctx, emailVerificationKey(userID), code, ttl).Err(); err != nil {
		return fmt.Errorf("保存验证码失败: %w", err)
	}

	subject := "邮箱验证"
	body := fmt.Sprintf(
		"<p>您好，感谢您注册我们的博客。</p>"+
			"<p>您的邮箱验证码是：<strong style=\"font-size:18px;\">%s</strong></p>"+
			"<p>验证码有效期为 %d 分钟，请勿泄露给他人。</p>"+
			"<p>如非本人操作，请忽略此邮件。</p>",
		code, cfg.CodeTTLMin,
	)

	if err := utils.SendEmail(email, subject, body); err != nil {
		return err
	}

	return nil
}

// VerifyEmail 校验用户提交的邮箱验证码。
// 校验通过后，将用户 email_verified 字段更新为 true。
func VerifyEmail(userID uint, code string) error {
	cfg := config.C.EmailVerification
	if !cfg.Enabled {
		return fmt.Errorf("邮箱验证未启用")
	}
	if database.Redis == nil {
		return fmt.Errorf("Redis 不可用，无法验证")
	}
	if strings.TrimSpace(code) == "" {
		return fmt.Errorf("验证码不能为空")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	key := emailVerificationKey(userID)
	storedCode, err := database.Redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("验证码已过期或不存在")
	}
	if err != nil {
		return fmt.Errorf("读取验证码失败: %w", err)
	}

	if !strings.EqualFold(storedCode, strings.TrimSpace(code)) {
		return fmt.Errorf("验证码错误")
	}

	// 更新用户验证状态
	if err := database.DB.Model(&model.User{}).Where("id = ?", userID).Update("email_verified", true).Error; err != nil {
		return fmt.Errorf("更新验证状态失败: %w", err)
	}

	// 验证成功后删除验证码
	_, err = database.Redis.Del(ctx, key).Result()
	return err
}

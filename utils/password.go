package utils

import (
	"errors"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 对密码进行哈希。
// bcrypt 是 Go 社区推荐的方式，会自动处理 salt，并能抵御彩虹表攻击。
func HashPassword(password string) (string, error) {
	// DefaultCost 为 10，数值越大越安全但越慢。生产环境通常用 10 或 12
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword 校验密码是否匹配哈希
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePasswordStrength 校验密码强度。
// 要求：长度 6-32，且同时包含大写字母、小写字母和数字。
func ValidatePasswordStrength(password string) error {
	if len(password) < 6 || len(password) > 32 {
		return errors.New("密码长度必须在 6-32 位之间")
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper {
		return errors.New("密码必须包含至少一个大写字母")
	}
	if !hasLower {
		return errors.New("密码必须包含至少一个小写字母")
	}
	if !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}
	return nil
}

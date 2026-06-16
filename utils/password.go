package utils

import (
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

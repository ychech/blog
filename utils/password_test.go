// package utils 的单元测试。
package utils

import "testing"

func TestHashPassword(t *testing.T) {
	password := "123456"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}
	if hash == "" {
		t.Error("哈希结果为空")
	}
	if hash == password {
		t.Error("哈希结果不应与明文相同")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "123456"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("密码哈希失败: %v", err)
	}

	if !CheckPassword(password, hash) {
		t.Error("正确密码应校验通过")
	}
	if CheckPassword("wrong", hash) {
		t.Error("错误密码不应校验通过")
	}
}

// package service 的单元测试。
//
// 使用 SQLite 内存数据库隔离测试环境，避免污染开发数据库。
package service

import (
	"blog/config"
	"blog/model"
	"testing"
)

func TestUserService_Register(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// 初始化默认配置，JWT 生成需要
	config.C = config.DefaultConfig()

	svc := NewUserService()
	resp, err := svc.Register(model.RegisterRequest{
		Username: "alice",
		Password: "123456",
		Nickname: "Alice",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}
	if resp.User.Username != "alice" {
		t.Errorf("用户名不匹配，期望 alice，得到 %s", resp.User.Username)
	}
	if resp.Token == "" {
		t.Error("注册后未返回 token")
	}

	// 重复注册应失败
	_, err = svc.Register(model.RegisterRequest{
		Username: "alice",
		Password: "123456",
	})
	if err == nil {
		t.Error("重复注册应返回错误")
	}
}

func TestUserService_Login(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = config.DefaultConfig()

	svc := NewUserService()
	if _, err := svc.Register(model.RegisterRequest{
		Username: "bob",
		Password: "123456",
	}); err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	// 正确密码登录
	resp, err := svc.Login(model.LoginRequest{
		Username: "bob",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("登录失败: %v", err)
	}
	if resp.User.Username != "bob" {
		t.Errorf("登录返回用户名不匹配")
	}

	// 错误密码登录
	_, err = svc.Login(model.LoginRequest{
		Username: "bob",
		Password: "wrong",
	})
	if err == nil {
		t.Error("错误密码应登录失败")
	}

	// 不存在的用户
	_, err = svc.Login(model.LoginRequest{
		Username: "notexist",
		Password: "123456",
	})
	if err == nil {
		t.Error("不存在用户应登录失败")
	}
}

func TestUserService_UpdateProfile(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	config.C = config.DefaultConfig()

	svc := NewUserService()
	resp, err := svc.Register(model.RegisterRequest{
		Username: "carol",
		Password: "123456",
	})
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	user, err := svc.UpdateProfile(resp.User.ID, model.UpdateProfileRequest{
		Nickname: "Carol New",
		Email:    "carol@example.com",
	})
	if err != nil {
		t.Fatalf("更新资料失败: %v", err)
	}
	if user.Nickname != "Carol New" {
		t.Errorf("昵称未更新")
	}
	if user.Email != "carol@example.com" {
		t.Errorf("邮箱未更新")
	}
}

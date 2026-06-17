// package config 的单元测试。
package config

import (
	"os"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.Host != DefaultServerHost {
		t.Errorf("默认服务器主机不匹配，期望 %s，得到 %s", DefaultServerHost, cfg.Server.Host)
	}
	if cfg.Server.Port != DefaultServerPort {
		t.Errorf("默认服务器端口不匹配")
	}
	if cfg.DB.Database != DefaultDBName {
		t.Errorf("默认数据库名不匹配")
	}
	if cfg.JWT.ExpireHour != DefaultJWTExpireHour {
		t.Errorf("默认 JWT 过期时间不匹配")
	}
}

func TestLoadWithEnvOverride(t *testing.T) {
	// 设置环境变量覆盖默认值
	os.Setenv("BLOG_SERVER_PORT", "9090")
	os.Setenv("BLOG_DB_NAME", "test_blog")
	defer func() {
		os.Unsetenv("BLOG_SERVER_PORT")
		os.Unsetenv("BLOG_DB_NAME")
	}()

	cfg, err := LoadE()
	if err != nil {
		t.Fatalf("加载配置失败: %v", err)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("环境变量未覆盖端口，期望 9090，得到 %s", cfg.Server.Port)
	}
	if cfg.DB.Database != "test_blog" {
		t.Errorf("环境变量未覆盖数据库名，期望 test_blog，得到 %s", cfg.DB.Database)
	}
}

func TestDSN(t *testing.T) {
	cfg := DBConfig{
		Host:     "127.0.0.1",
		Port:     "3306",
		User:     "root",
		Password: "123456",
		Database: "blog",
		Charset:  "utf8mb4",
	}
	expected := "root:123456@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local"
	if cfg.DSN() != expected {
		t.Errorf("DSN 不匹配，期望 %s，得到 %s", expected, cfg.DSN())
	}
}

func TestListenAddr(t *testing.T) {
	cfg := ServerConfig{Host: "0.0.0.0", Port: "8080"}
	if cfg.ListenAddr() != "0.0.0.0:8080" {
		t.Errorf("监听地址不匹配")
	}
}

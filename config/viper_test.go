// package config 的 Viper 加载测试。
package config

import (
	"os"
	"testing"
)

func TestLoadWithViper_Default(t *testing.T) {
	cfg, err := LoadWithViper(LoadOptions{})
	if err != nil {
		t.Fatalf("Viper 加载配置失败: %v", err)
	}
	if cfg.Server.Host != DefaultServerHost {
		t.Errorf("默认主机不匹配")
	}
}

func TestLoadWithViper_EnvOverride(t *testing.T) {
	os.Setenv("BLOG_APP_ENV", "test")
	os.Setenv("BLOG_SERVER_PORT", "9091")
	defer func() {
		os.Unsetenv("BLOG_APP_ENV")
		os.Unsetenv("BLOG_SERVER_PORT")
	}()

	cfg, err := LoadWithViper(LoadOptions{})
	if err != nil {
		t.Fatalf("Viper 加载配置失败: %v", err)
	}
	if cfg.Server.Port != "9091" {
		t.Errorf("环境变量未覆盖端口，期望 9091，得到 %s", cfg.Server.Port)
	}
}

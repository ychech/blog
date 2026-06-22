// package config 中的 validate.go 负责配置合法性校验。
//
// 校验目标：
//  1. 阻止程序因为缺少关键配置而启动。
//  2. 对潜在风险配置给出明确提示（如使用默认 JWT 密钥）。
//  3. 对越界或不合理的值进行修正（如上传大小 <= 0 时重置为默认值）。
//
// 注意：
//   - 校验函数不应执行文件系统或网络操作，只检查内存中的配置值。
//   - 对于使用默认密钥的情况，打印警告但不阻止启动，因为开发环境需要便捷启动。
package config

import (
	"fmt"
	"log"
)

// validate 校验 cfg 中的配置是否合法。
//
// 校验项包括：
//   - 服务端口不能为空
//   - 数据库主机、端口、用户名、数据库名不能为空
//   - 若使用默认 JWT 密钥则打印安全警告
//   - 最大上传文件大小若 <= 0 则重置为默认值
//
// 返回 error 表示配置不合法，调用者应当终止启动。
func validate(cfg *Config) error {
	// 服务端口必填
	if cfg.Server.Port == "" {
		return fmt.Errorf("server port 不能为空")
	}

	// 数据库连接信息必填
	if cfg.DB.Host == "" || cfg.DB.Port == "" {
		return fmt.Errorf("数据库主机或端口不能为空")
	}
	if cfg.DB.User == "" {
		return fmt.Errorf("数据库用户名不能为空")
	}
	if cfg.DB.Database == "" {
		return fmt.Errorf("数据库名不能为空")
	}

	// JWT 密钥安全提示
	// 使用默认密钥在开发环境是允许的，但生产环境必须替换，否则存在严重安全隐患。
	if cfg.JWT.Secret == DefaultJWTSecret {
		log.Println("[WARN] 正在使用默认 JWT_SECRET，生产环境请务必修改")
	}

	// 上传大小兜底
	if cfg.App.MaxUploadSize <= 0 {
		cfg.App.MaxUploadSize = DefaultMaxUploadSize
	}

	return nil
}

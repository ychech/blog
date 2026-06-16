// package config 中的 normalize.go 负责配置整理与自动补全。
//
// 设计原则：
//   1. 只做内存中的配置整理，不执行任何副作用操作（如创建目录、写文件、发网络请求）。
//   2. 统一配置格式，减少后续业务代码的兼容处理。
//   3. 对空值或不规范值进行兜底，确保配置总是处于一个可用状态。
//
// 典型整理项：
//   - 去除路径末尾多余的斜杠
//   - 数据库字符集统一转小写
//   - 必填字段若为空则填充默认值
package config

import (
	"strings"
)

// normalize 对 cfg 进行自动整理和补全。
//
// 注意：本函数保证幂等性，即多次调用 normalize(cfg) 结果相同。
// 本函数不修改文件系统，仅修改 cfg 的内存字段。
func normalize(cfg *Config) {
	// 去掉上传目录末尾多余的 /，避免拼接 URL 时出现双斜杠。
	// 例如 uploads/ → uploads，/tmp/uploads/ → /tmp/uploads
	cfg.App.UploadPath = strings.TrimRight(cfg.App.UploadPath, "/")

	// 数据库字符集统一转小写，避免 utf8mb4 与 UTF8MB4 被当作不同值处理。
	cfg.DB.Charset = strings.ToLower(cfg.DB.Charset)

	// 空值兜底：如果整理后某些关键字段仍为空，则使用默认值填充。
	if cfg.DB.Charset == "" {
		cfg.DB.Charset = DefaultDBCharset
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = DefaultJWTSecret
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = DefaultServerHost
	}
	if cfg.Server.Port == "" {
		cfg.Server.Port = DefaultServerPort
	}
}

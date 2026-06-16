// package config 中的 yaml.go 负责 YAML 配置文件的读取。
//
// YAML 配置适合作为项目的默认配置文件提交到仓库（不含敏感信息），
// 例如 config.example.yaml 展示了所有可配置项的结构。
//
// YAML 配置优先级高于 .env 文件，低于系统环境变量。
// 这意味着：
//   - 可以在 YAML 中定义通用默认值
//   - 在 .env 中定义开发环境覆盖值
//   - 在系统环境变量中定义生产环境最终值
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// applyYAMLFile 从指定的 YAML 配置文件读取配置，并覆盖到 cfg。
//
// 参数 path 是 YAML 文件路径。如果 path 为空，或者文件不存在，函数会静默返回，
// 不视为错误。这样 YAML 文件是可选的，项目可以只使用 .env 或环境变量运行。
//
// YAML 结构需要与 Config 结构体中的 yaml tag 对应，例如：
//   server:
//     host: 0.0.0.0
//     port: 8080
//   db:
//     host: 127.0.0.1
//     port: 3306
//
// 读取成功后，会直接通过 yaml.Unmarshal 反序列化到 cfg 结构体。
func applyYAMLFile(cfg *Config, path string) error {
	if path == "" {
		return nil
	}

	// 文件不存在是正常情况，不报错
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// 读取 YAML 文件内容
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 YAML 配置文件失败: %w", err)
	}

	// 反序列化到 Config 结构体
	// yaml.Unmarshal 会根据 yaml tag 自动映射字段
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("解析 YAML 配置文件失败: %w", err)
	}

	return nil
}

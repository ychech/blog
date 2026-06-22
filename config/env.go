// package config 中的 env.go 负责环境变量相关处理。
//
// 包括：
//  1. 从 .env 文件读取环境变量（开发时常用）。
//  2. 从系统环境变量读取配置（生产环境常用）。
//  3. 将扁平化的环境变量映射到 Config 结构体。
//
// 环境变量命名规则：
//
//	推荐使用 BLOG_ 前缀，例如 BLOG_DB_HOST、BLOG_JWT_SECRET。
//	为兼容旧写法，也支持无前缀版本，如 DB_HOST。
//	带前缀的变量优先级高于无前缀版本。
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// EnvPrefix 是环境变量名的项目前缀。
// 加上前缀可以避免与系统上其他 Go 项目或全局环境变量冲突。
const EnvPrefix = "BLOG_"

// applyEnvFile 从指定的 .env 文件读取键值对，并应用到 Config。
//
// 参数 path 是 .env 文件路径。如果 path 为空，或者文件不存在，函数会静默返回，
// 不视为错误。这样即使项目没有 .env 文件，也能靠默认值或环境变量启动。
//
// 读取成功后，会调用 applyEnvMap 把 .env 中的变量应用到 cfg。
func applyEnvFile(cfg *Config, path string) error {
	if path == "" {
		return nil
	}

	// 文件不存在是正常情况，不报错
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	// godotenv.Read 会把 .env 文件解析成 map[string]string
	vars, err := godotenv.Read(path)
	if err != nil {
		return fmt.Errorf("读取 .env 文件失败: %w", err)
	}

	applyEnvMap(cfg, vars)
	return nil
}

// applyEnv 从系统环境变量读取配置，并应用到 Config。
//
// 系统环境变量的优先级高于 .env 文件和 YAML 文件，
// 适合在容器、CI/CD 或生产服务器上通过 docker run -e 等方式注入敏感配置。
func applyEnv(cfg *Config) {
	applyEnvMap(cfg, envToMap())
}

// getEnv 从环境变量 map 中读取指定 key。
//
// 查找顺序：
//  1. 先查找带前缀的变量，例如 BLOG_DB_HOST
//  2. 如果找不到，再查找无前缀版本，例如 DB_HOST
//
// 返回值 ok 表示是否找到了该 key（即使值为空字符串也返回 true）。
func getEnv(vars map[string]string, key string) (string, bool) {
	if v, ok := vars[EnvPrefix+key]; ok {
		return v, true
	}
	if v, ok := vars[key]; ok {
		return v, true
	}
	return "", false
}

// applyEnvMap 将环境变量 map 中的值映射到 Config 结构体。
//
// 注意：
//   - 这里采用显式字段映射，而不是反射。虽然反射更通用，但显式映射代码更清晰，
//     编译期可检查，也便于后续扩展和维护。
//   - 对于密码等字段，允许值为空字符串（用 ok 判断存在性即可），因为空密码是合法配置。
//   - 对于字符串字段，只有非空时才覆盖默认值。
func applyEnvMap(cfg *Config, vars map[string]string) {
	// Server 配置
	if v, ok := getEnv(vars, "SERVER_HOST"); ok && v != "" {
		cfg.Server.Host = v
	}
	if v, ok := getEnv(vars, "SERVER_PORT"); ok && v != "" {
		cfg.Server.Port = v
	}

	// DB 配置
	if v, ok := getEnv(vars, "DB_HOST"); ok && v != "" {
		cfg.DB.Host = v
	}
	if v, ok := getEnv(vars, "DB_PORT"); ok && v != "" {
		cfg.DB.Port = v
	}
	if v, ok := getEnv(vars, "DB_USER"); ok && v != "" {
		cfg.DB.User = v
	}
	if v, ok := getEnv(vars, "DB_PASSWORD"); ok {
		cfg.DB.Password = v
	}
	if v, ok := getEnv(vars, "DB_NAME"); ok && v != "" {
		cfg.DB.Database = v
	}
	if v, ok := getEnv(vars, "DB_CHARSET"); ok && v != "" {
		cfg.DB.Charset = v
	}
	// 简单支持一个只读从库：BLOG_DB_REPLICA_HOST
	if v, ok := getEnv(vars, "DB_REPLICA_HOST"); ok && v != "" {
		replica := DBReplicaConfig{
			Host:     v,
			Port:     cfg.DB.Port,
			User:     cfg.DB.User,
			Password: cfg.DB.Password,
		}
		if rv, ok := getEnv(vars, "DB_REPLICA_PORT"); ok && rv != "" {
			replica.Port = rv
		}
		if rv, ok := getEnv(vars, "DB_REPLICA_USER"); ok && rv != "" {
			replica.User = rv
		}
		if rv, ok := getEnv(vars, "DB_REPLICA_PASSWORD"); ok {
			replica.Password = rv
		}
		cfg.DB.Replicas = []DBReplicaConfig{replica}
	}

	// Redis 配置
	if v, ok := getEnv(vars, "REDIS_HOST"); ok && v != "" {
		cfg.Redis.Host = v
	}
	if v, ok := getEnv(vars, "REDIS_PORT"); ok && v != "" {
		cfg.Redis.Port = v
	}
	if v, ok := getEnv(vars, "REDIS_PASSWORD"); ok {
		cfg.Redis.Password = v
	}
	if v, ok := getEnv(vars, "REDIS_DB"); ok && v != "" {
		cfg.Redis.DB = parseInt(v, cfg.Redis.DB)
	}

	// JWT 配置
	if v, ok := getEnv(vars, "JWT_SECRET"); ok && v != "" {
		cfg.JWT.Secret = v
	}
	if v, ok := getEnv(vars, "JWT_EXPIRE_HOUR"); ok && v != "" {
		cfg.JWT.ExpireHour = parseInt(v, cfg.JWT.ExpireHour)
	}

	// App 配置
	if v, ok := getEnv(vars, "UPLOAD_PATH"); ok && v != "" {
		cfg.App.UploadPath = v
	}
	if v, ok := getEnv(vars, "MAX_UPLOAD_SIZE"); ok && v != "" {
		cfg.App.MaxUploadSize = parseInt64(v, cfg.App.MaxUploadSize)
	}

	// RateLimit 配置
	if v, ok := getEnv(vars, "RATE_LIMIT_ENABLED"); ok && v != "" {
		cfg.RateLimit.Enabled = parseBool(v, cfg.RateLimit.Enabled)
	}
	if v, ok := getEnv(vars, "RATE_LIMIT_MODE"); ok && v != "" {
		cfg.RateLimit.Mode = v
	}
	if v, ok := getEnv(vars, "RATE_LIMIT_REQUESTS"); ok && v != "" {
		cfg.RateLimit.Requests = parseInt(v, cfg.RateLimit.Requests)
	}
	if v, ok := getEnv(vars, "RATE_LIMIT_WINDOW_SEC"); ok && v != "" {
		cfg.RateLimit.WindowSec = parseInt(v, cfg.RateLimit.WindowSec)
	}

	// Email 配置
	if v, ok := getEnv(vars, "EMAIL_HOST"); ok && v != "" {
		cfg.Email.Host = v
	}
	if v, ok := getEnv(vars, "EMAIL_PORT"); ok && v != "" {
		cfg.Email.Port = parseInt(v, cfg.Email.Port)
	}
	if v, ok := getEnv(vars, "EMAIL_USERNAME"); ok && v != "" {
		cfg.Email.Username = v
	}
	if v, ok := getEnv(vars, "EMAIL_PASSWORD"); ok {
		cfg.Email.Password = v
	}
	if v, ok := getEnv(vars, "EMAIL_FROM"); ok && v != "" {
		cfg.Email.From = v
	}
	if v, ok := getEnv(vars, "EMAIL_ENABLE_SSL"); ok && v != "" {
		cfg.Email.EnableSSL = parseBool(v, cfg.Email.EnableSSL)
	}

	// EmailVerification 配置
	if v, ok := getEnv(vars, "EMAIL_VERIFICATION_ENABLED"); ok && v != "" {
		cfg.EmailVerification.Enabled = parseBool(v, cfg.EmailVerification.Enabled)
	}
	if v, ok := getEnv(vars, "EMAIL_VERIFICATION_REQUIRED"); ok && v != "" {
		cfg.EmailVerification.Required = parseBool(v, cfg.EmailVerification.Required)
	}
	if v, ok := getEnv(vars, "EMAIL_VERIFICATION_CODE_TTL_MIN"); ok && v != "" {
		cfg.EmailVerification.CodeTTLMin = parseInt(v, cfg.EmailVerification.CodeTTLMin)
	}

	// Tracing 配置
	if v, ok := getEnv(vars, "TRACING_ENABLED"); ok && v != "" {
		cfg.Tracing.Enabled = parseBool(v, cfg.Tracing.Enabled)
	}
	if v, ok := getEnv(vars, "TRACING_ENDPOINT"); ok && v != "" {
		cfg.Tracing.Endpoint = v
	}
	if v, ok := getEnv(vars, "TRACING_SAMPLE_RATE"); ok && v != "" {
		var rate float64
		if _, err := fmt.Sscanf(v, "%f", &rate); err == nil {
			cfg.Tracing.SampleRate = rate
		}
	}
	if v, ok := getEnv(vars, "TRACING_SERVICE_NAME"); ok && v != "" {
		cfg.Tracing.ServiceName = v
	}

	// Meilisearch 配置
	if v, ok := getEnv(vars, "MEILISEARCH_ENABLED"); ok && v != "" {
		cfg.Meilisearch.Enabled = parseBool(v, cfg.Meilisearch.Enabled)
	}
	if v, ok := getEnv(vars, "MEILISEARCH_HOST"); ok && v != "" {
		cfg.Meilisearch.Host = v
	}
	if v, ok := getEnv(vars, "MEILISEARCH_API_KEY"); ok {
		cfg.Meilisearch.APIKey = v
	}
	if v, ok := getEnv(vars, "MEILISEARCH_INDEX"); ok && v != "" {
		cfg.Meilisearch.Index = v
	}
}

// envToMap 把当前进程的系统环境变量列表转换为 map。
//
// 环境变量格式为 KEY=VALUE，其中 VALUE 可能包含等号，
// 所以使用 strings.SplitN(e, "=", 2) 只分割一次。
func envToMap() map[string]string {
	result := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}
	return result
}

// parseInt 将字符串解析为 int，解析失败时返回 defaultValue。
// 用于解析 REDIS_DB、JWT_EXPIRE_HOUR 等整型环境变量。
func parseInt(s string, defaultValue int) int {
	var result int
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

// parseInt64 将字符串解析为 int64，解析失败时返回 defaultValue。
// 用于解析 MAX_UPLOAD_SIZE 等 int64 类型环境变量。
func parseInt64(s string, defaultValue int64) int64 {
	var result int64
	if _, err := fmt.Sscanf(s, "%d", &result); err != nil {
		return defaultValue
	}
	return result
}

// parseBool 将字符串解析为 bool，解析失败时返回 defaultValue。
// 用于解析 RATE_LIMIT_ENABLED 等布尔类型环境变量。
func parseBool(s string, defaultValue bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}

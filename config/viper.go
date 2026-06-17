// package config 中的 viper.go 负责基于 Viper 的多环境配置加载。
//
// 设计目标：
//   1. 支持按运行环境加载不同配置文件，例如 config.dev.yaml、config.prod.yaml。
//   2. 环境变量优先级高于配置文件。
//   3. 保留原有 .env / YAML / 环境变量的加载能力，便于渐进式迁移。
//
// 环境选择规则：
//   - 读取 APP_ENV 或 BLOG_APP_ENV 环境变量
//   - 默认值为 "dev"
//   - 根据环境值查找 config.{env}.yaml 文件
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// LoadWithViper 使用 Viper 加载配置。
//
// 加载顺序（后加载的覆盖先加载的）：
//   1. 硬编码默认值（defaultConfig）
//   2. config.{env}.yaml（如果存在）
//   3. config.yaml（如果存在，作为兜底）
//   4. .env 文件（如果存在）
//   5. 系统环境变量（最高优先级）
//
// 参数 opts 中的 EnvFile 和 YAMLFile 仍然有效；
// 如果 YAMLFile 为空，则根据 APP_ENV 自动选择配置文件。
func LoadWithViper(opts LoadOptions) (*Config, error) {
	cfg := defaultConfig()

	env := getAppEnv()
	viperInstance := viper.New()
	viperInstance.SetConfigType("yaml")

	// 1. 设置默认值
	setViperDefaults(viperInstance)

	// 2. 读取环境专属配置文件 config.{env}.yaml
	if opts.YAMLFile == "" {
		viperInstance.SetConfigName(fmt.Sprintf("config.%s", env))
		viperInstance.AddConfigPath(".")
		_ = viperInstance.ReadInConfig() // 文件不存在不视为错误
	} else {
		viperInstance.SetConfigFile(opts.YAMLFile)
		_ = viperInstance.ReadInConfig()
	}

	// 3. 再读取通用配置文件 config.yaml（如果存在）作为兜底
	commonViper := viper.New()
	commonViper.SetConfigType("yaml")
	commonViper.SetConfigName("config")
	commonViper.AddConfigPath(".")
	if err := commonViper.ReadInConfig(); err == nil {
		_ = viperInstance.MergeConfigMap(commonViper.AllSettings())
	}

	// 4. 读取 .env 文件
	if opts.EnvFile != "" {
		_ = loadEnvFile(opts.EnvFile)
	}

	// 5. 绑定环境变量
	// 注意顺序：先设置前缀和替换器，再绑定和启用 AutomaticEnv
	viperInstance.SetEnvPrefix(EnvPrefix)
	viperInstance.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	bindEnvVars(viperInstance)
	viperInstance.AutomaticEnv()

	// 6. 反序列化到 Config
	// 指定使用 yaml tag 进行字段映射，与配置文件格式保持一致
	if err := viperInstance.Unmarshal(cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "yaml"
	}); err != nil {
		return nil, fmt.Errorf("Viper 反序列化配置失败: %w", err)
	}

	// 7. 整理与校验
	normalize(cfg)
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// getAppEnv 获取当前运行环境。
// 优先读取 BLOG_APP_ENV，其次 APP_ENV，默认 dev。
func getAppEnv() string {
	if v := viper.GetString("BLOG_APP_ENV"); v != "" {
		return v
	}
	if v := viper.GetString("APP_ENV"); v != "" {
		return v
	}
	return "dev"
}

// setViperDefaults 将硬编码默认值注册到 Viper 实例。
func setViperDefaults(v *viper.Viper) {
	defaults := map[string]interface{}{
		"server.host": DefaultServerHost,
		"server.port": DefaultServerPort,
		"db.host":     DefaultDBHost,
		"db.port":     DefaultDBPort,
		"db.user":     DefaultDBUser,
		"db.password": DefaultDBPassword,
		"db.database": DefaultDBName,
		"db.charset":  DefaultDBCharset,
		"redis.host":  DefaultRedisHost,
		"redis.port":  DefaultRedisPort,
		"redis.db":    DefaultRedisDB,
		"jwt.secret":  DefaultJWTSecret,
		"jwt.expire_hour": DefaultJWTExpireHour,
		"app.upload_path":    DefaultUploadPath,
		"app.max_upload_size": DefaultMaxUploadSize,
		"rate_limit.enabled":   DefaultRateLimitEnabled,
		"rate_limit.requests":  DefaultRateLimitRequests,
		"rate_limit.window_sec": DefaultRateLimitWindowSec,
		"email.port":      DefaultEmailPort,
		"email.enable_ssl": DefaultEmailEnableSSL,
		"email_verification.enabled":     DefaultEmailVerificationEnabled,
		"email_verification.required":    DefaultEmailVerificationRequired,
		"email_verification.code_ttl_min": DefaultEmailVerificationCodeTTLMin,
	}
	for key, value := range defaults {
		v.SetDefault(key, value)
	}
}

// bindEnvVars 显式绑定关键环境变量到 Viper。
// 这样即使不使用 AutomaticEnv，也能正确读取带前缀的环境变量。
func bindEnvVars(v *viper.Viper) {
	keys := []string{
		"server.host", "server.port",
		"db.host", "db.port", "db.user", "db.password", "db.database", "db.charset",
		"redis.host", "redis.port", "redis.password", "redis.db",
		"jwt.secret", "jwt.expire_hour",
		"app.upload_path", "app.max_upload_size",
		"rate_limit.enabled", "rate_limit.mode", "rate_limit.requests", "rate_limit.window_sec",
		"email.host", "email.port", "email.username", "email.password", "email.from", "email.enable_ssl",
		"email_verification.enabled", "email_verification.required", "email_verification.code_ttl_min",
		"tracing.enabled", "tracing.endpoint", "tracing.sample_rate", "tracing.service_name",
		"meilisearch.enabled", "meilisearch.host", "meilisearch.api_key", "meilisearch.index",
	}
	for _, key := range keys {
		_ = v.BindEnv(key, EnvPrefix+strings.ToUpper(strings.ReplaceAll(key, ".", "_")))
	}
}

// loadEnvFile 读取 .env 文件并设置到系统环境变量。
// Viper 的 AutomaticEnv 会随后读取这些环境变量。
func loadEnvFile(path string) error {
	// 复用 godotenv 读取 .env 文件
	vars, err := godotenv.Read(path)
	if err != nil {
		return err
	}
	for key, value := range vars {
		_ = os.Setenv(key, value)
	}
	return nil
}

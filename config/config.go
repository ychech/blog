// package config 负责加载并统一管理应用配置。
//
// 设计目标：
//   1. 支持多种配置来源：硬编码默认值、.env 文件、YAML 文件、系统环境变量。
//   2. 配置优先级（从高到低）：系统环境变量 > YAML 文件 > .env 文件 > 硬编码默认值。
//   3. 提供统一的加载入口 config.Load()，加载失败直接退出程序，避免带着错误配置运行。
//   4. 提供 config.LoadE() 返回错误，方便单元测试或需要自定义处理的场景。
//   5. 所有默认值以常量形式集中定义，便于维护和测试。
//
// 使用方式：
//   // main.go 启动时调用
//   config.Load()
//
//   // 业务代码通过全局变量 config.C 访问
//   dsn := config.C.DB.DSN()
//   redisAddr := config.C.Redis.Addr()
//   jwtSecret := config.C.JWT.Secret
//
// 包内文件分工：
//   - config.go:   Config 结构体、加载入口、默认值、配置对象的方法
//   - env.go:      .env 文件读取与环境变量应用
//   - yaml.go:     YAML 配置文件读取
//   - normalize.go: 配置整理与自动补全
//   - validate.go: 配置合法性校验
package config

import (
	"fmt"
	"log"
)

// C 是全局配置实例，由 Load() 成功执行后初始化。
// 项目中的其他包通过 config.C.Xxx 访问配置，避免到处传参。
var C *Config

// Config 聚合应用运行所需的全部配置项。
// 使用 yaml tag 支持 YAML 反序列化，使用 json tag 便于日志输出时序列化。
type Config struct {
	Server           ServerConfig           `yaml:"server" json:"server"`                     // HTTP 服务监听配置
	DB               DBConfig               `yaml:"db" json:"db"`                             // 数据库连接配置
	Redis            RedisConfig            `yaml:"redis" json:"redis"`                       // Redis 连接配置
	JWT              JWTConfig              `yaml:"jwt" json:"jwt"`                           // JWT 签名与过期配置
	App              AppConfig              `yaml:"app" json:"app"`                           // 应用级配置（上传、静态资源等）
	OAuth            OAuthConfig            `yaml:"oauth" json:"oauth"`                       // 第三方 OAuth2 登录配置
	RateLimit        RateLimitConfig        `yaml:"rate_limit" json:"rate_limit"`             // 接口限流配置
	UserRateLimit    UserRateLimitConfig    `yaml:"user_rate_limit" json:"user_rate_limit"`   // 按用户维度限流配置
	Email            EmailConfig            `yaml:"email" json:"email"`                       // SMTP 邮件配置
	EmailVerification EmailVerificationConfig `yaml:"email_verification" json:"email_verification"` // 邮箱验证配置
	Tracing           TracingConfig          `yaml:"tracing" json:"tracing"`                                // 链路追踪配置
	Meilisearch       MeilisearchConfig      `yaml:"meilisearch" json:"meilisearch"`                        // Meilisearch 搜索配置
}

// ServerConfig 定义 HTTP 服务的监听地址。
type ServerConfig struct {
	Host string `yaml:"host" json:"host"` // 监听主机，0.0.0.0 表示监听所有网卡
	Port string `yaml:"port" json:"port"` // 监听端口
}

// DBConfig 定义 MySQL 数据库连接参数。
type DBConfig struct {
	Host     string            `yaml:"host" json:"host"`         // 数据库主机地址（主库）
	Port     string            `yaml:"port" json:"port"`         // 数据库端口
	User     string            `yaml:"user" json:"user"`         // 数据库用户名
	Password string            `yaml:"password" json:"password"` // 数据库密码
	Database string            `yaml:"database" json:"database"` // 数据库名
	Charset  string            `yaml:"charset" json:"charset"`   // 连接字符集
	Replicas []DBReplicaConfig `yaml:"replicas" json:"replicas"` // 只读从库配置列表
}

// DBReplicaConfig 定义只读从库连接参数。
type DBReplicaConfig struct {
	Host     string `yaml:"host" json:"host"`         // 从库主机地址
	Port     string `yaml:"port" json:"port"`         // 从库端口
	User     string `yaml:"user" json:"user"`         // 从库用户名
	Password string `yaml:"password" json:"password"` // 从库密码
}

// RedisConfig 定义 Redis 连接参数。
type RedisConfig struct {
	Host     string `yaml:"host" json:"host"`         // Redis 主机地址
	Port     string `yaml:"port" json:"port"`         // Redis 端口
	Password string `yaml:"password" json:"password"` // Redis 密码，无密码时为空
	DB       int    `yaml:"db" json:"db"`             // Redis 数据库编号
}

// JWTConfig 定义 JWT 令牌相关配置。
type JWTConfig struct {
	Secret     string `yaml:"secret" json:"secret"`           // 签名密钥，生产环境必须修改
	ExpireHour int    `yaml:"expire_hour" json:"expire_hour"` // Token 过期时间（小时）
}

// AppConfig 定义应用级别的业务配置。
type AppConfig struct {
	BaseURL       string `yaml:"base_url" json:"base_url"`               // 应用前端基地址，用于邮件链接
	UploadPath    string `yaml:"upload_path" json:"upload_path"`         // 上传文件保存目录
	MaxUploadSize int64  `yaml:"max_upload_size" json:"max_upload_size"` // 最大上传文件大小，单位：MB
}

// OAuthConfig 定义第三方 OAuth2 登录配置。
type OAuthConfig struct {
	GitHubEnabled      bool   `yaml:"github_enabled" json:"github_enabled"`           // 是否启用 GitHub 登录
	GitHubClientID     string `yaml:"github_client_id" json:"github_client_id"`       // GitHub OAuth App Client ID
	GitHubClientSecret string `yaml:"github_client_secret" json:"github_client_secret"` // GitHub OAuth App Client Secret
	GitHubRedirectURL  string `yaml:"github_redirect_url" json:"github_redirect_url"`  // GitHub 回调地址
}

// RateLimitConfig 定义接口限流配置。
// 基于客户端 IP 进行固定窗口计数，超过阈值后返回 429 Too Many Requests。
type RateLimitConfig struct {
	Enabled   bool   `yaml:"enabled" json:"enabled"`     // 是否启用限流
	Mode      string `yaml:"mode" json:"mode"`           // 限流模式：memory（内存）或 redis（分布式）
	Requests  int    `yaml:"requests" json:"requests"`   // 每个时间窗口内允许的最大请求数
	WindowSec int    `yaml:"window_sec" json:"window_sec"` // 时间窗口长度，单位：秒
}

// UserRateLimitConfig 定义按用户维度的接口限流配置。
type UserRateLimitConfig struct {
	Enabled   bool `yaml:"enabled" json:"enabled"`     // 是否启用
	Requests  int  `yaml:"requests" json:"requests"`   // 每个时间窗口内允许的最大请求数
	WindowSec int  `yaml:"window_sec" json:"window_sec"` // 时间窗口长度，单位：秒
}

// EmailConfig 定义 SMTP 邮件发送配置。
type EmailConfig struct {
	Host                   string `yaml:"host" json:"host"`                                       // SMTP 服务器地址
	Port                   int    `yaml:"port" json:"port"`                                       // SMTP 端口
	Username               string `yaml:"username" json:"username"`                               // 发件邮箱
	Password               string `yaml:"password" json:"password"`                               // 邮箱密码或授权码
	From                   string `yaml:"from" json:"from"`                                       // 发件人显示名称
	EnableSSL              bool   `yaml:"enable_ssl" json:"enable_ssl"`                           // 是否启用 SSL
	NotificationEmailEnabled bool `yaml:"notification_email_enabled" json:"notification_email_enabled"` // 是否启用站内通知邮件提醒
}

// EmailVerificationConfig 定义邮箱验证配置。
type EmailVerificationConfig struct {
	Enabled     bool `yaml:"enabled" json:"enabled"`         // 是否启用邮箱验证
	Required    bool `yaml:"required" json:"required"`       // 是否必须验证后才能登录
	CodeTTLMin  int  `yaml:"code_ttl_min" json:"code_ttl_min"` // 验证码有效期（分钟）
}

// TracingConfig 定义链路追踪配置。
type TracingConfig struct {
	Enabled     bool    `yaml:"enabled" json:"enabled"`           // 是否启用链路追踪
	Endpoint    string  `yaml:"endpoint" json:"endpoint"`         // OTLP 接收端地址，如 http://localhost:4318/v1/traces
	SampleRate  float64 `yaml:"sample_rate" json:"sample_rate"`   // 采样率，0.0-1.0
	ServiceName string  `yaml:"service_name" json:"service_name"` // 服务名称
}

// MeilisearchConfig 定义 Meilisearch 搜索引擎配置。
type MeilisearchConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"` // 是否启用 Meilisearch
	Host    string `yaml:"host" json:"host"`       // Meilisearch 服务地址
	APIKey  string `yaml:"api_key" json:"api_key"` // Meilisearch API Key
	Index   string `yaml:"index" json:"index"`     // 索引名称
}

// 默认配置常量。
// 所有硬编码默认值都集中在这里定义，方便统一修改和单元测试引用。
const (
	DefaultServerHost = "0.0.0.0"
	DefaultServerPort = "8080"

	DefaultDBHost     = "127.0.0.1"
	DefaultDBPort     = "3306"
	DefaultDBUser     = "root"
	DefaultDBPassword = "123456"
	DefaultDBName     = "blog"
	DefaultDBCharset  = "utf8mb4"

	DefaultRedisHost     = "127.0.0.1"
	DefaultRedisPort     = "6379"
	DefaultRedisPassword = ""
	DefaultRedisDB       = 0

	DefaultJWTSecret     = "your-secret-key-change-in-production"
	DefaultJWTExpireHour = 24

	DefaultUploadPath    = "uploads"
	DefaultMaxUploadSize = 10
	DefaultAppBaseURL    = "http://localhost:8080"

	DefaultRateLimitEnabled   = true
	DefaultRateLimitMode      = "memory"
	DefaultRateLimitRequests  = 100
	DefaultRateLimitWindowSec = 60

	DefaultEmailPort        = 587
	DefaultEmailEnableSSL   = true

	DefaultEmailVerificationEnabled    = false
	DefaultEmailVerificationRequired   = false
	DefaultEmailVerificationCodeTTLMin = 30

	DefaultTracingEnabled     = false
	DefaultTracingEndpoint    = "http://localhost:4318/v1/traces"
	DefaultTracingSampleRate  = 1.0
	DefaultTracingServiceName = "blog"

	DefaultMeilisearchEnabled = false
	DefaultMeilisearchHost    = "http://localhost:7700"
	DefaultMeilisearchIndex   = "posts"
)

// LoadOptions 是配置加载的可选参数，
// 用于自定义 .env 文件和 YAML 配置文件的路径。
type LoadOptions struct {
	EnvFile  string // .env 文件路径，为空表示不读取
	YAMLFile string // YAML 配置文件路径，为空表示不读取
}

// defaultConfig 返回硬编码的默认配置实例。
// 这是配置加载链的最底层兜底，确保即使没有任何外部配置，程序也能使用安全默认值启动。
func defaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: DefaultServerHost,
			Port: DefaultServerPort,
		},
		DB: DBConfig{
			Host:     DefaultDBHost,
			Port:     DefaultDBPort,
			User:     DefaultDBUser,
			Password: DefaultDBPassword,
			Database: DefaultDBName,
			Charset:  DefaultDBCharset,
		},
		Redis: RedisConfig{
			Host:     DefaultRedisHost,
			Port:     DefaultRedisPort,
			Password: DefaultRedisPassword,
			DB:       DefaultRedisDB,
		},
		JWT: JWTConfig{
			Secret:     DefaultJWTSecret,
			ExpireHour: DefaultJWTExpireHour,
		},
		App: AppConfig{
			BaseURL:       DefaultAppBaseURL,
			UploadPath:    DefaultUploadPath,
			MaxUploadSize: DefaultMaxUploadSize,
		},
		OAuth: OAuthConfig{
			GitHubEnabled:      false,
			GitHubClientID:     "",
			GitHubClientSecret: "",
			GitHubRedirectURL:  "http://localhost:8080/api/auth/oauth/github/callback",
		},
		RateLimit: RateLimitConfig{
			Enabled:   DefaultRateLimitEnabled,
			Mode:      DefaultRateLimitMode,
			Requests:  DefaultRateLimitRequests,
			WindowSec: DefaultRateLimitWindowSec,
		},
		UserRateLimit: UserRateLimitConfig{
			Enabled:   false,
			Requests:  60,
			WindowSec: 60,
		},
		Email: EmailConfig{
			Port:                     DefaultEmailPort,
			EnableSSL:                DefaultEmailEnableSSL,
			NotificationEmailEnabled: false,
		},
		EmailVerification: EmailVerificationConfig{
			Enabled:    DefaultEmailVerificationEnabled,
			Required:   DefaultEmailVerificationRequired,
			CodeTTLMin: DefaultEmailVerificationCodeTTLMin,
		},
		Tracing: TracingConfig{
			Enabled:     DefaultTracingEnabled,
			Endpoint:    DefaultTracingEndpoint,
			SampleRate:  DefaultTracingSampleRate,
			ServiceName: DefaultTracingServiceName,
		},
		Meilisearch: MeilisearchConfig{
			Enabled: DefaultMeilisearchEnabled,
			Host:    DefaultMeilisearchHost,
			Index:   DefaultMeilisearchIndex,
		},
	}
}

// DefaultConfig 返回默认配置实例。
// 主要用于测试场景，或在不需要从文件/环境变量加载时使用。
func DefaultConfig() *Config {
	return defaultConfig()
}

// Load 是程序启动时的标准配置加载入口。
//
// 它会调用 LoadE() 加载配置；如果加载失败，会调用 log.Fatal 打印错误并退出进程。
// 这样可以确保程序不会带着错误或不完整的配置继续运行。
//
// 通常在 main.go 的第一行调用：
//   config.Load()
func Load() {
	cfg, err := LoadE()
	if err != nil {
		log.Fatalf("配置加载失败: %v", err)
	}
	C = cfg
}

// LoadE 加载配置并返回 Config 指针。
//
// 与 Load() 的区别是：LoadE 出错时不会退出进程，而是把错误返回给调用者。
// 适合单元测试、工具脚本或需要自定义错误处理的场景。
//
// 默认读取当前目录下的 .env 和 config.yaml 文件。
func LoadE() (*Config, error) {
	return LoadWithOptions(LoadOptions{
		EnvFile:  ".env",
		YAMLFile: "config.yaml",
	})
}

// LoadWithOptions 是真正执行配置加载的函数。
//
// 加载顺序（后加载的会覆盖先加载的）：
//   1. 硬编码默认值（defaultConfig）
//   2. .env 文件（applyEnvFile）
//   3. YAML 配置文件（applyYAMLFile）
//   4. 系统环境变量（applyEnv）
//   5. 配置整理（normalize）
//   6. 合法性校验（validate）
//
// 参数 opts 允许调用者自定义 .env 和 YAML 文件的路径。
func LoadWithOptions(opts LoadOptions) (*Config, error) {
	cfg := defaultConfig()

	// 1. 读取 .env 文件中的环境变量
	if err := applyEnvFile(cfg, opts.EnvFile); err != nil {
		return nil, err
	}

	// 2. 读取 YAML 配置文件
	if err := applyYAMLFile(cfg, opts.YAMLFile); err != nil {
		return nil, err
	}

	// 3. 读取系统环境变量（最高优先级）
	applyEnv(cfg)

	// 4. 自动整理配置格式，补全依赖字段
	normalize(cfg)

	// 5. 校验配置是否合法
	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// DSN 根据 DBConfig 生成 GORM 使用的 MySQL DSN 连接字符串。
//
// DSN 格式：user:password@tcp(host:port)/dbname?charset=xxx&parseTime=True&loc=Local
// parseTime=True 确保 time.Time 类型能正确解析；loc=Local 使用本地时区。
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.Database, c.Charset,
	)
}

// Addr 返回 Redis 的连接地址，格式为 host:port。
func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// ListenAddr 返回 HTTP 服务的监听地址，格式为 host:port。
func (c ServerConfig) ListenAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

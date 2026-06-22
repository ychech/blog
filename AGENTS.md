# Agent Guide：practice 项目

本项目是一个 **Go + Gin + GORM** 个人博客后端，仓库地址 `https://github.com/ychech/blog`（位于 `practice/blog` 子目录）。本文件面向编码 Agent，补充 README 未涉及的协作约定、开发流程与常见陷阱。

## 1. 项目结构

所有业务代码在 `blog/` 目录下：

```
blog/
├── config/          # 配置加载（Viper / .env / YAML）
├── database/        # MySQL + Redis 连接初始化与迁移
├── docs/            # swaggo 自动生成的 Swagger 文档
├── handler/         # HTTP 处理器：解析参数、调用 service、返回 JSON
├── middleware/      # Gin 中间件：JWT、限流、审计、跨域、panic 恢复等
├── model/           # 数据模型、请求/响应结构体
├── router/          # 路由注册与分组
├── service/         # 业务逻辑、缓存、WebSocket Hub、邮件、搜索
├── uploads/         # 上传文件目录
└── utils/           # 通用工具：JWT、密码、响应、日志、邮件
```

分层约定：
- `handler` 不直接访问数据库，只负责参数绑定、权限校验（从 context 读取）、调用 service。
- `service` 承载业务规则，可直接使用 `database.DB` / `database.Redis`。
- `model` 只声明结构与常量，不包含业务逻辑。
- `utils` 保持通用，不依赖 service。

## 2. 开发流程

### 2.1 启动基础设施

本地开发推荐 Docker 启动依赖，然后在宿主机用 `go run main.go` 启动后端：

```bash
cd blog
# 方式一：使用 start.sh
./start.sh infra

# 方式二：直接使用 docker compose
docker compose up -d
```

启动的服务：
- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`
- Meilisearch：`127.0.0.1:7700`
- Jaeger：`127.0.0.1:16686`
- Prometheus：`127.0.0.1:9090`
- Grafana：`127.0.0.1:3000`

### 2.2 运行服务

```bash
cd blog
go run main.go
```

默认监听 `0.0.0.0:8080`，Swagger 地址 `http://localhost:8080/swagger/index.html`。

### 2.3 运行测试

```bash
cd blog
go test ./...
```

- service 层测试使用 `setupTestDB(t)` 创建独立的 SQLite 内存数据库，并自动迁移模型。
- `utils.Logger` 在 `setupTestDB` 中会自动初始化，避免异步任务中 nil logger panic。
- 涉及真实外部依赖（SMTP、Meilisearch、Redis）的测试使用 mock 或跳过未配置场景。

### 2.4 生成 Swagger 文档

修改 handler 注释后必须重新生成文档：

```bash
cd blog
go run github.com/swaggo/swag/cmd/swag@latest init
```

## 3. 编码约定

### 3.1 包与命名

- 包名使用小写单数，如 `service`、`handler`、`middleware`。
- 服务结构体使用 `XxxService` + `NewXxxService()` 工厂函数。
- handler 结构体使用 `XxxHandler` + `NewXxxHandler()` 工厂函数。
- 错误消息使用中文，面向国内用户与开发者。

### 3.2 数据库与 GORM

- 统一使用 `database.DB`。
- 写操作尽量在事务中完成（`database.DB.Transaction`）。
- 软删除通过 `gorm.DeletedAt` 实现，删除使用 `.Delete(&model)`。
- 模型变更后，**不需要手动写 migration**，`database/database.go` 启动时会 `AutoMigrate`。

### 3.3 错误处理与响应

- handler 层统一使用 `utils` 包中的响应函数：
  - `utils.Success(c, data)`
  - `utils.BadRequest(c, msg)`
  - `utils.Unauthorized(c, msg)`
  - `utils.Error(c, utils.CodeBusinessError, msg)`
- service 层返回 `error`，错误文本应清晰可读。
- 不要直接在 handler 中返回裸 `error` 给客户端。

### 3.4 异步任务

- 通知、邮件、缓存清理等建议使用 goroutine 异步执行。
- 异步函数内部若使用 `utils.Logger`，需确保日志器已初始化（测试里已处理）。
- 异步任务失败只记录日志，不阻塞主流程。

### 3.5 新增接口

1. 在 `model/model.go` 中补充请求/响应结构体（需要 Swagger 文档的）。
2. 在 `service/` 中实现业务逻辑。
3. 在 `handler/` 中实现 HTTP 入口，添加 swaggo 注释。
4. 在 `router/router.go` 中注册路由，注意公开/登录/管理员分组。
5. 如需后台任务，在 `main.go` 中启动/停止。
6. 补充 `service/*_test.go` 单元测试。
7. 重新生成 Swagger 文档。

### 3.6 路由分组

| 前缀/分组 | 中间件 | 用途 |
|---|---|---|
| `/api/auth/*` | 部分公开，部分 `JWTAuth` + `AdminAuth` | 认证与用户管理 |
| `/api/*` | 公开 | 文章列表、分类、标签、评论等 |
| `/api/*`（authorized） | `JWTAuth()` + `UserRateLimit()` | 写操作：发帖、评论、点赞、私信 |
| `/api/admin/*` | `JWTAuth()` + `AdminAuth()` | 管理员接口 |
| `/ws/notifications` | 无中间件，handler 内校验 token | WebSocket 实时通知 |

## 4. 配置

配置优先级（高到低）：
1. 系统环境变量 / `.env`
2. `config.{APP_ENV}.yaml`
3. `config.yaml`
4. 硬编码默认值（`config/config.go` 中的 `DefaultConfig()`）

环境变量推荐加 `BLOG_` 前缀（如 `BLOG_DB_HOST`），同时兼容无前缀版本。

新增配置项时：
- 在 `config/config.go` 的 `Config` 结构体中添加字段与 yaml/json tag。
- 在 `DefaultConfig()` 中提供默认值。
- 更新 `.env.example` 与 `config.example.yaml`。
- 更新 `README.md` 环境变量表。

## 5. 测试

### 5.1 service 层测试

使用 `setupTestDB(t)`：

```go
func TestXxx(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    // 创建测试数据
    user := model.User{Username: "test", Password: "hash"}
    database.DB.Create(&user)

    // 调用业务函数并断言
}
```

### 5.2 mock 外部依赖

对于邮件发送等外部调用，可在 service 包内暴露可替换的函数变量：

```go
var sendEmailFunc = utils.SendEmail
```

测试中临时替换并在 defer 中恢复。

### 5.3 handler/router 测试

目前项目 handler/router 层测试覆盖较少。如需补充，推荐使用 `httptest.NewRecorder` + `gin.CreateTestContext`，并替换 `database.DB` 为 SQLite 内存库。

## 6. 常见陷阱

- **`utils.Logger` nil**：main 启动顺序已保证日志器初始化；测试中使用 `setupTestDB` 会自动初始化。
- **Redis 未启动**：服务仍可启动，但缓存、限流、邮箱验证码、密码重置等功能会降级或报错。
- **Meilisearch 未启动**：搜索自动降级为 MySQL LIKE，不影响主流程。
- **WebSocket 跨域**：`handler/websocket.go` 的 `CheckOrigin` 默认允许 `localhost` 开发环境；生产环境会根据 `config.App.BaseURL` 限制来源。
- **管理员误删自己**：`AdminDeleteUser` / `AdminBatchDeleteUsers` 已过滤当前登录用户 ID。
- **gofmt**：提交前运行 `gofmt -w .`，保持代码格式一致。

## 7. 提交规范

建议 commit message 格式：

```
<type>(<scope>): <subject>

<body>
```

示例：

```
feat(notification): 新增 WebSocket 实时通知推送

- 引入 github.com/gorilla/websocket
- 实现用户维度 Hub 与心跳保活
- 通知创建后异步推送给在线用户
```

常见 type：`feat`、`fix`、`refactor`、`test`、`docs`、`chore`。

## 8. 后续扩展方向

- 用户通知偏好设置（按类型/渠道开关）
- Redis Pub/Sub 多实例 WebSocket 横向扩展
- 站点地图（sitemap.xml）与 RSS 订阅
- 文章导入/导出
- 内容审核（敏感词 / AI）
- 文章版本历史

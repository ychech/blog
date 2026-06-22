# 个人博客后端 (Gin + GORM + MySQL + Redis)

一个功能较完整的个人博客 RESTful API，使用 Go 语言编写。

## 技术栈

- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL
- **缓存**：Redis（可选，Redis 失败不影响主服务）
- **日志**：zap
- **认证**：JWT + bcrypt
- **前端**：Vue 3 + Vue Router + Axios + Vite
- **API 文档**：Swagger (swaggo)

## 已实现功能

### 用户与认证
1. **用户系统**：注册、登录、JWT 认证、登出、Token 刷新、拉黑旧 Token
2. **用户资料**：修改昵称/邮箱/头像、修改密码
3. **邮箱验证**：注册发送验证码，验证后激活邮箱；可配置是否必须验证后才能登录
4. **忘记/重置密码**：通过邮箱发送重置链接
5. **OAuth2 登录**：GitHub 第三方登录
6. **用户关注**：关注/取消关注、粉丝/关注列表
7. **用户角色**：普通用户 / 管理员 RBAC

### 文章与内容
8. **文章 CRUD**：创建、列表、详情、更新、删除
9. **文章搜索与筛选**：关键词搜索、按分类/标签/时间筛选、排序、分页
10. **文章状态**：草稿、已发布、定时发布
11. **草稿自动保存**：临时草稿缓存，避免误刷新丢失内容
12. **文章搜索优化**：接入 Meilisearch，关键词搜索优先走搜索引擎，失败降级 MySQL
13. **热门文章**：基于浏览量排序，Redis 缓存
14. **浏览量统计**：先写 Redis，定时同步到 MySQL，降低数据库写压力
15. **收藏夹与阅读历史**：登录用户可收藏文章、记录阅读历史
16. **个人 Feed**：基于关注用户生成文章动态流

### 评论与互动
17. **评论系统**：一级评论 + 嵌套回复、评论编辑、级联删除
18. **评论置顶/加精**：管理员可对评论置顶或加精
19. **评论举报**：用户可举报评论，管理员后台审核
20. **点赞系统**：文章点赞、评论点赞、批量点赞状态查询

### 通知与消息
21. **站内通知**：评论回复、点赞、关注、私信、勋章颁发等事件触发通知
22. **通知分类与过滤**：按类型过滤通知列表，支持未读数实时统计
23. **WebSocket 实时推送**：`/ws/notifications` 长连接，新通知实时推送给在线用户
24. **邮件通知渠道**：站内通知产生时异步发送邮件提醒给已验证邮箱
25. **站内私信**：用户间一对一私信，会话列表与未读数

### 管理后台
26. **分类/标签管理**：管理员创建、更新、删除
27. **文章批量操作**：管理员批量删除文章
28. **评论批量操作**：管理员批量删除评论
29. **用户管理**：管理员列表、详情、启用/禁用、修改角色、批量删除
30. **勋章系统**：管理员创建勋章、颁发/收回勋章；用户个人资料展示勋章
31. **审计日志**：自动记录管理员写操作，支持过滤与 CSV 导出
32. **站点统计**：管理员仪表盘统计用户数/文章数/评论数等

### 基础设施
33. **Redis 缓存**：分类列表、标签列表、热门文章、文章详情，失败自动降级
34. **统一响应格式**：统一错误码、统一 JSON 响应
35. **中间件**：请求日志、panic 恢复、跨域、JWT 认证、管理员权限、审计日志
36. **接口限流**：基于客户端 IP / 用户维度的固定窗口限流，支持 Redis 分布式限流
37. **健康检查**：`/health` 接口检查 MySQL、Redis 依赖状态
38. **优雅关闭**：收到 SIGINT/SIGTERM 后等待正在处理的请求完成再退出
39. **Prometheus 监控**：自动收集请求数、耗时、请求/响应大小指标，暴露 `/metrics` 接口
40. **链路追踪**：基于 OpenTelemetry + OTLP，集成 Jaeger/Tempo 定位慢请求
41. **MySQL 读写分离**：基于 GORM dbresolver 插件，读请求自动路由到只读从库
42. **单元测试**：service/config/utils 层单元测试，使用 SQLite 内存数据库隔离
43. **多环境配置**：基于 Viper 支持 `.env` / `config.yaml` / 环境变量多来源加载

## 项目结构

```
blog/
├── config/              # 配置管理
│   ├── config.go        # Config 结构体、Load/LoadE/LoadWithOptions 入口
│   ├── env.go           # .env 文件与环境变量处理
│   ├── yaml.go          # YAML 配置文件处理
│   ├── normalize.go     # 配置整理与自动补全
│   └── validate.go      # 配置校验
├── database/            # MySQL + Redis 连接初始化
├── docs/                # Swagger 自动生成的 API 文档
├── handler/             # HTTP 请求处理器
├── middleware/          # JWT 认证、日志、恢复、跨域、管理员权限、限流
├── model/               # 数据模型与请求/响应结构
├── router/              # 路由注册
├── service/             # 业务逻辑层 + Redis 缓存 + WebSocket Hub
├── utils/               # JWT、密码加密、统一响应、日志、邮件发送
├── uploads/             # 上传文件目录
├── frontend/            # Vue 3 前端项目
│   ├── src/
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── .env.example         # 环境变量示例
├── config.example.yaml  # YAML 配置文件示例
├── main.go
├── docker-compose.yml
└── README.md
```

## 架构与请求流程

```
┌─────────────┐     HTTP      ┌─────────────┐
│   客户端     │ ────────────▶ │   router    │
└─────────────┘               └──────┬──────┘
                                     │
              ┌──────────────────────┼──────────────────────┐
              ▼                      ▼                      ▼
        middleware          handler (参数校验)         静态资源
              │                      │
              ▼                      ▼
        JWT/日志/恢复            service (业务逻辑)
                                        │
                                        ▼
                              ┌─────────────────┐
                              │   database      │
                              │  MySQL / Redis  │
                              └─────────────────┘
```

### 分层职责

| 层 | 职责 | 代表文件 |
|---|---|---|
| **router** | 注册路由、挂载中间件、分组公开/受保护接口 | `router/router.go` |
| **middleware** | 认证、日志、跨域、panic 恢复 | `middleware/*.go` |
| **handler** | 解析 HTTP 参数、调用 service、返回 JSON | `handler/*.go` |
| **service** | 业务逻辑、数据库事务、缓存读写 | `service/*.go` |
| **database** | 连接 MySQL/Redis、自动迁移 | `database/database.go` |
| **model** | 数据模型、请求/响应结构体 | `model/model.go` |
| **utils** | 通用工具：JWT、密码、响应、日志 | `utils/*.go` |

## Docker 一键部署（推荐）

项目提供两套 Docker Compose 文件，适配不同网络环境：

| 文件 | 模式 | 说明 |
|---|---|---|
| `docker-compose.yml` | 基础设施 | 仅启动 MySQL / Redis / Meilisearch / Jaeger / Prometheus / Grafana，适合本地开发 |
| `docker-compose.full.yml` | 全栈部署 | 包含上述基础设施 + 构建并运行 Go 后端容器，适合服务器 |

### 本地开发（推荐）

因为国内网络可能拉不到 `golang:1.23-alpine` / `alpine:latest`，本地开发建议：

1. 用 Docker 启动基础设施：

```bash
cd blog
./start.sh infra
```

2. 在另一个终端用本地 Go 直接启动后端：

```bash
APP_ENV=dev go run main.go
```

这样 Go 后端连接 Docker 里的 MySQL / Redis / Meilisearch / Jaeger，无需拉取 Go 镜像。

### 服务器全容器化部署

如果服务器能正常拉取 Docker Hub 镜像：

```bash
cd blog
./start.sh full
```

这会构建并启动 Go 后端容器 + 全部基础设施。

### 启动的服务

| 服务 | 地址 | 说明 |
|---|---|---|
| 博客后端 | http://localhost:8080 | Go + Gin |
| Swagger | http://localhost:8080/swagger/index.html | API 文档 |
| Grafana | http://localhost:3000 | 账号 admin / admin |
| Prometheus | http://localhost:9090 | 指标采集 |
| Jaeger | http://localhost:16686 | 链路追踪 |
| Meilisearch | http://localhost:7700 | 全文搜索 |

### 常用命令

```bash
./start.sh infra   # 仅启动基础设施（默认）
./start.sh full    # 全栈启动（含后端镜像构建）
./start.sh down    # 停止并移除容器
./start.sh logs    # 查看基础设施日志
./start.sh build   # 仅构建 blog 后端镜像
```

## 快速开始

### 环境要求

- Go 1.20+
- MySQL 5.7+ / 8.0
- Redis（可选，用于缓存）

### 1. 启动 MySQL 和 Redis

#### 方式一：Docker（推荐，一键启动）

```bash
cd blog
docker compose up -d
```

这会同时启动 MySQL（3306）和 Redis（6379）。

#### 方式二：本地 MySQL + Redis

```bash
# MySQL
mysql -uroot -p123456 -e "CREATE DATABASE blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Redis
redis-server
```

### 2. 配置文件

项目支持三种配置来源，优先级从高到低：

1. **系统环境变量**（最高优先级）
2. **`config.yaml`** 文件
3. **`.env`** 文件
4. **硬编码默认值**（兜底）

已提供示例文件：

```bash
cp .env.example .env
# 或
cp config.example.yaml config.yaml
```

然后按需修改：

```bash
# .env 示例（推荐使用 BLOG_ 前缀，避免与其他应用冲突）
BLOG_DB_HOST=127.0.0.1
BLOG_DB_PORT=3306
BLOG_DB_USER=root
BLOG_DB_PASSWORD=123456
BLOG_DB_NAME=blog
BLOG_JWT_SECRET=your-production-secret
```

同时兼容无前缀版本（如 `DB_HOST`），但 `BLOG_` 前缀优先级更高。

配置入口在 `main.go` 中调用 `config.Load()`，加载失败会 `log.Fatal` 直接退出。

配置包按职责拆分为多个文件：

| 文件 | 职责 |
|---|---|
| `config/config.go` | `Config` 结构体、`Load`/`LoadE`/`LoadWithOptions`、默认值、配置方法 |
| `config/env.go` | `.env` 文件读取、系统环境变量应用 |
| `config/yaml.go` | `config.yaml` 读取 |
| `config/normalize.go` | 配置整理（去斜杠、小写、自动补全） |
| `config/validate.go` | 配置校验 |

所有默认值已抽离为 `config` 包中的常量（如 `DefaultDBHost`、`DefaultJWTSecret`），便于统一维护和单元测试。

### 3. 运行服务

```bash
cd blog
go run main.go
```

默认访问：http://localhost:8080/health

### 4. 查看 Redis 缓存

如果使用了 Docker Redis，可以进入容器查看缓存：

```bash
# 查看所有 key
docker exec -it blog-redis redis-cli keys '*'

# 查看热门文章缓存
docker exec -it blog-redis redis-cli get blog:hot_posts

# 进入 redis-cli 交互模式
docker exec -it blog-redis redis-cli
```

## Swagger API 文档

项目已接入 [swaggo](https://github.com/swaggo/swag)，启动服务后访问：

```
http://localhost:8080/swagger/index.html
```

如果修改了 handler 注释，需要重新生成文档：

```bash
cd blog
~/go/bin/swag init
```

## 管理员权限

- 用户表新增 `role` 字段，默认值为 `user`，管理员为 `admin`
- 管理员额外权限：
  - 创建 / 更新 / 删除分类
  - 创建 / 删除标签
  - 删除任意文章
  - 删除任意评论
- 普通用户只能管理自己的文章和评论

### 如何设置管理员

```bash
mysql -uroot -p123456 blog -e "UPDATE users SET role = 'admin' WHERE username = '你的用户名';"
```

## 前端（Vue 3 + Vite）

项目已包含一个基于 Vue 3 的简单博客前端，位于 `frontend/` 目录。

### 前端功能

- **首页**：文章列表、关键词搜索、分类/标签筛选、排序、分页、热门文章侧边栏、文章点赞数
- **文章详情**：展示文章、浏览量、点赞数、评论列表与嵌套回复；正文支持 Markdown 渲染与代码高亮
- **登录/注册**：同一个页面切换登录与注册
- **写文章 / 编辑文章**：登录后可发布/编辑文章，支持选择分类和标签，**写文章时可输入新标签自动创建**；内置 Markdown 实时预览；支持保存为草稿或立即发布
- **个人资料**：登录后可修改昵称、邮箱，上传头像，查看获得的勋章
- **文章点赞**：登录用户可对文章点赞/取消点赞
- **评论点赞**：登录用户可对评论点赞/取消点赞
- **管理员权限**：管理员可管理分类/标签/勋章/用户，可删除任意文章/评论
- **管理后台**：管理员后台包含仪表盘、分类管理、标签管理、勋章管理、用户列表、颁发勋章等功能
- **勋章 / NFT 奖励系统**：管理员可创建勋章（支持名称、描述、图标、NFT 合约地址、Token ID、Metadata URL），并颁发给指定用户；用户在个人资料页展示勋章
- **API 文档**：访问 `/swagger/index.html` 查看自动生成的 Swagger 文档

### 启动前端

```bash
cd blog/frontend
npm install
npm run dev
```

前端默认地址：http://localhost:5173

Vite 开发服务器已将 `/api` 与 `/uploads` 代理到后端的 `http://localhost:8080`，因此前后端可以独立运行。

### 前端项目结构

```
frontend/
├── src/
│   ├── api/           # Axios 请求封装与接口定义
│   ├── components/    # PostCard、CommentList 等复用组件
│   ├── router/        # Vue Router 配置
│   ├── stores/        # 简单全局用户状态
│   ├── views/         # 页面组件（首页、详情、登录、写文章）
│   ├── App.vue        # 根组件与导航栏
│   └── main.js        # 应用入口
├── index.html
├── package.json
└── vite.config.js
```

### 构建前端

```bash
cd blog/frontend
npm run build
```

构建产物输出到 `frontend/dist/`，可与后端一起部署。

### 运行测试

```bash
cd blog
go test ./...
```

当前已覆盖：
- `config`：默认配置、环境变量覆盖、DSN/Addr 方法
- `utils`：密码哈希与校验、国际化、JWT
- `service`：用户注册/登录/资料更新、文章 CRUD/批量删除/搜索、评论/评论举报、点赞、关注、私信、收藏、阅读历史、Feed、审计日志、站点统计、通知（含邮件与 WebSocket Hub）

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| DB_HOST | 127.0.0.1 | MySQL 地址 |
| DB_PORT | 3306 | MySQL 端口 |
| DB_USER | root | MySQL 用户名 |
| DB_PASSWORD | 123456 | MySQL 密码 |
| DB_NAME | blog | 数据库名 |
| REDIS_HOST | 127.0.0.1 | Redis 地址 |
| REDIS_PORT | 6379 | Redis 端口 |
| JWT_SECRET | your-secret-key-change-in-production | JWT 密钥 |
| JWT_EXPIRE_HOUR | 24 | Token 过期时间（小时） |
| UPLOAD_PATH | uploads | 上传文件保存目录 |
| MAX_UPLOAD_SIZE | 10 | 最大上传文件大小（MB） |
| RATE_LIMIT_ENABLED | true | 是否启用接口限流 |
| RATE_LIMIT_REQUESTS | 100 | 限流窗口内最大请求数 |
| RATE_LIMIT_WINDOW_SEC | 60 | 限流窗口长度（秒） |
| EMAIL_HOST | - | SMTP 服务器地址 |
| EMAIL_PORT | 587 | SMTP 端口 |
| EMAIL_USERNAME | - | 发件邮箱 |
| EMAIL_PASSWORD | - | 邮箱密码/授权码 |
| EMAIL_FROM | - | 发件人显示名称 |
| EMAIL_ENABLE_SSL | true | 是否启用 SMTP SSL |
| EMAIL_NOTIFICATION_EMAIL_ENABLED | false | 是否启用站内通知邮件提醒 |
| EMAIL_VERIFICATION_ENABLED | false | 是否启用邮箱验证 |
| EMAIL_VERIFICATION_REQUIRED | false | 是否必须验证后才能登录 |
| EMAIL_VERIFICATION_CODE_TTL_MIN | 30 | 验证码有效期（分钟） |

## 响应格式

所有接口统一返回如下 JSON 结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

- `code` 为 `0` 表示请求成功，`data` 为业务数据。
- `code` 非 `0` 表示出现错误，`message` 描述错误原因。

### 错误码说明

| 错误码 | 含义 | HTTP 状态码 | 典型场景 |
|--------|------|-------------|----------|
| 0 | 成功 | 200 | 正常返回 |
| 400 | 请求参数错误 | 400 | 参数校验失败、ID 格式错误 |
| 401 | 未授权 | 401 | 缺少 Token、Token 无效或过期 |
| 403 | 禁止访问 | 403 | 权限不足（预留） |
| 404 | 资源不存在 | 404 | 接口不存在、文章/分类未找到 |
| 429 | 请求过于频繁 | 429 | 触发接口限流 |
| 500 | 服务器内部错误 | 500 | 数据库异常、未知 panic |
| 1001 | 业务错误 | 200 | 用户名已存在、无权修改文章等 |

## 健康检查

服务启动后可通过 `/health` 检查运行状态：

```bash
curl http://localhost:8080/health
```

正常响应（HTTP 200）：

```json
{
  "status": "ok",
  "time": "2024-01-01T12:00:00+08:00",
  "dependencies": {
    "mysql": true,
    "redis": true
  }
}
```

当 MySQL 不可用时返回 HTTP 503，并将 `status` 设置为 `unhealthy`。
Redis 为可选依赖，连接失败不影响整体健康状态。

## 监控（Prometheus）

项目已接入 Prometheus 客户端库，启动服务后访问：

```bash
curl http://localhost:8080/metrics
```

### 已暴露指标

| 指标名 | 类型 | 说明 |
|---|---|---|
| `http_requests_total` | Counter | HTTP 请求总数，标签：`method`、`path`、`status` |
| `http_request_duration_seconds` | Histogram | HTTP 请求处理耗时，标签：`method`、`path` |
| `http_request_size_bytes` | Histogram | HTTP 请求体大小，标签：`method`、`path` |
| `http_response_size_bytes` | Histogram | HTTP 响应体大小，标签：`method`、`path` |

### 本地启动 Prometheus

创建 `prometheus.yml`：

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'blog'
    static_configs:
      - targets: ['host.docker.internal:8080']
```

使用 Docker 启动 Prometheus：

```bash
docker run -d \
  -p 9090:9090 \
  -v $(pwd)/prometheus.yml:/etc/prometheus/prometheus.yml \
  prom/prometheus
```

打开 http://localhost:9090 即可查询指标，例如：

```promql
rate(http_requests_total[1m])
```

### Grafana 可视化

1. 启动 Grafana：`docker run -d -p 3000:3000 grafana/grafana`
2. 添加 Prometheus 数据源：http://localhost:9090
3. 导入官方 **Gin** 或 **Go Processes** 仪表盘，或自建面板展示 QPS、P99 耗时、错误率等

## 多环境配置

项目使用 Viper 管理多环境配置，默认根据 `APP_ENV` 或 `BLOG_APP_ENV` 环境变量选择配置文件：

| 环境 | 配置文件 |
|---|---|
| dev（默认） | `config.dev.yaml` |
| prod | `config.prod.yaml` |
| test | `config.test.yaml` |

加载优先级（从高到低）：

1. 系统环境变量 / `.env` 文件
2. `config.{env}.yaml`
3. `config.yaml`
4. 硬编码默认值

启动生产环境：

```bash
APP_ENV=prod ./blog-server
```

## Meilisearch 搜索

项目集成 Meilisearch 提供全文搜索能力：

- 文章创建/更新/删除自动同步到 Meilisearch 索引
- `/api/posts` 搜索时优先使用 Meilisearch，失败时降级为 MySQL LIKE
- 支持中文分词、拼音、容错匹配（依赖 Meilisearch 配置）

启动 Meilisearch（Docker）：

```bash
docker run -d \
  --name meilisearch \
  -p 7700:7700 \
  -e MEILI_MASTER_KEY=your-master-key \
  -v $(pwd)/meili_data:/meili_data \
  getmeili/meilisearch:latest
```

启用搜索：

```bash
BLOG_MEILISEARCH_ENABLED=true \
BLOG_MEILISEARCH_HOST=http://localhost:7700 \
BLOG_MEILISEARCH_API_KEY=your-master-key \
./blog-server
```

## MySQL 读写分离

项目使用 GORM 的 `dbresolver` 插件支持读写分离：

- 写操作（INSERT/UPDATE/DELETE）自动路由到主库
- 读操作（SELECT）自动路由到配置的只读从库
- 支持多个从库，采用随机负载均衡策略

配置示例（`config.prod.yaml`）：

```yaml
db:
  host: mysql-master
  port: 3306
  user: blog
  password: ${DB_PASSWORD}
  database: blog
  charset: utf8mb4
  replicas:
    - host: mysql-replica-1
      port: 3306
      user: blog
      password: ${DB_PASSWORD}
    - host: mysql-replica-2
      port: 3306
      user: blog
      password: ${DB_PASSWORD}
```

## 链路追踪

项目已集成 OpenTelemetry，支持将 trace 数据推送到 Jaeger、Tempo 等兼容 OTLP 的收集器。

启动 Jaeger（Docker）：

```bash
docker run -d \
  --name jaeger \
  -p 16686:16686 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

启用追踪并启动服务：

```bash
BLOG_TRACING_ENABLED=true BLOG_TRACING_ENDPOINT=http://localhost:4318/v1/traces ./blog-server
```

打开 http://localhost:16686 即可查看调用链路、耗时分布和错误追踪。

## API 文档

### 认证

#### 注册

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "password": "123456",
    "nickname": "Alice",
    "email": "alice@example.com"
  }'
```

#### 登录

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"123456"}'
```

返回 `token`，后续需要登录的接口在 Header 中携带：

```
Authorization: Bearer <token>
```

#### 获取当前用户

```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer <token>"
```

#### 更新当前用户资料（需登录）

```bash
curl -X PUT http://localhost:8080/api/auth/me \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "nickname":"新昵称",
    "email":"new@example.com",
    "avatar":"/uploads/xxx.png"
  }'
```

#### 发送邮箱验证码（需登录）

启用邮箱验证后，注册时会自动发送验证码。也可手动触发：

```bash
curl -X POST http://localhost:8080/api/auth/send-verification-email \
  -H "Authorization: Bearer <token>"
```

#### 验证邮箱验证码（需登录）

```bash
curl -X POST http://localhost:8080/api/auth/verify-email \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"code":"123456"}'
```

### 通知（需登录）

#### WebSocket 实时通知

前端建立 WebSocket 连接（token 通过 query 参数传递）：

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/notifications?token=<jwt>');
ws.onmessage = (event) => {
  const msg = JSON.parse(event.data);
  console.log(msg.event, msg.data); // event: "notification"
};
```

#### 获取通知列表

```bash
# 全部通知
curl http://localhost:8080/api/notifications \
  -H "Authorization: Bearer <token>"

# 按类型过滤（comment_reply/post_like/comment_like/follow/message/badge_award）
curl "http://localhost:8080/api/notifications?type=follow" \
  -H "Authorization: Bearer <token>"
```

#### 获取未读通知数

```bash
curl http://localhost:8080/api/notifications/unread-count \
  -H "Authorization: Bearer <token>"
```

#### 标记通知为已读

```bash
# 单条已读
curl -X PUT http://localhost:8080/api/notifications/1/read \
  -H "Authorization: Bearer <token>"

# 全部已读
curl -X PUT http://localhost:8080/api/notifications/read-all \
  -H "Authorization: Bearer <token>"
```

### 审计日志（管理员）

#### 查询审计日志

```bash
curl http://localhost:8080/api/audit-logs?page_size=20 \
  -H "Authorization: Bearer <admin-token>"
```

管理员的 POST/PUT/DELETE/PATCH 操作会自动写入 audit_logs 表。

### 用户关注（需登录）

```bash
# 关注用户
curl -X POST http://localhost:8080/api/users/2/follow \
  -H "Authorization: Bearer <token>"

# 取消关注
curl -X DELETE http://localhost:8080/api/users/2/follow \
  -H "Authorization: Bearer <token>"

# 粉丝列表
curl http://localhost:8080/api/users/2/followers

# 关注列表
curl http://localhost:8080/api/users/2/following
```

### 私信（需登录）

```bash
# 发送私信
curl -X POST http://localhost:8080/api/messages \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"receiver_id":2,"content":"你好"}'

# 会话列表
curl http://localhost:8080/api/messages/conversations \
  -H "Authorization: Bearer <token>"

# 与某用户的聊天记录
curl http://localhost:8080/api/messages/2 \
  -H "Authorization: Bearer <token>"

# 未读私信数
curl http://localhost:8080/api/messages/unread-count \
  -H "Authorization: Bearer <token>"
```

### 收藏夹与阅读历史（需登录）

```bash
# 收藏文章
curl -X POST http://localhost:8080/api/posts/1/favorite \
  -H "Authorization: Bearer <token>"

# 取消收藏
curl -X DELETE http://localhost:8080/api/posts/1/favorite \
  -H "Authorization: Bearer <token>"

# 我的收藏
curl http://localhost:8080/api/auth/favorites \
  -H "Authorization: Bearer <token>"

# 我的阅读历史
curl http://localhost:8080/api/auth/read-history \
  -H "Authorization: Bearer <token>"
```

### 个人 Feed（需登录）

```bash
curl http://localhost:8080/api/feed \
  -H "Authorization: Bearer <token>"
```

### 评论举报

```bash
# 举报评论（需登录）
curl -X POST http://localhost:8080/api/comments/1/reports \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"reason":"垃圾广告"}'

# 管理员查看举报列表
curl http://localhost:8080/api/comment-reports \
  -H "Authorization: Bearer <admin-token>"

# 管理员处理举报
curl -X PUT http://localhost:8080/api/comment-reports/1/status \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"status":"resolved"}'
```

### 勋章系统（管理员）

```bash
# 创建勋章
curl -X POST http://localhost:8080/api/badges \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"name":"优秀作者","description":"发表 10 篇以上优质文章","icon_url":"/uploads/badge.png"}'

# 颁发勋章
curl -X POST http://localhost:8080/api/badges/award \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"user_id":2,"badge_id":1,"reason":"贡献突出"}'

# 收回勋章
curl -X DELETE http://localhost:8080/api/user-badges/1 \
  -H "Authorization: Bearer <admin-token>"

# 获取用户勋章
curl http://localhost:8080/api/users/2/badges
```

### 管理员批量操作

```bash
# 批量删除文章
curl -X POST http://localhost:8080/api/admin/posts/batch-delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"ids":[1,2,3]}'

# 批量删除评论
curl -X POST http://localhost:8080/api/admin/comments/batch-delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"ids":[1,2,3]}'

# 批量删除用户
curl -X POST http://localhost:8080/api/admin/users/batch-delete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"ids":[2,3]}'
```

### 分类

```bash
# 创建分类（需登录）
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"技术"}'

# 获取分类列表
curl http://localhost:8080/api/categories

# 更新分类（需登录）
curl -X PUT http://localhost:8080/api/categories/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"编程技术"}'

# 删除分类（需登录）
curl -X DELETE http://localhost:8080/api/categories/1 \
  -H "Authorization: Bearer <token>"
```

### 标签

```bash
# 创建标签（需登录）
curl -X POST http://localhost:8080/api/tags \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"name":"Go"}'

# 获取标签列表
curl http://localhost:8080/api/tags

# 删除标签（需登录）
curl -X DELETE http://localhost:8080/api/tags/1 \
  -H "Authorization: Bearer <token>"
```

### 文章

```bash
# 创建文章（需登录）
curl -X POST http://localhost:8080/api/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title":"我的第一篇博客",
    "summary":"这是一篇测试博客",
    "content":"这是博客正文内容...",
    "status":"published",
    "category_id":1,
    "tag_ids":[1,2]
  }'

# 获取文章列表
curl "http://localhost:8080/api/posts?page=1&page_size=10"

# 搜索文章
curl "http://localhost:8080/api/posts?keyword=Gin"

# 按分类筛选
curl "http://localhost:8080/api/posts?category_id=1"

# 按标签筛选
curl "http://localhost:8080/api/posts?tag_id=1"

# 按浏览量排序
curl "http://localhost:8080/api/posts?order_by=view_count"

# 获取文章详情
curl http://localhost:8080/api/posts/1

# 更新文章（需登录）
curl -X PUT http://localhost:8080/api/posts/1 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "title":"更新后的标题",
    "status":"draft"
  }'

# 删除文章（需登录）
curl -X DELETE http://localhost:8080/api/posts/1 \
  -H "Authorization: Bearer <token>"

# 热门文章
curl "http://localhost:8080/api/posts/hot?limit=10"

# 点赞/取消点赞（需登录）
curl -X POST http://localhost:8080/api/posts/1/like \
  -H "Authorization: Bearer <token>"

# 获取文章点赞状态（无需登录）
curl http://localhost:8080/api/posts/1/like
```

### 评论点赞

```bash
# 点赞/取消点赞某条评论（需登录）
curl -X POST http://localhost:8080/api/comments/1/like \
  -H "Authorization: Bearer <token>"

# 获取评论点赞状态
curl http://localhost:8080/api/comments/1/like
```

### 评论

```bash
# 创建评论（需登录）
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "post_id":1,
    "content":"写得不错！"
  }'

# 回复评论（需登录）
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "post_id":1,
    "parent_id":1,
    "content":"谢谢！"
  }'

# 获取文章评论列表
curl http://localhost:8080/api/posts/1/comments

# 删除评论（需登录）
curl -X DELETE http://localhost:8080/api/comments/1 \
  -H "Authorization: Bearer <token>"
```

### 文件上传

```bash
curl -X POST http://localhost:8080/api/uploads \
  -H "Authorization: Bearer <token>" \
  -F "file=@/path/to/image.png"
```

返回 `url`，可通过 `http://localhost:8080/uploads/<filename>` 访问。

## 开发与部署提示

### 热重载开发

推荐使用 `air` 进行热重载开发：

```bash
go install github.com/air-verse/air@latest
cd blog
air
```

### 生产环境 checklist

1. **修改 JWT 密钥**：将 `JWT_SECRET` 设置为高强度随机字符串。
2. **关闭 GORM 详细日志**：在 `database/database.go` 中将 `logger.Info` 改为 `logger.Silent` 或 `logger.Error`。
3. **设置 Gin 为 ReleaseMode**：在 `router/router.go` 中取消 `gin.SetMode(gin.ReleaseMode)` 的注释。
4. **使用反向代理**：通过 Nginx / Caddy 处理 HTTPS、静态文件与负载均衡。
5. **配置独立 MySQL/Redis**：避免使用 docker-compose 中的默认弱密码。
6. **限制上传文件大小**：根据需求调整 `MAX_UPLOAD_SIZE`。
7. **启用日志轮转**：生产环境建议将 zap 日志输出到文件，并配合 loki/logrotate 收集。
8. **调整限流阈值**：根据实际流量调整 `RATE_LIMIT_REQUESTS` 与 `RATE_LIMIT_WINDOW_SEC`。
9. **配置优雅关闭超时**：根据接口平均耗时调整 `main.go` 中的 `Shutdown` 超时时间。
10. **接入 Prometheus/Grafana**：生产环境部署 Prometheus 抓取 `/metrics`，配置告警规则（如 5xx 错误率、P99 耗时）。

### 常见问题

**Q1：启动时提示 MySQL 连接失败**

请检查：
- MySQL 是否已启动
- `DB_HOST`、`DB_PORT`、`DB_USER`、`DB_PASSWORD` 是否正确
- `blog` 数据库是否已创建

**Q2：Redis 连接失败是否影响服务运行？**

不影响。Redis 仅用于缓存，连接失败时服务会自动降级为直接查询数据库。

**Q3：为什么已登录接口返回 401？**

请确认请求头格式为：

```
Authorization: Bearer <你的 token>
```

注意 `Bearer` 与 token 之间有一个空格。

## 学习要点

1. **分层架构**：handler → service → database，职责清晰，便于单元测试与维护
2. **JWT 认证**：登录生成 token，中间件校验，通过 gin.Context 传递用户信息
3. **GORM 高级用法**：事务、关联查询、Preload、软删除、多对多关系
4. **RESTful API 设计**：资源 + HTTP 方法 + 状态码，接口语义清晰
5. **搜索与分页**：关键词 LIKE、分类/标签筛选、Limit/Offset、排序
6. **Redis 缓存**：分类/标签/热门文章缓存，更新时失效，失败自动降级
7. **文件上传**：multipart/form-data、静态文件服务、后缀与大小校验
8. **中间件**：日志、恢复、跨域、认证的组合使用
9. **统一错误处理**：统一响应结构、错误码，便于前端统一处理
10. **环境变量配置**：所有可变配置外置，便于不同环境部署
11. **前后端分离**：Vue 3 通过 Vite 代理调用后端 API，独立开发与部署
12. **Markdown 渲染与安全**：前端渲染 Markdown，禁用原始 HTML 防 XSS；代码块语法高亮
13. **文件上传与头像裁剪**：通过已有上传接口上传头像，前端即时预览
14. **点赞功能设计**：复合唯一索引防止重复点赞，切换式点赞/取消点赞
15. **文章状态机**：草稿与发布状态，不同角色可见性控制
16. **RBAC 权限模型**：基于角色的访问控制，管理员与普通用户权限分离
17. **Swagger 文档**：使用 swaggo 自动生成可交互的 API 文档
18. **前端权限控制**：路由守卫根据用户角色拦截管理员页面
19. **性能优化**：批量查询点赞数消除 N+1；JWT 携带角色信息避免管理员中间件重复查库；评论点赞状态批量接口；搜索输入防抖
20. **勋章 / NFT 奖励系统**：管理员可创建并颁发勋章，支持 NFT 元数据字段，用户在个人资料展示
21. **接口限流实现**：固定窗口计数、内存 map + 定时清理、429 错误码与 Retry-After 响应头
22. **优雅关闭**：http.Server.Shutdown + context 超时控制，保证发布不丢请求
23. **健康检查设计**：依赖探测区分关键依赖（MySQL）与可选依赖（Redis），返回对应 HTTP 状态码
24. **Prometheus 监控**：Counter/Histogram 指标收集，私有 Registry 避免全局冲突，/metrics 接口暴露
25. **邮箱验证流程**：Redis 存储验证码 + TTL，SMTP 发送邮件，登录时校验验证状态
26. **阅读量同步**：Redis 计数削峰，定时任务 + 优雅关闭时批量同步到 MySQL
27. **Markdown 目录生成**：markdown-it token 解析提取标题，自定义渲染器注入锚点 ID
28. **单元测试实践**：SQLite 内存数据库隔离 service 层测试，配置/工具函数独立测试
29. **Viper 多环境配置**：按 APP_ENV 自动选择配置文件，环境变量优先级高于文件
30. **通知系统设计**：评论回复异步创建通知，前端轮询未读数，支持标记已读
31. **分布式限流**：Redis + Lua 脚本原子化固定窗口限流，多实例共享限流状态
32. **链路追踪**：OpenTelemetry SDK + OTLP/HTTP，Gin 中间件自动注入 trace
33. **审计日志**：c.Next() 后读取用户上下文，异步记录管理员写操作
34. **读写分离**：GORM dbresolver 插件配置只读从库，读操作自动负载均衡
35. **Meilisearch 搜索**：文章变更同步索引，搜索接口优先搜索引擎，失败降级 MySQL
36. **WebSocket 实时推送**：用户维度 Hub、心跳保活、多端在线、通知实时下发
37. **邮件通知渠道**：站内通知异步发送邮件，仅发给已验证邮箱
38. **批量管理操作**：管理员批量删除文章/评论/用户
39. **用户关注与私信**：关注关系、粉丝/关注列表、一对一私信
40. **定时发布与草稿自动保存**：文章定时发布、临时草稿缓存
41. **评论举报与审核**：用户举报、管理员处理状态流转

## 后续可扩展

- [x] 用户头像上传与资料更新
- [x] 文章点赞
- [x] 评论点赞
- [x] Markdown 渲染与代码高亮
- [x] 文章草稿 / 发布状态
- [x] 后台管理权限（管理员角色）
- [x] OpenAPI / Swagger 文档
- [x] 阅读量统计持久化 + 定时同步到 Redis：先写 Redis 再定时落库，降低 DB 压力
- [x] 邮箱注册验证：发送验证邮件，激活后才能登录
- [x] Markdown TOC：自动生成文章目录
- [x] 单元测试与集成测试：为 service 层添加测试用例
- [x] 配置中心：使用 Viper 支持多环境配置文件
- [x] 评论回复通知：被回复时发送通知
- [x] 可观测性：接入 Prometheus + Grafana 监控与告警
- [x] 分布式限流：多实例部署时使用 Redis 实现全局限流
- [x] 链路追踪：接入 Jaeger/SkyWalking 定位慢请求
- [x] 操作审计日志：记录管理员关键操作
- [x] MySQL 读写分离：主库写、从库读
- [x] 文章搜索优化：接入 Elasticsearch 或 Meilisearch
- [x] WebSocket 实时通知推送
- [x] 站内通知邮件渠道
- [x] 管理员批量删除文章/评论/用户
- [x] 用户关注与私信
- [x] 评论举报与审核
- [ ] 用户通知偏好设置：按类型/渠道开关通知
- [ ] Redis Pub/Sub 多实例 WebSocket 横向扩展
- [ ] 站点地图（sitemap.xml）与 RSS 订阅
- [ ] 文章导入/导出（Markdown / YAML Front Matter）
- [ ] 内容审核（敏感词过滤 / AI 审核）
- [ ] 文章版本历史与 diff 对比

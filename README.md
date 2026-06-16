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

1. **用户系统**：注册、登录、JWT 认证、获取当前用户
2. **文章 CRUD**：创建、列表、详情、更新、删除
3. **文章搜索与筛选**：关键词搜索、按分类/标签筛选、排序、分页
4. **分类/标签 CRUD**
5. **评论系统**：一级评论 + 嵌套回复
6. **文件上传**：图片上传，静态文件访问
7. **Redis 缓存**：分类列表、标签列表、热门文章、文章详情
8. **统一响应格式**：统一错误码、统一 JSON 响应
9. **中间件**：请求日志、panic 恢复、跨域
10. **接口限流**：基于客户端 IP 的固定窗口限流，防止接口被刷
11. **健康检查**：`/health` 接口检查 MySQL、Redis 依赖状态
12. **优雅关闭**：收到 SIGINT/SIGTERM 后等待正在处理的请求完成再退出

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
├── middleware/          # JWT 认证、日志、恢复、跨域、管理员权限
├── model/               # 数据模型与请求/响应结构
├── router/              # 路由注册
├── service/             # 业务逻辑层 + Redis 缓存
├── utils/               # JWT、密码加密、统一响应、日志
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

## 后续可扩展

- [x] 用户头像上传与资料更新
- [x] 文章点赞
- [x] 评论点赞
- [x] Markdown 渲染与代码高亮
- [x] 文章草稿 / 发布状态
- [x] 后台管理权限（管理员角色）
- [x] OpenAPI / Swagger 文档
- [ ] 阅读量统计持久化 + 定时同步到 Redis：先写 Redis 再定时落库，降低 DB 压力
- [ ] 邮箱注册验证：发送验证邮件，激活后才能登录
- [ ] Markdown TOC：自动生成文章目录
- [ ] 单元测试与集成测试：为 service 层添加测试用例
- [ ] 配置中心：使用 Viper 支持多环境配置文件
- [ ] 评论回复通知：被回复时发送通知
- [ ] 分布式限流：多实例部署时使用 Redis 实现全局限流
- [ ] 可观测性：接入 Prometheus + Grafana 监控与告警
- [ ] 链路追踪：接入 Jaeger/SkyWalking 定位慢请求
- [ ] 文章搜索优化：接入 Elasticsearch 或 Meilisearch

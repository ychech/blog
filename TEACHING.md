# 博客后端实战教学文档

> 目标：带你从零开始理解并动手扩展一个基于 **Go + Gin + GORM + MySQL + Redis** 的个人博客 RESTful API。
>
> 本文档以 `blog/` 目录下的代码为准，循序渐进地讲解项目架构、核心模块与实战技巧。

---

## 目录

1. [适合谁与学习目标](#1-适合谁与学习目标)
2. [前置知识](#2-前置知识)
3. [环境搭建](#3-环境搭建)
4. [项目启动与验证](#4-项目启动与验证)
5. [项目结构总览](#5-项目结构总览)
6. [请求的一生：完整调用链](#6-请求的一生完整调用链)
7. [配置系统：Viper 多环境管理](#7-配置系统viper-多环境管理)
8. [数据库与 GORM](#8-数据库与-gorm)
9. [用户认证：JWT 与密码安全](#9-用户认证jwt-与密码安全)
10. [实战：实现一个完整接口](#10-实战实现一个完整接口)
11. [通知系统：从站内信到实时推送](#11-通知系统从站内信到实时推送)
12. [Redis 缓存与性能优化](#12-redis-缓存与性能优化)
13. [搜索：Meilisearch 集成](#13-搜索meilisearch-集成)
14. [可观测性：监控、追踪、审计](#14-可观测性监控追踪审计)
15. [测试：如何写单元测试](#15-测试如何写单元测试)
16. [部署与生产 checklist](#16-部署与生产-checklist)
17. [练习任务](#17-练习任务)
18. [常见问题](#18-常见问题)

---

## 1. 适合谁与学习目标

### 适合谁

- 已经学完 Go 基础语法，想做完整项目的学习者
- 了解 HTTP 和 SQL，但对 Web 框架、ORM、缓存、JWT 等模块缺乏实战经验
- 希望理解一个企业级后端项目是如何组织代码、处理错误、写测试的

### 学习目标

读完并动手实践后，你将能够：

- 理解 Gin 框架的路由、中间件、上下文用法
- 使用 GORM 完成模型定义、关联、事务、软删除
- 设计分层架构：`handler -> service -> database -> model`
- 实现 JWT 认证与 RBAC 权限控制
- 使用 Redis 做缓存、限流、计数
- 编写单元测试，使用 SQLite 内存库隔离测试数据
- 接入 Swagger、Prometheus、OpenTelemetry 等可观测性工具

---

## 2. 前置知识

在继续之前，建议你已经掌握：

- Go 语言：结构体、接口、切片、map、goroutine、channel、defer
- HTTP 协议：请求方法、状态码、Header、JSON
- 关系型数据库：表、主键、外键、索引、SQL 基础
- SQL 基础：SELECT / INSERT / UPDATE / DELETE
- Git 基础：clone、commit、push

---

## 3. 环境搭建

### 3.1 安装 Go

推荐 Go 1.22+。验证：

```bash
go version
```

### 3.2 安装 Docker

项目依赖 MySQL、Redis、Meilisearch、Jaeger、Prometheus、Grafana，全部通过 Docker 启动。

```bash
docker --version
docker compose version
```

### 3.3 克隆项目

```bash
git clone https://github.com/ychech/blog.git
cd blog
```

---

## 4. 项目启动与验证

### 4.1 启动基础设施

```bash
cd blog
./start.sh infra
# 或者：docker compose up -d
```

这会启动：

- MySQL：`127.0.0.1:3306`
- Redis：`127.0.0.1:6379`
- Meilisearch：`127.0.0.1:7700`
- Jaeger：`127.0.0.1:16686`
- Prometheus：`127.0.0.1:9090`
- Grafana：`127.0.0.1:3000`

### 4.2 启动后端

```bash
cd blog
go run main.go
```

看到日志：

```text
博客服务启动: http://0.0.0.0:8080
```

### 4.3 验证

```bash
curl http://localhost:8080/health
```

正常返回：

```json
{
  "status": "ok",
  "time": "2026-06-22T14:00:00+08:00",
  "dependencies": {
    "mysql": true,
    "redis": true
  }
}
```

打开 Swagger：

```text
http://localhost:8080/swagger/index.html
```

---

## 5. 项目结构总览

```
blog/
├── config/          # 配置加载（Viper / .env / YAML）
├── database/        # MySQL + Redis 初始化与迁移
├── docs/            # Swagger 自动生成的文档
├── handler/         # HTTP 入口：解析参数、调用 service
├── middleware/      # 中间件：JWT、限流、审计、跨域
├── model/           # 数据模型与请求/响应结构体
├── router/          # 路由注册
├── service/         # 业务逻辑、缓存、WebSocket、邮件、搜索
├── uploads/         # 上传文件目录
├── utils/           # 通用工具
├── frontend/        # Vue 3 前端（可选）
├── main.go          # 入口
└── docker-compose.yml
```

分层职责：

| 层 | 职责 | 不能做的事 |
|---|---|---|
| `router` | 注册路由、挂载中间件 | 不写业务逻辑 |
| `middleware` | 认证、日志、跨域、限流 | 不直接操作数据库 |
| `handler` | 解析请求、调用 service、返回 JSON | 不写 SQL |
| `service` | 业务规则、事务、缓存 | 不处理 HTTP |
| `database` | 连接初始化、自动迁移 | 不写业务 |
| `model` | 定义表结构与请求体 | 不写逻辑 |
| `utils` | 通用工具函数 | 不依赖 service |

---

## 6. 请求的一生：完整调用链

以 **发表评论** 为例：

```text
客户端
  │ POST /api/comments  Authorization: Bearer <token>
  ▼
router/router.go  ──▶  匹配到 authorized.POST("/comments", commentHandler.Create)
  │
  ▼
middleware.JWTAuth()  ──▶  校验 token，把 userID/username/role 写入 gin.Context
  │
  ▼
middleware.UserRateLimit()  ──▶  按用户限流
  │
  ▼
handler/comment.go:Create()
  │  从 context 读取 userID，绑定 JSON 参数
  │  调用 service.NewCommentService().Create(...)
  ▼
service/comment.go:Create()
  │  校验文章是否存在、父评论是否存在
  │  写入 comments 表
  │  如果是回复，异步发送通知
  ▼
database.DB.Create(&comment)
  │
  ▼
MySQL
```

这个链路体现了 **关注点分离**：每一层只关心自己的事。

---

## 7. 配置系统：Viper 多环境管理

### 7.1 配置来源

优先级从高到低：

1. 系统环境变量 / `.env`
2. `config.{APP_ENV}.yaml`
3. `config.yaml`
4. 硬编码默认值

### 7.2 读取示例

```go
cfg, err := config.LoadWithViper(config.LoadOptions{
    EnvFile: ".env",
})
config.C = cfg
```

之后全局访问：

```go
config.C.DB.DSN()
config.C.JWT.Secret
config.C.Redis.Addr()
```

### 7.3 动手：新增一个配置项

假设我们要加一个 `App.Name`：

1. 在 `config/config.go` 的 `AppConfig` 中添加：

```go
type AppConfig struct {
    BaseURL       string `yaml:"base_url" json:"base_url"`
    UploadPath    string `yaml:"upload_path" json:"upload_path"`
    MaxUploadSize int64  `yaml:"max_upload_size" json:"max_upload_size"`
    Name          string `yaml:"name" json:"name"` // 新增
}
```

2. 在 `defaultConfig()` 中给默认值：

```go
App: AppConfig{
    BaseURL:       DefaultAppBaseURL,
    UploadPath:    DefaultUploadPath,
    MaxUploadSize: DefaultMaxUploadSize,
    Name:          "我的博客",
}
```

3. 在 `.env.example` 和 `config.example.yaml` 中添加示例。

---

## 8. 数据库与 GORM

### 8.1 模型定义

以 `Category` 为例：

```go
type Category struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    Name      string         `json:"name" gorm:"size:100;not null;uniqueIndex"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}
```

- `gorm:"primaryKey"`：主键
- `gorm:"uniqueIndex"`：唯一索引
- `gorm.DeletedAt`：启用软删除

### 8.2 关系

文章与分类：

```go
type Post struct {
    CategoryID *uint    `json:"category_id,omitempty"`
    Category   Category `json:"category,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}
```

文章与标签多对多：

```go
type Post struct {
    Tags []Tag `json:"tags,omitempty" gorm:"many2many:post_tags;"`
}
```

### 8.3 常用操作

```go
// 查询
var post model.Post
database.DB.First(&post, 1)

// 条件查询
database.DB.Where("status = ?", "published").Find(&posts)

// 创建
database.DB.Create(&post)

// 更新
database.DB.Model(&post).Update("title", "新标题")

// 软删除
database.DB.Delete(&post)

// 事务
database.DB.Transaction(func(tx *gorm.DB) error {
    tx.Create(&a)
    tx.Create(&b)
    return nil
})
```

### 8.4 自动迁移

`database/database.go` 启动时会调用：

```go
db.AutoMigrate(
    &model.User{},
    &model.Post{},
    &model.Comment{},
    // ...
)
```

所以新增模型字段后，**不需要手写 migration**。

---

## 9. 用户认证：JWT 与密码安全

### 9.1 注册流程

```text
1. 用户提交 username + password
2. handler 绑定参数
3. service 校验密码强度
4. utils.HashPassword(password) 生成 bcrypt 哈希
5. database.DB.Create(&user)
6. 返回 JWT token
```

### 9.2 JWT 生成

```go
token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
```

载荷包含：用户 ID、用户名、角色、过期时间、JTI。

### 9.3 JWT 校验中间件

```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Authorization: Bearer <token> 读取
        // 解析 token
        // 检查是否被拉黑
        // 检查用户是否被禁用
        // 把 userID/username/role 写入 context
        c.Next()
    }
}
```

后续 handler 通过 `middleware.GetCurrentUserID(c)` 获取用户。

### 9.4 密码安全

- 使用 `bcrypt` 哈希，不存储明文
- 注册时校验密码强度（长度、字符多样性）
- 登录时用 `bcrypt.CompareHashAndPassword` 校验

---

## 10. 实战：实现一个完整接口

目标：实现 **创建分类** 接口。

### 10.1 定义模型与请求体

`model/model.go` 中已有：

```go
type Category struct {
    ID   uint   `json:"id" gorm:"primaryKey"`
    Name string `json:"name" gorm:"size:100;not null;uniqueIndex"`
    // ...
}

type CreateCategoryRequest struct {
    Name string `json:"name" binding:"required"`
}
```

### 10.2 实现 service

```go
// service/category.go
func (s *CategoryService) Create(req model.CreateCategoryRequest) (*model.Category, error) {
    name := strings.TrimSpace(req.Name)
    if name == "" {
        return nil, errors.New("分类名称不能为空")
    }
    category := model.Category{Name: name}
    if err := database.DB.Create(&category).Error; err != nil {
        if isDuplicateKeyError(err) {
            return nil, errors.New("分类名称已存在")
        }
        return nil, err
    }
    return &category, nil
}
```

### 10.3 实现 handler

```go
// handler/category.go
// @Summary 创建分类
// @Tags 分类
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreateCategoryRequest true "分类名称"
// @Success 201 {object} utils.Response{data=model.Category}
// @Router /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
    var req model.CreateCategoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, "请求参数错误: "+err.Error())
        return
    }
    category, err := h.service.Create(req)
    if err != nil {
        utils.Error(c, utils.CodeBusinessError, err.Error())
        return
    }
    utils.SuccessWithStatus(c, http.StatusCreated, category)
}
```

### 10.4 注册路由

```go
// router/router.go
admin := r.Group("/api")
admin.Use(middleware.JWTAuth(), middleware.AdminAuth())
{
    admin.POST("/categories", categoryHandler.Create)
}
```

### 10.5 生成 Swagger

```bash
go run github.com/swaggo/swag/cmd/swag@latest init
```

### 10.6 测试

```bash
curl -X POST http://localhost:8080/api/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <admin-token>" \
  -d '{"name":"Go语言"}'
```

---

## 11. 通知系统：从站内信到实时推送

### 11.1 站内通知

当用户 B 回复用户 A 的评论时：

```go
// service/comment.go
if parent.AuthorID != authorID {
    notifyAsync(func() error {
        return CreateCommentReplyNotification(parent.AuthorID, comment.ID, authorName, post.Title)
    })
}
```

`CreateNotification` 把通知写入 `notifications` 表：

```go
database.DB.Create(notification)
NotifyUserRealtime(userID, notification)   // WebSocket 推送
SendNotificationEmail(userID, notification) // 邮件提醒
```

### 11.2 WebSocket 实时推送

客户端连接：

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/notifications?token=<jwt>');
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log(msg.data); // 通知对象
};
```

服务端维护一个用户维度的 Hub：

```go
type Hub struct {
    clients    map[uint]map[*Client]struct{}
    register   chan *Client
    unregister chan *Client
    notify     chan *notificationTarget
}
```

通知到来时，向该用户的所有在线客户端广播。

### 11.3 邮件通知

```go
// service/notification_email.go
func sendNotificationEmailSync(userID uint, notification *model.Notification) {
    // 1. 检查全局开关
    // 2. 查询用户邮箱与验证状态
    // 3. 渲染邮件标题/正文
    // 4. 调用 utils.SendEmail
}
```

默认关闭，需在 `.env` 中开启：

```bash
EMAIL_NOTIFICATION_EMAIL_ENABLED=true
```

---

## 12. Redis 缓存与性能优化

### 12.1 缓存场景

- 热门文章列表
- 分类列表
- 标签列表
- 文章详情

### 12.2 缓存失效

以文章更新为例：

```go
func ClearPostCache(postID uint) {
    key := fmt.Sprintf("blog:post:%d", postID)
    database.Redis.Del(ctx, key)
    ClearHotPostsCache()
}
```

### 12.3 浏览量削峰

```text
用户访问文章
  │
  ▼
Redis INCR blog:view_count:<post_id>
  │
  ▼
定时任务（每 5 分钟）
  │
  ▼
读取 Redis 批量写入 MySQL posts.view_count
```

这样避免每次访问都写 MySQL。

---

## 13. 搜索：Meilisearch 集成

### 13.1 为什么用 Meilisearch

MySQL `LIKE` 搜索慢、不支持中文分词、容错差。Meilisearch 提供：

- 中文分词
- 拼写容错
- 过滤、排序、分页

### 13.2 同步索引

文章创建/更新/删除时：

```go
func AddPostIndex(post *model.Post) error { ... }
func UpdatePostIndex(post *model.Post) error { ... }
func DeletePostIndex(postID uint) error { ... }
```

### 13.3 搜索接口

```bash
curl "http://localhost:8080/api/posts?keyword=Gin&category_id=1"
```

优先走 Meilisearch，失败降级 MySQL LIKE。

---

## 14. 可观测性：监控、追踪、审计

### 14.1 Prometheus 指标

访问 `/metrics`：

```text
http_requests_total{method="GET",path="/api/posts",status="200"}
http_request_duration_seconds_bucket{method="GET",path="/api/posts",le="0.1"}
```

### 14.2 OpenTelemetry 链路追踪

每个请求都会生成 trace，推送到 Jaeger：

```bash
BLOG_TRACING_ENABLED=true \
BLOG_TRACING_ENDPOINT=http://localhost:4318/v1/traces \
go run main.go
```

打开 http://localhost:16686 查看调用链路。

### 14.3 审计日志

管理员的写操作自动记录到 `audit_logs`：

```go
// middleware/audit.go
func AuditLog() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next() // 等 handler 执行完
        // 读取 context 中的用户信息
        // 记录 action/resource/ip 等到 audit_logs
    }
}
```

---

## 15. 测试：如何写单元测试

### 15.1 service 层测试

使用 `setupTestDB(t)` 创建独立内存数据库：

```go
func TestLikeService_Toggle(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    user := model.User{Username: "likeuser", Password: "hash"}
    database.DB.Create(&user)

    post := model.Post{Title: "post", Content: "content", AuthorID: user.ID}
    database.DB.Create(&post)

    svc := NewLikeService()
    liked, err := svc.Toggle(post.ID, user.ID)
    if err != nil {
        t.Fatalf("点赞失败: %v", err)
    }
    if !liked {
        t.Error("第一次点赞应返回 true")
    }
}
```

### 15.2 mock 外部依赖

邮件发送：

```go
sendEmailFunc = func(_, _, _ string) error { return nil }
defer func() { sendEmailFunc = defaultSendEmailFunc }()
```

### 15.3 handler 层测试

使用 `httptest`：

```go
func TestCategoryHandler_Create(t *testing.T) {
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    // 构造请求、设置 context、调用 handler
}
```

---

## 16. 部署与生产 checklist

### 16.1 Docker 全栈部署

```bash
cd blog
./start.sh full
```

### 16.2 生产 checklist

1. 修改 `JWT_SECRET` 为高强度随机字符串
2. 关闭 Gin Debug 模式：`gin.SetMode(gin.ReleaseMode)`
3. 关闭 GORM 详细日志
4. 使用 Nginx / Caddy 反向代理 + HTTPS
5. 修改 MySQL / Redis 默认密码
6. 配置日志轮转
7. 调整限流阈值
8. 部署 Prometheus + Grafana

---

## 17. 练习任务

### 初级

1. 在本地跑通项目，注册一个用户并发布一篇文章
2. 使用 Swagger 测试评论、点赞、收藏接口
3. 为 `CategoryService.Update` 补充一个单元测试

### 中级

4. 实现 **文章版本历史**：每次更新文章时保存历史记录，提供 `/api/posts/:id/history` 查询
5. 实现 **用户通知偏好设置**：允许用户关闭某类通知的邮件/WebSocket
6. 为 `handler` 层补充集成测试

### 高级

7. 使用 Redis Pub/Sub 改造 WebSocket Hub，支持多实例部署
8. 实现 **站点地图** `/sitemap.xml` 与 **RSS** `/rss.xml`
9. 接入 AI 或敏感词库做内容审核

---

## 18. 常见问题

### Q1：MySQL 连接失败

检查：

- `docker compose ps` 确认 mysql 容器运行
- `.env` 中的 `DB_HOST/DB_PORT/DB_USER/DB_PASSWORD`
- `blog` 数据库是否已创建

### Q2：Redis 失败是否影响服务

不影响主流程。Redis 仅用于缓存、限流、验证码，失败时自动降级。

### Q3：Meilisearch 未启动怎么办

搜索接口自动降级为 MySQL LIKE，无需额外处理。

### Q4：如何成为管理员

```bash
mysql -uroot -p123456 blog -e "UPDATE users SET role = 'admin' WHERE username = '你的用户名';"
```

### Q5：为什么我的修改没有出现在 Swagger 里

修改 handler 注释后必须运行：

```bash
go run github.com/swaggo/swag/cmd/swag@latest init
```

---

## 结语

本项目覆盖了后端开发中绝大多数核心技能：

- 分层架构与 RESTful 设计
- 认证、授权、安全
- 数据库与 ORM
- 缓存、搜索、消息通知
- 可观测性与测试

建议你边读边改：从最小的功能开始，比如加一个字段、加一个接口、加一个测试，逐步深入。祝你学习愉快！

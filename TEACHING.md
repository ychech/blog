# 《Go 博客后端实战》课程讲义

> 本课程基于真实项目 `practice/blog` 编写，目标是用 16 节课带你从零开始理解并动手扩展一个企业级 Go 后端。
>
> 课程采用“讲一点、做一点、测一点”的方式，每节课都有明确的知识点、代码任务和课后作业。

---

## 课程信息

| 项目 | 说明 |
|---|---|
| 课程名称 | Go 博客后端实战 |
| 适用对象 | 已学完 Go 基础语法，希望做完整项目的学习者 |
| 总课时 | 16 节课，每节 45~60 分钟 |
| 实践环境 | macOS / Linux / Windows + Docker + Go 1.22+ |
| 项目地址 | `https://github.com/ychech/blog` |
| 后端入口 | `blog/main.go` |
| API 文档 | 启动后访问 `http://localhost:8080/swagger/index.html` |

---

## 如何使用本课程

### 面向讲师

- **课前**：按“课程大纲”准备 PPT，重点标注每节课的“核心知识点”和“动手任务”。
- **课中**：采用“讲 20 分钟 + 做 20 分钟 + 答疑 10 分钟”的节奏，鼓励学生边改代码边看效果。
- **课后**：布置“课后作业”，并在下次课前用 5 分钟回顾。
- **考核**：结合课堂任务、课后作业、结课扩展项目（详见第 16 节）。

### 面向学生

1. **课前准备**：确保 Docker、Go 已安装，项目已 clone。
2. **课堂模式**：先跟着讲师跑通代码，再尝试独立修改一个参数或字段。
3. **笔记建议**：每节课记录“3 个新概念 + 2 个代码片段 + 1 个疑问”。
4. **复习方式**：用“附录 C 常见错误排查”独立解决环境问题。

### 建议的教学节奏

```text
第 1~2 节   项目跑起来 + Go 基础回顾
第 3~5 节   Gin + GORM + 配置系统
第 6~8 节   分层架构 + 用户认证 + 文章 CRUD
第 9~11 节  评论/点赞/关注/私信 + Redis 缓存
第 12~14 节 通知系统 + 搜索 + 可观测性
第 15~16 节 测试 + 部署 + 扩展方向
```

---

## 学习路线图

```text
第 1~2 节   项目跑起来 + Go 基础回顾
第 3~5 节   Gin + GORM + 配置系统
第 6~8 节   分层架构 + 用户认证 + 文章 CRUD
第 9~11 节  评论/点赞/关注/私信 + Redis 缓存
第 12~14 节 通知系统 + 搜索 + 可观测性
第 15~16 节 测试 + 部署 + 扩展方向
```

### 知识图谱

```text
                    ┌─────────────┐
                    │   HTTP 请求  │
                    └──────┬──────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   router            middleware            handler
        │                  │                  │
        └──────────────────┼──────────────────┘
                           │
                       service 业务层
                           │
              ┌────────────┼────────────┐
              │            │            │
          database      Redis      Meilisearch
              │            │            │
              └────────────┴────────────┘
                           │
                        model
```

---

## 课程大纲

| 课节 | 主题 | 核心知识点 | 动手任务 | 课后作业 |
|---|---|---|---|---|
| 01 | 课程介绍与环境搭建 | 项目结构、Docker、Go 模块 | 跑通 `/health` | 阅读 `main.go` |
| 02 | 必备的 Go 语言基础 | 结构体、接口、错误处理、goroutine | 写一个迷你 HTTP 服务 | 整理 Go 基础笔记 |
| 03 | Gin 框架入门 | 路由、中间件、Context、参数绑定 | 添加一个 `/ping` 接口 | 学习 Swagger 注释 |
| 04 | 配置系统 | Viper、.env、YAML、默认值 | 新增一个配置项 | 画出配置优先级图 |
| 05 | 数据库与 GORM | 模型、迁移、关联、事务、软删除 | 添加一个 `Tag` 模型 | 写出 CRUD 代码 |
| 06 | 分层架构设计 | handler/service/database/model | 画出请求流程图 | 重构一个接口到 service |
| 07 | 用户认证 | JWT、bcrypt、中间件、Context | 手动生成并解析 JWT | 理解黑名单机制 |
| 08 | 文章 CRUD | RESTful、参数校验、Swagger | 实现文章更新接口测试 | 为文章加字段 |
| 09 | 评论系统 | 嵌套回复、级联删除、通知 | 实现评论回复 | 画出评论表关系 |
| 10 | 点赞与关注 | 复合唯一索引、切换式点赞 | 实现评论点赞 | 写关注状态查询 |
| 11 | Redis 缓存 | 缓存失效、计数削峰 | 查看 Redis 缓存 key | 实现缓存版本的热门文章 |
| 12 | 通知系统 | WebSocket、邮件、异步任务 | 用 WebSocket 接收通知 | 配置邮件通知 |
| 13 | 搜索与 Meilisearch | 全文搜索、索引同步、降级 | 搜索一篇文章 | 理解索引更新时机 |
| 14 | 可观测性 | Prometheus、OpenTelemetry、审计 | 查看 `/metrics` | 在 Jaeger 找一条 trace |
| 15 | 单元测试 | SQLite 内存库、mock、httptest | 为 service 写测试 | 补一个 handler 测试 |
| 16 | 部署与扩展 | Docker、生产 checklist、后续方向 | 用 Docker 启动全栈 | 规划一个扩展功能 |

---

## 第 01 节：课程介绍与环境搭建

### 本节课目标

1. 了解课程目标与项目功能
2. 在本地启动 MySQL、Redis 等依赖
3. 成功运行后端并访问 `/health`

### 项目功能总览

本项目是一个个人博客后端，已实现：

- 用户注册/登录/JWT 认证
- 文章 CRUD、草稿、定时发布
- 分类/标签管理
- 评论与嵌套回复
- 文章点赞、评论点赞
- 用户关注、私信
- 收藏夹、阅读历史、Feed
- 站内通知、WebSocket 实时推送、邮件通知
- 管理员后台：用户管理、勋章、审计日志、批量操作
- Redis 缓存、Meilisearch 搜索、Prometheus 监控、OpenTelemetry 链路追踪

### 动手：启动项目

#### 步骤 1：安装依赖

```bash
# 1. Go 1.22+
go version

# 2. Docker
docker --version
docker compose version
```

#### 步骤 2：克隆项目

```bash
git clone https://github.com/ychech/blog.git
cd blog
```

#### 步骤 3：启动基础设施

```bash
cd blog
./start.sh infra
```

如果提示无权限，先执行 `chmod +x start.sh`。

#### 步骤 4：启动后端

```bash
go run main.go
```

#### 步骤 5：验证

```bash
curl http://localhost:8080/health
```

预期返回：

```json
{
  "status": "ok",
  "dependencies": {
    "mysql": true,
    "redis": true
  }
}
```

### 思考题

1. 为什么 Redis 连接失败不会影响 `/health` 返回 `ok`？
2. `main.go` 中为什么要先 `config.Load()` 再初始化数据库？

### 课后作业

- 阅读 `main.go`，写出启动顺序（至少 5 个步骤）
- 访问 Swagger 文档，列出 5 个你感兴趣的接口

---

## 第 02 节：必备的 Go 语言基础

### 本节课目标

1. 回顾项目中反复用到的 Go 语法
2. 理解指针、结构体方法、接口在本项目中的使用
3. 写一个最小的 HTTP 服务

### 核心语法回顾

#### 结构体与方法

```go
type User struct {
    ID       uint
    Username string
}

// 值接收者
func (u User) String() string {
    return u.Username
}

// 指针接收者：可修改原对象
func (u *User) SetName(name string) {
    u.Username = name
}
```

#### 错误处理

```go
user, err := service.GetUserByID(1)
if err != nil {
    return err
}
```

#### 接口

```go
type Stringer interface {
    String() string
}
```

#### Goroutine 与 Channel

```go
go func() {
    // 异步任务
}()

ch := make(chan string)
ch <- "hello"
msg := <-ch
```

### 动手：写一个迷你 HTTP 服务

```go
package main

import (
    "fmt"
    "net/http"
)

func hello(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "Hello, Go!")
}

func main() {
    http.HandleFunc("/", hello)
    http.ListenAndServe("8081", nil)
}
```

运行后访问 `http://localhost:8081/`。

### 项目中的应用

- `service/notification.go` 使用 `go` 异步发送通知
- `handler` 使用 `gin.Context` 处理请求
- `model` 大量使用结构体与 tag

### 课后作业

- 解释 `func (s *UserService) GetUserByID(id uint)` 中 `*UserService` 为什么要用指针接收者
- 写出 `interface{}` 在本项目中的两个使用场景

---

## 第 03 节：Gin 框架入门

### 本节课目标

1. 理解 Gin 的路由、中间件和 Context
2. 能读懂 `router/router.go`
3. 能添加一个简单的接口

### Gin 是什么

Gin 是一个高性能的 Go Web 框架，特点：

- 路由快速（基于 httprouter）
- 中间件机制灵活
- Context 封装请求与响应

### 最小 Gin 程序

```go
package main

import "github.com/gin-gonic/gin"

func main() {
    r := gin.Default()
    r.GET("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{"message": "pong"})
    })
    r.Run()
}
```

### 路由分组

```go
api := r.Group("/api")
{
    api.GET("/posts", postHandler.List)
    api.GET("/posts/:id", postHandler.Get)
}

admin := r.Group("/api")
admin.Use(middleware.JWTAuth(), middleware.AdminAuth())
{
    admin.POST("/categories", categoryHandler.Create)
}
```

### 参数获取

```go
// URL 参数
id := c.Param("id")

// Query 参数
page := c.DefaultQuery("page", "1")

// JSON Body
var req model.CreatePostRequest
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
}
```

### 中间件

```go
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next() // 继续执行后续中间件和 handler
        duration := time.Since(start)
        fmt.Println("耗时:", duration)
    }
}
```

### 动手：添加 `/ping` 接口

在 `router/router.go` 中添加：

```go
r.GET("/ping", func(c *gin.Context) {
    utils.Success(c, gin.H{"message": "pong"})
})
```

访问 `http://localhost:8080/ping`。

### 课后作业

- 解释 `c.Next()` 和 `c.Abort()` 的区别
- 画出 `router/router.go` 中的三个路由分组：公开、需登录、管理员

---

## 第 04 节：配置系统

### 本节课目标

1. 理解项目如何管理多环境配置
2. 掌握 `.env`、YAML、环境变量的优先级
3. 能新增一个配置项

### 配置来源优先级

```text
系统环境变量 / .env  >  config.{env}.yaml  >  config.yaml  >  硬编码默认值
```

### 核心代码

`config/config.go`：

```go
type Config struct {
    Server ServerConfig `yaml:"server" json:"server"`
    DB     DBConfig     `yaml:"db" json:"db"`
    Redis  RedisConfig  `yaml:"redis" json:"redis"`
    JWT    JWTConfig    `yaml:"jwt" json:"jwt"`
    App    AppConfig    `yaml:"app" json:"app"`
    Email  EmailConfig  `yaml:"email" json:"email"`
    // ...
}
```

### 加载入口

```go
// main.go
cfg, err := config.LoadWithViper(config.LoadOptions{
    EnvFile: ".env",
})
config.C = cfg
```

### 使用配置

```go
addr := config.C.Server.ListenAddr()
secret := config.C.JWT.Secret
```

### 动手：新增 `App.Name` 配置

1. 修改 `config/config.go`：

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
    Name: "我的博客",
}
```

3. 在 `config.example.yaml` 和 `.env.example` 中添加示例。

### 课后作业

- 解释为什么环境变量优先级最高
- 写出读取 `BLOG_JWT_SECRET` 后，代码中如何访问

---

## 第 05 节：数据库与 GORM

### 本节课目标

1. 理解 GORM 模型定义
2. 掌握常用 CRUD、关联、事务、软删除
3. 能新增一个模型并自动迁移

### 模型示例

```go
type Post struct {
    ID         uint           `json:"id" gorm:"primaryKey"`
    Title      string         `json:"title" gorm:"size:255;not null"`
    Content    string         `json:"content" gorm:"type:text;not null"`
    AuthorID   uint           `json:"author_id" gorm:"not null"`
    Author     User           `json:"author,omitempty" gorm:"foreignKey:AuthorID"`
    CategoryID *uint          `json:"category_id,omitempty"`
    Category   Category       `json:"category,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
    Tags       []Tag          `json:"tags,omitempty" gorm:"many2many:post_tags;"`
    CreatedAt  time.Time      `json:"created_at"`
    UpdatedAt  time.Time      `json:"updated_at"`
    DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}
```

### 自动迁移

```go
// database/database.go
db.AutoMigrate(
    &model.User{},
    &model.Post{},
    &model.Comment{},
    // ...
)
```

### 常用操作

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

### 关联查询

```go
// 预加载作者和分类
database.DB.Preload("Author").Preload("Category").Preload("Tags").First(&post, 1)
```

### 动手：新增 `Tag` 模型字段

在 `model/model.go` 的 `Tag` 中添加 `Color` 字段：

```go
type Tag struct {
    ID        uint   `json:"id" gorm:"primaryKey"`
    Name      string `json:"name" gorm:"size:100;not null;uniqueIndex"`
    Color     string `json:"color" gorm:"size:20"` // 新增
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

重新运行 `go run main.go`，GORM 会自动添加该列。

### 课后作业

- 解释软删除和硬删除的区别
- 写出 `First`、`Take`、`Last` 三个方法的区别

---

## 第 06 节：分层架构设计

### 本节课目标

1. 理解 handler / service / database / model 的分层思想
2. 能画出一条请求的完整流程
3. 能将逻辑从 handler 下放到 service

### 分层职责

```text
客户端
  │
  ▼
router      路由分发
  │
  ▼
middleware  认证、日志、限流、审计
  │
  ▼
handler     解析参数、调用 service、返回响应
  │
  ▼
service     业务逻辑、事务、缓存、外部调用
  │
  ▼
database    数据库连接、迁移
  │
  ▼
model       结构体定义
```

### 为什么分层

- **可测试**：service 层可以脱离 HTTP 测试
- **可维护**：修改数据库不影响接口契约
- **可复用**：业务逻辑可被多个 handler 复用

### 反例

不要在 handler 中写 SQL：

```go
// ❌ 错误
database.DB.Create(&post)
```

应该调用 service：

```go
// ✅ 正确
post, err := postService.Create(userID, req)
```

### 动手：画出发表评论的流程

根据 `handler/comment.go` 和 `service/comment.go`，画出：

1. 请求进入哪个路由
2. 经过哪些中间件
3. handler 做了什么
4. service 做了什么
5. 最后返回什么

### 课后作业

- 解释 `handler` 为什么不直接操作 `database.DB`
- 找出一个 `service` 被多个 `handler` 调用的例子

---

## 第 07 节：用户认证

### 本节课目标

1. 理解 JWT 的生成与校验
2. 理解 bcrypt 密码哈希
3. 理解中间件如何把用户信息注入 Context

### JWT 流程

```text
登录
  │
  ▼
校验用户名密码
  │
  ▼
generate token（含 userID, username, role, exp, jti）
  │
  ▼
返回 token
```

### 生成 Token

```go
func GenerateToken(userID uint, username string, role model.UserRole) (string, error) {
    claims := &Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
            ID:        uuid.NewString(), // JTI，用于黑名单
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.C.JWT.Secret))
}
```

### 校验 Token

```go
func ParseToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.C.JWT.Secret), nil
    })
    // ...
}
```

### 中间件注入 Context

```go
// middleware/jwt.go
c.Set("userID", claims.UserID)
c.Set("username", claims.Username)
c.Set("role", claims.Role)
c.Set("jti", claims.ID)
c.Next()
```

### 密码哈希

```go
// 注册
hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// 登录
err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
```

### 动手：生成并解析一个 JWT

在临时 Go 文件中：

```go
config.C = config.DefaultConfig()
token, _ := utils.GenerateToken(1, "alice", model.UserRoleUser)
claims, _ := utils.ParseToken(token)
fmt.Println(claims.UserID, claims.Username)
```

### 课后作业

- 解释 JTI 的作用，为什么登出时要把它加入黑名单
- 比较 Cookie Session 和 JWT 的优缺点

---

## 第 08 节：文章 CRUD

### 本节课目标

1. 理解 RESTful API 设计
2. 掌握参数校验与 Swagger 注释
3. 实现一个带测试的文章更新接口

### RESTful 设计

| 方法 | 路径 | 含义 |
|---|---|---|
| GET | `/api/posts` | 列表 |
| GET | `/api/posts/:id` | 详情 |
| POST | `/api/posts` | 创建 |
| PUT | `/api/posts/:id` | 更新 |
| DELETE | `/api/posts/:id` | 删除 |

### 参数校验

```go
type CreatePostRequest struct {
    Title      string   `json:"title" binding:"required"`
    Content    string   `json:"content" binding:"required"`
    CategoryID *uint    `json:"category_id"`
    TagIDs     []uint   `json:"tag_ids"`
    Status     string   `json:"status"`
}
```

### Swagger 注释

```go
// Create 创建文章
// @Summary 创建文章
// @Tags 文章
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body model.CreatePostRequest true "文章信息"
// @Success 201 {object} utils.Response{data=model.Post}
// @Router /posts [post]
func (h *PostHandler) Create(c *gin.Context) { ... }
```

### 动手：为文章更新写测试

```go
func TestPostService_Update(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    user := model.User{Username: "author", Password: "hash"}
    database.DB.Create(&user)

    post := model.Post{Title: "old", Content: "content", AuthorID: user.ID, Status: model.PostStatusPublished}
    database.DB.Create(&post)

    svc := NewPostService()
    updated, err := svc.Update(post.ID, user.ID, false, model.UpdatePostRequest{
        Title:   ptr("new title"),
        Content: ptr("new content"),
    })
    if err != nil {
        t.Fatal(err)
    }
    if updated.Title != "new title" {
        t.Errorf("标题未更新")
    }
}

func ptr(s string) *string { return &s }
```

### 课后作业

- 解释 `binding:"required"` 在什么情况下会报错
- 为 `service/post.go` 的 `Delete` 方法写一个单元测试

---

## 第 09 节：评论系统

### 本节课目标

1. 理解嵌套评论的数据库设计
2. 理解级联删除
3. 实现评论回复并触发通知

### 评论表设计

```go
type Comment struct {
    ID         uint   `json:"id" gorm:"primaryKey"`
    PostID     uint   `json:"post_id"`
    ParentID   *uint  `json:"parent_id,omitempty"` // nil 表示一级评论
    AuthorID   uint   `json:"author_id"`
    AuthorName string `json:"author_name"`
    Content    string `json:"content"`
}
```

### 创建评论

```go
comment := model.Comment{
    PostID:   req.PostID,
    ParentID: req.ParentID,
    AuthorID: authorID,
    Content:  content,
}
database.DB.Create(&comment)

// 如果是回复他人，发送通知
if req.ParentID != nil && parent.AuthorID != authorID {
    go CreateCommentReplyNotification(parent.AuthorID, comment.ID, authorName, post.Title)
}
```

### 级联删除

删除一条评论时，软删除其所有后代回复：

```go
func (s *CommentService) deleteWithChildren(tx *gorm.DB, id uint) error {
    ids := s.collectDescendantIDs(tx, id)
    ids = append(ids, id)
    return tx.Where("id IN ?", ids).Delete(&model.Comment{}).Error
}
```

### 动手：画出评论树

给定数据：

| id | parent_id | content |
|---|---|---|
| 1 | nil | 文章写得很好 |
| 2 | 1 | 谢谢！ |
| 3 | 2 | 不客气 |
| 4 | nil | 收藏了 |

画出树状结构。

### 课后作业

- 解释为什么 `ParentID` 用 `*uint` 而不是 `uint`
- 写出批量删除评论时如何保证事务

---

## 第 10 节：点赞与关注

### 本节课目标

1. 理解复合唯一索引防止重复点赞
2. 理解切换式点赞/取消点赞
3. 理解关注关系的双向查询

### 点赞表

```go
type Like struct {
    ID     uint `gorm:"primaryKey"`
    PostID uint `gorm:"not null;uniqueIndex:idx_post_user"`
    UserID uint `gorm:"not null;uniqueIndex:idx_post_user"`
}
```

`uniqueIndex:idx_post_user` 保证一个用户只能点赞一篇文章一次。

### 切换式点赞

```go
func (s *LikeService) Toggle(postID, userID uint) (bool, error) {
    // INSERT ... ON DUPLICATE KEY UPDATE
    // 如果已存在则取消，不存在则点赞
}
```

### 关注关系

```go
type UserFollow struct {
    ID          uint `gorm:"primaryKey"`
    FollowerID  uint // 粉丝
    FollowingID uint // 被关注者
}
```

查询粉丝：

```go
database.DB.Where("following_id = ?", userID).Find(&followers)
```

查询关注：

```go
database.DB.Where("follower_id = ?", userID).Find(&following)
```

### 动手：实现一个“是否互相关注”接口

在 `service/user_follow.go` 中新增：

```go
func IsMutualFollow(aID, bID uint) bool {
    return IsFollowing(aID, bID) && IsFollowing(bID, aID)
}
```

### 课后作业

- 解释复合唯一索引和普通唯一索引的区别
- 如果取消点赞，是否需要删除通知？为什么本项目选择不删除？

---

## 第 11 节：Redis 缓存

### 本节课目标

1. 理解缓存的使用场景
2. 掌握缓存失效与更新策略
3. 理解浏览量削峰设计

### 缓存场景

| 数据 | 缓存 key | 说明 |
|---|---|---|
| 分类列表 | `blog:categories` |  rarely change |
| 标签列表 | `blog:tags` | rarely change |
| 热门文章 | `blog:hot_posts` | 定时刷新 |
| 文章详情 | `blog:post:<id>` | 更新时失效 |

### 缓存失效

```go
func ClearPostCache(postID uint) {
    key := fmt.Sprintf("blog:post:%d", postID)
    database.Redis.Del(ctx, key)
    ClearHotPostsCache()
}
```

### 浏览量削峰

```text
用户访问
  │
  ▼
Redis INCR blog:view_count:123
  │
  ▼
定时任务每 5 分钟
  │
  ▼
批量写入 MySQL
```

### 动手：查看 Redis 缓存

```bash
docker exec -it blog-redis redis-cli keys 'blog:*'
docker exec -it blog-redis redis-cli get blog:hot_posts
```

### 课后作业

- 解释“缓存穿透”和“缓存雪崩”
- 如果 Redis 挂了，项目还能正常运行吗？为什么？

---

## 第 12 节：通知系统

### 本节课目标

1. 理解站内通知的创建与存储
2. 理解 WebSocket 实时推送原理
3. 理解邮件通知的异步发送

### 通知模型

```go
type Notification struct {
    ID        uint             `json:"id"`
    UserID    uint             `json:"user_id"`
    Type      NotificationType `json:"type"`
    Title     string           `json:"title"`
    Content   string           `json:"content"`
    RelatedID uint             `json:"related_id"`
    IsRead    bool             `json:"is_read"`
}
```

### 通知类型

- `comment_reply`：评论回复
- `post_like`：文章被赞
- `comment_like`：评论被赞
- `follow`：被关注
- `message`：收到私信
- `badge_award`：获得勋章

### 创建通知

```go
func CreateNotification(userID uint, nType model.NotificationType, title, content string, relatedID uint) error {
    notification := &model.Notification{...}
    database.DB.Create(notification)
    NotifyUserRealtime(userID, notification)   // WebSocket
    SendNotificationEmail(userID, notification) // 邮件
    return nil
}
```

### WebSocket 连接

```javascript
const ws = new WebSocket('ws://localhost:8080/ws/notifications?token=<jwt>');
ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);
    console.log(msg.data);
};
```

### 邮件通知

```bash
# .env
EMAIL_NOTIFICATION_EMAIL_ENABLED=true
EMAIL_HOST=smtp.example.com
EMAIL_USERNAME=xxx
EMAIL_PASSWORD=xxx
```

### 动手：用 wscat 测试 WebSocket

```bash
npm install -g wscat
wscat -c "ws://localhost:8080/ws/notifications?token=<jwt>"
```

然后用另一个用户给你的账号发一条私信，观察 wscat 是否收到消息。

### 课后作业

- 解释为什么 WebSocket token 要放在 query 参数
- 邮件通知为什么要异步发送？

---

## 第 13 节：搜索与 Meilisearch

### 本节课目标

1. 理解为什么需要专门搜索引擎
2. 理解索引同步与降级策略
3. 能配置并测试搜索

### 为什么不用 MySQL LIKE

- 慢：全表扫描
- 不支持中文分词
- 不支持容错匹配

### Meilisearch 集成

文章变更时同步索引：

```go
func AddPostIndex(post *model.Post) error { ... }
func UpdatePostIndex(post *model.Post) error { ... }
func DeletePostIndex(postID uint) error { ... }
```

### 搜索降级

```go
posts, err := searchPosts(req)
if err != nil {
    // Meilisearch 失败，降级为 MySQL LIKE
    posts, err = searchPostsMySQL(req)
}
```

### 动手：搜索一篇文章

```bash
curl "http://localhost:8080/api/posts?keyword=Gin"
```

### 课后作业

- 解释“索引”在搜索引擎中的作用
- 如果删除文章时索引删除失败，会出现什么问题？如何解决？

---

## 第 14 节：可观测性

### 本节课目标

1. 理解 Prometheus 指标
2. 理解 OpenTelemetry 链路追踪
3. 理解审计日志的作用

### Prometheus

访问 `/metrics`：

```text
http_requests_total{method="GET",path="/api/posts",status="200"}
http_request_duration_seconds_bucket{method="GET",path="/api/posts"}
```

### OpenTelemetry + Jaeger

```bash
BLOG_TRACING_ENABLED=true \
BLOG_TRACING_ENDPOINT=http://localhost:4318/v1/traces \
go run main.go
```

打开 http://localhost:16686 查看 trace。

### 审计日志

管理员的写操作自动记录：

```bash
curl http://localhost:8080/api/audit-logs \
  -H "Authorization: Bearer <admin-token>"
```

### 动手：查看一条 trace

1. 启用 tracing
2. 调用一个接口
3. 在 Jaeger UI 中搜索服务名 `blog`
4. 查看调用耗时分布

### 课后作业

- 解释 Prometheus Counter 和 Histogram 的区别
- 审计日志为什么要用 `c.Next()` 后再记录？

---

## 第 15 节：单元测试

### 本节课目标

1. 理解 SQLite 内存库隔离测试
2. 掌握 service 层测试写法
3. 了解 mock 外部依赖

### 测试辅助函数

```go
func setupTestDB(t *testing.T) (cleanup func()) {
    db, _ := gorm.Open(sqlite.Open("file:memdb?mode=memory&cache=shared"), &gorm.Config{})
    db.AutoMigrate(&model.User{}, &model.Post{}, ...)
    database.DB = db
    config.C = config.DefaultConfig()
    return func() { database.DB = originalDB }
}
```

### 测试示例

```go
func TestLikeService_Toggle(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    user := model.User{Username: "u", Password: "hash"}
    database.DB.Create(&user)

    post := model.Post{Title: "p", Content: "c", AuthorID: user.ID}
    database.DB.Create(&post)

    svc := NewLikeService()
    liked, err := svc.Toggle(post.ID, user.ID)
    if err != nil || !liked {
        t.Fatal("点赞失败")
    }
}
```

### Mock 邮件

```go
sendEmailFunc = func(_, _, _ string) error { return nil }
defer func() { sendEmailFunc = defaultSendEmailFunc }()
```

### 动手：为 `CategoryService.Create` 写测试

```go
func TestCategoryService_Create(t *testing.T) {
    cleanup := setupTestDB(t)
    defer cleanup()

    svc := NewCategoryService()
    cat, err := svc.Create(model.CreateCategoryRequest{Name: "Go"})
    if err != nil {
        t.Fatal(err)
    }
    if cat.Name != "Go" {
        t.Error("名称不匹配")
    }
}
```

### 课后作业

- 解释 `defer cleanup()` 的作用
- 为什么要用 `count:1` 运行测试？

---

## 第 16 节：部署与扩展方向

### 本节课目标

1. 理解 Docker 全栈部署
2. 掌握生产环境 checklist
3. 规划一个自己的扩展功能

### Docker 全栈部署

```bash
cd blog
./start.sh full
```

这会构建 Go 后端镜像并启动所有依赖。

### 生产 checklist

- [ ] 修改 `JWT_SECRET`
- [ ] 设置 `gin.ReleaseMode`
- [ ] 关闭 GORM 详细日志
- [ ] 使用 Nginx + HTTPS
- [ ] 修改数据库/Redis 默认密码
- [ ] 配置日志轮转
- [ ] 部署 Prometheus/Grafana
- [ ] 调整限流阈值

### 后续扩展方向

| 方向 | 难度 | 说明 |
|---|---|---|
| 用户通知偏好设置 | ⭐⭐ | 按类型/渠道开关通知 |
| Redis Pub/Sub WebSocket | ⭐⭐⭐ | 多实例横向扩展 |
| Sitemap / RSS | ⭐⭐ | SEO 与订阅 |
| 文章导入导出 | ⭐⭐⭐ | Markdown / YAML |
| 内容审核 | ⭐⭐⭐⭐ | 敏感词 / AI |
| 文章版本历史 | ⭐⭐⭐⭐ | 每次更新保存快照 |

### 结课项目

选择以上任意一个方向，独立完成：

1. 设计模型与接口
2. 实现 service 与 handler
3. 写单元测试
4. 更新 Swagger 文档
5. 在 README 中补充说明

---

## 附录 A：数据库表速查

| 表名 | 核心字段 | 说明 |
|---|---|---|
| users | username, password, role, email_verified | 用户 |
| posts | title, content, author_id, category_id, status | 文章 |
| comments | post_id, parent_id, author_id, content | 评论 |
| likes | post_id, user_id | 文章点赞 |
| comment_likes | comment_id, user_id | 评论点赞 |
| categories | name | 分类 |
| tags | name | 标签 |
| post_tags | post_id, tag_id | 多对多关联 |
| notifications | user_id, type, title, content, is_read | 通知 |
| messages | sender_id, receiver_id, content, is_read | 私信 |
| user_follows | follower_id, following_id | 关注 |
| favorites | user_id, post_id | 收藏 |
| read_histories | user_id, post_id | 阅读历史 |
| badges | name, icon_url | 勋章 |
| user_badges | user_id, badge_id | 用户勋章 |
| audit_logs | user_id, action, resource, details, ip | 审计 |

## 附录 B：环境变量速查

| 变量 | 默认值 | 说明 |
|---|---|---|
| BLOG_DB_HOST | 127.0.0.1 | MySQL 地址 |
| BLOG_DB_PORT | 3306 | MySQL 端口 |
| BLOG_DB_USER | root | 用户名 |
| BLOG_DB_PASSWORD | 123456 | 密码 |
| BLOG_DB_NAME | blog | 数据库名 |
| BLOG_REDIS_HOST | 127.0.0.1 | Redis 地址 |
| BLOG_JWT_SECRET | your-secret-key-change-in-production | JWT 密钥 |
| BLOG_EMAIL_HOST | - | SMTP 服务器 |
| BLOG_EMAIL_NOTIFICATION_EMAIL_ENABLED | false | 是否发送通知邮件 |
| BLOG_MEILISEARCH_ENABLED | false | 是否启用搜索 |
| BLOG_TRACING_ENABLED | false | 是否启用链路追踪 |

## 附录 C：常见错误排查

### 1. `go run main.go` 提示数据库连接失败

检查 MySQL 容器是否运行：

```bash
docker ps
```

### 2. 测试报 `database table is locked`

这是因为 SQLite 内存库被并发访问。service 层测试已用独立数据库隔离，异步任务已关闭邮件通知。

### 3. 接口返回 401

检查 Header 格式：

```
Authorization: Bearer <token>
```

注意 `Bearer` 后有一个空格。

### 4. Swagger 不显示新接口

重新生成：

```bash
go run github.com/swaggo/swag/cmd/swag@latest init
```

---

## 附录 D：讲师备课清单

| 课节 | 备课重点 | 课堂演示命令 | 学生检查点 |
|---|---|---|---|
| 01 | 环境变量、Docker 基础 | `docker compose up -d`、`curl /health` | 能访问 `/health` |
| 02 | Go 基础语法回顾 | 运行迷你 HTTP 服务 | 能独立编译运行 |
| 03 | Gin 路由与中间件 | 添加 `/ping` | 能用 Postman 访问 |
| 04 | Viper 配置优先级 | 修改 `.env` 看效果 | 能新增配置项 |
| 05 | GORM 模型与迁移 | 新增 `Tag.Color` 字段 | 数据库自动加列 |
| 06 | 分层架构思想 | 画请求流程图 | 能解释 handler/service 分工 |
| 07 | JWT 与 bcrypt | 生成并解析 token | 理解黑名单机制 |
| 08 | RESTful + Swagger | 实现更新接口测试 | 能用 Swagger 调接口 |
| 09 | 嵌套评论设计 | 画评论树 | 理解级联删除 |
| 10 | 复合唯一索引 | 演示点赞切换 | 理解并发控制 |
| 11 | Redis 缓存策略 | `redis-cli` 查看 key | 理解缓存失效 |
| 12 | WebSocket 心跳 | `wscat` 接收通知 | 能建立 WS 连接 |
| 13 | 搜索与降级 | Meilisearch 搜索 | 理解索引同步 |
| 14 | 可观测性三支柱 | 查看 metrics/trace/审计 | 能在 Jaeger 找 trace |
| 15 | 测试与 mock | 写一个 service 测试 | 能通过 `go test` |
| 16 | 部署 checklist | Docker 全栈启动 | 能独立部署 |

---

## 附录 E：课堂互动问题与参考答案

### 问题 1：为什么 handler 不能直接操作 database.DB？

**参考答案**：
- 职责分离：handler 负责 HTTP，service 负责业务。
- 可测试：service 不依赖 HTTP，可用 SQLite 单独测试。
- 可复用：同一业务逻辑可被多个 handler 调用。

### 问题 2：JWT 和 Session 有什么区别？

**参考答案**：
- JWT 是无状态的，服务端不需要保存会话；Session 需要在服务端存储。
- JWT 适合分布式系统；Session 适合单体小应用。
- JWT 一旦签发难以撤销，所以本项目用 JTI 黑名单解决登出撤销问题。

### 问题 3：Redis 挂了项目还能用吗？

**参考答案**：
- 能。项目对 Redis 做了降级处理，缓存失败会回源到 MySQL。
- 但限流、验证码、浏览量同步等功能会受影响。

### 问题 4：WebSocket 为什么用 query 传 token？

**参考答案**：
- 浏览器 JavaScript 的 `WebSocket` 构造函数不支持自定义 HTTP Header。
- 因此只能把 token 放在 URL query 参数中。

### 问题 5：软删除有什么好处？

**参考答案**：
- 数据可恢复，避免误删造成数据丢失。
- 关联数据不会级联物理删除，保持数据完整性。
- 配合 GORM 的 `DeletedAt` 可自动过滤已删除记录。

---

## 附录 F：学生实验报告模板

```markdown
# 实验报告：第 X 节 XXXXX

## 实验目标

## 实验环境

- Go 版本：
- Docker 版本：
- 操作系统：

## 实验步骤

1. ...
2. ...

## 关键代码

```go
// 贴出关键代码
```

## 实验结果

截图或文字说明。

## 遇到的问题与解决方案

## 总结与收获
```

---

## 附录 G：中英文术语对照表

| 中文 | 英文 | 说明 |
|---|---|---|
| 路由 | Router | 请求分发 |
| 中间件 | Middleware | 请求处理链中的插件 |
| 处理器 | Handler | HTTP 请求处理函数 |
| 服务层 | Service Layer | 业务逻辑层 |
| 模型 | Model | 数据结构与请求体 |
| 对象关系映射 | ORM | 如 GORM |
| 软删除 | Soft Delete | 标记删除而非物理删除 |
| 限流 | Rate Limit | 控制请求频率 |
| 缓存 | Cache | 如 Redis |
| 链路追踪 | Tracing | 如 OpenTelemetry/Jaeger |
| 审计日志 | Audit Log | 操作记录 |
| 索引 | Index | 搜索引擎中的数据结构 |
| 负载均衡 | Load Balancing | 多实例分发请求 |
| 优雅关闭 | Graceful Shutdown | 服务退出前处理完请求 |

---

## 附录 H：扩展阅读与资源

### 官方文档

- Go 官方教程：https://go.dev/tour/
- Gin 框架：https://gin-gonic.com/docs/
- GORM 文档：https://gorm.io/docs/
- Swagger / Swaggo：https://github.com/swaggo/swag

### 项目内资源

- `README.md`：项目简介、部署、API 速查
- `AGENTS.md`：Agent 协作规范
- `docs/`：Swagger 自动生成的 API 文档
- `docker-compose.yml`：本地依赖定义

### 推荐书籍

- 《Go 程序设计语言》
- 《Go Web 编程》
- 《高性能 MySQL》
- 《Redis 设计与实现》

---

## 教学建议

### 课堂节奏

- 每节课：20 分钟讲解 + 20 分钟动手 + 10 分钟答疑
- 建议每 4 节课安排一次综合练习
- 鼓励学生自己改代码、跑测试、看日志

### 考核方式

1. 课堂动手任务完成度（40%）
2. 课后作业（30%）
3. 结课扩展项目（30%）

### 推荐学习顺序

1. 先跑通项目，建立信心
2. 再读 `handler` 和 `service`，理解接口
3. 然后读 `model` 和 `database`，理解数据
4. 最后读 `middleware` 和 `utils`，理解基础设施

祝教学顺利！

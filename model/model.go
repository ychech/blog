// package model 定义数据模型、请求/响应结构体以及分页相关工具。
// 所有数据库表结构都集中在此，便于统一维护与自动迁移。
package model

import (
	"time"

	"gorm.io/gorm"
)

// UserRole 用户角色常量
type UserRole string

const (
	UserRoleAdmin UserRole = "admin"
	UserRoleUser  UserRole = "user"
)

// User 用户模型，对应 users 表。
// Password 字段使用 json:"-"，避免在接口响应中泄露密码哈希。
type User struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Username       string         `json:"username" gorm:"size:50;not null;uniqueIndex"`
	Password       string         `json:"-" gorm:"size:255;not null"` // json:"-" 表示不序列化到 JSON
	Nickname       string         `json:"nickname" gorm:"size:100"`
	Email          string         `json:"email" gorm:"size:100"`
	EmailVerified  bool           `json:"email_verified" gorm:"default:false"`
	Avatar         string         `json:"avatar" gorm:"size:255"`
	Role           UserRole       `json:"role" gorm:"size:20;default:user"`
	IsActive       bool           `json:"is_active" gorm:"default:true;index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// Category 文章分类，对应 categories 表。
// 分类与文章是一对多关系：删除分类时，关联文章 category_id 会被置为 NULL。
type Category struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null;uniqueIndex"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Tag 文章标签，对应 tags 表。
// 标签与文章是多对多关系，中间表为 post_tags。
type Tag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"size:100;not null;uniqueIndex"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// PostStatus 文章状态常量
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusPublished PostStatus = "published"
	PostStatusScheduled PostStatus = "scheduled"
)

// Post 博客文章，对应 posts 表。
// 包含作者、分类、标签等关联；Summary 为空时会自动截取正文前 100 字。
// LikeCount 为临时字段，不存储在 posts 表中，由业务层查询时填充。
type Post struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	Title      string         `json:"title" gorm:"size:255;not null"`
	Summary    string         `json:"summary" gorm:"size:500"`
	Content    string         `json:"content" gorm:"type:text;not null"`
	CoverURL   string         `json:"cover_url" gorm:"size:500"`
	Status     PostStatus     `json:"status" gorm:"size:20;default:published;index"`
	PublishAt  *time.Time     `json:"publish_at,omitempty" gorm:"index"`
	AuthorID   uint           `json:"author_id" gorm:"not null"`
	Author     User           `json:"author,omitempty" gorm:"foreignKey:AuthorID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	CategoryID *uint          `json:"category_id,omitempty"`
	Category   Category       `json:"category,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Tags       []Tag          `json:"tags,omitempty" gorm:"many2many:post_tags;"`
	ViewCount  int64          `json:"view_count" gorm:"default:0"`
	LikeCount  int64          `json:"like_count" gorm:"-"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// Like 文章点赞，对应 likes 表。
// 使用复合唯一索引 (post_id, user_id) 防止重复点赞。
type Like struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	PostID    uint           `json:"post_id" gorm:"not null;uniqueIndex:idx_post_user"`
	UserID    uint           `json:"user_id" gorm:"not null;uniqueIndex:idx_post_user"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Badge 勋章 / NFT 奖励，对应 badges 表。
// 管理员可创建勋章，并颁发给指定用户。
type Badge struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Name            string         `json:"name" gorm:"size:100;not null"`
	Description     string         `json:"description" gorm:"size:500"`
	IconURL         string         `json:"icon_url" gorm:"size:500"`         // 勋章图标
	ContractAddress string         `json:"contract_address" gorm:"size:255"` // 可选：NFT 合约地址
	TokenID         string         `json:"token_id" gorm:"size:100"`         // 可选：NFT token ID
	MetadataURL     string         `json:"metadata_url" gorm:"size:500"`     // 可选：NFT metadata URL
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// UserBadge 用户获得的勋章，对应 user_badges 表。
// 使用复合唯一索引 (user_id, badge_id) 防止重复颁发。
type UserBadge struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_badge"`
	BadgeID   uint      `json:"badge_id" gorm:"not null;uniqueIndex:idx_user_badge"`
	Badge     Badge     `json:"badge" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Reason    string    `json:"reason" gorm:"size:255"` // 颁发原因
	CreatedAt time.Time `json:"created_at"`
}

// NotificationType 定义通知类型。
type NotificationType string

const (
	// NotificationTypeCommentReply 评论回复通知。
	NotificationTypeCommentReply NotificationType = "comment_reply"
)

// Notification 通知模型，对应 notifications 表。
// 用于存储用户收到的系统通知，例如评论回复、勋章颁发等。
type Notification struct {
	ID        uint             `json:"id" gorm:"primaryKey"`
	UserID    uint             `json:"user_id" gorm:"not null;index"`      // 接收通知的用户
	Type      NotificationType `json:"type" gorm:"size:50;not null"`       // 通知类型
	Title     string           `json:"title" gorm:"size:200;not null"`     // 通知标题
	Content   string           `json:"content" gorm:"type:text"`           // 通知内容
	RelatedID uint             `json:"related_id" gorm:"index"`            // 关联业务 ID（如评论 ID）
	IsRead    bool             `json:"is_read" gorm:"default:false;index"` // 是否已读
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// AuditLog 操作审计日志模型，对应 audit_logs 表。
// 用于记录管理员的关键操作，便于合规审计和问题追溯。
type AuditLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"user_id" gorm:"not null;index"`  // 操作用户 ID
	Username   string    `json:"username" gorm:"size:50"`        // 操作用户名
	Action     string    `json:"action" gorm:"size:50;not null"` // 操作类型：CREATE/UPDATE/DELETE/LOGIN
	Resource   string    `json:"resource" gorm:"size:50"`        // 操作对象：post/user/badge/...
	ResourceID uint      `json:"resource_id" gorm:"index"`       // 操作对象 ID
	Details    string    `json:"details" gorm:"type:text"`       // 操作详情（JSON 或描述）
	IP         string    `json:"ip" gorm:"size:50"`              // 操作者 IP
	CreatedAt  time.Time `json:"created_at"`
}

// CreateBadgeRequest 创建勋章请求
type CreateBadgeRequest struct {
	Name            string `json:"name" binding:"required"`
	Description     string `json:"description"`
	IconURL         string `json:"icon_url"`
	ContractAddress string `json:"contract_address"`
	TokenID         string `json:"token_id"`
	MetadataURL     string `json:"metadata_url"`
}

// UpdateBadgeRequest 更新勋章请求
type UpdateBadgeRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	IconURL         *string `json:"icon_url"`
	ContractAddress *string `json:"contract_address"`
	TokenID         *string `json:"token_id"`
	MetadataURL     *string `json:"metadata_url"`
}

// AwardBadgeRequest 颁发勋章请求
type AwardBadgeRequest struct {
	UserID uint   `json:"user_id" binding:"required"`
	BadgeID uint  `json:"badge_id" binding:"required"`
	Reason string `json:"reason"`
}

// CommentLike 评论点赞，对应 comment_likes 表。
// 使用复合唯一索引 (comment_id, user_id) 防止重复点赞。
type CommentLike struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CommentID uint           `json:"comment_id" gorm:"not null;uniqueIndex:idx_comment_user"`
	UserID    uint           `json:"user_id" gorm:"not null;uniqueIndex:idx_comment_user"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Comment 评论，对应 comments 表。
// ParentID 为 nil 表示一级评论；有值时表示回复某条评论。
// AuthorName 是冗余字段，用于减少列表查询时的 JOIN。
// LikeCount 为临时字段，不存储在 comments 表中，由业务层查询时填充。
type Comment struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	PostID     uint           `json:"post_id" gorm:"not null"`
	ParentID   *uint          `json:"parent_id,omitempty"` // 为空表示一级评论，否则为回复
	AuthorID   uint           `json:"author_id" gorm:"not null"`
	AuthorName string         `json:"author_name" gorm:"size:100;not null"` // 冗余字段，减少查询
	Content    string         `json:"content" gorm:"type:text;not null"`
	LikeCount  int64          `json:"like_count" gorm:"-"`
	IsPinned   bool           `json:"is_pinned" gorm:"default:false;index"`  // 是否置顶
	IsEssence  bool           `json:"is_essence" gorm:"default:false;index"` // 是否精华
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// RegisterRequest 用户注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
}

// LoginRequest 用户登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// ForgotPasswordRequest 忘记密码请求
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

// UpdateUserRoleRequest 更新用户角色请求（管理员）
type UpdateUserRoleRequest struct {
	Role UserRole `json:"role" binding:"required,oneof=admin user"`
}

// UpdateUserStatusRequest 更新用户启用状态请求（管理员）
type UpdateUserStatusRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

// VerifyEmailRequest 邮箱验证码验证请求
type VerifyEmailRequest struct {
	Code string `json:"code" binding:"required"`
}

// UpdateProfileRequest 更新用户资料请求
// 所有字段均为可选，使用指针区分"未传"和"传空"
type UpdateProfileRequest struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token    string `json:"token"`
	ExpireAt int64  `json:"expire_at"`
	User     User   `json:"user"`
}

// TokenResponse Token 刷新响应
type TokenResponse struct {
	Token string `json:"token"`
}

// TrendData 趋势数据点
type TrendData struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// AdminStats 后台仪表盘统计
type AdminStats struct {
	Counts CountsData    `json:"counts"`
	Trends TrendsData    `json:"trends"`
	HotPosts []Post      `json:"hot_posts"`
	TopTags  []Tag       `json:"top_tags"`
}

// CountsData 各类资源总数
type CountsData struct {
	Users    int64 `json:"users"`
	Posts    int64 `json:"posts"`
	Comments int64 `json:"comments"`
	Categories int64 `json:"categories"`
	Tags     int64 `json:"tags"`
	Badges   int64 `json:"badges"`
}

// TrendsData 近 7 天趋势
type TrendsData struct {
	Users    []TrendData `json:"users"`
	Posts    []TrendData `json:"posts"`
	Comments []TrendData `json:"comments"`
}

// CreatePostRequest 创建文章请求
type CreatePostRequest struct {
	Title      string     `json:"title" binding:"required"`
	Summary    string     `json:"summary"`
	Content    string     `json:"content" binding:"required"`
	CoverURL   string     `json:"cover_url"`
	Status     PostStatus `json:"status"`            // draft / published / scheduled，默认 published
	PublishAt  *time.Time `json:"publish_at"`        // 定时发布时间，status=scheduled 时必填
	CategoryID uint       `json:"category_id" binding:"required"`
	TagIDs     []uint     `json:"tag_ids"`
}

// UpdatePostRequest 更新文章请求
// 使用指针区分"未传"和"传空"
type UpdatePostRequest struct {
	Title      *string     `json:"title"`
	Summary    *string     `json:"summary"`
	Content    *string     `json:"content"`
	CoverURL   *string     `json:"cover_url"`
	Status     *PostStatus `json:"status"`
	PublishAt  *time.Time  `json:"publish_at"`
	CategoryID *uint       `json:"category_id"`
	TagIDs     []uint      `json:"tag_ids"`
}

// PostQuery 文章查询参数
type PostQuery struct {
	Page       int        `form:"page"`
	PageSize   int        `form:"page_size"`
	Keyword    string     `form:"keyword"`
	CategoryID uint       `form:"category_id"`
	TagID      uint       `form:"tag_id"`
	Status     PostStatus `form:"status"`   // draft / published / 空（默认只看已发布）
	OrderBy    string     `form:"order_by"` // created_at / view_count
	DateFrom   string     `form:"date_from"` // 开始日期，格式 2006-01-02
	DateTo     string     `form:"date_to"`   // 结束日期，格式 2006-01-02
}

// CreateCategoryRequest 创建分类请求
type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateTagRequest 创建标签请求
type CreateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateTagRequest 更新标签请求
type UpdateTagRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	PostID   uint   `json:"post_id" binding:"required"`
	ParentID *uint  `json:"parent_id"`
	Content  string `json:"content" binding:"required"`
}

// UpdateCommentRequest 更新评论请求
type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

// CommentReportStatus 举报状态
type CommentReportStatus string

const (
	CommentReportStatusPending  CommentReportStatus = "pending"
	CommentReportStatusApproved CommentReportStatus = "approved"
	CommentReportStatusRejected CommentReportStatus = "rejected"
)

// CommentReport 评论举报
type CommentReport struct {
	ID        uint                `json:"id" gorm:"primaryKey"`
	CommentID uint                `json:"comment_id" gorm:"not null;index"`
	UserID    uint                `json:"user_id" gorm:"not null;index"`
	Reason    string              `json:"reason" gorm:"type:text;not null"`
	Status    CommentReportStatus `json:"status" gorm:"size:20;default:pending;index"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

// CreateCommentReportRequest 创建评论举报请求
type CreateCommentReportRequest struct {
	Reason string `json:"reason" binding:"required"`
}

// UpdateCommentReportStatusRequest 更新举报状态请求
type UpdateCommentReportStatusRequest struct {
	Status CommentReportStatus `json:"status" binding:"required,oneof=pending approved rejected"`
}

// BatchDeleteRequest 批量删除请求
type BatchDeleteRequest struct {
	IDs []uint `json:"ids" binding:"required,min=1,max=100"`
}

// Message 站内私信模型
type Message struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SenderID   uint      `json:"sender_id" gorm:"not null;index"`
	ReceiverID uint      `json:"receiver_id" gorm:"not null;index"`
	Content    string    `json:"content" gorm:"type:text;not null"`
	IsRead     bool      `json:"is_read" gorm:"default:false;index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SendMessageRequest 发送私信请求
type SendMessageRequest struct {
	ReceiverID uint   `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

// Favorite 文章收藏
type Favorite struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null;uniqueIndex:idx_user_post_favorite"`
	PostID    uint      `json:"post_id" gorm:"not null;uniqueIndex:idx_user_post_favorite"`
	CreatedAt time.Time `json:"created_at"`
}

// ReadHistory 文章阅读历史
type ReadHistory struct {
	ID     uint      `json:"id" gorm:"primaryKey"`
	UserID uint      `json:"user_id" gorm:"not null;index"`
	PostID uint      `json:"post_id" gorm:"not null;index"`
	ReadAt time.Time `json:"read_at"`
}

// OAuthAccount 第三方 OAuth 账号绑定
type OAuthAccount struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"not null;index"`
	Provider       string    `json:"provider" gorm:"size:50;not null"`
	ProviderUserID string    `json:"provider_user_id" gorm:"size:100;not null"`
	AccessToken    string    `json:"-" gorm:"size:255"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// UserFollow 用户关注关系
type UserFollow struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FollowerID  uint      `json:"follower_id" gorm:"not null;uniqueIndex:idx_user_follow"`
	FollowingID uint      `json:"following_id" gorm:"not null;uniqueIndex:idx_user_follow"`
	CreatedAt   time.Time `json:"created_at"`
}

// Conversation 会话摘要
type Conversation struct {
	UserID        uint      `json:"user_id"`
	Username      string    `json:"username"`
	Nickname      string    `json:"nickname"`
	Avatar        string    `json:"avatar"`
	LastContent   string    `json:"last_content"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int64     `json:"unread_count"`
}

// AuditLogQuery 审计日志查询参数
type AuditLogQuery struct {
	Page      int       `form:"page"`
	PageSize  int       `form:"page_size"`
	Action    string    `form:"action"`
	Resource  string    `form:"resource"`
	UserID    uint      `form:"user_id"`
	StartTime time.Time `form:"-"`
	EndTime   time.Time `form:"-"`
}

// Pagination 分页参数
type Pagination struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// Normalize 设置默认值
func (p *Pagination) Normalize() {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 10
	}
	if p.PageSize > 100 {
		p.PageSize = 100
	}
}

// Offset 计算 SQL 偏移量
func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// ListResponse 通用列表响应
type ListResponse struct {
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
	Data  interface{} `json:"data"`
}

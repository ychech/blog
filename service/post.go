package service

import (
	"blog/database"
	"blog/model"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// PostService 文章服务，处理文章的增删改查、搜索筛选、排序分页与热门文章。
type PostService struct{}

// NewPostService 创建文章服务
func NewPostService() *PostService {
	return &PostService{}
}

// Create 创建文章
func (s *PostService) Create(authorID uint, req model.CreatePostRequest) (*model.Post, error) {
	// 基础校验
	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	summary := strings.TrimSpace(req.Summary)
	if err := validateNonEmptyTrimmed(title, "文章标题"); err != nil {
		return nil, err
	}
	if err := validateMaxLength(title, "文章标题", 255); err != nil {
		return nil, err
	}
	if err := validateNonEmptyTrimmed(content, "文章内容"); err != nil {
		return nil, err
	}
	if err := validateMaxLength(summary, "文章摘要", 500); err != nil {
		return nil, err
	}

	// 检查分类
	var category model.Category
	if err := database.DB.First(&category, req.CategoryID).Error; err != nil {
		return nil, fmt.Errorf("分类不存在")
	}

	status := req.Status
	if status != model.PostStatusDraft && status != model.PostStatusPublished {
		status = model.PostStatusPublished
	}

	post := model.Post{
		Title:      title,
		Summary:    summary,
		Content:    content,
		CoverURL:   req.CoverURL,
		Status:     status,
		AuthorID:   authorID,
		CategoryID: &req.CategoryID,
	}

	if post.Summary == "" && len([]rune(post.Content)) > 100 {
		post.Summary = string([]rune(post.Content)[:100])
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&post).Error; err != nil {
			return err
		}
		if len(req.TagIDs) > 0 {
			var tags []model.Tag
			if err := tx.Find(&tags, req.TagIDs).Error; err != nil {
				return err
			}
			if err := tx.Model(&post).Association("Tags").Append(&tags); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 同步 Meilisearch 索引
	if post.Status == model.PostStatusPublished {
		go IndexPost(&post)
	}

	ClearHotPostsCache()
	// 作者当然可以看自己的文章
	return s.GetByID(post.ID, post.AuthorID, false)
}

// List 获取文章列表（支持搜索、筛选、排序、分页）。
// 权限规则：未登录/普通用户默认只看已发布；登录用户额外看自己草稿；管理员看全部。
func (s *PostService) List(query model.PostQuery, currentUserID uint, isAdmin bool) (*model.ListResponse, error) {
	query.Page, query.PageSize = normalizePagination(query.Page, query.PageSize)

	db := database.DB.Model(&model.Post{})

	// 状态过滤：未指定时默认只显示已发布，但作者/管理员可查看更多
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	} else if !isAdmin {
		// 非管理员默认只显示已发布；登录用户额外显示自己的草稿
		if currentUserID > 0 {
			db = db.Where("status = ? OR (status = ? AND author_id = ?)",
				model.PostStatusPublished, model.PostStatusDraft, currentUserID)
		} else {
			db = db.Where("status = ?", model.PostStatusPublished)
		}
	}

	// 关键词搜索：优先使用 Meilisearch，失败时降级为 MySQL LIKE。
	// Meilisearch 只负责召回候选 ID，真正的状态/分类/标签过滤仍由数据库完成，
	// 避免搜索路径绕过原有筛选条件。
	if keyword := strings.TrimSpace(query.Keyword); keyword != "" {
		if ids, _, err := SearchPostIDs(keyword, 1000); err == nil {
			if len(ids) == 0 {
				return &model.ListResponse{
					Total: 0,
					Page:  query.Page,
					Size:  query.PageSize,
					Data:  []model.Post{},
				}, nil
			}
			db = db.Where("id IN ?", ids)
		} else {
			// Meilisearch 不可用时降级为 MySQL LIKE
			k := "%" + keyword + "%"
			db = db.Where("title LIKE ? OR content LIKE ?", k, k)
		}
	}

	// 按分类筛选
	if query.CategoryID > 0 {
		db = db.Where("category_id = ?", query.CategoryID)
	}

	// 按标签筛选：通过 post_tags 中间表
	if query.TagID > 0 {
		db = db.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
			Where("post_tags.tag_id = ?", query.TagID)
	}

	// 排序
	switch strings.ToLower(query.OrderBy) {
	case "view_count":
		db = db.Order("view_count DESC")
	default:
		db = db.Order("created_at DESC")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	var posts []model.Post
	if err := db.
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		Limit(query.PageSize).
		Offset((query.Page - 1) * query.PageSize).
		Find(&posts).Error; err != nil {
		return nil, err
	}

	// 批量填充点赞数（单次查询，避免 N+1）
	postIDs := make([]uint, len(posts))
	for i := range posts {
		postIDs[i] = posts[i].ID
	}
	likeCounts := NewLikeService().BatchGetLikeCounts(postIDs)
	for i := range posts {
		posts[i].LikeCount = likeCounts[posts[i].ID]
	}

	return &model.ListResponse{
		Total: total,
		Page:  query.Page,
		Size:  query.PageSize,
		Data:  posts,
	}, nil
}

// GetByID 获取文章详情。
// 权限规则：未发布文章仅作者或管理员可查看。
func (s *PostService) GetByID(id uint, currentUserID uint, isAdmin bool) (*model.Post, error) {
	var post model.Post
	if err := database.DB.
		Preload("Author").
		Preload("Category").
		Preload("Tags").
		First(&post, id).Error; err != nil {
		return nil, err
	}

	// 未发布文章仅作者或管理员可查看
	if post.Status != model.PostStatusPublished && post.AuthorID != currentUserID && !isAdmin {
		return nil, fmt.Errorf("文章不存在")
	}

	// 填充点赞数
	post.LikeCount = NewLikeService().BatchGetLikeCounts([]uint{id})[id]

	// 合并 Redis 未同步的浏览量增量
	viewCount, err := GetPostViewCount(id)
	if err == nil {
		post.ViewCount = viewCount
	}

	return &post, nil
}

// IncrementViewCount 增加文章浏览量。
// 实际逻辑委托给 viewcount.go，优先写入 Redis，Redis 不可用时直接写 MySQL。
func (s *PostService) IncrementViewCount(id uint) error {
	return IncrementPostViewCount(id)
}

// Update 更新文章。管理员可修改任意文章，普通用户只能修改自己的文章。
func (s *PostService) Update(id, currentUserID uint, isAdmin bool, req model.UpdatePostRequest) (*model.Post, error) {
	var post model.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		return nil, err
	}

	// 权限校验：普通用户只能修改自己的文章
	if !isAdmin && post.AuthorID != currentUserID {
		return nil, fmt.Errorf("无权修改该文章")
	}

	if req.Title != nil {
		post.Title = strings.TrimSpace(*req.Title)
		if err := validateNonEmptyTrimmed(post.Title, "文章标题"); err != nil {
			return nil, err
		}
		if err := validateMaxLength(post.Title, "文章标题", 255); err != nil {
			return nil, err
		}
	}
	if req.Summary != nil {
		post.Summary = strings.TrimSpace(*req.Summary)
		if err := validateMaxLength(post.Summary, "文章摘要", 500); err != nil {
			return nil, err
		}
	}
	if req.Content != nil {
		post.Content = strings.TrimSpace(*req.Content)
		if err := validateNonEmptyTrimmed(post.Content, "文章内容"); err != nil {
			return nil, err
		}
	}
	if req.CoverURL != nil {
		post.CoverURL = *req.CoverURL
	}
	if req.Status != nil {
		if *req.Status == model.PostStatusDraft || *req.Status == model.PostStatusPublished {
			post.Status = *req.Status
		}
	}
	if req.CategoryID != nil {
		var category model.Category
		if err := database.DB.First(&category, *req.CategoryID).Error; err != nil {
			return nil, fmt.Errorf("分类不存在")
		}
		post.CategoryID = req.CategoryID
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&post).Error; err != nil {
			return err
		}
		if req.TagIDs != nil {
			var tags []model.Tag
			if len(req.TagIDs) > 0 {
				if err := tx.Find(&tags, req.TagIDs).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&post).Association("Tags").Replace(&tags); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// 同步 Meilisearch 索引
	if post.Status == model.PostStatusPublished {
		go IndexPost(&post)
	} else {
		go DeletePostIndex(post.ID)
	}

	ClearHotPostsCache()
	ClearPostCache(id)
	// 更新者可以查看返回结果
	return s.GetByID(id, currentUserID, isAdmin)
}

// Delete 删除文章。管理员可删除任意文章，普通用户只能删除自己的文章。
func (s *PostService) Delete(id, currentUserID uint, isAdmin bool) error {
	var post model.Post
	if err := database.DB.First(&post, id).Error; err != nil {
		return err
	}
	if !isAdmin && post.AuthorID != currentUserID {
		return fmt.Errorf("无权删除该文章")
	}

	if err := database.DB.Delete(&post).Error; err != nil {
		return err
	}

	// 从 Meilisearch 删除索引
	go DeletePostIndex(id)

	ClearHotPostsCache()
	ClearPostCache(id)
	return nil
}

// GetHotPosts 获取热门文章（按浏览量排序）
func (s *PostService) GetHotPosts(limit int) ([]model.Post, error) {
	if data, ok := GetHotPostsCache(); ok {
		return data, nil
	}

	var posts []model.Post
	if err := database.DB.
		Where("status = ?", model.PostStatusPublished).
		Order("view_count DESC").
		Limit(limit).
		Preload("Category").
		Find(&posts).Error; err != nil {
		return nil, err
	}

	// 批量填充热门文章点赞数（单次查询，避免 N+1）
	postIDs := make([]uint, len(posts))
	for i := range posts {
		postIDs[i] = posts[i].ID
	}
	likeCounts := NewLikeService().BatchGetLikeCounts(postIDs)
	for i := range posts {
		posts[i].LikeCount = likeCounts[posts[i].ID]
	}

	SetHotPostsCache(posts)
	return posts, nil
}

func normalizePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

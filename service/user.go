// package service 实现业务逻辑，协调数据库访问、缓存、密码加密与 JWT 生成。
// handler 层不直接访问数据库，所有业务规则都在此层处理。
package service

import (
	"blog/config"
	"blog/database"
	"blog/model"
	"blog/utils"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserService 用户服务，处理注册、登录、用户信息查询等业务。
type UserService struct{}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{}
}

// Register 用户注册
func (s *UserService) Register(req model.RegisterRequest) (*model.LoginResponse, error) {
	// 密码加密
	hashedPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := model.User{
		Username: req.Username,
		Password: hashedPwd,
		Nickname: req.Nickname,
		Email:    req.Email,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		if isDuplicateKeyError(err) {
			return nil, errors.New("用户名已存在")
		}
		return nil, err
	}

	// 如果启用了邮箱验证且提供了邮箱，异步发送验证邮件
	if config.C.EmailVerification.Enabled && req.Email != "" {
		go func() {
			if err := SendVerificationEmail(user.ID, user.Email); err != nil {
				utils.Logger.Errorf("发送验证邮件失败: %v", err)
			}
		}()
	}

	return s.generateLoginResponse(&user)
}

// Login 用户登录
func (s *UserService) Login(req model.LoginRequest) (*model.LoginResponse, error) {
	var user model.User
	if err := database.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, errors.New("用户不存在或密码错误")
	}

	if !utils.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("用户不存在或密码错误")
	}

	// 如果配置要求必须验证邮箱，则未验证用户禁止登录
	if config.C.EmailVerification.Enabled && config.C.EmailVerification.Required && !user.EmailVerified {
		return nil, errors.New("邮箱未验证，请先验证邮箱后再登录")
	}

	return s.generateLoginResponse(&user)
}

// generateLoginResponse 生成登录响应
func (s *UserService) generateLoginResponse(user *model.User) (*model.LoginResponse, error) {
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, err
	}

	claims, err := utils.ParseToken(token)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		Token:    token,
		ExpireAt: claims.ExpiresAt.Unix(),
		User:     *user,
	}, nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// List 获取用户列表（管理员使用）
func (s *UserService) List(page, pageSize int) (*model.ListResponse, error) {
	page, pageSize = normalizePagination(page, pageSize)

	var total int64
	if err := database.DB.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, err
	}

	var users []model.User
	if err := database.DB.
		Order("created_at DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&users).Error; err != nil {
		return nil, err
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  users,
	}, nil
}

// GetStats 获取站点统计信息（管理员使用）
func (s *UserService) GetStats() (gin.H, error) {
	var stats struct {
		UserCount    int64
		PostCount    int64
		CommentCount int64
		CategoryCount int64
		TagCount     int64
		BadgeCount   int64
	}

	db := database.DB
	err := db.Model(&model.User{}).Count(&stats.UserCount).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&model.Post{}).Count(&stats.PostCount).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&model.Comment{}).Count(&stats.CommentCount).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&model.Category{}).Count(&stats.CategoryCount).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&model.Tag{}).Count(&stats.TagCount).Error
	if err != nil {
		return nil, err
	}
	err = db.Model(&model.Badge{}).Count(&stats.BadgeCount).Error
	if err != nil {
		return nil, err
	}

	return gin.H{
		"user_count":     stats.UserCount,
		"post_count":     stats.PostCount,
		"comment_count":  stats.CommentCount,
		"category_count": stats.CategoryCount,
		"tag_count":      stats.TagCount,
		"badge_count":    stats.BadgeCount,
	}, nil
}

// UpdateProfile 更新当前用户资料
func (s *UserService) UpdateProfile(id uint, req model.UpdateProfileRequest) (*model.User, error) {
	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}

	// 只更新传入的字段
	updates := map[string]interface{}{}
	if req.Nickname != "" {
		nickname := strings.TrimSpace(req.Nickname)
		if err := validateMaxLength(nickname, "昵称", 100); err != nil {
			return nil, err
		}
		updates["nickname"] = nickname
	}
	if req.Email != "" {
		if err := validateEmail(req.Email); err != nil {
			return nil, err
		}
		updates["email"] = req.Email
	}
	if req.Avatar != "" {
		if err := validateMaxLength(req.Avatar, "头像地址", 255); err != nil {
			return nil, err
		}
		updates["avatar"] = req.Avatar
	}

	if len(updates) == 0 {
		return &user, nil
	}

	if err := database.DB.Model(&user).Updates(updates).Error; err != nil {
		return nil, err
	}
	// 重新加载以返回最新值
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

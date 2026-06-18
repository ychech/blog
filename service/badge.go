// package service 实现业务逻辑，协调数据库访问、缓存、密码加密与 JWT 生成。
// handler 层不直接访问数据库，所有业务规则都在此层处理。
package service

import (
	"blog/database"
	"blog/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// BadgeService 勋章服务
type BadgeService struct{}

// NewBadgeService 创建勋章服务
func NewBadgeService() *BadgeService {
	return &BadgeService{}
}

// Create 创建勋章
func (s *BadgeService) Create(req model.CreateBadgeRequest) (*model.Badge, error) {
	badge := model.Badge{
		Name:            req.Name,
		Description:     req.Description,
		IconURL:         req.IconURL,
		ContractAddress: req.ContractAddress,
		TokenID:         req.TokenID,
		MetadataURL:     req.MetadataURL,
	}
	if err := database.DB.Create(&badge).Error; err != nil {
		return nil, err
	}
	return &badge, nil
}

// List 获取所有勋章
func (s *BadgeService) List() ([]model.Badge, error) {
	var badges []model.Badge
	if err := database.DB.Order("created_at DESC").Find(&badges).Error; err != nil {
		return nil, err
	}
	return badges, nil
}

// GetByID 根据 ID 获取勋章
func (s *BadgeService) GetByID(id uint) (*model.Badge, error) {
	var badge model.Badge
	if err := database.DB.First(&badge, id).Error; err != nil {
		return nil, err
	}
	return &badge, nil
}

// Update 更新勋章
func (s *BadgeService) Update(id uint, req model.UpdateBadgeRequest) (*model.Badge, error) {
	var badge model.Badge
	if err := database.DB.First(&badge, id).Error; err != nil {
		return nil, err
	}

	if req.Name != nil {
		badge.Name = *req.Name
	}
	if req.Description != nil {
		badge.Description = *req.Description
	}
	if req.IconURL != nil {
		badge.IconURL = *req.IconURL
	}
	if req.ContractAddress != nil {
		badge.ContractAddress = *req.ContractAddress
	}
	if req.TokenID != nil {
		badge.TokenID = *req.TokenID
	}
	if req.MetadataURL != nil {
		badge.MetadataURL = *req.MetadataURL
	}

	if err := database.DB.Save(&badge).Error; err != nil {
		return nil, err
	}
	return &badge, nil
}

// Delete 删除勋章
func (s *BadgeService) Delete(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 先删除用户勋章关联
		if err := tx.Where("badge_id = ?", id).Delete(&model.UserBadge{}).Error; err != nil {
			return err
		}
		// 再删除勋章
		if err := tx.Delete(&model.Badge{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// Award 颁发勋章给用户
func (s *BadgeService) Award(userID, badgeID uint, reason string) (*model.UserBadge, error) {
	var userBadge model.UserBadge
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 检查用户是否存在
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return errors.New("用户不存在")
		}

		// 检查勋章是否存在
		var badge model.Badge
		if err := tx.First(&badge, badgeID).Error; err != nil {
			return errors.New("勋章不存在")
		}

		userBadge = model.UserBadge{
			UserID:  userID,
			BadgeID: badgeID,
			Reason:  reason,
		}
		if err := tx.Create(&userBadge).Error; err != nil {
			if isDuplicateKeyError(err) {
				return fmt.Errorf("该用户已拥有 %s 勋章", badge.Name)
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// 预加载勋章信息返回
	if err := database.DB.Preload("Badge").First(&userBadge, userBadge.ID).Error; err != nil {
		return nil, err
	}

	notifyBadgeAwardNotification(userBadge.UserID, userBadge.ID, userBadge.Badge.Name)
	return &userBadge, nil
}

func notifyBadgeAwardNotification(userID, userBadgeID uint, badgeName string) {
	notifyAsync(func() error {
		return CreateBadgeAwardNotification(userID, userBadgeID, badgeName)
	})
}

// GetUserBadges 获取用户的勋章列表
func (s *BadgeService) GetUserBadges(userID uint) ([]model.UserBadge, error) {
	var userBadges []model.UserBadge
	if err := database.DB.
		Where("user_id = ?", userID).
		Preload("Badge").
		Order("created_at DESC").
		Find(&userBadges).Error; err != nil {
		return nil, err
	}
	return userBadges, nil
}

// Revoke 收回用户勋章（管理员）
func (s *BadgeService) Revoke(userBadgeID uint) error {
	var userBadge model.UserBadge
	if err := database.DB.First(&userBadge, userBadgeID).Error; err != nil {
		return err
	}
	return database.DB.Delete(&userBadge).Error
}

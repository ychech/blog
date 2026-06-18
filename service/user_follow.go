package service

import (
	"blog/database"
	"blog/model"
	"fmt"
)

// FollowUser 关注用户。
func FollowUser(followerID, followingID uint) error {
	if followerID == followingID {
		return fmt.Errorf("不能关注自己")
	}

	// 确认目标用户存在
	var target model.User
	if err := database.DB.First(&target, followingID).Error; err != nil {
		return fmt.Errorf("用户不存在")
	}

	follow := model.UserFollow{FollowerID: followerID, FollowingID: followingID}
	if err := database.DB.Create(&follow).Error; err != nil {
		if isDuplicateKeyError(err) {
			return fmt.Errorf("已经关注该用户")
		}
		return err
	}

	notifyFollowNotification(followerID, followingID)
	return nil
}

func notifyFollowNotification(followerID, followingID uint) {
	var follower model.User
	if err := database.DB.Select("id, nickname, username").First(&follower, followerID).Error; err != nil {
		return
	}
	nickname := follower.Nickname
	if nickname == "" {
		nickname = follower.Username
	}
	notifyAsync(func() error {
		return CreateFollowNotification(followingID, followerID, nickname)
	})
}

// UnfollowUser 取消关注。
func UnfollowUser(followerID, followingID uint) error {
	result := database.DB.Where("follower_id = ? AND following_id = ?", followerID, followingID).Delete(&model.UserFollow{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("未关注该用户")
	}
	return nil
}

// ListFollowers 查询用户的粉丝列表。
func ListFollowers(userID uint, page, pageSize int) (*model.ListResponse, error) {
	return listFollowUsers("following_id", userID, page, pageSize)
}

// ListFollowing 查询用户关注列表。
func ListFollowing(userID uint, page, pageSize int) (*model.ListResponse, error) {
	return listFollowUsers("follower_id", userID, page, pageSize)
}

func listFollowUsers(column string, userID uint, page, pageSize int) (*model.ListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	query := database.DB.Model(&model.UserFollow{}).Where(column+" = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var follows []model.UserFollow
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&follows).Error; err != nil {
		return nil, err
	}

	var users []model.User
	if len(follows) > 0 {
		ids := make([]uint, 0, len(follows))
		for _, f := range follows {
			if column == "follower_id" {
				ids = append(ids, f.FollowingID)
			} else {
				ids = append(ids, f.FollowerID)
			}
		}
		database.DB.Where("id IN ?", ids).Find(&users)
	}

	return &model.ListResponse{
		Total: total,
		Page:  page,
		Size:  pageSize,
		Data:  users,
	}, nil
}

// IsFollowing 查询是否已关注。
func IsFollowing(followerID, followingID uint) bool {
	var count int64
	database.DB.Model(&model.UserFollow{}).
		Where("follower_id = ? AND following_id = ?", followerID, followingID).
		Count(&count)
	return count > 0
}

// CountFollowers 查询粉丝数。
func CountFollowers(userID uint) int64 {
	var count int64
	database.DB.Model(&model.UserFollow{}).Where("following_id = ?", userID).Count(&count)
	return count
}

// CountFollowing 查询关注数。
func CountFollowing(userID uint) int64 {
	var count int64
	database.DB.Model(&model.UserFollow{}).Where("follower_id = ?", userID).Count(&count)
	return count
}

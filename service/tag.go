package service

import (
	"blog/database"
	"blog/model"
)

// TagService 标签服务，处理标签的增删查并维护标签缓存。
type TagService struct{}

// NewTagService 创建标签服务
func NewTagService() *TagService {
	return &TagService{}
}

// Create 创建标签
func (s *TagService) Create(name string) (*model.Tag, error) {
	tag := model.Tag{Name: name}
	if err := database.DB.Create(&tag).Error; err != nil {
		return nil, err
	}
	ClearTagCache()
	return &tag, nil
}

// List 获取标签列表
func (s *TagService) List() ([]model.Tag, error) {
	if data, ok := GetTagCache(); ok {
		return data, nil
	}

	var tags []model.Tag
	if err := database.DB.Find(&tags).Error; err != nil {
		return nil, err
	}

	SetTagCache(tags)
	return tags, nil
}

// Delete 删除标签
func (s *TagService) Delete(id uint) error {
	if err := database.DB.Delete(&model.Tag{}, id).Error; err != nil {
		return err
	}
	ClearTagCache()
	return nil
}

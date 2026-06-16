package service

import (
	"blog/database"
	"blog/model"
)

// CategoryService 分类服务，处理分类的增删改查并维护分类缓存。
type CategoryService struct{}

// NewCategoryService 创建分类服务
func NewCategoryService() *CategoryService {
	return &CategoryService{}
}

// Create 创建分类
func (s *CategoryService) Create(name string) (*model.Category, error) {
	category := model.Category{Name: name}
	if err := database.DB.Create(&category).Error; err != nil {
		return nil, err
	}
	ClearCategoryCache()
	return &category, nil
}

// List 获取分类列表
func (s *CategoryService) List() ([]model.Category, error) {
	// 先尝试从缓存读取
	if data, ok := GetCategoryCache(); ok {
		return data, nil
	}

	var categories []model.Category
	if err := database.DB.Find(&categories).Error; err != nil {
		return nil, err
	}

	SetCategoryCache(categories)
	return categories, nil
}

// Update 更新分类
func (s *CategoryService) Update(id uint, name string) (*model.Category, error) {
	var category model.Category
	if err := database.DB.First(&category, id).Error; err != nil {
		return nil, err
	}

	category.Name = name
	if err := database.DB.Save(&category).Error; err != nil {
		return nil, err
	}

	ClearCategoryCache()
	return &category, nil
}

// Delete 删除分类
func (s *CategoryService) Delete(id uint) error {
	if err := database.DB.Delete(&model.Category{}, id).Error; err != nil {
		return err
	}
	ClearCategoryCache()
	return nil
}

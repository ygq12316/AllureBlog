package service

import (
	"blog/internal/model"
	"blog/internal/repository"
	"blog/internal/util"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepo
}

func NewCategoryService(cr *repository.CategoryRepo) *CategoryService {
	return &CategoryService{categoryRepo: cr}
}

func (s *CategoryService) ListAll() ([]model.Category, error) {
	return s.categoryRepo.ListAll()
}

// Create 名称 → slug 的转换是业务规则，统一在 service 生成
func (s *CategoryService) Create(name string) (*model.Category, error) {
	cat := &model.Category{Name: name, Slug: util.Slugify(name)}
	if err := s.categoryRepo.Create(cat); err != nil {
		return nil, err
	}
	return cat, nil
}

func (s *CategoryService) Delete(id uint) error {
	return s.categoryRepo.Delete(id)
}

func (s *CategoryService) CountAll() (int64, error) {
	return s.categoryRepo.CountAll()
}

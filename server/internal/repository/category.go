package repository

import (
	"blog/internal/model"
	"gorm.io/gorm"
)

type CategoryRepo struct{ db *gorm.DB }

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db} }

func (r *CategoryRepo) Create(c *model.Category) error { return r.db.Create(c).Error }
func (r *CategoryRepo) Update(c *model.Category) error { return r.db.Save(c).Error }
func (r *CategoryRepo) Delete(id uint) error            { return r.db.Delete(&model.Category{}, id).Error }

func (r *CategoryRepo) FindByID(id uint) (*model.Category, error) {
	var c model.Category
	err := r.db.First(&c, id).Error
	return &c, err
}

func (r *CategoryRepo) FindBySlug(slug string) (*model.Category, error) {
	var c model.Category
	err := r.db.Where("slug = ?", slug).First(&c).Error
	return &c, err
}

func (r *CategoryRepo) ListAll() ([]model.Category, error) {
	var categories []model.Category
	err := r.db.Order("article_count DESC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepo) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Category{}).Count(&count).Error
	return count, err
}

func (r *CategoryRepo) IncrementCount(name string) {
	r.db.Model(&model.Category{}).Where("name = ?", name).UpdateColumn("article_count", gorm.Expr("article_count + ?", 1))
}

func (r *CategoryRepo) DecrementCount(name string) {
	r.db.Model(&model.Category{}).Where("name = ?", name).UpdateColumn("article_count", gorm.Expr("article_count - ?", 1))
}

package repository

import (
	"blog/internal/model"
	"gorm.io/gorm"
)

type TagRepo struct{ db *gorm.DB }

func NewTagRepo(db *gorm.DB) *TagRepo { return &TagRepo{db} }

func (r *TagRepo) Create(t *model.Tag) error { return r.db.Create(t).Error }
func (r *TagRepo) Delete(id uint) error       { return r.db.Delete(&model.Tag{}, id).Error }

func (r *TagRepo) ListAll() ([]model.Tag, error) {
	var tags []model.Tag
	err := r.db.Order("article_count DESC").Find(&tags).Error
	return tags, err
}

func (r *TagRepo) FindBySlug(slug string) (*model.Tag, error) {
	var t model.Tag
	err := r.db.Where("slug = ?", slug).First(&t).Error
	return &t, err
}

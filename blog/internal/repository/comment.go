package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
)

type CommentRepo struct{ db *gorm.DB }

func NewCommentRepo(db *gorm.DB) *CommentRepo { return &CommentRepo{db} }

func (r *CommentRepo) Create(c *model.Comment) error {
	return r.db.Create(c).Error
}

func (r *CommentRepo) Delete(id string) error {
	return r.db.Delete(&model.Comment{}, id).Error
}

func (r *CommentRepo) ListByNote(noteID uint, limit int) ([]model.CommentWithVisitor, error) {
	var results []model.CommentWithVisitor
	err := r.db.Table("comments").
		Select("comments.*, visitors.nickname, visitors.avatar_style, visitors.avatar_url, visitors.signature").
		Joins("LEFT JOIN visitors ON visitors.uuid = comments.visitor_uuid").
		Where("comments.note_id = ?", noteID).
		Order("comments.created_at ASC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

// FindByIDWithVisitor 查单条评论并带上访客资料（用于实时广播）
func (r *CommentRepo) FindByIDWithVisitor(id uint) (*model.CommentWithVisitor, error) {
	var result model.CommentWithVisitor
	err := r.db.Table("comments").
		Select("comments.*, visitors.nickname, visitors.avatar_style, visitors.avatar_url, visitors.signature").
		Joins("LEFT JOIN visitors ON visitors.uuid = comments.visitor_uuid").
		Where("comments.id = ?", id).
		First(&result).Error
	return &result, err
}

func (r *CommentRepo) CountByNote(noteID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Where("note_id = ?", noteID).Count(&count).Error
	return count, err
}

func (r *CommentRepo) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&model.Comment{}).Count(&count).Error
	return count, err
}

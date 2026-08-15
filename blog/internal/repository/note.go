package repository

import (
	"blog/internal/model"
	"gorm.io/gorm"
)

type NoteRepo struct{ db *gorm.DB }
func NewNoteRepo(db *gorm.DB) *NoteRepo { return &NoteRepo{db} }

func (r *NoteRepo) Create(n *model.Note) error { return r.db.Create(n).Error }
func (r *NoteRepo) Update(n *model.Note) error { return r.db.Save(n).Error }
func (r *NoteRepo) Delete(id uint) error { return r.db.Delete(&model.Note{}, id).Error }

func (r *NoteRepo) FindByID(id uint) (*model.Note, error) {
	var n model.Note
	err := r.db.First(&n, id).Error
	return &n, err
}

func (r *NoteRepo) ListPublished(page, pageSize int) ([]model.Note, int64, error) {
	var notes []model.Note
	var total int64
	q := r.db.Model(&model.Note{}).Where("is_published = ?", true)
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notes).Error
	return notes, total, err
}

func (r *NoteRepo) ListAll(page, pageSize int) ([]model.Note, int64, error) {
	var notes []model.Note
	var total int64
	r.db.Model(&model.Note{}).Count(&total)
	err := r.db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&notes).Error
	return notes, total, err
}

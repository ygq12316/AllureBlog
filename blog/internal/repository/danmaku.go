package repository

import (
	"blog/internal/model"

	"gorm.io/gorm"
)

type DanmakuRepo struct{ db *gorm.DB }

func NewDanmakuRepo(db *gorm.DB) *DanmakuRepo { return &DanmakuRepo{db} }

func (r *DanmakuRepo) Create(d *model.Danmaku) error {
	return r.db.Create(d).Error
}

func (r *DanmakuRepo) Delete(id string) error {
	return r.db.Delete(&model.Danmaku{}, id).Error
}

func (r *DanmakuRepo) ListRecent(limit int) ([]model.DanmakuWithVisitor, error) {
	var results []model.DanmakuWithVisitor
	err := r.db.Table("danmakus").
		Select("danmakus.*, visitors.nickname, visitors.avatar_style").
		Joins("LEFT JOIN visitors ON visitors.uuid = danmakus.visitor_uuid").
		Order("danmakus.created_at DESC").
		Limit(limit).
		Find(&results).Error
	return results, err
}

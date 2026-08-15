package repository

import (
	"blog/internal/model"
	"gorm.io/gorm"
)

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db} }

func (r *ConfigRepo) Get() (*model.BlogConfig, error) {
	var c model.BlogConfig
	err := r.db.First(&c).Error
	if err != nil {
		// 创建默认配置
		c = model.BlogConfig{AuthorName: "Allure"}
		r.db.Create(&c)
		return &c, nil
	}
	return &c, nil
}

func (r *ConfigRepo) Update(c *model.BlogConfig) error {
	return r.db.Save(c).Error
}
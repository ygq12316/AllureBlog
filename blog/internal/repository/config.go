package repository

import (
	"blog/internal/model"
	"errors"

	"gorm.io/gorm"
)

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db} }

// Get 读取博客配置；仅当记录不存在时建默认行，真实错误透传（不再吞错）
func (r *ConfigRepo) Get() (*model.BlogConfig, error) {
	var c model.BlogConfig
	err := r.db.First(&c).Error
	if err == nil {
		return &c, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c = model.BlogConfig{AuthorName: "Allure"}
		if err := r.db.Create(&c).Error; err != nil {
			return nil, err
		}
		return &c, nil
	}
	return nil, err
}

func (r *ConfigRepo) Update(c *model.BlogConfig) error {
	return r.db.Save(c).Error
}

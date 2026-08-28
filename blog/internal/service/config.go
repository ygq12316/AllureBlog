package service

import (
	"blog/internal/model"
	"blog/internal/repository"
)

type ConfigService struct {
	configRepo *repository.ConfigRepo
}

func NewConfigService(cr *repository.ConfigRepo) *ConfigService {
	return &ConfigService{configRepo: cr}
}

func (s *ConfigService) Get() (*model.BlogConfig, error) {
	return s.configRepo.Get()
}

// UpdateAuthor 先读后写的两步操作封装在此（配置表恒为单行）
func (s *ConfigService) UpdateAuthor(authorName, authorAvatar string) (*model.BlogConfig, error) {
	cfg, err := s.configRepo.Get()
	if err != nil {
		return nil, err
	}
	cfg.AuthorName = authorName
	cfg.AuthorAvatar = authorAvatar
	if err := s.configRepo.Update(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

package service

import (
	"blog/internal/model"
	"blog/internal/repository"
	"blog/internal/util"
)

// TagService 标签有两个来源：列表按已发布文章聚合（带出现次数），
// 词表本身的增删走 tags 表
type TagService struct {
	tagRepo     *repository.TagRepo
	articleRepo *repository.ArticleRepo
}

func NewTagService(tr *repository.TagRepo, ar *repository.ArticleRepo) *TagService {
	return &TagService{tagRepo: tr, articleRepo: ar}
}

func (s *TagService) ListAll() ([]model.Tag, error) {
	return s.articleRepo.GetAllTags()
}

// Create 名称 → slug 的转换与分类同规则，统一在 service 生成
func (s *TagService) Create(name string) (*model.Tag, error) {
	tag := &model.Tag{Name: name, Slug: util.Slugify(name)}
	if err := s.tagRepo.Create(tag); err != nil {
		return nil, err
	}
	return tag, nil
}

func (s *TagService) Delete(id uint) error {
	return s.tagRepo.Delete(id)
}

package service

import (
	"blog/internal/model"
	"blog/internal/repository"
	"blog/internal/util"
)

type ArticleService struct {
	articleRepo  *repository.ArticleRepo
	categoryRepo *repository.CategoryRepo
	tagRepo      *repository.TagRepo
}

func NewArticleService(ar *repository.ArticleRepo, cr *repository.CategoryRepo, tr *repository.TagRepo) *ArticleService {
	return &ArticleService{articleRepo: ar, categoryRepo: cr, tagRepo: tr}
}

func (s *ArticleService) Create(title, content, category, tags string, isPublished bool) (*model.Article, error) {
	slug := util.Slugify(title)
	// 查重不过滤发布状态:草稿撞名同样会触发 uniqueIndex,必须加后缀
	if s.articleRepo.ExistsBySlug(slug) {
		slug = slug + "-" + util.RandomSuffix(6)
	}

	html := util.RenderMarkdown(content)
	excerpt := util.ExtractExcerpt(html, 120)

	a := &model.Article{
		Title:       title,
		Slug:        slug,
		Content:     content,
		HTML:        html,
		Excerpt:     excerpt,
		Category:    category,
		Tags:        tags,
		IsPublished: isPublished,
	}

	if err := s.articleRepo.Create(a); err != nil {
		return nil, err
	}

	if category != "" {
		if cat, _ := s.categoryRepo.FindBySlug(util.Slugify(category)); cat != nil {
			s.categoryRepo.IncrementCount(cat.Name)
		}
	}

	return a, nil
}

func (s *ArticleService) Update(id uint, title, content, category, tags string, isPublished bool) (*model.Article, error) {
	a, err := s.articleRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// 分类变更时同步计数
	if a.Category != category {
		if a.Category != "" {
			s.categoryRepo.DecrementCount(a.Category)
		}
		if category != "" {
			if cat, _ := s.categoryRepo.FindBySlug(util.Slugify(category)); cat != nil {
				s.categoryRepo.IncrementCount(cat.Name)
			}
		}
	}

	a.Title = title
	a.Content = content
	a.HTML = util.RenderMarkdown(content)
	a.Excerpt = util.ExtractExcerpt(a.HTML, 120)
	a.Category = category
	a.Tags = tags
	a.IsPublished = isPublished

	if err := s.articleRepo.Update(a); err != nil {
		return nil, err
	}
	return a, nil
}

func (s *ArticleService) Delete(id uint) error {
	a, err := s.articleRepo.FindByID(id)
	if err != nil {
		return err
	}
	// 与 Create 的 Increment 对称:分类存在才减,避免计数漂移为负
	if a.Category != "" {
		if cat, _ := s.categoryRepo.FindBySlug(util.Slugify(a.Category)); cat != nil {
			s.categoryRepo.DecrementCount(cat.Name)
		}
	}
	return s.articleRepo.Delete(id)
}

func (s *ArticleService) GetBySlug(slug string) (*model.Article, error) {
	return s.articleRepo.FindBySlug(slug)
}

func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	return s.articleRepo.FindByID(id)
}

func (s *ArticleService) ListPublished(page, pageSize int) ([]model.Article, int64, error) {
	return s.articleRepo.ListPublished(page, pageSize)
}

func (s *ArticleService) ListAll(page, pageSize int) ([]model.Article, int64, error) {
	return s.articleRepo.ListAll(page, pageSize)
}

func (s *ArticleService) ListByCategory(category string, page, pageSize int) ([]model.Article, int64, error) {
	return s.articleRepo.ListByCategory(category, page, pageSize)
}

func (s *ArticleService) Search(q string, page, pageSize int) ([]model.Article, int64, error) {
	return s.articleRepo.Search(q, page, pageSize)
}

func (s *ArticleService) CountAll() (int64, error) {
	return s.articleRepo.CountAll()
}

func (s *ArticleService) GetAllTags() ([]model.Tag, error) {
	return s.articleRepo.GetAllTags()
}

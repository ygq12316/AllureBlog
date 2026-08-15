package repository

import (
	"blog/internal/model"
	"sort"
	"strings"

	"gorm.io/gorm"
)

type ArticleRepo struct{ db *gorm.DB }

func NewArticleRepo(db *gorm.DB) *ArticleRepo { return &ArticleRepo{db} }

func (r *ArticleRepo) Create(a *model.Article) error { return r.db.Create(a).Error }
func (r *ArticleRepo) Update(a *model.Article) error { return r.db.Save(a).Error }
func (r *ArticleRepo) Delete(id uint) error           { return r.db.Delete(&model.Article{}, id).Error }

func (r *ArticleRepo) FindByID(id uint) (*model.Article, error) {
	var a model.Article
	err := r.db.First(&a, id).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ArticleRepo) FindBySlug(slug string) (*model.Article, error) {
	var a model.Article
	err := r.db.Where("slug = ? AND is_published = ?", slug, true).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *ArticleRepo) ListPublished(page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	q := r.db.Model(&model.Article{}).Where("is_published = ?", true)
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error
	return articles, total, err
}

func (r *ArticleRepo) ListAll(page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	r.db.Model(&model.Article{}).Count(&total)
	err := r.db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error
	return articles, total, err
}

func (r *ArticleRepo) ListByCategory(category string, page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64
	q := r.db.Model(&model.Article{}).Where("category = ? AND is_published = ?", category, true)
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&articles).Error
	return articles, total, err
}

func (r *ArticleRepo) GetAllTags() ([]model.Tag, error) {
	// 从 articles 的 tags 字段提取所有标签并统计出现次数
	var tagStrs []string
	err := r.db.Model(&model.Article{}).
		Where("is_published = ? AND tags IS NOT NULL AND tags != ''", true).
		Pluck("tags", &tagStrs).Error
	if err != nil {
		return nil, err
	}

	counts := map[string]int{}
	for _, s := range tagStrs {
		for _, name := range strings.Split(s, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				counts[name]++
			}
		}
	}

	var tags []model.Tag
	for name, count := range counts {
		tags = append(tags, model.Tag{Name: name, Slug: name, ArticleCount: count})
	}
	// map 遍历无序，显式排序保证标签列表稳定
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].ArticleCount != tags[j].ArticleCount {
			return tags[i].ArticleCount > tags[j].ArticleCount
		}
		return tags[i].Name < tags[j].Name
	})
	return tags, nil
}

func (r *ArticleRepo) ListAllPublished() ([]model.Article, error) {
	var articles []model.Article
	err := r.db.Where("is_published = ?", true).Order("created_at DESC").Find(&articles).Error
	return articles, err
}

func (r *ArticleRepo) Search(q string, page, pageSize int) ([]model.Article, int64, error) {
	var articles []model.Article
	var ids []uint

	// 为每个词添加前缀匹配通配符
	q = strings.TrimSpace(q)
	terms := strings.Fields(q)
	for i, t := range terms {
		if !strings.ContainsAny(t, `*"`) {
			terms[i] = `"` + t + `"*`
		}
	}
	q = strings.Join(terms, " ")

	rows, err := r.db.Raw(
		"SELECT rowid FROM articles_fts WHERE articles_fts MATCH ? ORDER BY rank LIMIT ? OFFSET ?",
		q, pageSize, (page-1)*pageSize,
	).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var id uint
		rows.Scan(&id)
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return articles, 0, nil
	}

	r.db.Where("id IN ? AND is_published = ?", ids, true).Order("created_at DESC").Find(&articles)
	return articles, int64(len(articles)), nil
}

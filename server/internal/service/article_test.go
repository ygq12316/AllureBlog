package service

import (
	"testing"

	"blog/internal/model"
	"blog/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newArticleTestSvc(t *testing.T) *ArticleService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return NewArticleService(repository.NewArticleRepo(db), repository.NewCategoryRepo(db), repository.NewTagRepo(db))
}

// 回归:与草稿撞 slug 必须加后缀,而不是触发 uniqueIndex 裸 500
func TestCreateSlugCollisionWithDraft(t *testing.T) {
	svc := newArticleTestSvc(t)

	draft, err := svc.Create("同名文章", "内容一", "", "", false)
	if err != nil {
		t.Fatalf("创建草稿: %v", err)
	}
	second, err := svc.Create("同名文章", "内容二", "", "", false)
	if err != nil {
		t.Fatalf("同名再创建不应报错: %v", err)
	}
	if second.Slug == draft.Slug {
		t.Errorf("slug 冲突未被消解: %q == %q", second.Slug, draft.Slug)
	}
}

// 回归:删除文章时分类不存在则不计数,避免 article_count 漂移为负
func TestDeleteDoesNotDecrementMissingCategory(t *testing.T) {
	svc := newArticleTestSvc(t)

	// 文章带一个从未注册过的分类(Create 时不会 Increment,Delete 也不应 Decrement)
	a, err := svc.Create("文章", "内容", "不存在的分类", "", true)
	if err != nil {
		t.Fatalf("创建: %v", err)
	}
	if err := svc.Delete(a.ID); err != nil {
		t.Fatalf("删除: %v", err)
	}
}

package database

import (
	"blog/internal/model"
	"fmt"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(
		&model.Article{},
		&model.Category{},
		&model.Tag{},
		&model.Note{},
		&model.Visitor{},
		&model.Comment{},
		&model.Danmaku{},
		&model.BlogConfig{},
	); err != nil {
		return nil, fmt.Errorf("自动迁移失败: %w", err)
	}

	// exec 记录所有 DDL 错误，不再静默
	exec := func(sql string) {
		if e := db.Exec(sql).Error; e != nil {
			log.Printf("database: %q 执行失败: %v", sql, e)
		}
	}

	exec("PRAGMA journal_mode=WAL")
	exec("PRAGMA busy_timeout=5000")

	// 修复 username 唯一约束（仅非空值唯一）
	exec("DROP INDEX IF EXISTS idx_visitors_username")
	exec("CREATE UNIQUE INDEX IF NOT EXISTS unq_visitors_username ON visitors(username) WHERE username != ''")

	// FTS5：仅当索引缺 tags 列时才重建（重建需全量重灌，不能放在每次启动路径）
	var cols []struct{ Name string }
	db.Raw("PRAGMA table_info(articles_fts)").Scan(&cols)
	hasTags := false
	for _, c := range cols {
		if c.Name == "tags" {
			hasTags = true
			break
		}
	}

	if !hasTags {
		exec(`DROP TABLE IF EXISTS articles_fts`)
		exec(`DROP TRIGGER IF EXISTS articles_ai`)
		exec(`DROP TRIGGER IF EXISTS articles_ad`)
		exec(`DROP TRIGGER IF EXISTS articles_au`)

		exec(`CREATE VIRTUAL TABLE articles_fts USING fts5(title, content, tags, content=articles, content_rowid=id)`)
		exec(`CREATE TRIGGER articles_ai AFTER INSERT ON articles BEGIN
			INSERT INTO articles_fts(rowid, title, content, tags) VALUES (new.id, new.title, new.content, new.tags); END`)
		exec(`CREATE TRIGGER articles_ad AFTER DELETE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, title, content, tags) VALUES('delete', old.id, old.title, old.content, old.tags); END`)
		exec(`CREATE TRIGGER articles_au AFTER UPDATE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, title, content, tags) VALUES('delete', old.id, old.title, old.content, old.tags);
			INSERT INTO articles_fts(rowid, title, content, tags) VALUES (new.id, new.title, new.content, new.tags); END`)

		exec(`INSERT INTO articles_fts(rowid, title, content, tags) SELECT id, title, content, tags FROM articles`)
	}

	return db, nil
}

package repository

import (
	"testing"

	"blog/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// gormOpenMem 打开单连接内存库（不迁移）
// :memory: 下每个连接是独立库，必须强制单连接
func gormOpenMem() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gormOpenMem()
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	if err := db.AutoMigrate(&model.Note{}, &model.Comment{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	return db
}

func TestNoteRepoListAllCommentCount(t *testing.T) {
	db := newTestDB(t)
	nr := NewNoteRepo(db)
	cr := NewCommentRepo(db)

	n1 := &model.Note{Content: "随笔一", HTML: "随笔一", IsPublished: true}
	n2 := &model.Note{Content: "随笔二", HTML: "随笔二", IsPublished: false}
	if err := nr.Create(n1); err != nil {
		t.Fatal(err)
	}
	if err := nr.Create(n2); err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 2; i++ {
		if err := cr.Create(&model.Comment{NoteID: n1.ID, VisitorUUID: "u1", Content: "评论"}); err != nil {
			t.Fatal(err)
		}
	}

	notes, total, err := nr.ListAll(1, 50)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if total != 2 {
		t.Errorf("total = %d, want 2", total)
	}
	byID := map[uint]int64{}
	for _, n := range notes {
		byID[n.ID] = n.CommentCount
	}
	if byID[n1.ID] != 2 {
		t.Errorf("随笔 %d comment_count = %d, want 2", n1.ID, byID[n1.ID])
	}
	if byID[n2.ID] != 0 {
		t.Errorf("随笔 %d comment_count = %d, want 0", n2.ID, byID[n2.ID])
	}
}

func TestNoteRepoListPublishedCommentCount(t *testing.T) {
	db := newTestDB(t)
	nr := NewNoteRepo(db)
	cr := NewCommentRepo(db)

	pub := &model.Note{Content: "已发布", HTML: "已发布", IsPublished: true}
	draft := &model.Note{Content: "草稿", HTML: "草稿", IsPublished: false}
	if err := nr.Create(pub); err != nil {
		t.Fatal(err)
	}
	if err := nr.Create(draft); err != nil {
		t.Fatal(err)
	}
	if err := cr.Create(&model.Comment{NoteID: pub.ID, VisitorUUID: "u1", Content: "评论"}); err != nil {
		t.Fatal(err)
	}

	notes, total, err := nr.ListPublished(1, 50)
	if err != nil {
		t.Fatalf("ListPublished: %v", err)
	}
	if total != 1 || len(notes) != 1 {
		t.Fatalf("total=%d len=%d, want 1/1（草稿不应出现）", total, len(notes))
	}
	if notes[0].CommentCount != 1 {
		t.Errorf("comment_count = %d, want 1", notes[0].CommentCount)
	}
}

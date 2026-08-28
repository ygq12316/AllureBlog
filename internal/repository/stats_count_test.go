package repository

import (
	"testing"

	"blog/internal/model"
)

func TestCommentRepoCountAll(t *testing.T) {
	db := newTestDB(t) // note_test.go 中的助手（同包共享）
	cr := NewCommentRepo(db)

	n := &model.Note{Content: "随笔", HTML: "随笔", IsPublished: true}
	if err := NewNoteRepo(db).Create(n); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := cr.Create(&model.Comment{NoteID: n.ID, VisitorUUID: "u1", Content: "评论"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := cr.CountAll()
	if err != nil {
		t.Fatalf("CountAll: %v", err)
	}
	if got != 3 {
		t.Errorf("CountAll = %d, want 3", got)
	}
}

func TestVisitorRepoCountAll(t *testing.T) {
	db, err := gormOpenMem()
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Visitor{}); err != nil {
		t.Fatal(err)
	}
	vr := NewVisitorRepo(db)
	for i := 0; i < 2; i++ {
		if err := vr.Create(&model.Visitor{UUID: "u" + string(rune('1'+i)), Nickname: "访客", AvatarStyle: "lorelei"}); err != nil {
			t.Fatal(err)
		}
	}
	got, err := vr.CountAll()
	if err != nil {
		t.Fatalf("CountAll: %v", err)
	}
	if got != 2 {
		t.Errorf("CountAll = %d, want 2", got)
	}
}

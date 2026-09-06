package service

import (
	"strconv"
	"testing"

	"blog/internal/model"
	"blog/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newCommentTestSvc(t *testing.T) *CommentService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.Comment{}, &model.Visitor{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	visitorRepo := repository.NewVisitorRepo(db)
	// 两个登录账号 + 一个匿名身份：评论只允许登录用户（访客模式已移除）
	for _, v := range []*model.Visitor{
		{UUID: "u1", Username: "alice", Nickname: "甲"},
		{UUID: "u2", Username: "bob", Nickname: "乙"},
		{UUID: "u3", Username: "carol", Nickname: "丙"},
		{UUID: "ghost", Username: "", Nickname: "匿名路人"},
	} {
		if err := visitorRepo.Create(v); err != nil {
			t.Fatalf("建访客: %v", err)
		}
	}
	return NewCommentService(repository.NewCommentRepo(db), visitorRepo)
}

// 回归:回复的回复必须归根到根评论,层级恒为两层,且返回被回复人昵称
func TestCreateReplyFlattensToRoot(t *testing.T) {
	svc := newCommentTestSvc(t)

	root, _, err := svc.Create(1, "u1", "根评论", nil)
	if err != nil {
		t.Fatalf("建根评论: %v", err)
	}
	reply, _, err := svc.Create(1, "u2", "一级回复", &root.ID)
	if err != nil {
		t.Fatalf("建回复: %v", err)
	}
	reply2, replyTo, err := svc.Create(1, "u3", "回复的回复", &reply.ID)
	if err != nil {
		t.Fatalf("建回复的回复: %v", err)
	}

	if reply2.ParentID == nil || *reply2.ParentID != root.ID {
		t.Errorf("回复的回复应归根到根评论 %d, 实际 %v", root.ID, reply2.ParentID)
	}
	if replyTo != "乙" {
		t.Errorf("replyTo 应为被回复人昵称「乙」, 实际 %q", replyTo)
	}
}

// 回归:不能回复其他随笔下的评论
func TestCreateReplyRejectsCrossNote(t *testing.T) {
	svc := newCommentTestSvc(t)

	root, _, err := svc.Create(1, "u1", "根评论", nil)
	if err != nil {
		t.Fatalf("建根评论: %v", err)
	}
	other := uint(2)
	if _, _, err := svc.Create(other, "u2", "跨随笔回复", &root.ID); err == nil {
		t.Fatal("跨随笔回复应报错")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("应为 ValidationError, 实际 %T", err)
	}
}

// 回复不存在的父评论应报业务错误而非 500
func TestCreateReplyRejectsMissingParent(t *testing.T) {
	svc := newCommentTestSvc(t)

	missing := uint(999)
	if _, _, err := svc.Create(1, "u1", "回复", &missing); err == nil {
		t.Fatal("父评论不存在应报错")
	} else if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("应为 ValidationError, 实际 %T", err)
	}
}

// 回归:访客模式已移除,匿名身份(uuid 无 username)不可评论
func TestCreateRejectsAnonymousVisitor(t *testing.T) {
	svc := newCommentTestSvc(t)

	anon, _, err := svc.Create(1, "ghost", "匿名评论", nil)
	if err == nil {
		t.Fatalf("匿名访客不应能评论: %+v", anon)
	}
	if _, ok := err.(*ValidationError); !ok {
		t.Fatalf("应为 ValidationError, 实际 %T", err)
	}
}

// 回归:删除根评论应级联删除其全部回复
func TestDeleteCascadeRemovesReplies(t *testing.T) {
	svc := newCommentTestSvc(t)

	root, _, _ := svc.Create(1, "u1", "根评论", nil)
	r1, _, _ := svc.Create(1, "u2", "回复一", &root.ID)
	if _, _, err := svc.Create(1, "u3", "回复二", &r1.ID); err != nil {
		t.Fatalf("建嵌套回复: %v", err)
	}

	if err := svc.Delete(strconv.FormatUint(uint64(root.ID), 10)); err != nil {
		t.Fatalf("删除根评论: %v", err)
	}

	comments, total, err := svc.ListByNote(1)
	if err != nil {
		t.Fatalf("列评论: %v", err)
	}
	if len(comments) != 0 || total != 0 {
		t.Errorf("级联删除后应剩 0 条, 实际 %d 条 (total=%d)", len(comments), total)
	}
}

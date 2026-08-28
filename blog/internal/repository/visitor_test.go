package repository

import (
	"errors"
	"testing"

	"blog/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newVisitorTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开内存库失败: %v", err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	if err := db.AutoMigrate(&model.Visitor{}); err != nil {
		t.Fatalf("自动迁移失败: %v", err)
	}
	// username 的部分唯一索引由 database.InitDB 手工维护,测试库需同样建上
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS unq_visitors_username ON visitors(username) WHERE username != ''").Error; err != nil {
		t.Fatalf("建唯一索引失败: %v", err)
	}
	return db
}

func TestIsUniqueViolation(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil 错误", nil, false},
		{"GORM 重复键", gorm.ErrDuplicatedKey, true},
		{"SQLite 唯一约束消息", errors.New("UNIQUE constraint failed: visitors.username"), true},
		{"其他错误", errors.New("no such table: visitors"), false},
		{"包裹的唯一约束", errors.New("创建访客失败: UNIQUE constraint failed: visitors.uuid"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isUniqueViolation(tc.err); got != tc.want {
				t.Errorf("isUniqueViolation(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestRegisterDuplicateReturnsSentinel(t *testing.T) {
	db := newVisitorTestDB(t)
	r := NewVisitorRepo(db)

	v := &model.Visitor{UUID: "u-1", Username: "小明"}
	if err := r.Register(v, "pass1234"); err != nil {
		t.Fatalf("首次注册: %v", err)
	}

	// 重名应翻译为 ErrDuplicate sentinel,HTTP 层靠 errors.Is 判定 409
	dup := &model.Visitor{UUID: "u-2", Username: "小明"}
	if err := r.Register(dup, "pass5678"); !errors.Is(err, ErrDuplicate) {
		t.Fatalf("重名注册 err = %v, want ErrDuplicate", err)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	db := newVisitorTestDB(t)
	r := NewVisitorRepo(db)
	if err := r.Register(&model.Visitor{UUID: "u-1", Username: "小明"}, "pass1234"); err != nil {
		t.Fatalf("注册: %v", err)
	}
	if _, err := r.Login("小明", "wrong-pass"); err == nil {
		t.Error("错误密码不应登录成功")
	}
	if _, err := r.Login("小明", "pass1234"); err != nil {
		t.Errorf("正确密码应登录成功: %v", err)
	}
}

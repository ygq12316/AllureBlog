package handler

import (
	"errors"
	"testing"

	"gorm.io/gorm"
)

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

func TestAllowedUploadExts(t *testing.T) {
	allowed := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	for _, ext := range allowed {
		if !allowedUploadExts[ext] {
			t.Errorf("扩展名 %s 应在白名单", ext)
		}
	}
	dangerous := []string{".svg", ".html", ".exe", ".php", ".sh", ".js"}
	for _, ext := range dangerous {
		if allowedUploadExts[ext] {
			t.Errorf("扩展名 %s 不应在白名单", ext)
		}
	}
}

package handler

import (
	"testing"
)

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

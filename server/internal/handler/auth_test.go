package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"blog/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// 统一账号体系：登录入口在 VisitorHandler（账号密码+bcrypt），
// Auth 只负责签发与角色鉴权，本文件覆盖令牌与中间件矩阵

func newTestAuth() *Auth {
	return NewAuth(AuthConfig{Secret: "test-secret", User: "admin", Pass: "test-pass"}, nil)
}

func newProtectedR(a *Auth) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", a.Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	return r
}

func doReq(r *gin.Engine, authHeader string) int {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestIssueTokenCarriesRole(t *testing.T) {
	a := newTestAuth()
	admin := &model.Visitor{UUID: "admin_admin", Username: "admin", Role: "admin"}
	user := &model.Visitor{UUID: "acct_1", Username: "bob", Role: "user"}

	adminToken, err := a.IssueToken(admin)
	if err != nil {
		t.Fatalf("签发管理员令牌: %v", err)
	}
	userToken, _ := a.IssueToken(user)

	r := newProtectedR(a)
	if code := doReq(r, "Bearer "+adminToken); code != http.StatusNoContent {
		t.Errorf("管理员 token: code = %d, want 204", code)
	}
	if code := doReq(r, "Bearer "+userToken); code != http.StatusForbidden {
		t.Errorf("普通用户 token: code = %d, want 403", code)
	}
}

func TestMiddlewareRejects(t *testing.T) {
	a := newTestAuth()
	r := newProtectedR(a)

	if code := doReq(r, ""); code != http.StatusUnauthorized {
		t.Errorf("无 token: code = %d, want 401", code)
	}
	if code := doReq(r, "Bearer garbage"); code != http.StatusUnauthorized {
		t.Errorf("垃圾 token: code = %d, want 401", code)
	}

	// 旧版令牌（无 role 声明）：统一角色后必须被拒，防止旧令牌越权进后台
	legacy := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "admin",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	legacyStr, _ := legacy.SignedString([]byte("test-secret"))
	if code := doReq(r, "Bearer "+legacyStr); code != http.StatusForbidden {
		t.Errorf("无角色旧令牌: code = %d, want 403", code)
	}

	// 过期 token
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "admin", "role": "admin",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	expiredStr, _ := expired.SignedString([]byte("test-secret"))
	if code := doReq(r, "Bearer "+expiredStr); code != http.StatusUnauthorized {
		t.Errorf("过期 token: code = %d, want 401", code)
	}

	// 异密钥签发
	other := NewAuth(AuthConfig{Secret: "other-secret", User: "admin", Pass: "x"}, nil)
	otherToken, _ := other.IssueToken(&model.Visitor{Username: "admin", Role: "admin"})
	if code := doReq(r, "Bearer "+otherToken); code != http.StatusUnauthorized {
		t.Errorf("异密钥 token: code = %d, want 401", code)
	}
}

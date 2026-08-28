package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth 的测试不再需要环境变量：配置经构造注入，这正是去全局化的目的

func newTestAuth() *Auth {
	return NewAuth(AuthConfig{Secret: "test-secret", User: "admin", Pass: "test-pass"}, nil)
}

func postLogin(t *testing.T, a *Auth, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/login", a.Login)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestLoginSuccess(t *testing.T) {
	w := postLogin(t, newTestAuth(), `{"username":"admin","password":"test-pass"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("code = %d, want 200", w.Code)
	}
	var resp struct {
		Token string `json:"token"`
		User  string `json:"user"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应: %v", err)
	}
	if resp.Token == "" || resp.User != "admin" {
		t.Errorf("token/user 不符合预期: %+v", resp)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	w := postLogin(t, newTestAuth(), `{"username":"admin","password":"wrong"}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestLoginEmptyCredentials(t *testing.T) {
	// 空账号密码经常数时间比对后按「账号或密码错误」401（与错误凭据一致）
	w := postLogin(t, newTestAuth(), `{}`)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code = %d, want 401", w.Code)
	}
}

func TestLoginMalformedJSON(t *testing.T) {
	w := postLogin(t, newTestAuth(), `{`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code = %d, want 400", w.Code)
	}
}

func TestMiddleware(t *testing.T) {
	a := newTestAuth()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", a.Middleware(), func(c *gin.Context) { c.Status(http.StatusNoContent) })

	do := func(authHeader string) int {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		return w.Code
	}

	if code := do(""); code != http.StatusUnauthorized {
		t.Errorf("无 token: code = %d, want 401", code)
	}

	// 用 Login 签发的真 token
	lw := postLogin(t, a, `{"username":"admin","password":"test-pass"}`)
	var resp struct{ Token string `json:"token"` }
	json.Unmarshal(lw.Body.Bytes(), &resp)
	if code := do("Bearer " + resp.Token); code != http.StatusNoContent {
		t.Errorf("有效 token: code = %d, want 204", code)
	}

	// 签名密钥不匹配（另一个 Auth 实例签发）
	other := NewAuth(AuthConfig{Secret: "other-secret", User: "admin", Pass: "test-pass"}, nil)
	ow := postLogin(t, other, `{"username":"admin","password":"test-pass"}`)
	var oresp struct{ Token string `json:"token"` }
	json.Unmarshal(ow.Body.Bytes(), &oresp)
	if code := do("Bearer " + oresp.Token); code != http.StatusUnauthorized {
		t.Errorf("异密钥 token: code = %d, want 401", code)
	}

	// 过期 token
	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": "admin",
		"exp":  time.Now().Add(-time.Hour).Unix(),
	})
	expiredStr, _ := expired.SignedString([]byte("test-secret"))
	if code := do("Bearer " + expiredStr); code != http.StatusUnauthorized {
		t.Errorf("过期 token: code = %d, want 401", code)
	}
}

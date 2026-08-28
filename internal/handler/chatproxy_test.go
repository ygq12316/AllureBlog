package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestChatProxyForwards(t *testing.T) {
	var gotPath, gotHeader string
	up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotHeader = r.Header.Get("X-Test")
		w.Header().Set("X-Upstream", "yes")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer up.Close()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Any("/chat/*path", ChatProxy(up.URL))
	// ReverseProxy 依赖 CloseNotifier，须走真实 HTTP server（ResponseRecorder 会 panic）
	ts := httptest.NewServer(r)
	defer ts.Close()

	req, _ := http.NewRequest(http.MethodGet, ts.URL+"/chat/health", nil)
	req.Header.Set("X-Test", "1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("状态码 = %d, want 200", resp.StatusCode)
	}
	if gotPath != "/chat/health" {
		t.Errorf("上游收到路径 = %q, want /chat/health", gotPath)
	}
	if gotHeader != "1" {
		t.Error("请求头未透传到上游")
	}
	if resp.Header.Get("X-Upstream") != "yes" {
		t.Error("上游响应头未透传")
	}
	if string(body) != `{"status":"ok"}` {
		t.Errorf("响应体 = %q, want {\"status\":\"ok\"}", body)
	}
}

func TestChatProxyBadURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// url.Parse 对明显非法的 control 字符会报错
	r.Any("/chat/*path", ChatProxy("http://loc\x00alhost:8000"))
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/chat/ws")
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("无效地址应返回 503, got %d", resp.StatusCode)
	}
}

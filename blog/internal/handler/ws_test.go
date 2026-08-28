package handler

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func newWSServer(t *testing.T, hub *Hub) *httptest.Server {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/ws/:id", hub.HandleNoteWS)
	srv := httptest.NewServer(r)
	t.Cleanup(srv.Close)
	return srv
}

func dialWS(t *testing.T, srv *httptest.Server, noteID string) *websocket.Conn {
	t.Helper()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws/" + noteID
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("WS 握手失败: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}

func TestHubBroadcastDelivers(t *testing.T) {
	hub := NewHub()
	srv := newWSServer(t, hub)
	conn := dialWS(t, srv, "7")

	// 等注册完成
	time.Sleep(50 * time.Millisecond)
	hub.Broadcast(7, map[string]interface{}{"type": "comment", "comment": "你好"})

	var msg struct {
		Type    string `json:"type"`
		Comment string `json:"comment"`
	}
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("读取广播: %v", err)
	}
	if msg.Type != "comment" || msg.Comment != "你好" {
		t.Errorf("广播内容不符: %+v", msg)
	}
}

func TestHubBroadcastIsolatesNotes(t *testing.T) {
	hub := NewHub()
	srv := newWSServer(t, hub)
	connA := dialWS(t, srv, "1")
	connB := dialWS(t, srv, "2")
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast(1, map[string]string{"type": "comment"})

	connA.SetReadDeadline(time.Now().Add(2 * time.Second))
	var m map[string]string
	if err := connA.ReadJSON(&m); err != nil {
		t.Fatalf("订阅随笔1应收到广播: %v", err)
	}

	// 随笔2的订阅者不应收到（读超时即正确）
	connB.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if err := connB.ReadJSON(&m); err == nil {
		t.Error("订阅随笔2不应收到随笔1的广播")
	}
}

// 慢连接满缓冲时 Broadcast 必须立即返回，不阻塞调用方（发评论请求）
func TestHubBroadcastNeverBlocks(t *testing.T) {
	hub := NewHub()
	srv := newWSServer(t, hub)
	conn := dialWS(t, srv, "3") // 故意永不读,缓冲将被填满
	defer conn.Close()
	time.Sleep(50 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		for i := 0; i < wsSendBuf*4; i++ {
			hub.Broadcast(3, map[string]int{"n": i})
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Broadcast 被慢连接阻塞")
	}
}

func TestHubUnregisterOnClose(t *testing.T) {
	hub := NewHub()
	srv := newWSServer(t, hub)
	conn := dialWS(t, srv, "9")
	time.Sleep(50 * time.Millisecond)
	conn.Close()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		hub.mu.RLock()
		n := len(hub.clients[9])
		hub.mu.RUnlock()
		if n == 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Error("断连后订阅未清理")
}

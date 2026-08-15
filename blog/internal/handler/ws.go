package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	// 仅允许同源连接（非浏览器客户端无 Origin 头，放行）
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		return err == nil && u.Host == r.Host
	},
}

// client 封装连接与写锁：gorilla/websocket 不允许并发写同一连接
type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *client) writeJSON(msg interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteJSON(msg)
}

// wsHub 管理每个随笔的 WebSocket 客户端
type wsHub struct {
	mu      sync.RWMutex
	clients map[uint]map[*client]bool
}

var defaultWSHub = &wsHub{clients: make(map[uint]map[*client]bool)}

// HandleNoteWS /api/notes/:id/ws
func HandleNoteWS(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	cl := &client{conn: conn}

	defaultWSHub.mu.Lock()
	if defaultWSHub.clients[uint(noteID)] == nil {
		defaultWSHub.clients[uint(noteID)] = make(map[*client]bool)
	}
	defaultWSHub.clients[uint(noteID)][cl] = true
	defaultWSHub.mu.Unlock()

	defer func() {
		defaultWSHub.mu.Lock()
		delete(defaultWSHub.clients[uint(noteID)], cl)
		if len(defaultWSHub.clients[uint(noteID)]) == 0 {
			delete(defaultWSHub.clients, uint(noteID))
		}
		defaultWSHub.mu.Unlock()
		conn.Close()
	}()

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

// broadcastNoteComment 向某篇随笔的所有订阅者推送新评论
func broadcastNoteComment(noteID uint, msg interface{}) {
	defaultWSHub.mu.RLock()
	clients := make([]*client, 0, len(defaultWSHub.clients[noteID]))
	for cl := range defaultWSHub.clients[noteID] {
		clients = append(clients, cl)
	}
	defaultWSHub.mu.RUnlock()

	for _, cl := range clients {
		if err := cl.writeJSON(msg); err != nil {
			// 写失败的连接交给 HandleNoteWS 的读循环退出后清理
			continue
		}
	}
}

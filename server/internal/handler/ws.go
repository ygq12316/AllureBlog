package handler

import (
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	wsWriteWait  = 5 * time.Second  // 写超时：半开连接拖不死广播方
	wsPongWait   = 60 * time.Second // 读超时：靠 ping/pong 续期
	wsPingPeriod = 45 * time.Second // 心跳间隔，须小于读超时
	wsSendBuf    = 32               // 每连接发送缓冲；满则丢消息（由读超时/写超时最终清连接）
)

var wsUpgrader = websocket.Upgrader{
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

// wsClient 单连接：所有写出都经 send 通道由 writer goroutine 串行完成，
// 满足 gorilla/websocket 不允许并发写同一连接的约束。
type wsClient struct {
	conn *websocket.Conn
	send chan interface{}
}

// Hub 管理每个随笔的 WebSocket 订阅者。由 main 组装注入，非包级全局。
type Hub struct {
	mu      sync.RWMutex
	clients map[uint]map[*wsClient]bool
}

func NewHub() *Hub {
	return &Hub{clients: make(map[uint]map[*wsClient]bool)}
}

// Broadcast 向某篇随笔的所有订阅者推送。非阻塞：慢连接满缓冲即丢消息，
// 半开/卡死连接由写超时断连，绝不阻塞调用方（发评论的请求 goroutine）。
func (h *Hub) Broadcast(noteID uint, msg interface{}) {
	h.mu.RLock()
	snapshot := make([]*wsClient, 0, len(h.clients[noteID]))
	for cl := range h.clients[noteID] {
		snapshot = append(snapshot, cl)
	}
	h.mu.RUnlock()

	for _, cl := range snapshot {
		select {
		case cl.send <- msg:
		default:
		}
	}
}

// HandleNoteWS GET /api/notes/:id/ws — 单篇随笔房间
func (h *Hub) HandleNoteWS(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	h.serve(uint(noteID), c)
}

// HandleGlobalWS GET /api/ws — 全局房间（room 0，弹幕实时广播）
func (h *Hub) HandleGlobalWS(c *gin.Context) {
	h.serve(0, c)
}

// serve 升级连接并加入房间；客户端只读不发消息，读循环仅用于感知断连与 pong 续期
func (h *Hub) serve(room uint, c *gin.Context) {
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[ws] 房间 %d 握手失败: %v", room, err)
		return
	}
	cl := &wsClient{conn: conn, send: make(chan interface{}, wsSendBuf)}

	h.mu.Lock()
	if h.clients[room] == nil {
		h.clients[room] = make(map[*wsClient]bool)
	}
	h.clients[room][cl] = true
	h.mu.Unlock()

	done := make(chan struct{})
	go cl.writer(done)

	// 读循环：客户端不主动发消息，仅用于感知断连与 pong 续期
	conn.SetReadLimit(4096)
	conn.SetReadDeadline(time.Now().Add(wsPongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(wsPongWait))
		return nil
	})
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}

	// 断连清理：摘除订阅、通知 writer 退出并关连接
	h.mu.Lock()
	delete(h.clients[room], cl)
	if len(h.clients[room]) == 0 {
		delete(h.clients, room)
	}
	h.mu.Unlock()
	close(done)
}

// writer 独占连接写出；done 关闭或写超时即退出并关闭连接。
func (cl *wsClient) writer(done <-chan struct{}) {
	ticker := time.NewTicker(wsPingPeriod)
	defer func() {
		ticker.Stop()
		cl.conn.Close()
	}()
	for {
		select {
		case msg := <-cl.send:
			cl.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := cl.conn.WriteJSON(msg); err != nil {
				return
			}
		case <-ticker.C:
			cl.conn.SetWriteDeadline(time.Now().Add(wsWriteWait))
			if err := cl.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}

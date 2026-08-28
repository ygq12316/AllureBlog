package handler

import (
	"blog/internal/service"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// VisitorHandler 访客公开接口（身份/评论/弹幕）。
// 业务规则（长度上限、访客存在性、缺省值）全部在 service 层
type VisitorHandler struct {
	visitorSvc *service.VisitorService
	commentSvc *service.CommentService
	danmakuSvc *service.DanmakuService
	hub        *Hub
}

func NewVisitorHandler(vs *service.VisitorService, cs *service.CommentService, ds *service.DanmakuService, hub *Hub) *VisitorHandler {
	return &VisitorHandler{visitorSvc: vs, commentSvc: cs, danmakuSvc: ds, hub: hub}
}

// POST /api/visitor (匿名访客身份更新)
func (h *VisitorHandler) RegisterAnonymous(c *gin.Context) {
	var req struct {
		UUID        string `json:"uuid" binding:"required"`
		Nickname    string `json:"nickname"`
		AvatarStyle string `json:"avatar_style"`
		AvatarURL   string `json:"avatar_url"`
		Signature   string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	v, err := h.visitorSvc.RegisterAnonymous(req.UUID, req.Nickname, req.AvatarStyle, req.AvatarURL, req.Signature)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v})
}

// POST /api/visitor/register (账号注册)
func (h *VisitorHandler) RegisterWithAccount(c *gin.Context) {
	var req struct {
		UUID     string `json:"uuid" binding:"required"`
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	v, err := h.visitorSvc.RegisterWithAccount(req.UUID, req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUsernameTaken):
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		default:
			respondError(c, err)
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v, "message": "注册成功"})
}

// POST /api/visitor/login (账号登录)
func (h *VisitorHandler) LoginAccount(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	v, err := h.visitorSvc.Login(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v})
}

// GET /api/visitor/:uuid
func (h *VisitorHandler) GetVisitor(c *gin.Context) {
	v, err := h.visitorSvc.GetByUUID(c.Param("uuid"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v})
}

// GET /api/notes/:id/comments
func (h *VisitorHandler) ListComments(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	comments, total, err := h.commentSvc.ListByNote(uint(noteID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"comments": comments, "total": total})
}

// POST /api/notes/:id/comments
func (h *VisitorHandler) CreateComment(c *gin.Context) {
	noteID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var req struct {
		VisitorUUID string `json:"visitor_uuid" binding:"required"`
		Content     string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	fresh, err := h.commentSvc.Create(uint(noteID), req.VisitorUUID, req.Content)
	if err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "访客不存在，请先注册"})
			return
		}
		respondError(c, err)
		return
	}

	// 广播给所有订阅该随笔的用户（service 已返回带访客资料的整条）
	h.hub.Broadcast(uint(noteID), gin.H{"type": "comment", "comment": fresh})

	c.JSON(http.StatusOK, gin.H{"comment": fresh})
}

// GET /api/danmaku
func (h *VisitorHandler) ListDanmaku(c *gin.Context) {
	list, err := h.danmakuSvc.ListRecent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"danmaku": list})
}

// POST /api/danmaku
func (h *VisitorHandler) CreateDanmaku(c *gin.Context) {
	var req struct {
		VisitorUUID string `json:"visitor_uuid" binding:"required"`
		Content     string `json:"content" binding:"required"`
		Color       string `json:"color"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	danmaku, err := h.danmakuSvc.Create(req.VisitorUUID, req.Content, req.Color)
	if err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "访客不存在，请先注册"})
			return
		}
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"danmaku": danmaku})
}

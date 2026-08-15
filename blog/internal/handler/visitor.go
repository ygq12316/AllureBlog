package handler

import (
	"blog/internal/model"
	"blog/internal/service"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type VisitorHandler struct {
	visitorSvc *service.VisitorService
}

func NewVisitorHandler(vs *service.VisitorService) *VisitorHandler {
	return &VisitorHandler{visitorSvc: vs}
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
	if req.AvatarStyle == "" {
		req.AvatarStyle = "lorelei"
	}
	v := &model.Visitor{
		UUID:        req.UUID,
		Nickname:    req.Nickname,
		AvatarStyle: req.AvatarStyle,
		AvatarURL:   req.AvatarURL,
		Signature:   req.Signature,
	}
	if err := h.visitorSvc.Register(v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
	if len(req.Username) < 2 || len(req.Username) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名2-20个字符"})
		return
	}
	if len(req.Password) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码至少4位"})
		return
	}
	// 直接依赖数据库唯一索引判定重名，避免 check-then-act 竞态
	v := &model.Visitor{
		UUID:        req.UUID,
		Username:    req.Username,
		Nickname:    req.Username,
		AvatarStyle: "lorelei",
	}
	if err := h.visitorSvc.RegisterWithPassword(v, req.Password); err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "用户名已存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v, "message": "注册成功"})
}

// isUniqueViolation 判定唯一约束冲突（GORM 未开启 TranslateError 时走消息匹配兜底）
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
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
	uuid := c.Param("uuid")
	v, err := h.visitorSvc.GetByUUID(uuid)
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
	comments, total, err := h.visitorSvc.ListComments(uint(noteID), 100)
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
	if len(req.Content) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "评论不能超过500字"})
		return
	}
	if !h.visitorSvc.VisitorExists(req.VisitorUUID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "访客不存在，请先注册"})
		return
	}
	comment, err := h.visitorSvc.CreateComment(uint(noteID), req.VisitorUUID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 广播给所有订阅该随笔的用户（单查该条，避免全量拉取再筛选）
	fresh, ferr := h.visitorSvc.GetCommentWithVisitor(comment.ID)
	if ferr == nil {
		broadcastNoteComment(uint(noteID), gin.H{"type": "comment", "comment": fresh})
	}

	c.JSON(http.StatusOK, gin.H{"comment": comment})
}

// GET /api/danmaku
func (h *VisitorHandler) ListDanmaku(c *gin.Context) {
	list, err := h.visitorSvc.ListDanmaku(50)
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
	if len(req.Content) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "弹幕不能超过100字"})
		return
	}
	if req.Color == "" {
		req.Color = "#b8944c"
	}
	if !h.visitorSvc.VisitorExists(req.VisitorUUID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "访客不存在，请先注册"})
		return
	}
	danmaku, err := h.visitorSvc.CreateDanmaku(req.VisitorUUID, req.Content, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"danmaku": danmaku})
}

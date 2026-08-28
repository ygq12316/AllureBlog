package handler

import (
	"blog/internal/service"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// AdminHandler 管理后台 + 公开读接口。只依赖 service 层，
// 不直接触碰 repository（上传是纯文件操作，无需 service）
type AdminHandler struct {
	articleSvc  *service.ArticleService
	noteSvc     *service.NoteService
	categorySvc *service.CategoryService
	tagSvc      *service.TagService
	commentSvc  *service.CommentService
	danmakuSvc  *service.DanmakuService
	visitorSvc  *service.VisitorService
	configSvc   *service.ConfigService
}

func NewAdminHandler(as *service.ArticleService, ns *service.NoteService, cs *service.CategoryService, ts *service.TagService, cms *service.CommentService, ds *service.DanmakuService, vs *service.VisitorService, cfgs *service.ConfigService) *AdminHandler {
	return &AdminHandler{articleSvc: as, noteSvc: ns, categorySvc: cs, tagSvc: ts, commentSvc: cms, danmakuSvc: ds, visitorSvc: vs, configSvc: cfgs}
}

// -- Articles --

func (h *AdminHandler) ListArticles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "20"))

	// Support category filter
	if cat := c.Query("category"); cat != "" {
		articles, total, err := h.articleSvc.ListByCategory(cat, page, perPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total})
		return
	}

	// Support slug lookup (for post detail page)
	if slug := c.Query("slug"); slug != "" {
		a, err := h.articleSvc.GetBySlug(slug)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"article": a, "total": 1})
		return
	}

	// ?all=true for admin listing (includes drafts)
	if c.Query("all") == "true" {
		articles, total, err := h.articleSvc.ListAll(page, perPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total})
		return
	}

	articles, total, err := h.articleSvc.ListPublished(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total})
}

func (h *AdminHandler) SearchArticles(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"articles": []interface{}{}, "total": 0})
		return
	}
	articles, total, err := h.articleSvc.Search(q, 1, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"articles": articles, "total": total})
}

func (h *AdminHandler) GetArticle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	a, err := h.articleSvc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AdminHandler) CreateArticle(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		Category    string `json:"category"`
		Tags        string `json:"tags"`
		IsPublished bool   `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.articleSvc.Create(req.Title, req.Content, req.Category, req.Tags, req.IsPublished)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, a)
}

func (h *AdminHandler) UpdateArticle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Title       string `json:"title"`
		Content     string `json:"content"`
		Category    string `json:"category"`
		Tags        string `json:"tags"`
		IsPublished bool   `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	a, err := h.articleSvc.Update(uint(id), req.Title, req.Content, req.Category, req.Tags, req.IsPublished)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, a)
}

func (h *AdminHandler) DeleteArticle(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.articleSvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Notes --

func (h *AdminHandler) ListNotes(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))

	if c.Query("all") == "true" {
		notes, total, err := h.noteSvc.ListAll(page, perPage)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"notes": notes, "total": total})
		return
	}

	notes, total, err := h.noteSvc.ListPublished(page, perPage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"notes": notes, "total": total})
}

func (h *AdminHandler) GetNote(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	n, err := h.noteSvc.GetByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, n)
}

func (h *AdminHandler) CreateNote(c *gin.Context) {
	var req struct {
		Content     string `json:"content"`
		Images      string `json:"images"`
		IsPublished bool   `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.noteSvc.Create(req.Content, req.Images, req.IsPublished)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, n)
}

func (h *AdminHandler) UpdateNote(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req struct {
		Content     string `json:"content"`
		Images      string `json:"images"`
		IsPublished bool   `json:"is_published"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	n, err := h.noteSvc.Update(uint(id), req.Content, req.Images, req.IsPublished)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, n)
}

func (h *AdminHandler) DeleteNote(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.noteSvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Categories --

func (h *AdminHandler) ListCategories(c *gin.Context) {
	categories, err := h.categorySvc.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cat, err := h.categorySvc.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.categorySvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Tags --

func (h *AdminHandler) ListTags(c *gin.Context) {
	tags, err := h.tagSvc.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func (h *AdminHandler) CreateTag(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag, err := h.tagSvc.Create(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *AdminHandler) DeleteTag(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.tagSvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Upload --

// UploadsDir 上传目录，main 启动时按 web 根位置注入（相对服务器工作目录）
var UploadsDir = "web/static/uploads"

// 允许上传的图片扩展名（svg 可携带脚本，不放开）
var allowedUploadExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
}

const maxUploadSize = 10 << 20 // 10MB

func (h *AdminHandler) UploadFile(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请选择文件（不超过10MB）"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedUploadExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg/jpeg/png/gif/webp 图片"})
		return
	}

	// 检测真实文件类型，防止伪造扩展名
	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		return
	}
	buf := make([]byte, 512)
	n, _ := src.Read(buf)
	src.Close()
	if n > 0 && !strings.HasPrefix(http.DetectContentType(buf[:n]), "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件内容不是图片"})
		return
	}

	// Generate unique filename
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(UploadsDir, name)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}

// -- Stats --

func (h *AdminHandler) Stats(c *gin.Context) {
	articleTotal, _ := h.articleSvc.CountAll()
	noteTotal, _ := h.noteSvc.CountAll()
	categoryTotal, _ := h.categorySvc.CountAll()
	commentTotal, _ := h.commentSvc.CountAll()
	visitorTotal, _ := h.visitorSvc.CountAll()
	c.JSON(http.StatusOK, gin.H{
		"article_count":  articleTotal,
		"note_count":     noteTotal,
		"category_count": categoryTotal,
		"comment_count":  commentTotal,
		"visitor_count":  visitorTotal,
	})
}

// -- Comment Management --

func (h *AdminHandler) DeleteComment(c *gin.Context) {
	id := c.Param("id")
	if err := h.commentSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Danmaku Management --

func (h *AdminHandler) DeleteDanmaku(c *gin.Context) {
	id := c.Param("id")
	if err := h.danmakuSvc.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Visitor Management --

func (h *AdminHandler) ListVisitors(c *gin.Context) {
	visitors, err := h.visitorSvc.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitors": visitors})
}

func (h *AdminHandler) UpdateVisitor(c *gin.Context) {
	uuid := c.Param("uuid")
	var req struct {
		Nickname  string `json:"nickname"`
		Signature string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	v, err := h.visitorSvc.UpdateProfile(uuid, req.Nickname, req.Signature)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v})
}

func (h *AdminHandler) DeleteVisitor(c *gin.Context) {
	uuid := c.Param("uuid")
	if err := h.visitorSvc.Delete(uuid); err != nil {
		if errors.Is(err, service.ErrProtectedVisitor) {
			c.JSON(http.StatusForbidden, gin.H{"error": "不能删除管理员账号"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Blog Config --

func (h *AdminHandler) GetConfig(c *gin.Context) {
	cfg, err := h.configSvc.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "读取配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

func (h *AdminHandler) UpdateConfig(c *gin.Context) {
	var req struct {
		AuthorName   string `json:"author_name"`
		AuthorAvatar string `json:"author_avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	cfg, err := h.configSvc.UpdateAuthor(req.AuthorName, req.AuthorAvatar)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存配置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

package handler

import (
	"blog/internal/model"
	"blog/internal/repository"
	"blog/internal/service"
	"blog/internal/util"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// requireEnv 启动即失败：安全相关配置不允许静默使用默认值
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("环境变量 %s 未配置，拒绝启动", key)
	}
	return v
}

var (
	jwtSecret = []byte(requireEnv("JWT_SECRET"))
	adminUser = func() string {
		if u := os.Getenv("ADMIN_USER"); u != "" {
			return u
		}
		return "admin"
	}()
	adminPass = requireEnv("ADMIN_PASSWORD")
)

// 管理员登录时自动创建访客记录
var visitorRepoForLogin *repository.VisitorRepo

func InitLogin(vr *repository.VisitorRepo) { visitorRepoForLogin = vr }

func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入账号和密码"})
		return
	}
	// 常数时间比对，避免时序侧信道
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(adminUser)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(req.Password), []byte(adminPass)) == 1
	if !userOK || !passOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	// 自动创建/更新管理员的访客记录（管理员既是后台用户也是前台访客）
	if visitorRepoForLogin != nil {
		v, err := visitorRepoForLogin.FindByUUID("admin_" + req.Username)
		if err != nil {
			v = &model.Visitor{
				UUID:        "admin_" + req.Username,
				Username:    req.Username,
				Nickname:    req.Username,
				AvatarStyle: "lorelei",
			}
			visitorRepoForLogin.Register(v, req.Password)
		} else {
			visitorRepoForLogin.UpdatePassword(v.UUID, req.Password)
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": req.Username,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "user": req.Username})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return jwtSecret, nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
			c.Abort()
			return
		}
		c.Next()
	}
}

type AdminHandler struct {
	articleSvc   *service.ArticleService
	noteSvc      *service.NoteService
	categoryRepo *repository.CategoryRepo
	tagRepo      *repository.TagRepo
	commentRepo  *repository.CommentRepo
	danmakuRepo  *repository.DanmakuRepo
	visitorRepo  *repository.VisitorRepo
	configRepo   *repository.ConfigRepo
}

func NewAdminHandler(as *service.ArticleService, ns *service.NoteService, cr *repository.CategoryRepo, tr *repository.TagRepo, cmr *repository.CommentRepo, dr *repository.DanmakuRepo, vr *repository.VisitorRepo, cfg *repository.ConfigRepo) *AdminHandler {
	return &AdminHandler{articleSvc: as, noteSvc: ns, categoryRepo: cr, tagRepo: tr, commentRepo: cmr, danmakuRepo: dr, visitorRepo: vr, configRepo: cfg}
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
	categories, err := h.categoryRepo.ListAll()
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
	cat := &model.Category{Name: req.Name, Slug: util.Slugify(req.Name)}
	if err := h.categoryRepo.Create(cat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.categoryRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Tags --

func (h *AdminHandler) ListTags(c *gin.Context) {
	tags, err := h.articleSvc.GetAllTags()
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
	tag := &model.Tag{Name: req.Name, Slug: util.Slugify(req.Name)}
	if err := h.tagRepo.Create(tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, tag)
}

func (h *AdminHandler) DeleteTag(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := h.tagRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Upload --

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
	savePath := filepath.Join("web", "static", "uploads", name)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": "/uploads/" + name})
}

// -- Stats --

func (h *AdminHandler) Stats(c *gin.Context) {
	_, articleTotal, _ := h.articleSvc.ListAll(1, 1)
	_, noteTotal, _ := h.noteSvc.ListAll(1, 1)
	categories, _ := h.categoryRepo.ListAll()
	c.JSON(http.StatusOK, gin.H{
		"article_count":  articleTotal,
		"note_count":     noteTotal,
		"category_count": len(categories),
	})
}

// -- Comment Management --

func (h *AdminHandler) DeleteComment(c *gin.Context) {
	id := c.Param("id")
	if err := h.commentRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Danmaku Management --

func (h *AdminHandler) DeleteDanmaku(c *gin.Context) {
	id := c.Param("id")
	if err := h.danmakuRepo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Visitor Management --

func (h *AdminHandler) ListVisitors(c *gin.Context) {
	visitors, err := h.visitorRepo.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitors": visitors})
}

func (h *AdminHandler) UpdateVisitor(c *gin.Context) {
	uuid := c.Param("uuid")
	v, err := h.visitorRepo.FindByUUID(uuid)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "未找到"})
		return
	}
	var req struct {
		Nickname  string `json:"nickname"`
		Signature string `json:"signature"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	v.Nickname = req.Nickname
	v.Signature = req.Signature
	if err := h.visitorRepo.Update(v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"visitor": v})
}

func (h *AdminHandler) DeleteVisitor(c *gin.Context) {
	uuid := c.Param("uuid")
	if strings.HasPrefix(uuid, "admin_") {
		c.JSON(http.StatusForbidden, gin.H{"error": "不能删除管理员账号"})
		return
	}
	if err := h.visitorRepo.DeleteByUUID(uuid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// -- Blog Config --

func (h *AdminHandler) GetConfig(c *gin.Context) {
	cfg, _ := h.configRepo.Get()
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
	cfg, _ := h.configRepo.Get()
	cfg.AuthorName = req.AuthorName
	cfg.AuthorAvatar = req.AuthorAvatar
	h.configRepo.Update(cfg)
	c.JSON(http.StatusOK, gin.H{"config": cfg})
}

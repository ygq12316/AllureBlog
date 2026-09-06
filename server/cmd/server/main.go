package main

import (
	"blog/internal/database"
	"blog/internal/handler"
	"blog/internal/repository"
	"blog/internal/service"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// requireEnv 启动即失败：安全相关配置不允许静默使用默认值。
// 环境变量读取集中在 main,handler 等包 import 无副作用。
func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("环境变量 %s 未配置，拒绝启动", key)
	}
	return v
}

func main() {
	// 仓库根定位：生产/根目录启动 cwd 即根；开发 go run（cwd=server/）回退上级。
	// db 与 web 静态资源的相对路径都从根算起，保证两种启动方式行为一致
	root := "."
	if _, err := os.Stat(filepath.Join(root, "web")); err != nil {
		root = ".."
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = filepath.Join(root, "blog.db")
	}

	db, err := database.InitDB(dbPath)
	if err != nil {
		log.Fatal("数据库初始化失败:", err)
	}

	articleRepo := repository.NewArticleRepo(db)
	categoryRepo := repository.NewCategoryRepo(db)
	tagRepo := repository.NewTagRepo(db)
	noteRepo := repository.NewNoteRepo(db)
	visitorRepo := repository.NewVisitorRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	danmakuRepo := repository.NewDanmakuRepo(db)
	configRepo := repository.NewConfigRepo(db)

	// 业务层：全部领域都有 service，handler 不再直接触碰 repository
	articleSvc := service.NewArticleService(articleRepo, categoryRepo, tagRepo)
	noteSvc := service.NewNoteService(noteRepo)
	categorySvc := service.NewCategoryService(categoryRepo)
	tagSvc := service.NewTagService(tagRepo, articleRepo)
	commentSvc := service.NewCommentService(commentRepo, visitorRepo)
	danmakuSvc := service.NewDanmakuService(danmakuRepo, visitorRepo)
	visitorSvc := service.NewVisitorService(visitorRepo)
	configSvc := service.NewConfigService(configRepo)

	adminHandler := handler.NewAdminHandler(articleSvc, noteSvc, categorySvc, tagSvc, commentSvc, danmakuSvc, visitorSvc, configSvc)
	wsHub := handler.NewHub()

	adminUser := os.Getenv("ADMIN_USER")
	if adminUser == "" {
		adminUser = "admin"
	}
	// 统一账号体系：环境凭据仅作管理员账号的启动种子，登录一律走账号密码 + 角色鉴权
	auth := handler.NewAuth(handler.AuthConfig{
		Secret: requireEnv("JWT_SECRET"),
		User:   adminUser,
		Pass:   requireEnv("ADMIN_PASSWORD"),
	}, visitorSvc)
	auth.SeedAdmin()

	visitorHandler := handler.NewVisitorHandler(visitorSvc, commentSvc, danmakuSvc, wsHub, auth)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Visitor public API
	r.POST("/api/visitor", visitorHandler.RegisterAnonymous)
	r.POST("/api/visitor/register", visitorHandler.RegisterWithAccount)
	r.POST("/api/visitor/login", visitorHandler.LoginAccount)
	r.GET("/api/visitor/:uuid", visitorHandler.GetVisitor)

	// Public upload (for avatar)
	r.POST("/api/upload", adminHandler.UploadFile)

	// Public blog config
	r.GET("/api/config", adminHandler.GetConfig)

	// 笔墨精灵 agent 反向代理 — 浏览器同源连 /chat/ws 靠它转发（含 WebSocket）；
	// 生产由 Caddy 先行拦截 /chat/*，正常不会走到这里
	agentURL := os.Getenv("AGENT_URL")
	if agentURL == "" {
		agentURL = "http://localhost:8000"
	}
	r.Any("/chat/*path", handler.ChatProxy(agentURL))

	// API routes - GET public, POST/PUT/DELETE protected
	api := r.Group("/api")
	{
		api.GET("/articles", adminHandler.ListArticles)
		api.GET("/articles/search", adminHandler.SearchArticles)
		api.GET("/articles/:id", adminHandler.GetArticle)
		api.GET("/notes", adminHandler.ListNotes)
		api.GET("/notes/:id", adminHandler.GetNote)
		api.GET("/categories", adminHandler.ListCategories)
		api.GET("/tags", adminHandler.ListTags)
		api.GET("/stats", adminHandler.Stats)
		api.GET("/notes/:id/comments", visitorHandler.ListComments)
		api.POST("/notes/:id/comments", visitorHandler.CreateComment)
		api.GET("/notes/:id/ws", wsHub.HandleNoteWS)
		api.GET("/ws", wsHub.HandleGlobalWS)
		api.GET("/danmaku", visitorHandler.ListDanmaku)
		api.POST("/danmaku", visitorHandler.CreateDanmaku)
	}

	// Protected API (auth required)
	authGroup := r.Group("/api")
	authGroup.Use(auth.Middleware())
	{
		authGroup.POST("/articles", adminHandler.CreateArticle)
		authGroup.PUT("/articles/:id", adminHandler.UpdateArticle)
		authGroup.DELETE("/articles/:id", adminHandler.DeleteArticle)
		authGroup.POST("/notes", adminHandler.CreateNote)
		authGroup.PUT("/notes/:id", adminHandler.UpdateNote)
		authGroup.DELETE("/notes/:id", adminHandler.DeleteNote)
		authGroup.POST("/categories", adminHandler.CreateCategory)
		authGroup.DELETE("/categories/:id", adminHandler.DeleteCategory)
		authGroup.POST("/tags", adminHandler.CreateTag)
		authGroup.DELETE("/tags/:id", adminHandler.DeleteTag)
		authGroup.DELETE("/admin/comments/:id", adminHandler.DeleteComment)
		authGroup.DELETE("/admin/danmaku/:id", adminHandler.DeleteDanmaku)
		authGroup.GET("/admin/visitors", adminHandler.ListVisitors)
		authGroup.PUT("/admin/visitors/:uuid", adminHandler.UpdateVisitor)
		authGroup.DELETE("/admin/visitors/:uuid", adminHandler.DeleteVisitor)
		authGroup.GET("/admin/config", adminHandler.GetConfig)
		authGroup.PUT("/admin/config", adminHandler.UpdateConfig)
	}

	// Serve uploaded files
	webDir := filepath.Join(root, "web")
	handler.UploadsDir = filepath.Join(webDir, "static", "uploads")
	if err := os.MkdirAll(handler.UploadsDir, 0755); err != nil {
		log.Fatal("创建上传目录失败:", err)
	}
	// 禁目录列表、缺失文件 404（而非落入 SPA 回退返回 index.html）
	r.GET("/uploads/*filepath", func(c *gin.Context) {
		p := c.Param("filepath")
		if p == "/" || strings.Contains(p, "..") {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(filepath.Join(handler.UploadsDir, filepath.FromSlash(p)))
	})

	// Serve Vue SPA static files
	distDir := filepath.Join(webDir, "dist")
	r.Static("/assets", filepath.Join(distDir, "assets"))
	r.StaticFile("/TagCloud.min.js", filepath.Join(distDir, "TagCloud.min.js"))
	r.StaticFile("/favicon.ico", filepath.Join(distDir, "favicon.ico"))

	// SPA fallback: serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		c.File(filepath.Join(distDir, "index.html"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("笔墨 · Ink & Code 已启动 → http://localhost:%s\n", port)
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("启动失败:", err)
	}
}

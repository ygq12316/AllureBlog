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
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "blog.db"
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

	articleSvc := service.NewArticleService(articleRepo, categoryRepo, tagRepo)
	noteSvc := service.NewNoteService(noteRepo)
	visitorSvc := service.NewVisitorService(visitorRepo, commentRepo, danmakuRepo)

	adminHandler := handler.NewAdminHandler(articleSvc, noteSvc, categoryRepo, tagRepo, commentRepo, danmakuRepo, visitorRepo, configRepo)
	visitorHandler := handler.NewVisitorHandler(visitorSvc)

	// 管理员登录时自动创建访客记录
	handler.InitLogin(visitorRepo)

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Public API (no auth needed)
	r.POST("/api/login", handler.Login)

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
		api.GET("/notes/:id/ws", handler.HandleNoteWS)
		api.GET("/danmaku", visitorHandler.ListDanmaku)
		api.POST("/danmaku", visitorHandler.CreateDanmaku)
	}

	// Protected API (auth required)
	auth := r.Group("/api")
	auth.Use(handler.AuthMiddleware())
	{
		auth.POST("/articles", adminHandler.CreateArticle)
		auth.PUT("/articles/:id", adminHandler.UpdateArticle)
		auth.DELETE("/articles/:id", adminHandler.DeleteArticle)
		auth.POST("/notes", adminHandler.CreateNote)
		auth.PUT("/notes/:id", adminHandler.UpdateNote)
		auth.DELETE("/notes/:id", adminHandler.DeleteNote)
		auth.POST("/categories", adminHandler.CreateCategory)
		auth.DELETE("/categories/:id", adminHandler.DeleteCategory)
		auth.POST("/tags", adminHandler.CreateTag)
		auth.DELETE("/tags/:id", adminHandler.DeleteTag)
		auth.DELETE("/admin/comments/:id", adminHandler.DeleteComment)
		auth.DELETE("/admin/danmaku/:id", adminHandler.DeleteDanmaku)
		auth.GET("/admin/visitors", adminHandler.ListVisitors)
		auth.PUT("/admin/visitors/:uuid", adminHandler.UpdateVisitor)
		auth.DELETE("/admin/visitors/:uuid", adminHandler.DeleteVisitor)
		auth.GET("/admin/config", adminHandler.GetConfig)
		auth.PUT("/admin/config", adminHandler.UpdateConfig)
	}

	// Serve uploaded files
	if err := os.MkdirAll("web/static/uploads", 0755); err != nil {
		log.Fatal("创建上传目录失败:", err)
	}
	r.Static("/uploads", "web/static/uploads")

	// Serve Vue SPA static files
	distDir := "web/admin/dist"
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

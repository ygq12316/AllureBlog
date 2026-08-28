package handler

import (
	"blog/internal/service"
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig 鉴权配置：由 main 组装时注入。
// handler 包不再有包级 init 副作用（环境变量读取与缺失即 Fatal 均在 main 完成）。
type AuthConfig struct {
	Secret string // JWT 签名密钥
	User   string // 管理员账号
	Pass   string // 管理员密码
}

// Auth 管理员登录签发 JWT 与写操作鉴权。
// interface 只有两个方法（Login / Middleware），常数时间比对、bcrypt 同步、
// 7 天有效期全部藏在 implementation 里。
type Auth struct {
	cfg        AuthConfig
	visitorSvc *service.VisitorService // 可为 nil：仅用于登录时同步管理员访客记录
}

func NewAuth(cfg AuthConfig, vs *service.VisitorService) *Auth {
	return &Auth{cfg: cfg, visitorSvc: vs}
}

// Login POST /api/login
func (a *Auth) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请输入账号和密码"})
		return
	}
	// 常数时间比对，避免时序侧信道
	userOK := subtle.ConstantTimeCompare([]byte(req.Username), []byte(a.cfg.User)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(req.Password), []byte(a.cfg.Pass)) == 1
	if !userOK || !passOK {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "账号或密码错误"})
		return
	}

	// 自动创建/更新管理员的访客记录（管理员既是后台用户也是前台访客）
	if a.visitorSvc != nil {
		a.visitorSvc.SyncAdmin(req.Username, req.Password)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": req.Username,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString([]byte(a.cfg.Secret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "生成令牌失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenStr, "user": req.Username})
}

// Middleware 校验 Bearer token，供所有写操作路由使用
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "请先登录"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) { return []byte(a.cfg.Secret), nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "登录已过期"})
			c.Abort()
			return
		}
		c.Next()
	}
}

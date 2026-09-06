package handler

import (
	"blog/internal/model"
	"blog/internal/service"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig 鉴权配置：由 main 组装时注入。
// 环境凭据仅作管理员账号的启动种子（SyncAdmin），登录统一走账号密码 + 角色校验。
type AuthConfig struct {
	Secret string // JWT 签名密钥
	User   string // 管理员账号名（种子）
	Pass   string // 管理员密码（种子）
}

// Auth 统一账号体系的 JWT 签发与管理端鉴权。
// 角色存于 visitors.role（user/admin），令牌携带 role，Middleware 只放行 admin。
type Auth struct {
	cfg        AuthConfig
	visitorSvc *service.VisitorService
}

func NewAuth(cfg AuthConfig, vs *service.VisitorService) *Auth {
	return &Auth{cfg: cfg, visitorSvc: vs}
}

// SeedAdmin 启动时以环境凭据同步管理员账号（存在则刷新密码与角色）。
// 环境变量从「登录凭据」退化为「引导种子」：此后登录一律走账号体系。
func (a *Auth) SeedAdmin() {
	if a.visitorSvc != nil {
		a.visitorSvc.SyncAdmin(a.cfg.User, a.cfg.Pass)
		log.Printf("管理员账号 %s 已同步（来源：环境变量种子）", a.cfg.User)
	}
}

// IssueToken 为登录成功的用户签发 7 天期 JWT，携带角色供管理端鉴权
func (a *Auth) IssueToken(v *model.Visitor) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user": v.Username,
		"uuid": v.UUID,
		"role": v.Role,
		"exp":  time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return token.SignedString([]byte(a.cfg.Secret))
}

// Middleware 校验 Bearer token 且角色必须为 admin，供后台写操作路由使用。
// 普通用户（role=user）持有合法令牌也进不了管理端。
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
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["role"] != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可执行此操作"})
			c.Abort()
			return
		}
		c.Next()
	}
}

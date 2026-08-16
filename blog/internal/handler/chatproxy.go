package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

// ChatProxy 返回 /chat/* 到笔墨精灵 agent 的反向代理（含 WebSocket）
//
// 本机直跑（Go 直服 dist 的 :8080）时浏览器同源连 /chat/ws 靠它转发；
// 生产环境由 Caddy 先行拦截 /chat/*，正常不会走到这里。
func ChatProxy(agentURL string) gin.HandlerFunc {
	target, err := url.Parse(agentURL)
	if err != nil {
		// 地址配置无效直接 503，日志由 gin 记录
		return func(c *gin.Context) {
			c.String(http.StatusServiceUnavailable, "AGENT_URL 配置无效")
		}
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	return func(c *gin.Context) {
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

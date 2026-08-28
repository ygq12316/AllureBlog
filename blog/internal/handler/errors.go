package handler

import (
	"blog/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// respondError 统一错误响应：业务校验失败 → 400（Message 原样给用户），其余 → 500。
// 特定状态码（403/404/409）由各 handler 用 errors.Is 判领域哨兵后自行返回
func respondError(c *gin.Context, err error) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		c.JSON(http.StatusBadRequest, gin.H{"error": ve.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

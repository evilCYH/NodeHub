package middleware

import (
	"net/http"

	"github.com/evilCYH/NodeHub/internal/server/resp"
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Warnf("Panic recovered: %v", recovered)
		resp.Error(c, http.StatusInternalServerError, "An unexpected error occurred")
		c.Abort()
	})
}

package middleware

import (
	"github.com/evilCYH/NodeHub/internal/utils/log"
	"github.com/gin-gonic/gin"
)

// 日志中间件
func Logging() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {

		log.Debugf("%s %d %s %s %s",
			param.Method,
			param.StatusCode,
			param.ClientIP,
			param.Latency,
			param.Path,
		)
		return ""
	})
}

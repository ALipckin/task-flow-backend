package middleware

import (
	"TaskRestApiService/logger"
	"github.com/gin-gonic/gin"
	"time"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		remoteAddr := c.ClientIP()
		userAgent := c.Request.UserAgent()

		logger.LogRequest(
			method,
			path,
			remoteAddr,
			userAgent,
			statusCode,
			duration,
		)
	}
}

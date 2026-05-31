package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		slog.Info("request",
			"status", c.Writer.Status(),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"ip", c.ClientIP(),
			"duration_ms", duration.Milliseconds(),
		)
	}
}

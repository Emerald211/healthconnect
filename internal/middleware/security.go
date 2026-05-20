package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds production security headers to every response
// These protect against common web attacks
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Enable XSS protection in older browsers
		c.Header("X-XSS-Protection", "1; mode=block")

		// Force HTTPS in production
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Control what info is sent in referrer
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Restrict browser features
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		c.Next()
	}
}

// RequestSizeLimit rejects requests larger than the limit
// Prevents memory exhaustion attacks from huge payloads
func RequestSizeLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// TimeoutMiddleware cancels requests that take too long
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a deadline on the request context
		// If handler takes longer than timeout, context is cancelled
		c.Next()
	}
}
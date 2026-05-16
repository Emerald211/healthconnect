package middleware

import (
	"net/http"
	"strings"

	"github.com/Emerald211/healthconnect/internal/config"
	"github.com/Emerald211/healthconnect/pkg/jwt"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing_token",
				"message": "authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_token_format",
				"message": "authorization header format must be: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := jwt.VerifyToken(tokenString, cfg.JWTSecret)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid_token",
				"message": "token is invalid or expired",
			})

			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()

	}

}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")

		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
				"message": "not authenticated",
			})
			c.Abort()
			return
		}

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "forbidden",
				"message": "you do not have permission to access this resource",
			})

			c.Abort()
			return
		}

		c.Next()

	}
}

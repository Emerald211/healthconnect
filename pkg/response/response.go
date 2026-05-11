package response

import (
	"fmt"
	"net/http"

	"github.com/Emerald211/healthconnect/internal/domain"
	"github.com/gin-gonic/gin"
)

func Success(c *gin.Context, statusCode int, data any) {
	c.JSON(statusCode, gin.H{
		"success": true,
		"data":    data,
	})
}

func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*domain.AppError); ok {
		c.JSON(appErr.Status, gin.H{
			"success": false,
			"error":   appErr.Code,
			"message": appErr.Message,
		})

		return
	}

	fmt.Printf("ERROR: %v\n", err)


	c.JSON(http.StatusInternalServerError, gin.H{
		"success": false,
		"error": "internal_error",
		"message": "something went wrong",
	})
}

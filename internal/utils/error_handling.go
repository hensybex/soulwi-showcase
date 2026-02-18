// internal/utils/error_handling.go

package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// RespondWithError sends a JSON error response
func RespondWithError(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{"error": message})
}

// RespondWithSuccess sends a JSON success response
func RespondWithSuccess(c *gin.Context, code int, payload interface{}) {
	c.JSON(code, gin.H{"message": "Success", "data": payload})
}

// GetUintParam retrieves a uint parameter from the URL path
func GetUintParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	id, err := strconv.ParseUint(idStr, 10, 32)
	return uint(id), err
}

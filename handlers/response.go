package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// APIResponse defines the uniform structure for all JSON responses.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// SendSuccess formats and sends a HTTP 200 OK JSON success response.
func SendSuccess(c *gin.Context, message string, data interface{}) {
	// If data is nil, we pass an empty array to maintain JSON array consistency
	if data == nil {
		data = []interface{}{}
	}
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError formats and sends a HTTP JSON error response with the provided status code.
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

// GetPaginationParams extracts and validates page and limit parameters from URL query, returning limit and offset.
func GetPaginationParams(c *gin.Context) (int, int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 50
	}

	offset := (page - 1) * limit
	return limit, offset
}


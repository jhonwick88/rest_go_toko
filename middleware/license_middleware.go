package middleware

import (
	"net/http"
	"rest_go_toko/services"

	"github.com/gin-gonic/gin"
)

// RequireLicense is a Gin middleware that ensures a valid license exists before proceeding.
func RequireLicense() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Allow CORS preflight requests to pass through
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		if !services.HasValidLicense() {
			c.AbortWithStatusJSON(http.StatusPaymentRequired, gin.H{
				"error":   "License Not Found or Expired",
				"message": "Please activate your TokoPintar POS Server License to access the APIs.",
			})
			return
		}

		c.Next()
	}
}

package handlers

import (
	"net/http"

	"rest_go_toko/services"

	"github.com/gin-gonic/gin"
)

// ActivateLicense handles the POST request to activate the license on this backend.
func ActivateLicense(c *gin.Context) {
	var req struct {
		LicenseKey string `json:"license_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "license_key is required"})
		return
	}

	err := services.ActivateLicense(req.LicenseKey)
	if err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"error":   "Activation failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "License activated successfully",
	})
}

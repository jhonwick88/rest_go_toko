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

	claims, _ := services.GetLicenseClaims()
	var features map[string]interface{}
	if claims != nil {
		features = claims.Features
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "License activated successfully",
		"features": features,
	})
}


// GetLicenseStatus returns the current license status and features
func GetLicenseStatus(c *gin.Context) {
	isValid := services.HasValidLicense()
	if !isValid {
		c.JSON(http.StatusOK, gin.H{
			"is_activated": false,
		})
		return
	}

	claims, err := services.GetLicenseClaims()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"is_activated": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"is_activated": true,
		"features":     claims.Features,
		"license_id":   claims.LicenseID,
		"customer_id":  claims.CustomerID,
	})
}

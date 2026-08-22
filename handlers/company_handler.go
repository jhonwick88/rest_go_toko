package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetCompany returns the global company settings
func GetCompany(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `SELECT 
			COALESCE(COMPANYNAME, ''), 
			COALESCE(ADDRESSLINE1, ''), 
			COALESCE(PHONENO, ''), 
			COALESCE(RECEIPT_HEADER, ''), 
			COALESCE(RECEIPT_FOOTER, ''), 
			COALESCE(ADMIN_PIN, '1234'), 
			COALESCE(ENABLE_TAX, 0), 
			COALESCE(TAX_PERCENTAGE, 0.0), 
			COALESCE(ENABLE_SERVICE_CHARGE, 0), 
			COALESCE(SERVICE_CHARGE_PERCENTAGE, 0.0) 
		FROM COMPANY LIMIT 1`

		var company models.Company
		var enableTaxInt, enableServiceInt int
		err := db.QueryRow(query).Scan(
			&company.CompanyName,
			&company.AddressLine1,
			&company.PhoneNo,
			&company.ReceiptHeader,
			&company.ReceiptFooter,
			&company.AdminPin,
			&enableTaxInt,
			&company.TaxPercentage,
			&enableServiceInt,
			&company.ServiceChargePercentage,
		)
		company.EnableTax = enableTaxInt == 1
		company.EnableServiceCharge = enableServiceInt == 1

		if err != nil {
			if err == sql.ErrNoRows {
				// Return default if empty
				company = models.Company{
					CompanyName: "Toko Pintar",
					AddressLine1: "Alamat Toko",
					PhoneNo: "081234567890",
					AdminPin: "1234",
				}
				SendSuccess(c, "Data default perusahaan (belum ada data di DB)", company)
				return
			}
			log.Printf("[Error] Query GetCompany failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data perusahaan")
			return
		}

		SendSuccess(c, "Data perusahaan ditemukan", company)
	}
}

// UpdateCompany updates the global company settings
func UpdateCompany(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.Company
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		// Ensure a row exists first
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM COMPANY").Scan(&count)
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal memverifikasi data perusahaan")
			return
		}

		enableTaxInt := 0
		if input.EnableTax {
			enableTaxInt = 1
		}
		enableServiceInt := 0
		if input.EnableServiceCharge {
			enableServiceInt = 1
		}

		if count == 0 {
			// Insert
			query := `INSERT INTO COMPANY 
			(COMPANYNAME, ADDRESSLINE1, PHONENO, RECEIPT_HEADER, RECEIPT_FOOTER, ADMIN_PIN, ENABLE_TAX, TAX_PERCENTAGE, ENABLE_SERVICE_CHARGE, SERVICE_CHARGE_PERCENTAGE) 
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
			
			_, err = db.Exec(query, input.CompanyName, input.AddressLine1, input.PhoneNo, input.ReceiptHeader, input.ReceiptFooter, input.AdminPin, enableTaxInt, input.TaxPercentage, enableServiceInt, input.ServiceChargePercentage)
		} else {
			// Update
			query := `UPDATE COMPANY SET 
			COMPANYNAME = ?, 
			ADDRESSLINE1 = ?, 
			PHONENO = ?, 
			RECEIPT_HEADER = ?, 
			RECEIPT_FOOTER = ?, 
			ADMIN_PIN = ?, 
			ENABLE_TAX = ?, 
			TAX_PERCENTAGE = ?, 
			ENABLE_SERVICE_CHARGE = ?, 
			SERVICE_CHARGE_PERCENTAGE = ?`
			
			_, err = db.Exec(query, input.CompanyName, input.AddressLine1, input.PhoneNo, input.ReceiptHeader, input.ReceiptFooter, input.AdminPin, enableTaxInt, input.TaxPercentage, enableServiceInt, input.ServiceChargePercentage)
		}

		if err != nil {
			log.Printf("[Error] UpdateCompany failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan data perusahaan")
			return
		}

		SendSuccess(c, "Data perusahaan berhasil disimpan", input)
	}
}

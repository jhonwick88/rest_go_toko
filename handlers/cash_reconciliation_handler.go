package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCashReconciliations returns all cash reconciliations
func GetCashReconciliations(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)

		query := "SELECT ID, DATE, CASHIER_NAME, SYSTEM_REVENUE, ACTUAL_DRAWER_CASH, DIFFERENCE, ACCURACY_RATE, NOTES FROM CASH_RECONCILIATIONS ORDER BY DATE DESC LIMIT ? OFFSET ?"
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			log.Printf("[Error] Query GetCashReconciliations failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data tutup kasir")
			return
		}
		defer rows.Close()

		var recons []models.CashReconciliation
		for rows.Next() {
			var r models.CashReconciliation
			var date, cashierName, notes sql.NullString
			var accRate sql.NullFloat64

			if err := rows.Scan(&r.ID, &date, &cashierName, &r.ExpectedAmount, &r.ActualAmount, &r.Difference, &accRate, &notes); err != nil {
				SendError(c, http.StatusInternalServerError, "Gagal membaca data tutup kasir")
				return
			}
			r.Date = date.String
			r.Cashier = cashierName.String
			r.AccuracyRate = accRate.Float64
			r.Note = notes.String
			recons = append(recons, r)
		}

		if recons == nil {
			recons = []models.CashReconciliation{}
		}
		SendSuccess(c, "Data ditemukan", recons)
	}
}

// CreateCashReconciliation creates a new cash reconciliation record
func CreateCashReconciliation(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CashReconciliation
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		id := uuid.New().String()
		query := "INSERT INTO CASH_RECONCILIATIONS (ID, DATE, CASHIER_NAME, SYSTEM_REVENUE, ACTUAL_DRAWER_CASH, DIFFERENCE, ACCURACY_RATE, NOTES) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
		
		_, err := db.Exec(query, id, input.Date, input.Cashier, input.ExpectedAmount, input.ActualAmount, input.Difference, input.AccuracyRate, input.Note)
		if err != nil {
			log.Printf("[Error] CreateCashReconciliation failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan laporan tutup kasir")
			return
		}

		input.ID = id
		SendSuccess(c, "Laporan tutup kasir berhasil disimpan", input)
	}
}

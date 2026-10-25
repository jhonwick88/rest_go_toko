package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetCashMovements returns cash movement records (cash out/in)
func GetCashMovements(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		startDate := c.Query("start_date")
		endDate := c.Query("end_date")

		var conditions []string
		var args []interface{}

		if startDate != "" {
			startDay := startDate
			if len(startDay) >= 10 {
				startDay = startDay[:10]
			}
			conditions = append(conditions, "substr(DATE, 1, 10) >= ?")
			args = append(args, startDay)
		}
		if endDate != "" {
			endDay := endDate
			if len(endDay) >= 10 {
				endDay = endDay[:10]
			}
			conditions = append(conditions, "substr(DATE, 1, 10) <= ?")
			args = append(args, endDay)
		}

		whereClause := ""
		if len(conditions) > 0 {
			whereClause = "WHERE " + strings.Join(conditions, " AND ") + " "
		}

		limit, offset := GetPaginationParams(c)
		query := fmt.Sprintf("SELECT ID, DATE, CASHIER_NAME, TYPE, CATEGORY, AMOUNT, NOTES FROM CASH_MOVEMENTS %sORDER BY DATE DESC LIMIT ? OFFSET ?", whereClause)
		args = append(args, limit, offset)

		rows, err := db.Query(query, args...)
		if err != nil {
			log.Printf("[Error] Query GetCashMovements failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data pengeluaran kas")
			return
		}
		defer rows.Close()

		var movements []models.CashMovement
		for rows.Next() {
			var m models.CashMovement
			var date, cashierName, movType, category, notes sql.NullString
			var amount sql.NullFloat64

			if err := rows.Scan(&m.ID, &date, &cashierName, &movType, &category, &amount, &notes); err != nil {
				log.Printf("[Error] Scan GetCashMovements failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data pengeluaran kas")
				return
			}
			m.Date = date.String
			m.CashierName = cashierName.String
			m.Type = movType.String
			m.Category = category.String
			m.Amount = amount.Float64
			m.Notes = notes.String

			movements = append(movements, m)
		}

		if movements == nil {
			movements = []models.CashMovement{}
		}

		SendSuccess(c, "Data kas keluar berhasil diambil", movements)
	}
}

// CreateCashMovement records a new cash out or cash in transaction
func CreateCashMovement(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.CashMovement
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		if input.Amount <= 0 {
			SendError(c, http.StatusBadRequest, "Nominal kas keluar harus lebih dari 0")
			return
		}

		if input.Type == "" {
			input.Type = "OUT"
		}
		if input.Category == "" {
			input.Category = "Operasional"
		}
		if input.Date == "" {
			input.Date = time.Now().Format("2006-01-02 15:04:05")
		}

		id := uuid.New().String()
		query := "INSERT INTO CASH_MOVEMENTS (ID, DATE, CASHIER_NAME, TYPE, CATEGORY, AMOUNT, NOTES) VALUES (?, ?, ?, ?, ?, ?, ?)"

		_, err := db.Exec(query, id, input.Date, input.CashierName, input.Type, input.Category, input.Amount, input.Notes)
		if err != nil {
			log.Printf("[Error] CreateCashMovement failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mencatat kas keluar")
			return
		}

		input.ID = id
		SendSuccess(c, "Pengeluaran kas berhasil dicatat", input)
	}
}

// DeleteCashMovement deletes a cash movement record
func DeleteCashMovement(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			SendError(c, http.StatusBadRequest, "ID tidak valid")
			return
		}

		result, err := db.Exec("DELETE FROM CASH_MOVEMENTS WHERE ID = ?", id)
		if err != nil {
			log.Printf("[Error] DeleteCashMovement failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menghapus catatan kas keluar")
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			SendError(c, http.StatusNotFound, "Data kas keluar tidak ditemukan")
			return
		}

		SendSuccess(c, "Data kas keluar berhasil dihapus", nil)
	}
}

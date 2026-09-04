package handlers

import (
	"database/sql"
	"net/http"
	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetStockLedger returns the stock ledger for a specific item.
func GetStockLedger(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemNo := c.Param("itemno")
		limit, offset := GetPaginationParams(c)

		rows, err := db.Query("SELECT ID, ITEMNO, TX_TYPE, QTY, TIMESTAMP, NOTES, REFERENCE_ID FROM STOCK_LEDGER WHERE ITEMNO = ? ORDER BY TIMESTAMP DESC LIMIT ? OFFSET ?", itemNo, limit, offset)
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data kartu stok: "+err.Error())
			return
		}
		defer rows.Close()

		var ledgers []models.StockLedger
		for rows.Next() {
			var sl models.StockLedger
			var notes, refID sql.NullString
			if err := rows.Scan(&sl.ID, &sl.ItemNo, &sl.TxType, &sl.Qty, &sl.Timestamp, &notes, &refID); err != nil {
				SendError(c, http.StatusInternalServerError, "Gagal membaca data kartu stok")
				return
			}
			sl.Notes = notes.String
			sl.ReferenceID = refID.String
			ledgers = append(ledgers, sl)
		}
		SendSuccess(c, "Data kartu stok berhasil diambil", ledgers)
	}
}

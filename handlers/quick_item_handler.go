package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetQuickItems returns all quick items
func GetQuickItems(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := `SELECT q.ID, q.ITEM_NO, q.ITEM_NAME, q.ICON_NAME, q.COLOR_HEX, q.DISPLAY_ORDER, q.IS_ACTIVE, COALESCE(i.DEF_UNITPRICE1, 0)
			FROM QUICK_ITEMS q
			LEFT JOIN ITEM i ON q.ITEM_NO = i.ITEMNO
			ORDER BY q.DISPLAY_ORDER ASC`
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("[Error] Query GetQuickItems failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data quick items")
			return
		}
		defer rows.Close()

		var items []models.QuickItem
		for rows.Next() {
			var q models.QuickItem
			var icon, color sql.NullString
			var price sql.NullFloat64
			if err := rows.Scan(&q.ID, &q.ItemNo, &q.ItemName, &icon, &color, &q.DisplayOrder, &q.IsActive, &price); err != nil {
				SendError(c, http.StatusInternalServerError, "Gagal membaca data quick items")
				return
			}
			q.IconName = icon.String
			q.ColorHex = color.String
			q.Price = price.Float64
			items = append(items, q)
		}

		if items == nil {
			items = []models.QuickItem{}
		}
		SendSuccess(c, "Data ditemukan", items)
	}
}

// CreateQuickItem creates a new quick item
func CreateQuickItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.QuickItem
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		if input.ID == "" {
			input.ID = "QI-" + input.ItemNo
		}

		query := "INSERT INTO QUICK_ITEMS (ID, ITEM_NO, ITEM_NAME, ICON_NAME, COLOR_HEX, DISPLAY_ORDER, IS_ACTIVE) VALUES (?, ?, ?, ?, ?, ?, ?)"
		_, err := db.Exec(query, input.ID, input.ItemNo, input.ItemName, input.IconName, input.ColorHex, input.DisplayOrder, input.IsActive)
		if err != nil {
			log.Printf("[Error] CreateQuickItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menambahkan quick item baru")
			return
		}

		SendSuccess(c, "Quick item berhasil ditambahkan", input)
	}
}

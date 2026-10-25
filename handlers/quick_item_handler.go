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

// CreateQuickItem creates or upserts a new quick item
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

		query := `INSERT INTO QUICK_ITEMS (ID, ITEM_NO, ITEM_NAME, ICON_NAME, COLOR_HEX, DISPLAY_ORDER, IS_ACTIVE) 
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(ID) DO UPDATE SET
				ITEM_NO = excluded.ITEM_NO,
				ITEM_NAME = excluded.ITEM_NAME,
				ICON_NAME = excluded.ICON_NAME,
				COLOR_HEX = excluded.COLOR_HEX,
				DISPLAY_ORDER = excluded.DISPLAY_ORDER,
				IS_ACTIVE = excluded.IS_ACTIVE`
		_, err := db.Exec(query, input.ID, input.ItemNo, input.ItemName, input.IconName, input.ColorHex, input.DisplayOrder, input.IsActive)
		if err != nil {
			log.Printf("[Error] CreateQuickItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan quick item: "+err.Error())
			return
		}

		SendSuccess(c, "Quick item berhasil disimpan", input)
	}
}

// UpdateQuickItem updates an existing quick item by ID
func UpdateQuickItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			SendError(c, http.StatusBadRequest, "Parameter id wajib diisi")
			return
		}

		var input models.QuickItem
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		query := `UPDATE QUICK_ITEMS SET ITEM_NO = ?, ITEM_NAME = ?, ICON_NAME = ?, COLOR_HEX = ?, DISPLAY_ORDER = ?, IS_ACTIVE = ? WHERE ID = ?`
		res, err := db.Exec(query, input.ItemNo, input.ItemName, input.IconName, input.ColorHex, input.DisplayOrder, input.IsActive, id)
		if err != nil {
			log.Printf("[Error] UpdateQuickItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memperbarui quick item: "+err.Error())
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			// If not found by ID, try upserting with ID
			input.ID = id
			upsertQuery := `INSERT INTO QUICK_ITEMS (ID, ITEM_NO, ITEM_NAME, ICON_NAME, COLOR_HEX, DISPLAY_ORDER, IS_ACTIVE) VALUES (?, ?, ?, ?, ?, ?, ?)`
			_, _ = db.Exec(upsertQuery, id, input.ItemNo, input.ItemName, input.IconName, input.ColorHex, input.DisplayOrder, input.IsActive)
		}

		input.ID = id
		SendSuccess(c, "Quick item berhasil diperbarui", input)
	}
}

// DeleteQuickItem deletes a quick item by ID or ITEM_NO
func DeleteQuickItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			SendError(c, http.StatusBadRequest, "Parameter id wajib diisi")
			return
		}

		res, err := db.Exec("DELETE FROM QUICK_ITEMS WHERE ID = ? OR ITEM_NO = ?", id, id)
		if err != nil {
			log.Printf("[Error] DeleteQuickItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menghapus quick item: "+err.Error())
			return
		}

		rowsAffected, _ := res.RowsAffected()
		log.Printf("[Info] DeleteQuickItem '%s' executed, rows affected: %d", id, rowsAffected)

		SendSuccess(c, "Quick item berhasil dihapus", gin.H{"id": id, "rows_affected": rowsAffected})
	}
}

// SaveQuickItemsBatch replaces or resequences a batch of quick items
func SaveQuickItemsBatch(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var items []models.QuickItem
		if err := c.ShouldBindJSON(&items); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		tx, err := db.Begin()
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal memulai transaksi database")
			return
		}

		for _, item := range items {
			if item.ID == "" {
				item.ID = "QI-" + item.ItemNo
			}
			query := `INSERT INTO QUICK_ITEMS (ID, ITEM_NO, ITEM_NAME, ICON_NAME, COLOR_HEX, DISPLAY_ORDER, IS_ACTIVE)
				VALUES (?, ?, ?, ?, ?, ?, ?)
				ON CONFLICT(ID) DO UPDATE SET
					ITEM_NO = excluded.ITEM_NO,
					ITEM_NAME = excluded.ITEM_NAME,
					ICON_NAME = excluded.ICON_NAME,
					COLOR_HEX = excluded.COLOR_HEX,
					DISPLAY_ORDER = excluded.DISPLAY_ORDER,
					IS_ACTIVE = excluded.IS_ACTIVE`
			_, err := tx.Exec(query, item.ID, item.ItemNo, item.ItemName, item.IconName, item.ColorHex, item.DisplayOrder, item.IsActive)
			if err != nil {
				tx.Rollback()
				log.Printf("[Error] SaveQuickItemsBatch failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal menyimpan batch quick items: "+err.Error())
				return
			}
		}

		if err := tx.Commit(); err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan transaksi batch")
			return
		}

		SendSuccess(c, "Batch quick items berhasil disimpan", gin.H{"count": len(items)})
	}
}

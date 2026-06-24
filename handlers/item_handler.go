package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetItems returns a Gin handler to fetch all items from database.
func GetItems(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)

		query := "SELECT FIRST ? SKIP ? ITEMNO, ITEMUPC, ITEMNAME, CATEGORYID, DEF_UNITPRICE1 FROM ITEM ORDER BY ITEMNO ASC"
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			log.Printf("[Error] Query GetItems failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengakses database untuk mengambil produk: "+err.Error())
			return
		}
		defer rows.Close()

		var items []models.Item
		for rows.Next() {
			var (
				itemNo     sql.NullString
				itemUPC    sql.NullString
				itemName   sql.NullString
				catID      sql.NullInt64
				price      sql.NullFloat64
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &catID, &price); err != nil {
				log.Printf("[Error] Scan item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data produk")
				return
			}
			items = append(items, models.Item{
				ItemNo:     strings.TrimSpace(itemNo.String),
				ItemUPC:    strings.TrimSpace(itemUPC.String),
				ItemName:   strings.TrimSpace(itemName.String),
				CategoryID: int(catID.Int64),
				Price:      price.Float64,
			})
		}

		if err = rows.Err(); err != nil {
			log.Printf("[Error] Rows error in GetItems: %v", err)
			SendError(c, http.StatusInternalServerError, "Koneksi database bermasalah saat membaca produk")
			return
		}

		if items == nil {
			items = []models.Item{}
		}

		SendSuccess(c, "Data ditemukan", items)
	}
}

// GetItemByNo returns a Gin handler to fetch a single item by its ITEMNO.
func GetItemByNo(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemNoParam := c.Param("itemno")
		if itemNoParam == "" {
			SendError(c, http.StatusBadRequest, "Parameter itemno wajib diisi")
			return
		}

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, CATEGORYID, DEF_UNITPRICE1 FROM ITEM WHERE ITEMNO = ?"
		row := db.QueryRow(query, itemNoParam)

		var (
			itemNo     sql.NullString
			itemUPC    sql.NullString
			itemName   sql.NullString
			catID      sql.NullInt64
			price      sql.NullFloat64
		)
		err := row.Scan(&itemNo, &itemUPC, &itemName, &catID, &price)
		if err != nil {
			if err == sql.ErrNoRows {
				SendError(c, http.StatusNotFound, "Produk tidak ditemukan")
				return
			}
			log.Printf("[Error] Query GetItemByNo failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengakses database untuk mengambil detail produk: "+err.Error())
			return
		}

		item := models.Item{
			ItemNo:     strings.TrimSpace(itemNo.String),
			ItemUPC:    strings.TrimSpace(itemUPC.String),
			ItemName:   strings.TrimSpace(itemName.String),
			CategoryID: int(catID.Int64),
			Price:      price.Float64,
		}

		SendSuccess(c, "Data ditemukan", item)
	}
}

// SearchItems returns a Gin handler to search items by name (case-insensitive LIKE).
func SearchItems(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryParam := c.Query("q")
		if queryParam == "" {
			SendError(c, http.StatusBadRequest, "Parameter pencarian 'q' wajib diisi")
			return
		}

		limit, offset := GetPaginationParams(c)

		// Perform case-insensitive LIKE query using LOWER.
		// Firebird parameter placeholder is ?
		query := "SELECT FIRST ? SKIP ? ITEMNO, ITEMUPC, ITEMNAME, CATEGORYID, DEF_UNITPRICE1 FROM ITEM WHERE LOWER(ITEMNAME) LIKE ? ORDER BY ITEMNAME ASC"
		searchPattern := "%" + strings.ToLower(queryParam) + "%"

		rows, err := db.Query(query, limit, offset, searchPattern)
		if err != nil {
			log.Printf("[Error] Query SearchItems failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengakses database untuk mencari produk: "+err.Error())
			return
		}
		defer rows.Close()

		var items []models.Item
		for rows.Next() {
			var (
				itemNo     sql.NullString
				itemUPC    sql.NullString
				itemName   sql.NullString
				catID      sql.NullInt64
				price      sql.NullFloat64
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &catID, &price); err != nil {
				log.Printf("[Error] Scan item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data produk")
				return
			}
			items = append(items, models.Item{
				ItemNo:     strings.TrimSpace(itemNo.String),
				ItemUPC:    strings.TrimSpace(itemUPC.String),
				ItemName:   strings.TrimSpace(itemName.String),
				CategoryID: int(catID.Int64),
				Price:      price.Float64,
			})
		}

		if err = rows.Err(); err != nil {
			log.Printf("[Error] Rows error in SearchItems: %v", err)
			SendError(c, http.StatusInternalServerError, "Koneksi database bermasalah saat membaca hasil pencarian")
			return
		}

		if items == nil {
			items = []models.Item{}
		}

		SendSuccess(c, "Data ditemukan", items)
	}
}

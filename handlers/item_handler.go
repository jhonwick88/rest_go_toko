package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetItems returns a Gin handler to fetch all items from database.
func GetItems(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, DEF_UNITPRICE1, OBQUANTITY, NOTES FROM ITEM ORDER BY ITEMNO ASC LIMIT ? OFFSET ?"
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
				itemNo        sql.NullString
				itemUPC       sql.NullString
				itemName      sql.NullString
				itemNameShort sql.NullString
				catID         sql.NullInt64
				price         sql.NullFloat64
				obQty         sql.NullFloat64
				notes         sql.NullString
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &price, &obQty, &notes); err != nil {
				log.Printf("[Error] Scan item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data produk")
				return
			}
			items = append(items, models.Item{
				ItemNo:        strings.TrimSpace(itemNo.String),
				ItemUPC:       strings.TrimSpace(itemUPC.String),
				ItemName:      strings.TrimSpace(itemName.String),
				ItemNameShort: strings.TrimSpace(itemNameShort.String),
				CategoryID:    int(catID.Int64),
				DefUnitPrice1: price.Float64,
				ObQuantity:    obQty.Float64,
				Notes:         strings.TrimSpace(notes.String),
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

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, DEF_UNITPRICE1, OBQUANTITY, NOTES FROM ITEM WHERE ITEMNO = ?"
		row := db.QueryRow(query, itemNoParam)

		var (
			itemNo        sql.NullString
			itemUPC       sql.NullString
			itemName      sql.NullString
			itemNameShort sql.NullString
			catID         sql.NullInt64
			price         sql.NullFloat64
			obQty         sql.NullFloat64
			notes         sql.NullString
		)
		err := row.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &price, &obQty, &notes)
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
			ItemNo:        strings.TrimSpace(itemNo.String),
			ItemUPC:       strings.TrimSpace(itemUPC.String),
			ItemName:      strings.TrimSpace(itemName.String),
			ItemNameShort: strings.TrimSpace(itemNameShort.String),
			CategoryID:    int(catID.Int64),
			DefUnitPrice1: price.Float64,
			ObQuantity:    obQty.Float64,
			Notes:         strings.TrimSpace(notes.String),
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

		// Perform case-insensitive LIKE query using LOWER on name, no, and upc.
		// Firebird parameter placeholder is ?
		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, DEF_UNITPRICE1, OBQUANTITY, NOTES FROM ITEM WHERE LOWER(ITEMNAME) LIKE ? OR LOWER(ITEMNO) LIKE ? OR LOWER(ITEMUPC) LIKE ? ORDER BY ITEMNAME ASC LIMIT ? OFFSET ?"
		searchPattern := "%" + strings.ToLower(queryParam) + "%"

		rows, err := db.Query(query, searchPattern, searchPattern, searchPattern, limit, offset)
		if err != nil {
			log.Printf("[Error] Query SearchItems failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengakses database untuk mencari produk: "+err.Error())
			return
		}
		defer rows.Close()

		var items []models.Item
		for rows.Next() {
			var (
				itemNo        sql.NullString
				itemUPC       sql.NullString
				itemName      sql.NullString
				itemNameShort sql.NullString
				catID         sql.NullInt64
				price         sql.NullFloat64
				obQty         sql.NullFloat64
				notes         sql.NullString
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &price, &obQty, &notes); err != nil {
				log.Printf("[Error] Scan item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data produk")
				return
			}
			items = append(items, models.Item{
				ItemNo:        strings.TrimSpace(itemNo.String),
				ItemUPC:       strings.TrimSpace(itemUPC.String),
				ItemName:      strings.TrimSpace(itemName.String),
				ItemNameShort: strings.TrimSpace(itemNameShort.String),
				CategoryID:    int(catID.Int64),
				DefUnitPrice1: price.Float64,
				ObQuantity:    obQty.Float64,
				Notes:         strings.TrimSpace(notes.String),
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

// UpdateItem updates an item's ITEMNO and ITEMUPC in database.
func UpdateItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemNoParam := c.Param("itemno")
		if itemNoParam == "" {
			SendError(c, http.StatusBadRequest, "Parameter itemno wajib diisi")
			return
		}

		type UpdateItemInput struct {
			NewItemNo     string  `json:"new_itemno" binding:"required"`
			ItemUPC       string  `json:"itemupc"`
			DefUnitPrice1 float64 `json:"def_unitprice1"`
			ItemName      string  `json:"itemname"`
			ItemNameShort string  `json:"itemname_short"`
			Notes         string  `json:"notes"`
		}

		var input UpdateItemInput
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
			return
		}

		// Clean inputs
		newItemNo := strings.TrimSpace(input.NewItemNo)
		newItemUPC := strings.TrimSpace(input.ItemUPC)
		newItemName := strings.TrimSpace(input.ItemName)

		if newItemNo == "" {
			SendError(c, http.StatusBadRequest, "new_itemno tidak boleh kosong")
			return
		}

		if newItemName == "" {
			SendError(c, http.StatusBadRequest, "itemname tidak boleh kosong")
			return
		}

		// First, check if the original item exists in the database
		var exists int
		checkQuery := "SELECT 1 FROM ITEM WHERE ITEMNO = ? LIMIT 1"
		err := db.QueryRow(checkQuery, itemNoParam).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				SendError(c, http.StatusNotFound, "Produk yang ingin diubah tidak ditemukan")
				return
			}
			log.Printf("[Error] Query check exists failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memverifikasi keberadaan produk: "+err.Error())
			return
		}

		// Execute update
		var upcToSave interface{}
		if newItemUPC == "" {
			upcToSave = sql.NullString{String: "", Valid: false}
		} else {
			upcToSave = newItemUPC
		}

		updateQuery := "UPDATE ITEM SET ITEMNO = ?, ITEMUPC = ?, DEF_UNITPRICE1 = ?, ITEMNAME = ?, ITEMNAME_SHORT = ?, NOTES = ? WHERE ITEMNO = ?"
		_, err = db.Exec(updateQuery, newItemNo, upcToSave, input.DefUnitPrice1, newItemName, input.ItemNameShort, input.Notes, itemNoParam)
		if err != nil {
			log.Printf("[Error] Query UpdateItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memperbarui database produk: "+err.Error())
			return
		}

		// Retrieve updated item details to send back
		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, DEF_UNITPRICE1, OBQUANTITY, NOTES FROM ITEM WHERE ITEMNO = ?"
		row := db.QueryRow(query, newItemNo)

		var (
			resItemNo        sql.NullString
			resItemUPC       sql.NullString
			resItemName      sql.NullString
			resItemNameShort sql.NullString
			resCatID         sql.NullInt64
			resPrice         sql.NullFloat64
			resObQty         sql.NullFloat64
			resNotes         sql.NullString
		)
		err = row.Scan(&resItemNo, &resItemUPC, &resItemName, &resItemNameShort, &resCatID, &resPrice, &resObQty, &resNotes)
		if err != nil {
			log.Printf("[Error] Query GetItemByNo after update failed: %v", err)
			SendSuccess(c, "Produk berhasil diperbarui", models.Item{
				ItemNo:        newItemNo,
				ItemUPC:       newItemUPC,
				ItemName:      newItemName,
				ItemNameShort: input.ItemNameShort,
				Notes:         input.Notes,
			})
			return
		}

		updatedItem := models.Item{
			ItemNo:        strings.TrimSpace(resItemNo.String),
			ItemUPC:       strings.TrimSpace(resItemUPC.String),
			ItemName:      strings.TrimSpace(resItemName.String),
			ItemNameShort: strings.TrimSpace(resItemNameShort.String),
			CategoryID:    int(resCatID.Int64),
			DefUnitPrice1: resPrice.Float64,
			ObQuantity:    resObQty.Float64,
			Notes:         strings.TrimSpace(resNotes.String),
		}

		SendSuccess(c, "Produk berhasil diperbarui", updatedItem)
	}
}

// CreateItem creates a new item in the database.
func CreateItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type CreateItemInput struct {
			ItemNo        string   `json:"itemno"`
			ItemName      string   `json:"itemname" binding:"required"`
			ItemNameShort string   `json:"itemname_short"`
			ItemUPC       string   `json:"itemupc"`
			CategoryID    int      `json:"categoryid"`
			DefUnitPrice1 float64  `json:"def_unitprice1"`
			ItemType      *int     `json:"itemtype"`
			ObQuantity    *float64 `json:"obquantity"`
			Notes         string   `json:"notes"`
		}

		var input CreateItemInput
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
			return
		}

		itemName := strings.TrimSpace(input.ItemName)
		if itemName == "" {
			SendError(c, http.StatusBadRequest, "Nama produk tidak boleh kosong")
			return
		}

		itemNo := strings.TrimSpace(input.ItemNo)
		if itemNo == "" {
			itemNo = "SKU-" + time.Now().Format("0206150405")
		}

		itemUPC := strings.TrimSpace(input.ItemUPC)
		var upcToSave interface{}
		if itemUPC == "" {
			upcToSave = sql.NullString{String: "", Valid: false}
		} else {
			upcToSave = itemUPC
		}

		itemType := 0
		if input.ItemType != nil {
			itemType = *input.ItemType
		}

		obQuantity := 10.0
		if input.ObQuantity != nil {
			obQuantity = *input.ObQuantity
		}

		// Check if ITEMNO already exists to avoid primary/unique key issues
		var exists int
		checkQuery := "SELECT 1 FROM ITEM WHERE ITEMNO = ? LIMIT 1"
		db.QueryRow(checkQuery, itemNo).Scan(&exists)
		if exists > 0 {
			SendError(c, http.StatusBadRequest, "SKU "+itemNo+" sudah terdaftar")
			return
		}

		// Insert product without explicit ID, letting SQLite autoincrement it
		query := "INSERT INTO ITEM (ITEMNO, ITEMNAME, ITEMNAME_SHORT, ITEMUPC, CATEGORYID, DEF_UNITPRICE1, ITEMTYPE, OBQUANTITY, NOTES, SUPPLIERID, UNIT1) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, NULL)"
		args := []interface{}{itemNo, itemName, input.ItemNameShort, upcToSave, input.CategoryID, input.DefUnitPrice1, itemType, obQuantity, input.Notes}

		_, err := db.Exec(query, args...)
		if err != nil {
			log.Printf("[Error] Query CreateItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menambahkan produk baru ke database: "+err.Error())
			return
		}

		// Retrieve created item details to send back
		var (
			resItemNo        sql.NullString
			resItemUPC       sql.NullString
			resItemName      sql.NullString
			resItemNameShort sql.NullString
			resCatID         sql.NullInt64
			resPrice         sql.NullFloat64
			resObQty         sql.NullFloat64
			resNotes         sql.NullString
		)
		db.QueryRow("SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, DEF_UNITPRICE1, OBQUANTITY, NOTES FROM ITEM WHERE ITEMNO = ?", itemNo).
			Scan(&resItemNo, &resItemUPC, &resItemName, &resItemNameShort, &resCatID, &resPrice, &resObQty, &resNotes)

		createdItem := models.Item{
			ItemNo:        strings.TrimSpace(resItemNo.String),
			ItemUPC:       strings.TrimSpace(resItemUPC.String),
			ItemName:      strings.TrimSpace(resItemName.String),
			ItemNameShort: strings.TrimSpace(resItemNameShort.String),
			CategoryID:    int(resCatID.Int64),
			DefUnitPrice1: resPrice.Float64,
			ObQuantity:    resObQty.Float64,
			Notes:         strings.TrimSpace(resNotes.String),
		}

		SendSuccess(c, "Produk berhasil ditambahkan", createdItem)
	}
}

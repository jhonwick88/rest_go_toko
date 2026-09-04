package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"rest_go_toko/models"
	"rest_go_toko/services"

	"github.com/gin-gonic/gin"
)

// GetItems returns a Gin handler to fetch all items from database.
func GetItems(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE1, DEF_UNITPRICE2, DEF_UNITPRICE3, OBQUANTITY, NOTES FROM ITEM ORDER BY ITEMNO ASC LIMIT ? OFFSET ?"
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
				attribute     sql.NullString
				unit1         sql.NullString
				unit2         sql.NullString
				unit3         sql.NullString
				ratio2        sql.NullFloat64
				ratio3        sql.NullFloat64
				price1        sql.NullFloat64
				price2        sql.NullFloat64
				price3        sql.NullFloat64
				obQty         sql.NullFloat64
				notes         sql.NullString
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &attribute, &unit1, &unit2, &unit3, &ratio2, &ratio3, &price1, &price2, &price3, &obQty, &notes); err != nil {
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
				Attribute:     strings.TrimSpace(attribute.String),
				Unit1:         strings.TrimSpace(unit1.String),
				Unit2:         strings.TrimSpace(unit2.String),
				Unit3:         strings.TrimSpace(unit3.String),
				Ratio2:        ratio2.Float64,
				Ratio3:        ratio3.Float64,
				DefUnitPrice1: price1.Float64,
				DefUnitPrice2: price2.Float64,
				DefUnitPrice3: price3.Float64,
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

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE1, DEF_UNITPRICE2, DEF_UNITPRICE3, OBQUANTITY, NOTES FROM ITEM WHERE ITEMNO = ?"
		row := db.QueryRow(query, itemNoParam)

		var (
			itemNo        sql.NullString
			itemUPC       sql.NullString
			itemName      sql.NullString
			itemNameShort sql.NullString
			catID         sql.NullInt64
			attribute     sql.NullString
			unit1         sql.NullString
			unit2         sql.NullString
			unit3         sql.NullString
			ratio2        sql.NullFloat64
			ratio3        sql.NullFloat64
			price1        sql.NullFloat64
			price2        sql.NullFloat64
			price3        sql.NullFloat64
			obQty         sql.NullFloat64
			notes         sql.NullString
		)
		err := row.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &attribute, &unit1, &unit2, &unit3, &ratio2, &ratio3, &price1, &price2, &price3, &obQty, &notes)
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
			Attribute:     strings.TrimSpace(attribute.String),
			Unit1:         strings.TrimSpace(unit1.String),
			Unit2:         strings.TrimSpace(unit2.String),
			Unit3:         strings.TrimSpace(unit3.String),
			Ratio2:        ratio2.Float64,
			Ratio3:        ratio3.Float64,
			DefUnitPrice1: price1.Float64,
			DefUnitPrice2: price2.Float64,
			DefUnitPrice3: price3.Float64,
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
		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE1, DEF_UNITPRICE2, DEF_UNITPRICE3, OBQUANTITY, NOTES FROM ITEM WHERE LOWER(ITEMNAME) LIKE ? OR LOWER(ITEMNO) LIKE ? OR LOWER(ITEMUPC) LIKE ? ORDER BY ITEMNAME ASC LIMIT ? OFFSET ?"
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
				attribute     sql.NullString
				unit1         sql.NullString
				unit2         sql.NullString
				unit3         sql.NullString
				ratio2        sql.NullFloat64
				ratio3        sql.NullFloat64
				price1        sql.NullFloat64
				price2        sql.NullFloat64
				price3        sql.NullFloat64
				obQty         sql.NullFloat64
				notes         sql.NullString
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &attribute, &unit1, &unit2, &unit3, &ratio2, &ratio3, &price1, &price2, &price3, &obQty, &notes); err != nil {
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
				Attribute:     strings.TrimSpace(attribute.String),
				Unit1:         strings.TrimSpace(unit1.String),
				Unit2:         strings.TrimSpace(unit2.String),
				Unit3:         strings.TrimSpace(unit3.String),
				Ratio2:        ratio2.Float64,
				Ratio3:        ratio3.Float64,
				DefUnitPrice1: price1.Float64,
				DefUnitPrice2: price2.Float64,
				DefUnitPrice3: price3.Float64,
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
			Attribute     string  `json:"attribute"`
			Unit1         string  `json:"unit1"`
			Unit2         string  `json:"unit2"`
			Unit3         string  `json:"unit3"`
			Ratio2        float64 `json:"ratio2"`
			Ratio3        float64 `json:"ratio3"`
			DefUnitPrice2 float64 `json:"def_unitprice2"`
			DefUnitPrice3 float64 `json:"def_unitprice3"`
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

		updateQuery := "UPDATE ITEM SET ITEMNO = ?, ITEMUPC = ?, DEF_UNITPRICE1 = ?, ITEMNAME = ?, ITEMNAME_SHORT = ?, NOTES = ?, ATTRIBUTE = ?, UNIT1 = ?, UNIT2 = ?, UNIT3 = ?, RATIO2 = ?, RATIO3 = ?, DEF_UNITPRICE2 = ?, DEF_UNITPRICE3 = ? WHERE ITEMNO = ?"
		_, err = db.Exec(updateQuery, newItemNo, upcToSave, input.DefUnitPrice1, newItemName, input.ItemNameShort, input.Notes, input.Attribute, input.Unit1, input.Unit2, input.Unit3, input.Ratio2, input.Ratio3, input.DefUnitPrice2, input.DefUnitPrice3, itemNoParam)
		if err != nil {
			log.Printf("[Error] Query UpdateItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memperbarui database produk: "+err.Error())
			return
		}

		// Retrieve updated item details to send back
		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE1, DEF_UNITPRICE2, DEF_UNITPRICE3, OBQUANTITY, NOTES FROM ITEM WHERE ITEMNO = ?"
		row := db.QueryRow(query, newItemNo)

		var (
			resItemNo        sql.NullString
			resItemUPC       sql.NullString
			resItemName      sql.NullString
			resItemNameShort sql.NullString
			resCatID         sql.NullInt64
			resAttribute     sql.NullString
			resUnit1         sql.NullString
			resUnit2         sql.NullString
			resUnit3         sql.NullString
			resRatio2        sql.NullFloat64
			resRatio3        sql.NullFloat64
			resPrice1        sql.NullFloat64
			resPrice2        sql.NullFloat64
			resPrice3        sql.NullFloat64
			resObQty         sql.NullFloat64
			resNotes         sql.NullString
		)
		err = row.Scan(&resItemNo, &resItemUPC, &resItemName, &resItemNameShort, &resCatID, &resAttribute, &resUnit1, &resUnit2, &resUnit3, &resRatio2, &resRatio3, &resPrice1, &resPrice2, &resPrice3, &resObQty, &resNotes)
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
			Attribute:     strings.TrimSpace(resAttribute.String),
			Unit1:         strings.TrimSpace(resUnit1.String),
			Unit2:         strings.TrimSpace(resUnit2.String),
			Unit3:         strings.TrimSpace(resUnit3.String),
			Ratio2:        resRatio2.Float64,
			Ratio3:        resRatio3.Float64,
			DefUnitPrice1: resPrice1.Float64,
			DefUnitPrice2: resPrice2.Float64,
			DefUnitPrice3: resPrice3.Float64,
			ObQuantity:    resObQty.Float64,
			Notes:         strings.TrimSpace(resNotes.String),
		}

		SendSuccess(c, "Produk berhasil diperbarui", updatedItem)
	}
}

// CreateItem creates a new item in the database.
func CreateItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// --- LICENSE CHECK ---
		features := services.GetLicenseFeatures()
		if features != nil {
			if maxProducts, ok := features["max_products"].(float64); ok && maxProducts > 0 {
				var count int
				db.QueryRow("SELECT COUNT(*) FROM ITEM").Scan(&count)
				if float64(count) >= maxProducts {
					SendError(c, 403, "Batas maksimal produk dari lisensi Anda telah tercapai.")
					return
				}
			}
		}
		// --- END LICENSE CHECK ---

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
			Attribute     string   `json:"attribute"`
			Unit1         string   `json:"unit1"`
			Unit2         string   `json:"unit2"`
			Unit3         string   `json:"unit3"`
			Ratio2        float64  `json:"ratio2"`
			Ratio3        float64  `json:"ratio3"`
			DefUnitPrice2 float64  `json:"def_unitprice2"`
			DefUnitPrice3 float64  `json:"def_unitprice3"`
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
		query := "INSERT INTO ITEM (ITEMNO, ITEMNAME, ITEMNAME_SHORT, ITEMUPC, CATEGORYID, DEF_UNITPRICE1, ITEMTYPE, OBQUANTITY, NOTES, SUPPLIERID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE2, DEF_UNITPRICE3) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?, ?, ?, ?, ?, ?)"
		args := []interface{}{itemNo, itemName, input.ItemNameShort, upcToSave, input.CategoryID, input.DefUnitPrice1, itemType, obQuantity, input.Notes, input.Attribute, input.Unit1, input.Unit2, input.Unit3, input.Ratio2, input.Ratio3, input.DefUnitPrice2, input.DefUnitPrice3}

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
			resAttribute     sql.NullString
			resUnit1         sql.NullString
			resUnit2         sql.NullString
			resUnit3         sql.NullString
			resRatio2        sql.NullFloat64
			resRatio3        sql.NullFloat64
			resPrice1        sql.NullFloat64
			resPrice2        sql.NullFloat64
			resPrice3        sql.NullFloat64
			resObQty         sql.NullFloat64
			resNotes         sql.NullString
		)
		db.QueryRow("SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE1, DEF_UNITPRICE2, DEF_UNITPRICE3, OBQUANTITY, NOTES FROM ITEM WHERE ITEMNO = ?", itemNo).
			Scan(&resItemNo, &resItemUPC, &resItemName, &resItemNameShort, &resCatID, &resAttribute, &resUnit1, &resUnit2, &resUnit3, &resRatio2, &resRatio3, &resPrice1, &resPrice2, &resPrice3, &resObQty, &resNotes)

		createdItem := models.Item{
			ItemNo:        strings.TrimSpace(resItemNo.String),
			ItemUPC:       strings.TrimSpace(resItemUPC.String),
			ItemName:      strings.TrimSpace(resItemName.String),
			ItemNameShort: strings.TrimSpace(resItemNameShort.String),
			CategoryID:    int(resCatID.Int64),
			Attribute:     strings.TrimSpace(resAttribute.String),
			Unit1:         strings.TrimSpace(resUnit1.String),
			Unit2:         strings.TrimSpace(resUnit2.String),
			Unit3:         strings.TrimSpace(resUnit3.String),
			Ratio2:        resRatio2.Float64,
			Ratio3:        resRatio3.Float64,
			DefUnitPrice1: resPrice1.Float64,
			DefUnitPrice2: resPrice2.Float64,
			DefUnitPrice3: resPrice3.Float64,
			ObQuantity:    resObQty.Float64,
			Notes:         strings.TrimSpace(resNotes.String),
		}

		SendSuccess(c, "Produk berhasil ditambahkan", createdItem)
	}
}

// DeleteItem deletes an item from the database.
func DeleteItem(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemNoParam := c.Param("itemno")
		if itemNoParam == "" {
			SendError(c, http.StatusBadRequest, "Parameter itemno wajib diisi")
			return
		}

		// Check if exists
		var exists int
		err := db.QueryRow("SELECT 1 FROM ITEM WHERE ITEMNO = ? LIMIT 1", itemNoParam).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				SendError(c, http.StatusNotFound, "Produk tidak ditemukan")
				return
			}
			SendError(c, http.StatusInternalServerError, "Gagal memverifikasi produk: "+err.Error())
			return
		}

		_, err = db.Exec("DELETE FROM ITEM WHERE ITEMNO = ?", itemNoParam)
		if err != nil {
			log.Printf("[Error] Query DeleteItem failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menghapus produk: "+err.Error())
			return
		}

		SendSuccess(c, "Produk berhasil dihapus", nil)
	}
}

// UpdateStock updates an item's stock (OBQUANTITY).
func UpdateStock(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		itemNoParam := c.Param("itemno")
		if itemNoParam == "" {
			SendError(c, http.StatusBadRequest, "Parameter itemno wajib diisi")
			return
		}

		type UpdateStockInput struct {
			ObQuantity float64 `json:"obquantity"`
			Notes      string  `json:"notes"`
		}

		var input UpdateStockInput
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid: "+err.Error())
			return
		}

		var oldQty float64
		_ = db.QueryRow("SELECT OBQUANTITY FROM ITEM WHERE ITEMNO = ?", itemNoParam).Scan(&oldQty)

		_, err := db.Exec("UPDATE ITEM SET OBQUANTITY = ? WHERE ITEMNO = ?", input.ObQuantity, itemNoParam)
		if err != nil {
			log.Printf("[Error] Query UpdateStock failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memperbarui stok: "+err.Error())
			return
		}

		diff := input.ObQuantity - oldQty
		if diff != 0 {
			hasStockMovement := false
			if claims, err := services.GetLicenseClaims(); err == nil && claims != nil {
				if smFlag, ok := claims.Features["stock_movement"].(bool); ok {
					hasStockMovement = smFlag
				}
			}

			if hasStockMovement {
				nowDate := time.Now().Format(time.RFC3339)
				noteStr := "Manual Adjustment"
				if input.Notes != "" {
					noteStr = "Stok Opname: " + input.Notes
				}
				_, err = db.Exec("INSERT INTO STOCK_LEDGER (ITEMNO, TX_TYPE, QTY, NOTES, TIMESTAMP) VALUES (?, 'ADJ', ?, ?, ?)", itemNoParam, diff, noteStr, nowDate)
				if err != nil {
					log.Printf("[Error] Query insert stock ledger failed: %v", err)
					// Proceed, don't fail the whole request
				}
			}
		}

		SendSuccess(c, "Stok berhasil diperbarui", gin.H{"obquantity": input.ObQuantity})
	}
}

// GetItemsByBarcode returns a list of items matching a specific barcode (ITEMUPC) or ITEMNO
func GetItemsByBarcode(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		barcode := c.Param("barcode")
		if barcode == "" {
			SendError(c, http.StatusBadRequest, "Parameter barcode wajib diisi")
			return
		}

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, ITEMNAME_SHORT, CATEGORYID, ATTRIBUTE, UNIT1, UNIT2, UNIT3, RATIO2, RATIO3, DEF_UNITPRICE1, DEF_UNITPRICE2, DEF_UNITPRICE3, OBQUANTITY, NOTES FROM ITEM WHERE ITEMUPC = ? OR ITEMNO = ?"
		rows, err := db.Query(query, barcode, barcode)
		if err != nil {
			log.Printf("[Error] Query GetItemsByBarcode failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mencari barcode: "+err.Error())
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
				attribute     sql.NullString
				unit1         sql.NullString
				unit2         sql.NullString
				unit3         sql.NullString
				ratio2        sql.NullFloat64
				ratio3        sql.NullFloat64
				price1        sql.NullFloat64
				price2        sql.NullFloat64
				price3        sql.NullFloat64
				obQty         sql.NullFloat64
				notes         sql.NullString
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &itemNameShort, &catID, &attribute, &unit1, &unit2, &unit3, &ratio2, &ratio3, &price1, &price2, &price3, &obQty, &notes); err != nil {
				log.Printf("[Error] Scan item failed: %v", err)
				continue
			}
			items = append(items, models.Item{
				ItemNo:        strings.TrimSpace(itemNo.String),
				ItemUPC:       strings.TrimSpace(itemUPC.String),
				ItemName:      strings.TrimSpace(itemName.String),
				ItemNameShort: strings.TrimSpace(itemNameShort.String),
				CategoryID:    int(catID.Int64),
				Attribute:     strings.TrimSpace(attribute.String),
				Unit1:         strings.TrimSpace(unit1.String),
				Unit2:         strings.TrimSpace(unit2.String),
				Unit3:         strings.TrimSpace(unit3.String),
				Ratio2:        ratio2.Float64,
				Ratio3:        ratio3.Float64,
				DefUnitPrice1: price1.Float64,
				DefUnitPrice2: price2.Float64,
				DefUnitPrice3: price3.Float64,
				ObQuantity:    obQty.Float64,
				Notes:         strings.TrimSpace(notes.String),
			})
		}

		if items == nil {
			items = []models.Item{}
		}
		
		SendSuccess(c, "Data ditemukan", items)
	}
}


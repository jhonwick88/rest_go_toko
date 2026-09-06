package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetCategories returns a Gin handler to fetch all categories from database.
func GetCategories(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := "SELECT ID, NAME FROM ITEM_CATEGORY ORDER BY NAME ASC"
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("[Error] Query GetCategories failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengakses database untuk mengambil kategori: "+err.Error())
			return
		}
		defer rows.Close()

		var categories []models.Category
		for rows.Next() {
			var (
				id   sql.NullInt64
				name sql.NullString
			)
			if err := rows.Scan(&id, &name); err != nil {
				log.Printf("[Error] Scan category failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data kategori")
				return
			}
			categories = append(categories, models.Category{
				ID:   int(id.Int64),
				Name: stringsTrim(name.String),
			})
		}

		if err = rows.Err(); err != nil {
			log.Printf("[Error] Rows error in GetCategories: %v", err)
			SendError(c, http.StatusInternalServerError, "Koneksi database bermasalah saat membaca kategori")
			return
		}

		// If no category found, categories will be nil. Map to empty slice for clean JSON.
		if categories == nil {
			categories = []models.Category{}
		}

		SendSuccess(c, "Data ditemukan", categories)
	}
}

// GetCategoryItems returns a Gin handler to fetch all items that belong to a category.
func GetCategoryItems(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		categoryID, err := strconv.Atoi(idParam)
		if err != nil {
			SendError(c, http.StatusBadRequest, "ID kategori tidak valid")
			return
		}

		limit, offset := GetPaginationParams(c)

		query := "SELECT ITEMNO, ITEMUPC, ITEMNAME, CATEGORYID, DEF_UNITPRICE1, OBQUANTITY FROM ITEM WHERE CATEGORYID = ? ORDER BY ITEMNAME ASC LIMIT ? OFFSET ?"
		rows, err := db.Query(query, categoryID, limit, offset)
		if err != nil {
			log.Printf("[Error] Query GetCategoryItems failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengakses database untuk mengambil produk: "+err.Error())
			return
		}
		defer rows.Close()

		var items []models.Item
		for rows.Next() {
			var (
				itemNo   sql.NullString
				itemUPC  sql.NullString
				itemName sql.NullString
				catID    sql.NullInt64
				price    sql.NullFloat64
				obQty    sql.NullFloat64
			)
			if err := rows.Scan(&itemNo, &itemUPC, &itemName, &catID, &price, &obQty); err != nil {
				log.Printf("[Error] Scan item failed: %v", err)
				SendError(c, http.StatusInternalServerError, "Gagal membaca data produk")
				return
			}
			items = append(items, models.Item{
				ItemNo:        stringsTrim(itemNo.String),
				ItemUPC:       stringsTrim(itemUPC.String),
				ItemName:      stringsTrim(itemName.String),
				CategoryID:    int(catID.Int64),
				DefUnitPrice1: price.Float64,
				ObQuantity:    obQty.Float64,
			})
		}

		if err = rows.Err(); err != nil {
			log.Printf("[Error] Rows error in GetCategoryItems: %v", err)
			SendError(c, http.StatusInternalServerError, "Koneksi database bermasalah saat membaca produk")
			return
		}

		if items == nil {
			items = []models.Item{}
		}

		SendSuccess(c, "Data ditemukan", items)
	}
}

// CreateCategory handles adding a new category
func CreateCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.Category
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
			return
		}

		if strings.TrimSpace(input.Name) == "" {
			SendError(c, http.StatusBadRequest, "Nama kategori tidak boleh kosong")
			return
		}

		query := "INSERT INTO ITEM_CATEGORY (NAME) VALUES (?)"
		result, err := db.Exec(query, strings.TrimSpace(input.Name))
		if err != nil {
			log.Printf("[Error] Query CreateCategory failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menambahkan kategori: "+err.Error())
			return
		}

		id, _ := result.LastInsertId()
		input.ID = int(id)

		SendSuccess(c, "Kategori berhasil ditambahkan", input)
	}
}

// UpdateCategory handles updating an existing category
func UpdateCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		categoryID, err := strconv.Atoi(idParam)
		if err != nil {
			SendError(c, http.StatusBadRequest, "ID kategori tidak valid")
			return
		}

		var input models.Category
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Data tidak valid: "+err.Error())
			return
		}

		if strings.TrimSpace(input.Name) == "" {
			SendError(c, http.StatusBadRequest, "Nama kategori tidak boleh kosong")
			return
		}

		query := "UPDATE ITEM_CATEGORY SET NAME = ? WHERE ID = ?"
		res, err := db.Exec(query, strings.TrimSpace(input.Name), categoryID)
		if err != nil {
			log.Printf("[Error] Query UpdateCategory failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengupdate kategori: "+err.Error())
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			SendError(c, http.StatusNotFound, "Kategori tidak ditemukan")
			return
		}

		input.ID = categoryID
		SendSuccess(c, "Kategori berhasil diupdate", input)
	}
}

// DeleteCategory handles deleting a category
func DeleteCategory(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		categoryID, err := strconv.Atoi(idParam)
		if err != nil {
			SendError(c, http.StatusBadRequest, "ID kategori tidak valid")
			return
		}

		// Check if category is used by any items
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM ITEM WHERE CATEGORYID = ?", categoryID).Scan(&count)
		if err != nil {
			log.Printf("[Error] Check items for category failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengecek penggunaan kategori")
			return
		}
		if count > 0 {
			SendError(c, http.StatusConflict, "Kategori tidak dapat dihapus karena masih digunakan oleh produk")
			return
		}

		query := "DELETE FROM ITEM_CATEGORY WHERE ID = ?"
		res, err := db.Exec(query, categoryID)
		if err != nil {
			log.Printf("[Error] Query DeleteCategory failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menghapus kategori: "+err.Error())
			return
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected == 0 {
			SendError(c, http.StatusNotFound, "Kategori tidak ditemukan")
			return
		}

		SendSuccess(c, "Kategori berhasil dihapus", nil)
	}
}

// Helper function to trim spaces (Firebird CHAR fields are space-padded to their max length)
func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

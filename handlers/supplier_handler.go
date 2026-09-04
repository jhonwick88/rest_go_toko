package handlers

import (
	"database/sql"
	"net/http"
	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
)

// GetSuppliers returns all suppliers.
func GetSuppliers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT SUPPLIERID, NAME, CONTACT, ADDRESS, NOTES FROM SUPPLIERS")
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data supplier: "+err.Error())
			return
		}
		defer rows.Close()

		var suppliers []models.Supplier
		for rows.Next() {
			var s models.Supplier
			var contact, address, notes sql.NullString
			if err := rows.Scan(&s.SupplierID, &s.Name, &contact, &address, &notes); err != nil {
				SendError(c, http.StatusInternalServerError, "Gagal membaca data supplier")
				return
			}
			s.Contact = contact.String
			s.Address = address.String
			s.Notes = notes.String
			suppliers = append(suppliers, s)
		}
		SendSuccess(c, "Data supplier berhasil diambil", suppliers)
	}
}

// CreateSupplier adds a new supplier.
func CreateSupplier(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var s models.Supplier
		if err := c.ShouldBindJSON(&s); err != nil {
			SendError(c, http.StatusBadRequest, "Data tidak valid")
			return
		}

		res, err := db.Exec("INSERT INTO SUPPLIERS (NAME, CONTACT, ADDRESS, NOTES) VALUES (?, ?, ?, ?)", s.Name, s.Contact, s.Address, s.Notes)
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal menambahkan supplier: "+err.Error())
			return
		}

		id, _ := res.LastInsertId()
		s.SupplierID = int(id)
		SendSuccess(c, "Supplier berhasil ditambahkan", s)
	}
}

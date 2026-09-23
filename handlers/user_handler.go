package handlers

import (
	"rest_go_toko/services"
	"database/sql"
	"log"
	"net/http"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUsers returns all users
func GetUsers(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := "SELECT UID, USERNAME, FULLNAME, ROLE, PIN, IS_ACTIVE FROM USERS ORDER BY USERNAME ASC"
		rows, err := db.Query(query)
		if err != nil {
			log.Printf("[Error] Query GetUsers failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data user")
			return
		}
		defer rows.Close()

		var users []models.User
		for rows.Next() {
			var u models.User
			var pin sql.NullString
			if err := rows.Scan(&u.UID, &u.Username, &u.FullName, &u.Role, &pin, &u.IsActive); err != nil {
				SendError(c, http.StatusInternalServerError, "Gagal membaca data user")
				return
			}
			u.PIN = pin.String
			users = append(users, u)
		}

		if users == nil {
			users = []models.User{}
		}
		SendSuccess(c, "Data ditemukan", users)
	}
}

// CreateUser creates a new user
func CreateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.User
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		// --- LICENSE CHECK ---
		features := services.GetLicenseFeatures()
		if features != nil {
			// Check max_cashiers or max_users
			var maxLimit float64 = 0
			if mc, ok := features["max_cashiers"].(float64); ok && mc > 0 {
				maxLimit = mc
			} else if mu, ok := features["max_users"].(float64); ok && mu > 0 {
				maxLimit = mu
			}

			if maxLimit > 0 {
				var count int
				if input.Role == "cashier" {
					db.QueryRow("SELECT COUNT(*) FROM USERS WHERE ROLE = 'cashier' AND IS_ACTIVE = 1").Scan(&count)
				} else {
					db.QueryRow("SELECT COUNT(*) FROM USERS WHERE IS_ACTIVE = 1").Scan(&count)
				}
				if float64(count) >= maxLimit {
					SendError(c, http.StatusForbidden, "Batas maksimal kasir/pengguna dari lisensi Anda telah tercapai.")
					return
				}
			}
		}
		// --- END LICENSE CHECK ---

		// --- PIN UNIQUE CHECK ---
		if input.PIN != "" {
			var pinCount int
			db.QueryRow("SELECT COUNT(*) FROM USERS WHERE PIN = ?", input.PIN).Scan(&pinCount)
			if pinCount > 0 {
				SendError(c, http.StatusBadRequest, "PIN ini sudah digunakan oleh pengguna lain")
				return
			}
		}
		// --- END PIN UNIQUE CHECK ---

		uid := uuid.New().String()
		query := "INSERT INTO USERS (UID, USERNAME, FULLNAME, ROLE, PIN, IS_ACTIVE) VALUES (?, ?, ?, ?, ?, ?)"
		_, err := db.Exec(query, uid, input.Username, input.FullName, input.Role, input.PIN, 1)
		if err != nil {
			log.Printf("[Error] CreateUser failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menambahkan user baru")
			return
		}

		input.UID = uid
		input.IsActive = 1
		SendSuccess(c, "User berhasil ditambahkan", input)
	}
}

// UpdateUser updates an existing user
func UpdateUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			SendError(c, http.StatusBadRequest, "UID user diperlukan")
			return
		}

		var input models.User
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		// Check if user exists
		var existingRole string
		err := db.QueryRow("SELECT ROLE FROM USERS WHERE UID = ?", uid).Scan(&existingRole)
		if err != nil {
			if err == sql.ErrNoRows {
				SendError(c, http.StatusNotFound, "User tidak ditemukan")
				return
			}
			SendError(c, http.StatusInternalServerError, "Gagal memeriksa data user")
			return
		}

		// --- PIN UNIQUE CHECK ---
		if input.PIN != "" {
			var pinCount int
			db.QueryRow("SELECT COUNT(*) FROM USERS WHERE PIN = ? AND UID != ?", input.PIN, uid).Scan(&pinCount)
			if pinCount > 0 {
				SendError(c, http.StatusBadRequest, "PIN ini sudah digunakan oleh pengguna lain")
				return
			}
		}
		// --- END PIN UNIQUE CHECK ---

		var query string
		var args []interface{}
		if input.PIN != "" {
			query = "UPDATE USERS SET USERNAME = ?, FULLNAME = ?, ROLE = ?, PIN = ?, IS_ACTIVE = ? WHERE UID = ?"
			args = []interface{}{input.Username, input.FullName, input.Role, input.PIN, input.IsActive, uid}
		} else {
			query = "UPDATE USERS SET USERNAME = ?, FULLNAME = ?, ROLE = ?, IS_ACTIVE = ? WHERE UID = ?"
			args = []interface{}{input.Username, input.FullName, input.Role, input.IsActive, uid}
		}

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("[Error] UpdateUser failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal memperbarui user")
			return
		}

		input.UID = uid
		SendSuccess(c, "User berhasil diperbarui", input)
	}
}

// DeleteUser deletes an existing user, protecting admin role from deletion
func DeleteUser(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.Param("uid")
		if uid == "" {
			SendError(c, http.StatusBadRequest, "UID user diperlukan")
			return
		}

		// Check user role
		var role string
		err := db.QueryRow("SELECT ROLE FROM USERS WHERE UID = ?", uid).Scan(&role)
		if err != nil {
			if err == sql.ErrNoRows {
				SendError(c, http.StatusNotFound, "User tidak ditemukan")
				return
			}
			SendError(c, http.StatusInternalServerError, "Gagal memeriksa data user")
			return
		}

		if role == "admin" {
			SendError(c, http.StatusBadRequest, "Akun Administrator tidak dapat dihapus")
			return
		}

		_, err = db.Exec("DELETE FROM USERS WHERE UID = ?", uid)
		if err != nil {
			log.Printf("[Error] DeleteUser failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menghapus user")
			return
		}

		SendSuccess(c, "User berhasil dihapus", nil)
	}
}

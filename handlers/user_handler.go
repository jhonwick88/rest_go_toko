package handlers

import (
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

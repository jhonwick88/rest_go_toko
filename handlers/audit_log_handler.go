package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"time"

	"rest_go_toko/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetAuditLogs returns all audit logs
func GetAuditLogs(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, offset := GetPaginationParams(c)

		query := "SELECT ID, TIMESTAMP, USER_ID, USERNAME, ACTION, DETAILS FROM AUDIT_LOGS ORDER BY TIMESTAMP DESC LIMIT ? OFFSET ?"
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			log.Printf("[Error] Query GetAuditLogs failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal mengambil data log audit")
			return
		}
		defer rows.Close()

		var logs []models.AuditLog
		for rows.Next() {
			var l models.AuditLog
			var timestamp, userID, username, action, details sql.NullString

			if err := rows.Scan(&l.ID, &timestamp, &userID, &username, &action, &details); err != nil {
				SendError(c, http.StatusInternalServerError, "Gagal membaca data log audit")
				return
			}
			l.Timestamp = timestamp.String
			l.UserID = userID.String
			l.Username = username.String
			l.Action = action.String
			l.Details = details.String
			logs = append(logs, l)
		}

		if logs == nil {
			logs = []models.AuditLog{}
		}
		SendSuccess(c, "Data ditemukan", logs)
	}
}

// CreateAuditLog creates a new audit log
func CreateAuditLog(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.AuditLog
		if err := c.ShouldBindJSON(&input); err != nil {
			SendError(c, http.StatusBadRequest, "Payload request tidak valid")
			return
		}

		id := uuid.New().String()
		timestamp := input.Timestamp
		if timestamp == "" {
			timestamp = time.Now().Format(time.RFC3339)
		}

		query := "INSERT INTO AUDIT_LOGS (ID, TIMESTAMP, USER_ID, USERNAME, ACTION, DETAILS) VALUES (?, ?, ?, ?, ?, ?)"
		
		_, err := db.Exec(query, id, timestamp, input.UserID, input.Username, input.Action, input.Details)
		if err != nil {
			log.Printf("[Error] CreateAuditLog failed: %v", err)
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan log audit")
			return
		}

		input.ID = id
		SendSuccess(c, "Log audit berhasil disimpan", input)
	}
}

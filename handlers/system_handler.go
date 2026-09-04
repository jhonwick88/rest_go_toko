package handlers

import (
	"database/sql"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// BackupDatabase serves the SQLite database file for download
func BackupDatabase(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			SendError(c, http.StatusInternalServerError, "DB_PATH tidak dikonfigurasi.")
			return
		}

		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			SendError(c, http.StatusNotFound, "File database tidak ditemukan.")
			return
		}

		c.Header("Content-Disposition", `attachment; filename="toko_pintar.db"`)
		c.Header("Content-Type", "application/octet-stream")
		c.File(dbPath)
	}
}

// RestoreDatabase accepts an uploaded database file and replaces the current one
func RestoreDatabase(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			SendError(c, http.StatusInternalServerError, "DB_PATH tidak dikonfigurasi.")
			return
		}

		// Parse the uploaded file
		file, header, err := c.Request.FormFile("database")
		if err != nil {
			SendError(c, http.StatusBadRequest, "Gagal mengunggah file database.")
			return
		}
		defer file.Close()

		// Validate file extension loosely
		if len(header.Filename) < 3 || header.Filename[len(header.Filename)-3:] != ".db" {
			SendError(c, http.StatusBadRequest, "Format file harus .db")
			return
		}

		// Close current database connection so we can overwrite the file (Windows lock issue)
		_ = db.Close()

		// Remove WAL and SHM files to prevent corruption
		_ = os.Remove(dbPath + "-wal")
		_ = os.Remove(dbPath + "-shm")

		// Create the destination file
		out, err := os.Create(dbPath)
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal menimpa database (mungkin sedang digunakan). Silakan restart aplikasi dan ulangi.")
			return
		}
		defer out.Close()

		// Copy the uploaded file contents
		_, err = io.Copy(out, file)
		if err != nil {
			SendError(c, http.StatusInternalServerError, "Gagal menyimpan file database yang baru.")
			return
		}

		// Send success response
		SendSuccess(c, "Restore berhasil! Peladen akan dimatikan otomatis dalam 2 detik. Silakan restart TokoPintar Server secara manual.", nil)

		// Schedule a graceful exit to force a restart
		go func() {
			time.Sleep(2 * time.Second)
			os.Exit(0)
		}()
	}
}

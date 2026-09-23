package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"rest_go_toko/config"

	_ "modernc.org/sqlite"
)

// InitDB initializes the connection pool to the SQLite database.
func InitDB(cfg *config.Config) (*sql.DB, error) {
	log.Printf("Connecting to SQLite database at %s...", cfg.DBPath)

	// Open connection to SQLite
	db, err := sql.Open("sqlite", cfg.DBPath)
	if err != nil {
		return nil, fmt.Errorf("error opening sqlite connection: %w", err)
	}

	// Enable Foreign Keys for SQLite
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys = ON;"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// Configure Connection Pooling
	db.SetMaxOpenConns(10)                  // Maximum number of open connections to the database.
	db.SetMaxIdleConns(5)                   // Maximum number of connections in the idle connection pool.
	db.SetConnMaxLifetime(1 * time.Hour)    // Maximum amount of time a connection may be reused.
	db.SetConnMaxIdleTime(15 * time.Minute) // Maximum amount of time a connection may be idle.

	// Test connection
	// Ensure required tables exist
	migrationQuery := `
	CREATE TABLE IF NOT EXISTS CASH_MOVEMENTS (
		ID TEXT PRIMARY KEY,
		DATE TEXT NOT NULL,
		CASHIER_NAME TEXT NOT NULL,
		TYPE TEXT NOT NULL DEFAULT 'OUT',
		CATEGORY TEXT NOT NULL,
		AMOUNT REAL NOT NULL,
		NOTES TEXT
	);
	CREATE TABLE IF NOT EXISTS CASH_RECONCILIATIONS (
		ID TEXT PRIMARY KEY,
		DATE TEXT NOT NULL,
		CASHIER_NAME TEXT,
		SYSTEM_REVENUE REAL,
		ACTUAL_DRAWER_CASH REAL,
		DIFFERENCE REAL,
		ACCURACY_RATE REAL,
		NOTES TEXT
	);
	`
	if _, err := db.Exec(migrationQuery); err != nil {
		log.Printf("[Warning] Failed to run table migrations: %v", err)
	}

	log.Println("SQLite database connection established successfully")
	return db, nil
}


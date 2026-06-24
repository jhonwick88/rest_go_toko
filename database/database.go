package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"rest_go_toko/config"

	_ "github.com/nakagami/firebirdsql"
)

// InitDB initializes the connection pool to the Firebird database.
func InitDB(cfg *config.Config) (*sql.DB, error) {
	// Normalize Windows file path backslashes to forward slashes for URI compatibility.
	normalizedPath := strings.ReplaceAll(cfg.DBPath, "\\", "/")

	// Construct connection string: user:password@host:port/path?charset=UTF8
	// Format is compatible with github.com/nakagami/firebirdsql driver.
	dsn := fmt.Sprintf("%s:%s@%s:%s/%s?charset=UTF8",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		normalizedPath,
	)

	log.Printf("Connecting to Firebird database at %s:%s/%s...", cfg.DBHost, cfg.DBPort, normalizedPath)

	db, err := sql.Open("firebirdsql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening firebird connection: %w", err)
	}

	// Configure Connection Pooling
	db.SetMaxOpenConns(10)                  // Maximum number of open connections to the database.
	db.SetMaxIdleConns(5)                   // Maximum number of connections in the idle connection pool.
	db.SetConnMaxLifetime(1 * time.Hour)    // Maximum amount of time a connection may be reused.
	db.SetConnMaxIdleTime(15 * time.Minute) // Maximum amount of time a connection may be idle.

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("error pinging firebird database: %w", err)
	}

	log.Println("Firebird database connection established successfully")
	return db, nil
}

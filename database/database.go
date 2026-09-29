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

	// Migrate SALES_ITEMS to have CATEGORY_ID and CATEGORY_NAME
	_, _ = db.Exec("ALTER TABLE SALES_ITEMS ADD COLUMN CATEGORY_ID INTEGER DEFAULT 0")
	_, _ = db.Exec("ALTER TABLE SALES_ITEMS ADD COLUMN CATEGORY_NAME TEXT DEFAULT ''")

	// Auto-backfill CATEGORY_ID and CATEGORY_NAME in SALES_ITEMS from master ITEM & ITEM_CATEGORY for old records
	backfillQuery := `
	UPDATE SALES_ITEMS
	SET 
		CATEGORY_ID = COALESCE((SELECT i.CATEGORYID FROM ITEM i WHERE LOWER(i.ITEMNO) = LOWER(SALES_ITEMS.ITEMNO) OR (SALES_ITEMS.ITEMUPC != '' AND LOWER(i.ITEMUPC) = LOWER(SALES_ITEMS.ITEMUPC)) LIMIT 1), 0),
		CATEGORY_NAME = COALESCE((SELECT c.NAME FROM ITEM i JOIN ITEM_CATEGORY c ON i.CATEGORYID = c.ID WHERE LOWER(i.ITEMNO) = LOWER(SALES_ITEMS.ITEMNO) OR (SALES_ITEMS.ITEMUPC != '' AND LOWER(i.ITEMUPC) = LOWER(SALES_ITEMS.ITEMUPC)) LIMIT 1), 'Tanpa Kategori')
	WHERE CATEGORY_ID = 0 OR CATEGORY_NAME = '' OR CATEGORY_NAME = 'Tanpa Kategori';
	`
	_, _ = db.Exec(backfillQuery)

	// Create Indexes for High Performance Queries
	indexQueries := `
	CREATE INDEX IF NOT EXISTS idx_sales_date ON SALES(DATE);
	CREATE INDEX IF NOT EXISTS idx_sales_invoice_no ON SALES(INVOICE_NO);
	CREATE INDEX IF NOT EXISTS idx_sales_status ON SALES(STATUS);
	CREATE INDEX IF NOT EXISTS idx_sales_items_invoice ON SALES_ITEMS(INVOICE_NO);
	CREATE INDEX IF NOT EXISTS idx_sales_items_itemno ON SALES_ITEMS(ITEMNO);
	CREATE INDEX IF NOT EXISTS idx_item_itemno ON ITEM(ITEMNO);
	CREATE INDEX IF NOT EXISTS idx_item_itemupc ON ITEM(ITEMUPC);
	CREATE INDEX IF NOT EXISTS idx_item_categoryid ON ITEM(CATEGORYID);
	CREATE INDEX IF NOT EXISTS idx_stock_ledger_itemno ON STOCK_LEDGER(ITEMNO);
	CREATE INDEX IF NOT EXISTS idx_cash_movements_date ON CASH_MOVEMENTS(DATE);
	CREATE INDEX IF NOT EXISTS idx_cash_reconciliations_date ON CASH_RECONCILIATIONS(DATE);
	`
	if _, err := db.Exec(indexQueries); err != nil {
		log.Printf("[Warning] Failed to create database indexes: %v", err)
	}

	log.Println("SQLite database connection established successfully")
	return db, nil
}


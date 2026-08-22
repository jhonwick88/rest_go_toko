package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "H:\\AMAN\\toko_pintar.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("PRAGMA table_info(SALES)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("SALES Schema:")
	for rows.Next() {
		var cid int
		var name, typeName string
		var notnull int
		var dfltValue *string
		var pk int
		rows.Scan(&cid, &name, &typeName, &notnull, &dfltValue, &pk)
		fmt.Printf("- %s %s\n", name, typeName)
	}
}

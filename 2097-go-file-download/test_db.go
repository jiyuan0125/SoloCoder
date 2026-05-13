package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./test_files.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schema := `
	CREATE TABLE IF NOT EXISTS files (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		stored_path TEXT NOT NULL,
		size INTEGER NOT NULL,
		upload_time DATETIME NOT NULL,
		expire_time DATETIME NOT NULL,
		max_downloads INTEGER NOT NULL,
		current_download INTEGER NOT NULL DEFAULT 0
	);
	`
	_, err = db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := db.Query("PRAGMA table_info(files)")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("Table columns:")
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt_value sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("  %d: %s (%s)\n", cid, name, ctype)
	}
}

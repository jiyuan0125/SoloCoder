package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	sourceDB, err := sql.Open("sqlite", "source_test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sourceDB.Close()

	_, err = sourceDB.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		age INTEGER,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		log.Fatal(err)
	}

	for i := 1; i <= 20; i++ {
		_, err = sourceDB.Exec(`INSERT INTO users (name, email, age, created_at) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			20+i,
			time.Now().AddDate(0, 0, -i),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Source database created: source_test.db with 20 users")

	targetDB, err := sql.Open("sqlite", "target_test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer targetDB.Close()

	_, err = targetDB.Exec(`CREATE TABLE IF NOT EXISTS members (
		member_id INTEGER PRIMARY KEY,
		full_name TEXT NOT NULL,
		contact_email TEXT NOT NULL,
		years_old INTEGER,
		join_time TEXT
	)`)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Target database created: target_test.db")
}

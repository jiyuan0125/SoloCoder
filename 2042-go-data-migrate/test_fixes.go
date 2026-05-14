package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	os.Remove("source_r1.db")
	os.Remove("target_r1.db")
	os.Remove("data_migrate_r1.db")

	sourceDB, err := sql.Open("sqlite", "source_r1.db")
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

	now := time.Now()
	for i := 1; i <= 10; i++ {
		_, err = sourceDB.Exec(`INSERT INTO users (name, email, age, created_at) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			20+i,
			now.AddDate(0, 0, -i),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Source database created: source_r1.db with 10 users")

	targetDB, err := sql.Open("sqlite", "target_r1.db")
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

	fmt.Println("Target database created: target_r1.db")
	fmt.Println("\nTest files ready!")
	fmt.Println("Port: 8101")
	fmt.Println("\nNow you can test:")
	fmt.Println("1. Create config and task - should migrate 10 records")
	fmt.Println("2. Add 5 more records to source")
	fmt.Println("3. Create new task - should only migrate 5 new records (incremental)")
}

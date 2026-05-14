package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

func main() {
	sourceDB, err := sql.Open("sqlite", "source_r1.db")
	if err != nil {
		log.Fatal(err)
	}
	defer sourceDB.Close()

	now := time.Now()
	for i := 11; i <= 15; i++ {
		_, err = sourceDB.Exec(`INSERT INTO users (name, email, age, created_at) VALUES (?, ?, ?, ?)`,
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			20+i,
			now.Add(time.Duration(i-10)*time.Second),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("Added 5 new records to source database (ID 11-15)")
}

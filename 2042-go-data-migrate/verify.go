package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

func main() {
	targetDB, err := sql.Open("sqlite", "target_test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer targetDB.Close()

	rows, err := targetDB.Query("SELECT * FROM members ORDER BY member_id")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	fmt.Println("=== Migrated Data in Target Database ===")
	fmt.Println("member_id | full_name | contact_email | years_old | join_time")
	fmt.Println("----------|-----------|---------------|-----------|----------")
	
	for rows.Next() {
		var id int
		var name, email, joinTime string
		var age sql.NullInt64
		if err := rows.Scan(&id, &name, &email, &age, &joinTime); err != nil {
			log.Fatal(err)
		}
		ageVal := "NULL"
		if age.Valid {
			ageVal = fmt.Sprintf("%d", age.Int64)
		}
		fmt.Printf("%9d | %-9s | %-25s | %-9s | %s", id, name, email, ageVal, joinTime)
	}
}

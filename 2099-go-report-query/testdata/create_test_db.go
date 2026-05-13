package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`DROP TABLE IF EXISTS employees`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE employees (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			age INTEGER,
			department TEXT,
			salary REAL,
			city TEXT,
			hire_date TEXT
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	employees := []struct {
		name       string
		age        int
		department string
		salary     float64
		city       string
		hireDate   string
	}{
		{"张三", 28, "研发部", 15000.50, "北京", "2020-03-15"},
		{"李四", 32, "市场部", 12000.00, "上海", "2019-07-20"},
		{"王五", 25, "研发部", 13500.00, "北京", "2021-01-10"},
		{"赵六", 35, "人事部", 18000.00, "广州", "2018-05-05"},
		{"钱七", 29, "研发部", 16000.00, "深圳", "2020-11-30"},
		{"孙八", 27, "市场部", 11000.00, "上海", "2021-06-18"},
		{"周九", 31, "研发部", 17000.00, "北京", "2019-09-22"},
		{"吴十", 26, "人事部", 14000.00, "广州", "2021-04-08"},
	}

	stmt, err := db.Prepare(`
		INSERT INTO employees (name, age, department, salary, city, hire_date)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	for _, emp := range employees {
		_, err = stmt.Exec(emp.name, emp.age, emp.department, emp.salary, emp.city, emp.hireDate)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("测试数据库已创建: test.db")
	fmt.Println("表名: employees")
	fmt.Println("数据已插入: 8 条员工记录")
}

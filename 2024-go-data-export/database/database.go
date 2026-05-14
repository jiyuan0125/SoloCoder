package database

import (
	"database/sql"
	"log"
	"os"
	"time"

	"data-export/config"
	"data-export/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() {
	os.MkdirAll(config.ExportDir, 0755)

	dsn := config.DBPath + "?_journal=WAL&_busy_timeout=5000&_fk=1"

	var err error
	DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get sql.DB:", err)
	}

	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := enableWAL(sqlDB); err != nil {
		log.Printf("Warning: failed to enable WAL: %v", err)
	}

	DB.AutoMigrate(&models.ExportTask{})

	createSampleTables()
}

func enableWAL(sqlDB *sql.DB) error {
	_, err := sqlDB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("PRAGMA synchronous=NORMAL;")
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec("PRAGMA busy_timeout=5000;")
	return err
}

func createSampleTables() {
	DB.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		phone TEXT,
		age INTEGER,
		created_at DATETIME
	)`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER,
		order_no TEXT,
		amount REAL,
		status TEXT,
		created_at DATETIME
	)`)

	var count int64
	DB.Table("users").Count(&count)
	if count == 0 {
		now := time.Now()
		users := []map[string]interface{}{
			{"name": "张三", "email": "zhangsan@example.com", "phone": "13812345678", "age": 25, "created_at": now},
			{"name": "李四", "email": "lisi@test.com", "phone": "13987654321", "age": 30, "created_at": now},
			{"name": "王五", "email": "wangwu@company.org", "phone": "13611112222", "age": 28, "created_at": now},
			{"name": "赵六", "email": "zhaoliu@mail.cn", "phone": "13733334444", "age": 35, "created_at": now},
			{"name": "钱七", "email": "qianqi@example.com", "phone": "13555556666", "age": 22, "created_at": now},
		}
		for _, u := range users {
			DB.Table("users").Create(u)
		}

		orders := []map[string]interface{}{
			{"user_id": 1, "order_no": "ORD2024001", "amount": 199.99, "status": "paid", "created_at": now},
			{"user_id": 2, "order_no": "ORD2024002", "amount": 599.00, "status": "pending", "created_at": now},
			{"user_id": 1, "order_no": "ORD2024003", "amount": 89.50, "status": "paid", "created_at": now},
		}
		for _, o := range orders {
			DB.Table("orders").Create(o)
		}
		log.Println("Sample data inserted")
	}
}

func TableExists(tableName string) bool {
	var count int
	DB.Raw("SELECT count(*) FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&count)
	return count > 0
}

func GetTableColumns(tableName string) ([]string, error) {
	rows, err := DB.Raw("PRAGMA table_info(" + tableName + ")").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt_value *string
		rows.Scan(&cid, &name, &ctype, &notnull, &dflt_value, &pk)
		columns = append(columns, name)
	}
	return columns, nil
}

func CountRows(tableName string, filter string) (int64, error) {
	query := "SELECT COUNT(*) FROM " + tableName
	if filter != "" {
		query += " WHERE " + filter
	}
	var count int64
	err := DB.Raw(query).Scan(&count).Error
	return count, err
}

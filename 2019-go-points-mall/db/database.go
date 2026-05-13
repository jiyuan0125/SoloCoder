package db

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL&_busy_timeout=10000")
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(1)

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = runMigrations(); err != nil {
		return err
	}

	log.Println("数据库初始化完成")
	return nil
}

func runMigrations() error {
	schema, err := os.ReadFile(filepath.Join("db", "schema.sql"))
	if err != nil {
		return err
	}

	_, err = DB.Exec(string(schema))
	return err
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

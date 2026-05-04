package database

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = createTables(); err != nil {
		return err
	}

	if err = createDefaultAdmin(); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

func createTables() error {
	createFeedbackTable := `
	CREATE TABLE IF NOT EXISTS feedbacks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id TEXT NOT NULL,
		feedback_type TEXT NOT NULL CHECK(feedback_type IN ('feature', 'bug', 'complaint')),
		rating INTEGER NOT NULL CHECK(rating >= 1 AND rating <= 5),
		description TEXT NOT NULL,
		status TEXT NOT NULL DEFAULT 'pending' CHECK(status IN ('pending', 'processing', 'closed')),
		internal_note TEXT DEFAULT '',
		processing_note TEXT DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createAdminTable := `
	CREATE TABLE IF NOT EXISTS admins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createFeedbackIndex := `
	CREATE INDEX IF NOT EXISTS idx_feedbacks_user_id ON feedbacks(user_id);
	CREATE INDEX IF NOT EXISTS idx_feedbacks_type ON feedbacks(feedback_type);
	CREATE INDEX IF NOT EXISTS idx_feedbacks_status ON feedbacks(status);
	CREATE INDEX IF NOT EXISTS idx_feedbacks_created_at ON feedbacks(created_at);
	`

	if _, err := DB.Exec(createFeedbackTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createAdminTable); err != nil {
		return err
	}

	if _, err := DB.Exec(createFeedbackIndex); err != nil {
		return err
	}

	return nil
}

func createDefaultAdmin() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		defaultUsername := "admin"
		defaultPassword := "admin123"

		_, err = DB.Exec(
			"INSERT INTO admins (username, password) VALUES (?, ?)",
			defaultUsername,
			defaultPassword,
		)
		if err != nil {
			return err
		}

		log.Println("Default admin user created: username=admin, password=admin123")
	}

	return nil
}

func CloseDB() {
	if DB != nil {
		DB.Close()
	}
}

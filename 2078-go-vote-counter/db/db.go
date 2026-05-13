package db

import (
	"database/sql"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

var (
	DB    *sql.DB
	mutex sync.Mutex
)

type Poll struct {
	ID            int64
	Title         string
	IsSingleChoice bool
	MaxChoices    int
	IsAnonymous   bool
	Deadline      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Option struct {
	ID    int64
	PollID int64
	Text  string
	Votes int
}

type Vote struct {
	ID         int64
	PollID     int64
	ParticipantID string
	OptionIDs  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	if err = DB.Ping(); err != nil {
		return err
	}

	if err = createTables(); err != nil {
		return err
	}

	return nil
}

func CloseDB() error {
	return DB.Close()
}

func createTables() error {
	createPollTable := `
	CREATE TABLE IF NOT EXISTS polls (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		is_single_choice BOOLEAN NOT NULL,
		max_choices INTEGER NOT NULL,
		is_anonymous BOOLEAN NOT NULL,
		deadline DATETIME NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);`

	createOptionTable := `
	CREATE TABLE IF NOT EXISTS options (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		poll_id INTEGER NOT NULL,
		text TEXT NOT NULL,
		votes INTEGER DEFAULT 0,
		FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE
	);`

	createVoteTable := `
	CREATE TABLE IF NOT EXISTS votes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		poll_id INTEGER NOT NULL,
		participant_id TEXT NOT NULL,
		option_ids TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE
	);`

	createVoteIndex := `
	CREATE INDEX IF NOT EXISTS idx_votes_poll_participant ON votes(poll_id, participant_id);
	`

	for _, query := range []string{createPollTable, createOptionTable, createVoteTable, createVoteIndex} {
		if _, err := DB.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func Lock() {
	mutex.Lock()
}

func Unlock() {
	mutex.Unlock()
}

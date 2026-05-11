package errtest

import "errors"

type DB struct {
	dsn string
}

func NewDB(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, errors.New("empty dsn")
	}
	return &DB{dsn: dsn}, nil
}

type Service struct {
	db *DB
}

func NewService(db *DB) *Service {
	return &Service{db: db}
}

package config

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"confman/pkg/model"
)

func InitDB(dbPath string) (*gorm.DB, error) {
	log.Printf("Initializing database at %s", dbPath)
	
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	log.Println("Running database migrations...")
	
	err = db.AutoMigrate(
		&model.User{},
		&model.Meeting{},
		&model.Paper{},
		&model.PaperAuthor{},
		&model.Review{},
		&model.Session{},
		&model.Schedule{},
		&model.Reminder{},
		&model.AuditLog{},
	)
	if err != nil {
		return nil, err
	}

	log.Println("Database migrations completed")
	return db, nil
}

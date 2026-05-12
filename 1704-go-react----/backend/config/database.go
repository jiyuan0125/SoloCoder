package config

import (
	"log"
	"os"

	"rehab-system/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./rehab.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")

	DB.AutoMigrate(
		&models.Patient{},
		&models.Plan{},
		&models.Exercise{},
		&models.Task{},
		&models.TrainingRecord{},
		&models.DifficultyLog{},
		&models.Assessment{},
		&models.AssessmentIndicator{},
	)

	log.Println("Database migration completed")
}

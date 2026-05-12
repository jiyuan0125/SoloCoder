package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"trial-management-system/internal/models"
)

var DB *gorm.DB

func Init(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(
		&models.Protocol{},
		&models.Site{},
		&models.Visit{},
		&models.Subject{},
		&models.VisitRecord{},
		&models.AdverseEvent{},
		&models.Todo{},
		&models.Department{},
		&models.Budget{},
		&models.BudgetItem{},
		&models.BudgetAlert{},
		&models.SystemA{},
		&models.SystemB{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database initialized successfully")
}

package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"smart-exam/models"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("exam.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	err = DB.AutoMigrate(
		&models.KnowledgePoint{},
		&models.Question{},
		&models.Exam{},
		&models.ExamQuestion{},
		&models.WrongAnswer{},
		&models.KnowledgeMastery{},
		&models.Bill{},
		&models.BillItem{},
	)
	if err != nil {
		log.Fatal("failed to migrate database")
	}

	log.Println("Database initialized successfully")
}

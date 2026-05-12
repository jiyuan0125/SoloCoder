package db

import (
	"log"

	"medical-quality-system/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("medical_quality.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	err = DB.AutoMigrate(
		&model.Indicator{},
		&model.IndicatorTargetHistory{},
		&model.IndicatorData{},
		&model.PDCA{},
		&model.PDCAPhaseDetail{},
		&model.Todo{},
		&model.Meeting{},
		&model.ActionItem{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database initialized successfully")
}

package database

import (
	"hospital-pharmacy/pkg/models"
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("pharmacy.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&models.DrugCategory{},
		&models.Drug{},
		&models.StockItem{},
		&models.StockTransaction{},
		&models.StockAlert{},
		&models.Prescription{},
		&models.PrescriptionItem{},
	)
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

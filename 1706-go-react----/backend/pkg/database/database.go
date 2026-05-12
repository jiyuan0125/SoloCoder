package database

import (
	"hospital-pharmacy/pkg/models"
	"log"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init() error {
	var err error
	DB, err = gorm.Open(sqlite.Open("pharmacy.db?_busy_timeout=5000&_journal_mode=WAL"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

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

	log.Println("Database initialized successfully (WAL mode, single connection)")
	return nil
}

package database

import (
	"medical-device-manager/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath string) error {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	DB = db

	err = db.AutoMigrate(
		&models.Device{},
		&models.CalibrationAgency{},
		&models.CalibrationRecord{},
		&models.MaintenancePlan{},
		&models.WorkOrder{},
	)
	if err != nil {
		return err
	}

	return nil
}

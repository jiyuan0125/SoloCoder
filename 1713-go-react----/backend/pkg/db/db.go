package db

import (
	"ohims/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&models.Enterprise{},
		&models.HazardFactor{},
		&models.Worker{},
		&models.WorkerHazardFactor{},
		&models.Examination{},
		&models.ExamItemResult{},
		&models.TodoItem{},
		&models.EnterpriseReport{},
		&models.StatisticsCache{},
	)
	return err
}

func GetDB() *gorm.DB {
	return DB
}

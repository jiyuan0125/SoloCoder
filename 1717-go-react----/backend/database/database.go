package database

import (
	"log"

	"hospital-infection/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	var err error
	DB, err = gorm.Open(sqlite.Open("hospital_infection.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = DB.AutoMigrate(
		&models.Department{},
		&models.Patient{},
		&models.InfectionCase{},
		&models.PreventionMeasure{},
		&models.TargetMonitoring{},
		&models.Alert{},
		&models.Report{},
		&models.ApprovalHistory{},
		&models.PriceHistory{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&models.Department{}).Count(&count)
	if count > 0 {
		return
	}

	departments := []models.Department{
		{Name: "内科"},
		{Name: "外科"},
		{Name: "儿科"},
		{Name: "妇产科"},
		{Name: "ICU"},
		{Name: "新生儿科"},
		{Name: "烧伤科"},
		{Name: "血液科"},
		{Name: "骨科"},
		{Name: "感染管理科"},
	}

	for _, dept := range departments {
		DB.Create(&dept)
	}

	log.Println("Database seeded with departments")
}

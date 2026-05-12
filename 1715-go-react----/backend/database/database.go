package database

import (
	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"epidemic-management/models"
)

var DB *gorm.DB

func Init() {
	var err error
	DB, err = gorm.Open(sqlite.Open("epidemic.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database:", err)
	}
	DB.AutoMigrate(
		&models.Disease{},
		&models.OutbreakReport{},
		&models.Investigation{},
		&models.Contact{},
		&models.Vaccine{},
		&models.VaccinationRecord{},
		&models.Todo{},
		&models.Alert{},
		&models.WeeklyReport{},
	)
	seedData()
}

func seedData() {
	var count int64
	DB.Model(&models.Disease{}).Count(&count)
	if count > 0 {
		return
	}
	diseases := []models.Disease{
		{Name: "鼠疫", Class: models.ClassA, MinIncubation: 1, MaxIncubation: 6, ReportDeadline: 2},
		{Name: "霍乱", Class: models.ClassA, MinIncubation: 1, MaxIncubation: 5, ReportDeadline: 2},
		{Name: "传染性非典型肺炎", Class: models.ClassB, MinIncubation: 2, MaxIncubation: 14, ReportDeadline: 24},
		{Name: "艾滋病", Class: models.ClassB, MinIncubation: 90, MaxIncubation: 365, ReportDeadline: 24},
		{Name: "病毒性肝炎", Class: models.ClassB, MinIncubation: 15, MaxIncubation: 180, ReportDeadline: 24},
		{Name: "脊髓灰质炎", Class: models.ClassB, MinIncubation: 3, MaxIncubation: 35, ReportDeadline: 24},
		{Name: "人感染高致病性禽流感", Class: models.ClassB, MinIncubation: 1, MaxIncubation: 7, ReportDeadline: 24},
		{Name: "流行性感冒", Class: models.ClassC, MinIncubation: 1, MaxIncubation: 4, ReportDeadline: 24},
		{Name: "流行性腮腺炎", Class: models.ClassC, MinIncubation: 8, MaxIncubation: 30, ReportDeadline: 24},
		{Name: "风疹", Class: models.ClassC, MinIncubation: 14, MaxIncubation: 21, ReportDeadline: 24},
		{Name: "新冠肺炎", Class: models.ClassB, MinIncubation: 2, MaxIncubation: 14, ReportDeadline: 24},
		{Name: "乙肝", Class: models.ClassB, MinIncubation: 45, MaxIncubation: 160, ReportDeadline: 24},
	}
	for _, d := range diseases {
		DB.Create(&d)
	}

	vaccines := []models.Vaccine{
		{Name: "乙肝疫苗", Manufacturer: "某制药", BatchNumber: "B2024001", Schedule: "0-1-6", Doses: 3, IntervalDays: "0,30,180"},
		{Name: "流感疫苗", Manufacturer: "某生物", BatchNumber: "F2024001", Schedule: "每年1剂", Doses: 1, IntervalDays: "0"},
		{Name: "新冠疫苗", Manufacturer: "某科兴", BatchNumber: "C2024001", Schedule: "0-28", Doses: 2, IntervalDays: "0,28"},
	}
	for _, v := range vaccines {
		DB.Create(&v)
	}
}

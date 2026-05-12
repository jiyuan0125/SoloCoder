package storage

import (
	"fmt"
	"log"
	"time"

	"health-supervision-system/internal/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = autoMigrate()
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	err = seedTemplates()
	if err != nil {
		return fmt.Errorf("failed to seed templates: %w", err)
	}

	return nil
}

func autoMigrate() error {
	return DB.AutoMigrate(
		&models.SupervisedUnit{},
		&models.InspectionItemTemplate{},
		&models.InspectionRecord{},
		&models.InspectionItem{},
		&models.HealthOpinion{},
		&models.Penalty{},
		&models.Complaint{},
		&models.PublicNotice{},
		&models.AuditLog{},
	)
}

func seedTemplates() error {
	var count int64
	DB.Model(&models.InspectionItemTemplate{}).Count(&count)
	if count > 0 {
		return nil
	}

	templates := []models.InspectionItemTemplate{
		{UnitType: models.UnitTypeHospital, ItemName: "传染病防控"},
		{UnitType: models.UnitTypeHospital, ItemName: "消毒隔离"},
		{UnitType: models.UnitTypeHospital, ItemName: "医疗废物处理"},
		{UnitType: models.UnitTypeHospital, ItemName: "放射诊疗管理"},
		{UnitType: models.UnitTypeClinic, ItemName: "传染病防控"},
		{UnitType: models.UnitTypeClinic, ItemName: "消毒隔离"},
		{UnitType: models.UnitTypeClinic, ItemName: "医疗废物处理"},
		{UnitType: models.UnitTypeBeautyShop, ItemName: "卫生设施"},
		{UnitType: models.UnitTypeBeautyShop, ItemName: "消毒设备运行"},
		{UnitType: models.UnitTypeBeautyShop, ItemName: "用品用具消毒"},
		{UnitType: models.UnitTypeBeautyShop, ItemName: "从业人员健康证"},
		{UnitType: models.UnitTypeHotel, ItemName: "客房卫生"},
		{UnitType: models.UnitTypeHotel, ItemName: "消毒设施"},
		{UnitType: models.UnitTypeHotel, ItemName: "饮用水卫生"},
		{UnitType: models.UnitTypeHotel, ItemName: "公共用品消毒"},
		{UnitType: models.UnitTypeSwimmingPool, ItemName: "水质检测"},
		{UnitType: models.UnitTypeSwimmingPool, ItemName: "池水更换记录"},
		{UnitType: models.UnitTypeSwimmingPool, ItemName: "强制淋浴设施"},
		{UnitType: models.UnitTypeSwimmingPool, ItemName: "消毒设备运行情况"},
		{UnitType: models.UnitTypeSchool, ItemName: "教学环境卫生"},
		{UnitType: models.UnitTypeSchool, ItemName: "饮用水卫生"},
		{UnitType: models.UnitTypeSchool, ItemName: "传染病防控"},
		{UnitType: models.UnitTypeSchool, ItemName: "教室采光照明"},
		{UnitType: models.UnitTypeWaterSupply, ItemName: "水质检测"},
		{UnitType: models.UnitTypeWaterSupply, ItemName: "水源防护"},
		{UnitType: models.UnitTypeWaterSupply, ItemName: "消毒设施运行"},
		{UnitType: models.UnitTypeWaterSupply, ItemName: "从业人员健康证"},
	}

	for _, t := range templates {
		if err := DB.Create(&t).Error; err != nil {
			log.Printf("Warning: failed to create template %s: %v", t.ItemName, err)
		}
	}

	return nil
}

func CreateAuditLog(userID uint, username, action, resource string, resourceID *uint, description, ipAddress string) {
	log := models.AuditLog{
		UserID:      userID,
		Username:    username,
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		IPAddress:  ipAddress,
		CreatedAt:  time.Now(),
	}
	DB.Create(&log)
}

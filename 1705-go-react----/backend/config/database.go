package config

import (
	"telemedicine/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(sqlite.Open("telemedicine.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	DB.AutoMigrate(
		&models.User{},
		&models.Consultation{},
		&models.Patient{},
		&models.Exam{},
		&models.Message{},
		&models.Prescription{},
		&models.PrescriptionItem{},
		&models.ForbiddenDrug{},
		&models.ImageTemplate{},
		&models.ImageData{},
		&models.DrugPriceHistory{},
	)

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&models.ForbiddenDrug{}).Count(&count)
	if count == 0 {
		forbiddenDrugs := []models.ForbiddenDrug{
			{Name: "阿莫西林", Category: "抗生素", Reason: "远程问诊禁止使用抗生素"},
			{Name: "头孢氨苄", Category: "抗生素", Reason: "远程问诊禁止使用抗生素"},
			{Name: "安定", Category: "精神类", Reason: "远程问诊禁止使用精神类药品"},
			{Name: "阿普唑仑", Category: "精神类", Reason: "远程问诊禁止使用精神类药品"},
		}
		DB.Create(&forbiddenDrugs)
	}

	DB.Model(&models.ImageTemplate{}).Count(&count)
	if count == 0 {
		templates := []models.ImageTemplate{
			{
				Type:     "胸部X光",
				Fields:   "肺部纹理,心影大小,纵隔形态,膈面情况,其他描述",
				IsActive: true,
			},
			{
				Type:     "CT",
				Fields:   "检查部位,层面描述,病灶位置,大小形态,密度特征,增强表现,周围组织,其他描述",
				IsActive: true,
			},
			{
				Type:     "MRI",
				Fields:   "检查部位,扫描序列,信号特征,病灶位置,大小形态,周围组织,其他描述",
				IsActive: true,
			},
		}
		DB.Create(&templates)
	}

	DB.Model(&models.User{}).Count(&count)
	if count == 0 {
		users := []models.User{
			{Username: "grassroot1", Password: "$2a$10$zfcBX8lLe5.BKMUUGRhyE.TKeEEJf0PsjYMczdYcElSCLilhHx9R2", Name: "基层医生张", Role: "grassroot"},
			{Username: "expert1", Password: "$2a$10$zfcBX8lLe5.BKMUUGRhyE.TKeEEJf0PsjYMczdYcElSCLilhHx9R2", Name: "专家李", Role: "expert"},
			{Username: "admin", Password: "$2a$10$zfcBX8lLe5.BKMUUGRhyE.TKeEEJf0PsjYMczdYcElSCLilhHx9R2", Name: "管理员", Role: "admin"},
		}
		DB.Create(&users)
	}
}

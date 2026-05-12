package repositories

import (
	"hospital-pharmacy/pkg/database"
	"hospital-pharmacy/pkg/models"

	"gorm.io/gorm"
)

func CreateDrug(drug *models.Drug) error {
	return database.DB.Create(drug).Error
}

func GetDrugByID(id uint) (*models.Drug, error) {
	var drug models.Drug
	err := database.DB.Preload("Category").First(&drug, id).Error
	return &drug, err
}

func GetDrugByCode(code string) (*models.Drug, error) {
	var drug models.Drug
	err := database.DB.Preload("Category").Where("drug_code = ?", code).First(&drug).Error
	return &drug, err
}

func ListDrugs(search string) ([]models.Drug, error) {
	var drugs []models.Drug
	query := database.DB.Preload("Category")
	if search != "" {
		query = query.Where("generic_name LIKE ? OR drug_code LIKE ? OR brand_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	err := query.Find(&drugs).Error
	return drugs, err
}

func UpdateDrug(drug *models.Drug) error {
	return database.DB.Save(drug).Error
}

func UpdateDrugWithTx(tx *gorm.DB, drug *models.Drug) error {
	return tx.Save(drug).Error
}

func DeleteDrug(id uint) error {
	return database.DB.Delete(&models.Drug{}, id).Error
}

func CreateCategory(category *models.DrugCategory) error {
	return database.DB.Create(category).Error
}

func ListCategories() ([]models.DrugCategory, error) {
	var categories []models.DrugCategory
	err := database.DB.Find(&categories).Error
	return categories, err
}

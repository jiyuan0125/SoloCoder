package services

import (
	"errors"
	"hospital-pharmacy/pkg/models"
	"hospital-pharmacy/pkg/repositories"
	"hospital-pharmacy/pkg/utils"
)

func CreateDrug(drug *models.Drug) error {
	if err := utils.ValidateDrugCode(drug.DrugCode); err != nil {
		return err
	}
	if err := utils.ValidatePrices(drug.RetailPrice, drug.PurchasePrice); err != nil {
		return err
	}
	if err := utils.ValidateStockQuantity(drug.CurrentStock); err != nil {
		return err
	}

	existing, _ := repositories.GetDrugByCode(drug.DrugCode)
	if existing.ID > 0 {
		return errors.New("药品编码已存在")
	}

	return repositories.CreateDrug(drug)
}

func UpdateDrug(drug *models.Drug) error {
	existing, err := repositories.GetDrugByID(drug.ID)
	if err != nil {
		return errors.New("药品不存在")
	}

	if drug.DrugCode != existing.DrugCode {
		if err := utils.ValidateDrugCode(drug.DrugCode); err != nil {
			return err
		}
	}
	if err := utils.ValidatePrices(drug.RetailPrice, drug.PurchasePrice); err != nil {
		return err
	}

	return repositories.UpdateDrug(drug)
}

func GetDrug(id uint) (*models.Drug, error) {
	return repositories.GetDrugByID(id)
}

func ListDrugs(search string) ([]models.Drug, error) {
	return repositories.ListDrugs(search)
}

func DeleteDrug(id uint) error {
	return repositories.DeleteDrug(id)
}

func CreateCategory(category *models.DrugCategory) error {
	return repositories.CreateCategory(category)
}

func ListCategories() ([]models.DrugCategory, error) {
	return repositories.ListCategories()
}

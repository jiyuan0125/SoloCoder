package repositories

import (
	"hospital-pharmacy/pkg/database"
	"hospital-pharmacy/pkg/models"
)

func CreatePrescription(prescription *models.Prescription) error {
	return database.DB.Create(prescription).Error
}

func GetPrescriptionByID(id uint) (*models.Prescription, error) {
	var prescription models.Prescription
	err := database.DB.Preload("Items.Drug").First(&prescription, id).Error
	return &prescription, err
}

func ListPrescriptions(status string) ([]models.Prescription, error) {
	var prescriptions []models.Prescription
	query := database.DB.Preload("Items")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Find(&prescriptions).Error
	return prescriptions, err
}

func UpdatePrescription(prescription *models.Prescription) error {
	return database.DB.Save(prescription).Error
}

func GetLatestPrescriptionSeq(today string) (int64, error) {
	var count int64
	pattern := "CF" + today + "%"
	err := database.DB.Model(&models.Prescription{}).
		Where("prescription_no LIKE ?", pattern).
		Count(&count).Error
	return count, err
}

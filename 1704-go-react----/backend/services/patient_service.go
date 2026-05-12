package services

import (
	"rehab-system/config"
	"rehab-system/models"
	"time"

	"gorm.io/gorm"
)

type PatientService struct{}

func NewPatientService() *PatientService {
	return &PatientService{}
}

func (s *PatientService) List() ([]models.Patient, error) {
	var patients []models.Patient
	err := config.DB.Where("status = ?", "active").Find(&patients).Error
	return patients, err
}

func (s *PatientService) Get(id uint) (*models.Patient, error) {
	var patient models.Patient
	err := config.DB.Preload("Plans").First(&patient, id).Error
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (s *PatientService) Create(patient *models.Patient) error {
	patient.CreatedAt = time.Now()
	patient.UpdatedAt = time.Now()
	patient.Status = "active"
	return config.DB.Create(patient).Error
}

func (s *PatientService) Update(id uint, patient *models.Patient) (*models.Patient, error) {
	var existing models.Patient
	if err := config.DB.First(&existing, id).Error; err != nil {
		return nil, err
	}

	existing.Name = patient.Name
	existing.BirthDate = patient.BirthDate
	existing.Gender = patient.Gender
	existing.Phone = patient.Phone
	existing.UpdatedAt = time.Now()

	if err := config.DB.Save(&existing).Error; err != nil {
		return nil, err
	}
	return &existing, nil
}

func (s *PatientService) Delete(id uint) error {
	return config.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Delete(&models.Patient{}, id).Error; err != nil {
			return err
		}
		return tx.Where("patient_id = ?", id).Delete(&models.Plan{}).Error
	})
}

func (s *PatientService) GetActivePlans(patientID uint) ([]models.Plan, error) {
	var plans []models.Plan
	err := config.DB.Preload("Exercises").
		Where("patient_id = ? AND status = ?", patientID, "active").
		Find(&plans).Error
	return plans, err
}

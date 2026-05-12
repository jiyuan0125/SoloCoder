package services

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"epidemic-management/database"
	"epidemic-management/models"
)

func ListVaccines() ([]models.Vaccine, error) {
	var vaccines []models.Vaccine
	err := database.DB.Find(&vaccines).Error
	return vaccines, err
}

func CreateVaccine(vaccine *models.Vaccine) error {
	return database.DB.Create(vaccine).Error
}

func RecordVaccination(record *models.VaccinationRecord) error {
	var vaccine models.Vaccine
	if err := database.DB.Where("name = ?", record.VaccineName).First(&vaccine).Error; err != nil {
		return fmt.Errorf("未找到该疫苗信息")
	}

	var existing models.VaccinationRecord
	err := database.DB.Where("recipient_id = ? AND vaccine_name = ? AND dose_number = ?",
		record.RecipientID, record.VaccineName, record.DoseNumber).First(&existing).Error
	if err == nil {
		return errors.New("该剂次已接种")
	}

	if !isWithinTimeWindow(record, &vaccine, record.VaccinationDate) {
		return fmt.Errorf("超出允许时间范围")
	}

	return database.DB.Create(record).Error
}

func isWithinTimeWindow(record *models.VaccinationRecord, vaccine *models.Vaccine, vacDate time.Time) bool {
	if record.DoseNumber == 1 {
		return true
	}

	intervals := strings.Split(vaccine.IntervalDays, ",")
	if len(intervals) < record.DoseNumber-1 {
		return true
	}

	var prevRecord models.VaccinationRecord
	err := database.DB.Where("recipient_id = ? AND vaccine_name = ? AND dose_number = ?",
		record.RecipientID, record.VaccineName, record.DoseNumber-1).Order("vaccination_date desc").First(&prevRecord).Error
	if err != nil {
		return true
	}

	intervalDays, _ := strconv.Atoi(intervals[record.DoseNumber-1])
	expectedDate := prevRecord.VaccinationDate.AddDate(0, 0, intervalDays)

	earliest := expectedDate.AddDate(0, 0, -7)
	latest := expectedDate.AddDate(0, 0, 30)

	return !vacDate.Before(earliest) && !vacDate.After(latest)
}

func ListVaccinationRecords(recipientID string) ([]models.VaccinationRecord, error) {
	var records []models.VaccinationRecord
	query := database.DB.Order("vaccination_date desc")
	if recipientID != "" {
		query = query.Where("recipient_id = ?", recipientID)
	}
	err := query.Find(&records).Error
	return records, err
}

func ListAllVaccinationRecords() ([]models.VaccinationRecord, error) {
	var records []models.VaccinationRecord
	err := database.DB.Order("vaccination_date desc").Find(&records).Error
	return records, err
}

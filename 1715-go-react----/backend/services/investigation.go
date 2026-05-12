package services

import (
	"fmt"
	"time"

	"epidemic-management/database"
	"epidemic-management/models"
)

func CreateInvestigation(inv *models.Investigation) error {
	if err := database.DB.Create(inv).Error; err != nil {
		return err
	}
	var report models.OutbreakReport
	database.DB.First(&report, inv.ReportID)
	report.InvestigationID = &inv.ID
	report.Status = "investigating"
	database.DB.Save(&report)
	return nil
}

func ListInvestigations() ([]models.Investigation, error) {
	var invs []models.Investigation
	err := database.DB.Order("created_at desc").Find(&invs).Error
	return invs, err
}

func GetInvestigation(id uint) (*models.Investigation, error) {
	var inv models.Investigation
	err := database.DB.First(&inv, id).Error
	return &inv, err
}

func CreateContact(contact *models.Contact, diseaseName string) error {
	var disease models.Disease
	if err := database.DB.Where("name = ?", diseaseName).First(&disease).Error; err != nil {
		return fmt.Errorf("未找到该传染病信息")
	}

	contact.ObservationStart = contact.LastContactDate.AddDate(0, 0, 1)
	contact.ObservationEnd = contact.LastContactDate.AddDate(0, 0, disease.MaxIncubation)
	contact.Status = models.StatusNormal

	if err := database.DB.Create(contact).Error; err != nil {
		return err
	}

	CreateTodo(&models.Todo{
		Title:       "密接者追踪",
		Description: fmt.Sprintf("新密接者: %s", contact.Name),
		Priority:    models.PriorityHigh,
		Status:    models.TodoPending,
		RelatedType: "contact",
		RelatedID: contact.ID,
	})

	return nil
}

func ListContactsByInvestigation(invID uint) ([]models.Contact, error) {
	var contacts []models.Contact
	err := database.DB.Where("investigation_id = ?", invID).Order("created_at desc").Find(&contacts).Error
	return contacts, err
}

func GetContact(id uint) (*models.Contact, error) {
	var contact models.Contact
	err := database.DB.First(&contact, id).Error
	return &contact, err
}

func UpdateContactStatus(contactID uint, newStatus models.ContactStatus) error {
	var contact models.Contact
	if err := database.DB.First(&contact, contactID).Error; err != nil {
		return err
	}

	if !isValidStatusTransition(contact.Status, newStatus) {
		return fmt.Errorf("无效的状态流转")
	}

	oldStatus := contact.Status
	contact.Status = newStatus

	if newStatus == models.StatusSymptoms {
		CreateTodo(&models.Todo{
			Title:       "采样送检",
			Description: fmt.Sprintf("密接者%s出现症状", contact.Name),
			Priority:    models.PriorityHigh,
			Status:    models.TodoPending,
			RelatedType: "contact",
			RelatedID: contact.ID,
		})
	}

	if newStatus == models.StatusConfirmed && !contact.IsNewCase == false {
		contact.IsNewCase = true

		var inv models.Investigation
		database.DB.First(&inv, contact.InvestigationID)

		var report models.OutbreakReport
		database.DB.First(&report, inv.ReportID)

		newReport := &models.OutbreakReport{
			PatientName: contact.Name,
			PatientID:   fmt.Sprintf("CONV-%d", contact.ID),
			DiseaseName: report.DiseaseName,
			Region:    report.Region,
			District:  report.District,
			OnsetTime: time.Now(),
			ReportTime: time.Now(),
		}
		CreateOutbreakReport(newReport)

		var subContacts []models.Contact
		database.DB.Where("parent_contact_id = ?", contact.ID).Find(&subContacts)
		for _, sc := range subContacts {
			CreateTodo(&models.Todo{
				Title:       "密接者追踪",
				Description: fmt.Sprintf("次级密接者: %s", sc.Name),
				Priority:    models.PriorityHigh,
				Status:    models.TodoPending,
				RelatedType: "contact",
				RelatedID: sc.ID,
			})
		}
	}

	database.DB.Save(&contact)

	_ = oldStatus
	return nil
}

func isValidStatusTransition(current, next models.ContactStatus) bool {
	validTransitions := map[models.ContactStatus][]models.ContactStatus{
		models.StatusNormal:   {models.StatusSymptoms, models.StatusExcluded},
		models.StatusSymptoms: {models.StatusConfirmed, models.StatusExcluded},
		models.StatusConfirmed: {},
		models.StatusExcluded: {},
	}
	allowed, exists := validTransitions[current]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

func CheckAndExpireContacts() error {
	now := time.Now()
	var contacts []models.Contact
	database.DB.Where("status IN ? AND observation_end < ?", []string{"正常", "出现症状"}, now).Find(&contacts)

	for _, contact := range contacts {
		contact.Status = models.StatusExcluded
		database.DB.Save(&contact)
	}
	return nil
}

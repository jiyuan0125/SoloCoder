package services

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"trial-management-system/internal/database"
	"trial-management-system/internal/models"
)

func CreateProtocol(protocol *models.Protocol) error {
	return database.DB.Create(protocol).Error
}

func GetProtocolByNumber(number string) (*models.Protocol, error) {
	var protocol models.Protocol
	err := database.DB.Preload("Sites").Preload("Visits").Where("protocol_number = ?", number).First(&protocol).Error
	if err != nil {
		return nil, err
	}
	return &protocol, nil
}

func GetProtocolByID(id uuid.UUID) (*models.Protocol, error) {
	var protocol models.Protocol
	err := database.DB.Preload("Sites").Preload("Visits").Where("id = ?", id).First(&protocol).Error
	if err != nil {
		return nil, err
	}
	return &protocol, nil
}

func ListProtocols() ([]models.Protocol, error) {
	var protocols []models.Protocol
	err := database.DB.Preload("Sites").Find(&protocols).Error
	return protocols, err
}

func UpdateProtocol(id uuid.UUID, updates map[string]interface{}) error {
	return database.DB.Model(&models.Protocol{}).Where("id = ?", id).Updates(updates).Error
}

func CreateSite(site *models.Site) error {
	return database.DB.Create(site).Error
}

func CreateVisit(visit *models.Visit) error {
	return database.DB.Create(visit).Error
}

func GetVisitsByProtocol(protocolID uuid.UUID) ([]models.Visit, error) {
	var visits []models.Visit
	err := database.DB.Where("protocol_id = ?", protocolID).Order("visit_order").Find(&visits).Error
	return visits, err
}

func GetSiteByID(id uuid.UUID) (*models.Site, error) {
	var site models.Site
	err := database.DB.Where("id = ?", id).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func GetSiteByCode(protocolID uuid.UUID, siteCode string) (*models.Site, error) {
	var site models.Site
	err := database.DB.Where("protocol_id = ? AND site_code = ?", protocolID, siteCode).First(&site).Error
	if err != nil {
		return nil, err
	}
	return &site, nil
}

func IsDuplicateProtocol(number string) bool {
	var count int64
	database.DB.Model(&models.Protocol{}).Where("protocol_number = ?", number).Count(&count)
	return count > 0
}

func GetVisitByID(id uuid.UUID) (*models.Visit, error) {
	var visit models.Visit
	err := database.DB.Where("id = ?", id).First(&visit).Error
	if err != nil {
		return nil, err
	}
	return &visit, nil
}

func CreateSubject(subject *models.Subject) error {
	return database.DB.Create(subject).Error
}

func GetSubjectByID(id uuid.UUID) (*models.Subject, error) {
	var subject models.Subject
	err := database.DB.Where("id = ?", id).First(&subject).Error
	if err != nil {
		return nil, err
	}
	return &subject, nil
}

func GetSubjectByRandomID(randomID string) (*models.Subject, error) {
	var subject models.Subject
	err := database.DB.Where("randomization_id = ?", randomID).First(&subject).Error
	if err != nil {
		return nil, err
	}
	return &subject, nil
}

func ListSubjectsByProtocol(protocolID uuid.UUID) ([]models.Subject, error) {
	var subjects []models.Subject
	err := database.DB.Where("protocol_id = ?", protocolID).Find(&subjects).Error
	return subjects, err
}

func UpdateSubject(id uuid.UUID, updates map[string]interface{}) error {
	return database.DB.Model(&models.Subject{}).Where("id = ?", id).Updates(updates).Error
}

func GetNextSeqNumber(siteID uuid.UUID) (int, error) {
	var maxSeq int
	err := database.DB.Model(&models.Subject{}).Where("site_id = ?", siteID).Select("COALESCE(MAX(seq_number), 0)").Scan(&maxSeq).Error
	return maxSeq + 1, err
}

func CountSubjectsByProtocol(protocolID uuid.UUID) (int64, error) {
	var count int64
	err := database.DB.Model(&models.Subject{}).Where("protocol_id = ?", protocolID).Count(&count).Error
	return count, err
}

func IsDuplicateRandomID(randomID string) bool {
	var count int64
	database.DB.Model(&models.Subject{}).Where("randomization_id = ?", randomID).Count(&count)
	return count > 0
}

func CreateVisitRecord(record *models.VisitRecord) error {
	return database.DB.Create(record).Error
}

func GetVisitRecordsBySubject(subjectID uuid.UUID) ([]models.VisitRecord, error) {
	var records []models.VisitRecord
	err := database.DB.Where("subject_id = ?", subjectID).Find(&records).Error
	return records, err
}

func IsDuplicateVisitRecord(subjectID, visitID uuid.UUID) bool {
	var count int64
	database.DB.Model(&models.VisitRecord{}).Where("subject_id = ? AND visit_id = ?", subjectID, visitID).Count(&count)
	return count > 0
}

func CreateAdverseEvent(ae *models.AdverseEvent) error {
	return database.DB.Create(ae).Error
}

func ListAdverseEventsBySubject(subjectID uuid.UUID) ([]models.AdverseEvent, error) {
	var aes []models.AdverseEvent
	err := database.DB.Where("subject_id = ?", subjectID).Find(&aes).Error
	return aes, err
}

func ListAdverseEventsByProtocol(protocolID uuid.UUID) ([]models.AdverseEvent, error) {
	var aes []models.AdverseEvent
	err := database.DB.Joins("JOIN subjects ON subjects.id = adverse_events.subject_id").Where("subjects.protocol_id = ?", protocolID).Find(&aes).Error
	return aes, err
}

func GetAdverseEventByID(id uuid.UUID) (*models.AdverseEvent, error) {
	var ae models.AdverseEvent
	err := database.DB.Where("id = ?", id).First(&ae).Error
	if err != nil {
		return nil, err
	}
	return &ae, nil
}

func UpdateAdverseEvent(id uuid.UUID, updates map[string]interface{}) error {
	return database.DB.Model(&models.AdverseEvent{}).Where("id = ?", id).Updates(updates).Error
}

func CreateTodo(todo *models.Todo) error {
	return database.DB.Create(todo).Error
}

func ListTodos() ([]models.Todo, error) {
	var todos []models.Todo
	err := database.DB.Order("created_at DESC").Find(&todos).Error
	return todos, err
}

func UpdateTodo(id uuid.UUID, updates map[string]interface{}) error {
	return database.DB.Model(&models.Todo{}).Where("id = ?", id).Updates(updates).Error
}

func CheckOverdueTodos() {
	now := time.Now()
	database.DB.Model(&models.Todo{}).Where("status = ? AND due_date < ?", models.TodoPending, now).Update("status", models.TodoOverdue)
}

func ProtocolExists(id uuid.UUID) bool {
	var count int64
	database.DB.Model(&models.Protocol{}).Where("id = ?", id).Count(&count)
	return count > 0
}

func SubjectExists(id uuid.UUID) bool {
	var count int64
	database.DB.Model(&models.Subject{}).Where("id = ?", id).Count(&count)
	return count > 0
}

func GetAllSubjectsWithDetails() ([]models.Subject, error) {
	var subjects []models.Subject
	err := database.DB.Preload("Records").Preload("AEs").Find(&subjects).Error
	return subjects, err
}

func GetSubjectsBySite(siteID uuid.UUID) ([]models.Subject, error) {
	var subjects []models.Subject
	err := database.DB.Where("site_id = ?", siteID).Find(&subjects).Error
	return subjects, err
}

func GetSubjectsWithProtocolSite(protocolID, siteID uuid.UUID) ([]models.Subject, error) {
	var subjects []models.Subject
	query := database.DB.Preload("Records").Preload("AEs").Where("protocol_id = ?", protocolID)
	if siteID != uuid.Nil {
		query = query.Where("site_id = ?", siteID)
	}
	err := query.Find(&subjects).Error
	return subjects, err
}

func ListSAEByProtocol(protocolID uuid.UUID) ([]models.AdverseEvent, error) {
	var aes []models.AdverseEvent
	err := database.DB.Joins("JOIN subjects ON subjects.id = adverse_events.subject_id").Where("subjects.protocol_id = ? AND adverse_events.is_sae = ?", protocolID, true).Find(&aes).Error
	return aes, err
}

func ListSAEBySite(siteID uuid.UUID) ([]models.AdverseEvent, error) {
	var aes []models.AdverseEvent
	err := database.DB.Joins("JOIN subjects ON subjects.id = adverse_events.subject_id").Where("subjects.site_id = ? AND adverse_events.is_sae = ?", siteID, true).Find(&aes).Error
	return aes, err
}

func GetDB() *gorm.DB {
	return database.DB
}

func CreateDepartment(dept *models.Department) error {
	return database.DB.Create(dept).Error
}

func ListDepartments() ([]models.Department, error) {
	var depts []models.Department
	err := database.DB.Find(&depts).Error
	return depts, err
}

func GetDepartmentByID(id uuid.UUID) (*models.Department, error) {
	var dept models.Department
	err := database.DB.Where("id = ?", id).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

func CreateBudget(budget *models.Budget) error {
	return database.DB.Create(budget).Error
}

func ListBudgets() ([]models.Budget, error) {
	var budgets []models.Budget
	err := database.DB.Preload("Items").Find(&budgets).Error
	return budgets, err
}

func GetBudgetByID(id uuid.UUID) (*models.Budget, error) {
	var budget models.Budget
	err := database.DB.Preload("Items").Where("id = ?", id).First(&budget).Error
	if err != nil {
		return nil, err
	}
	return &budget, nil
}

func UpdateBudget(id uuid.UUID, updates map[string]interface{}) error {
	return database.DB.Model(&models.Budget{}).Where("id = ?", id).Updates(updates).Error
}

func CreateBudgetItem(item *models.BudgetItem) error {
	return database.DB.Create(item).Error
}

func GetBudgetItemByID(id uuid.UUID) (*models.BudgetItem, error) {
	var item models.BudgetItem
	err := database.DB.Where("id = ?", id).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func UpdateBudgetItem(id uuid.UUID, updates map[string]interface{}) error {
	return database.DB.Model(&models.BudgetItem{}).Where("id = ?", id).Updates(updates).Error
}

func ListBudgetItems(budgetID uuid.UUID) ([]models.BudgetItem, error) {
	var items []models.BudgetItem
	err := database.DB.Where("budget_id = ?", budgetID).Find(&items).Error
	return items, err
}

func CreateBudgetAlert(alert *models.BudgetAlert) error {
	return database.DB.Create(alert).Error
}

func ListBudgetAlerts() ([]models.BudgetAlert, error) {
	var alerts []models.BudgetAlert
	err := database.DB.Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func CreateSystemA(a *models.SystemA) error {
	return database.DB.Create(a).Error
}

func ListSystemA() ([]models.SystemA, error) {
	var items []models.SystemA
	err := database.DB.Preload("Items").Find(&items).Error
	return items, err
}

func GetSystemAByID(id uuid.UUID) (*models.SystemA, error) {
	var a models.SystemA
	err := database.DB.Preload("Items").Where("id = ?", id).First(&a).Error
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func CreateSystemB(b *models.SystemB) error {
	return database.DB.Create(b).Error
}

func ListSystemB() ([]models.SystemB, error) {
	var items []models.SystemB
	err := database.DB.Find(&items).Error
	return items, err
}

func GetSystemBByID(id uuid.UUID) (*models.SystemB, error) {
	var b models.SystemB
	err := database.DB.Where("id = ?", id).First(&b).Error
	if err != nil {
		return nil, err
	}
	return &b, nil
}

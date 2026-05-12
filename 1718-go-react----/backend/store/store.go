package store

import (
	"clinical-path-backend/models"
	"fmt"
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	nextPathID int64
	nextStageID int64
	nextItemID int64
	nextPatientID int64
	nextEnrollmentID int64
	nextDailyOrderID int64
	nextVariationID int64

	paths          map[int64]*models.ClinicalPath
	pathsByCode    map[string]*models.ClinicalPath
	stages         map[int64][]*models.PathStage
	stagesByID     map[int64]*models.PathStage
	items          map[int64][]*models.OrderItem
	itemsByID      map[int64]*models.OrderItem
	patients       map[int64]*models.Patient
	patientsByNo   map[string]*models.Patient
	enrollments    map[int64]*models.PatientEnrollment
	enrollmentsByPatient map[int64][]*models.PatientEnrollment
	enrollmentsByPath map[int64][]*models.PatientEnrollment
	dailyOrders    map[int64][]*models.DailyOrder
	dailyOrdersByID map[int64]*models.DailyOrder
	variations     map[int64][]*models.VariationRecord
	variationsByID map[int64]*models.VariationRecord
}

func NewStore() *Store {
	s := &Store{
		nextPathID: 1,
		nextStageID: 1,
		nextItemID: 1,
		nextPatientID: 1,
		nextEnrollmentID: 1,
		nextDailyOrderID: 1,
		nextVariationID: 1,
		paths:          make(map[int64]*models.ClinicalPath),
		pathsByCode:    make(map[string]*models.ClinicalPath),
		stages:         make(map[int64][]*models.PathStage),
		stagesByID:     make(map[int64]*models.PathStage),
		items:          make(map[int64][]*models.OrderItem),
		itemsByID:      make(map[int64]*models.OrderItem),
		patients:       make(map[int64]*models.Patient),
		patientsByNo:   make(map[string]*models.Patient),
		enrollments:    make(map[int64]*models.PatientEnrollment),
		enrollmentsByPatient: make(map[int64][]*models.PatientEnrollment),
		enrollmentsByPath: make(map[int64][]*models.PatientEnrollment),
		dailyOrders:    make(map[int64][]*models.DailyOrder),
		dailyOrdersByID: make(map[int64]*models.DailyOrder),
		variations:     make(map[int64][]*models.VariationRecord),
		variationsByID: make(map[int64]*models.VariationRecord),
	}
	s.initSampleData()
	return s
}

func (s *Store) initSampleData() {
	now := time.Now()
	
	path1 := &models.ClinicalPath{
		ID:        1,
		Code:      "CP-APP-001",
		Name:      "急性阑尾炎",
		ICDCodes:  "K35,K35.0,K35.1,K35.2,K35.3,K35.8,K35.9",
		MinDays:   5,
		MaxDays:   7,
		MinCost:   8000.0,
		MaxCost:   15000.0,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.paths[1] = path1
	s.pathsByCode["CP-APP-001"] = path1
	s.nextPathID = 2

	stage1 := &models.PathStage{ID: 1, PathID: 1, Name: "术前准备", Days: 2, Order: 1}
	stage2 := &models.PathStage{ID: 2, PathID: 1, Name: "手术日", Days: 1, Order: 2}
	stage3 := &models.PathStage{ID: 3, PathID: 1, Name: "术后恢复", Days: 3, Order: 3}
	s.stages[1] = []*models.PathStage{stage1, stage2, stage3}
	s.stagesByID[1] = stage1
	s.stagesByID[2] = stage2
	s.stagesByID[3] = stage3
	s.nextStageID = 4

	items := []*models.OrderItem{
		{ID: 1, StageID: 1, PathID: 1, Name: "血常规", Category: "检查检验", IsRequired: true},
		{ID: 2, StageID: 1, PathID: 1, Name: "凝血功能", Category: "检查检验", IsRequired: true},
		{ID: 3, StageID: 1, PathID: 1, Name: "腹部B超", Category: "检查检验", IsRequired: true},
		{ID: 4, StageID: 1, PathID: 1, Name: "禁食水", Category: "饮食", IsRequired: true},
		{ID: 5, StageID: 2, PathID: 1, Name: "阑尾切除术", Category: "手术", IsRequired: true},
		{ID: 6, StageID: 2, PathID: 1, Name: "头孢类抗生素", Category: "用药", IsRequired: true},
		{ID: 7, StageID: 2, PathID: 1, Name: "镇痛药物", Category: "用药", IsRequired: false},
		{ID: 8, StageID: 3, PathID: 1, Name: "术后换药", Category: "护理", IsRequired: true},
		{ID: 9, StageID: 3, PathID: 1, Name: "流质饮食", Category: "饮食", IsRequired: true},
		{ID: 10, StageID: 3, PathID: 1, Name: "复查血常规", Category: "检查检验", IsRequired: false},
	}
	s.items[1] = items
	for _, item := range items {
		s.itemsByID[item.ID] = item
	}
	s.nextItemID = 11

	patient1 := &models.Patient{
		ID:            1,
		HospitalNo:    "Z20260510001",
		Name:          "张三",
		Diagnosis:     "急性阑尾炎(K35)",
		AdmissionDate: now.AddDate(0, 0, -3),
		Status:        "active",
	}
	patient2 := &models.Patient{
		ID:            2,
		HospitalNo:    "Z20260511002",
		Name:          "李四",
		Diagnosis:     "急性单纯性阑尾炎(K35.0)",
		AdmissionDate: now.AddDate(0, 0, -1),
		Status:        "active",
	}
	patient3 := &models.Patient{
		ID:            3,
		HospitalNo:    "Z20260512003",
		Name:          "王五",
		Diagnosis:     "上呼吸道感染",
		AdmissionDate: now.AddDate(0, 0, -2),
		Status:        "active",
	}
	s.patients[1] = patient1
	s.patients[2] = patient2
	s.patients[3] = patient3
	s.patientsByNo["Z20260510001"] = patient1
	s.patientsByNo["Z20260511002"] = patient2
	s.patientsByNo["Z20260512003"] = patient3
	s.nextPatientID = 4
}

func (s *Store) CreatePath(path *models.ClinicalPath) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pathsByCode[path.Code]; exists {
		return fmt.Errorf("路径编号已存在")
	}
	if path.MinDays > path.MaxDays {
		return fmt.Errorf("标准住院天数下限不能大于上限")
	}
	if path.MinCost > path.MaxCost {
		return fmt.Errorf("标准费用下限不能大于上限")
	}

	path.ID = s.nextPathID
	now := time.Now()
	path.CreatedAt = now
	path.UpdatedAt = now

	s.paths[path.ID] = path
	s.pathsByCode[path.Code] = path
	s.nextPathID++

	return nil
}

func (s *Store) UpdatePath(path *models.ClinicalPath) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.paths[path.ID]
	if !ok {
		return fmt.Errorf("路径不存在")
	}

	if existing.Code != path.Code {
		if _, exists := s.pathsByCode[path.Code]; exists {
			return fmt.Errorf("路径编号已存在")
		}
		delete(s.pathsByCode, existing.Code)
	}

	if path.MinDays > path.MaxDays {
		return fmt.Errorf("标准住院天数下限不能大于上限")
	}
	if path.MinCost > path.MaxCost {
		return fmt.Errorf("标准费用下限不能大于上限")
	}

	existing.Code = path.Code
	existing.Name = path.Name
	existing.ICDCodes = path.ICDCodes
	existing.MinDays = path.MinDays
	existing.MaxDays = path.MaxDays
	existing.MinCost = path.MinCost
	existing.MaxCost = path.MaxCost
	existing.UpdatedAt = time.Now()

	s.pathsByCode[existing.Code] = existing

	return nil
}

func (s *Store) ListPaths() []*models.ClinicalPath {
	s.mu.RLock()
	defer s.mu.RUnlock()

	paths := make([]*models.ClinicalPath, 0, len(s.paths))
	for _, p := range s.paths {
		paths = append(paths, p)
	}
	return paths
}

func (s *Store) GetPath(id int64) (*models.ClinicalPath, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path, ok := s.paths[id]
	if !ok {
		return nil, fmt.Errorf("路径不存在")
	}
	return path, nil
}

func (s *Store) CreateStage(stage *models.PathStage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.paths[stage.PathID]
	if !ok {
		return fmt.Errorf("路径不存在")
	}

	stage.ID = s.nextStageID
	s.stages[stage.PathID] = append(s.stages[stage.PathID], stage)
	s.stagesByID[stage.ID] = stage
	s.nextStageID++

	return nil
}

func (s *Store) ListStages(pathID int64) []*models.PathStage {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stages := make([]*models.PathStage, len(s.stages[pathID]))
	copy(stages, s.stages[pathID])
	return stages
}

func (s *Store) CreateItem(item *models.OrderItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.stagesByID[item.StageID]
	if !ok {
		return fmt.Errorf("阶段不存在")
	}

	item.ID = s.nextItemID
	s.items[item.PathID] = append(s.items[item.PathID], item)
	s.itemsByID[item.ID] = item
	s.nextItemID++

	return nil
}

func (s *Store) ListItems(pathID int64) []*models.OrderItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*models.OrderItem, len(s.items[pathID]))
	copy(items, s.items[pathID])
	return items
}

func (s *Store) ListItemsByStage(stageID int64) []*models.OrderItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]*models.OrderItem, 0)
	for _, item := range s.itemsByID {
		if item.StageID == stageID {
			items = append(items, item)
		}
	}
	return items
}

func (s *Store) ListPatients() []*models.Patient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patients := make([]*models.Patient, 0, len(s.patients))
	for _, p := range s.patients {
		patients = append(patients, p)
	}
	return patients
}

func (s *Store) GetPatient(id int64) (*models.Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patient, ok := s.patients[id]
	if !ok {
		return nil, fmt.Errorf("患者不存在")
	}
	return patient, nil
}

func (s *Store) GetPatientByNo(hospitalNo string) (*models.Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patient, ok := s.patientsByNo[hospitalNo]
	if !ok {
		return nil, fmt.Errorf("患者不存在")
	}
	return patient, nil
}

func (s *Store) CreatePatient(patient *models.Patient) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.patientsByNo[patient.HospitalNo]; exists {
		return fmt.Errorf("住院号已存在")
	}

	patient.ID = s.nextPatientID
	patient.Status = "active"
	s.patients[patient.ID] = patient
	s.patientsByNo[patient.HospitalNo] = patient
	s.nextPatientID++

	return nil
}

func (s *Store) HasActiveEnrollment(patientID int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, e := range s.enrollmentsByPatient[patientID] {
		if e.Status == "active" || e.Status == "enrolled" {
			return true
		}
	}
	return false
}

func (s *Store) CreateEnrollment(enrollment *models.PatientEnrollment, orders []*models.DailyOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.enrollmentsByPatient[enrollment.PatientID] {
		if e.Status == "active" || e.Status == "enrolled" {
			return fmt.Errorf("患者已在其他路径中")
		}
	}

	enrollment.ID = s.nextEnrollmentID
	enrollment.Status = "active"
	enrollment.VariationCount = 0
	enrollment.SuggestExit = false
	now := time.Now()
	enrollment.CreatedAt = now

	s.enrollments[enrollment.ID] = enrollment
	s.enrollmentsByPatient[enrollment.PatientID] = append(s.enrollmentsByPatient[enrollment.PatientID], enrollment)
	s.enrollmentsByPath[enrollment.PathID] = append(s.enrollmentsByPath[enrollment.PathID], enrollment)
	s.nextEnrollmentID++

	for _, order := range orders {
		order.ID = s.nextDailyOrderID
		order.EnrollmentID = enrollment.ID
		s.dailyOrders[enrollment.ID] = append(s.dailyOrders[enrollment.ID], order)
		s.dailyOrdersByID[order.ID] = order
		s.nextDailyOrderID++
	}

	return nil
}

func (s *Store) ListEnrollments() []*models.PatientEnrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	enrollments := make([]*models.PatientEnrollment, 0, len(s.enrollments))
	for _, e := range s.enrollments {
		enrollments = append(enrollments, e)
	}
	return enrollments
}

func (s *Store) GetEnrollment(id int64) (*models.PatientEnrollment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	enrollment, ok := s.enrollments[id]
	if !ok {
		return nil, fmt.Errorf("入径记录不存在")
	}
	return enrollment, nil
}

func (s *Store) ListDailyOrders(enrollmentID int64) []*models.DailyOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]*models.DailyOrder, len(s.dailyOrders[enrollmentID]))
	copy(orders, s.dailyOrders[enrollmentID])
	return orders
}

func (s *Store) GetDailyOrder(id int64) (*models.DailyOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.dailyOrdersByID[id]
	if !ok {
		return nil, fmt.Errorf("医嘱不存在")
	}
	return order, nil
}

func (s *Store) UpdateDailyOrder(order *models.DailyOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.dailyOrdersByID[order.ID]
	if !ok {
		return fmt.Errorf("医嘱不存在")
	}

	existing.IsExecuted = order.IsExecuted
	existing.ExecutedAt = order.ExecutedAt
	existing.IsVariation = order.IsVariation

	return nil
}

func (s *Store) AddVariationAndUpdateEnrollment(variation *models.VariationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	enrollment, ok := s.enrollments[variation.EnrollmentID]
	if !ok {
		return fmt.Errorf("入径记录不存在")
	}

	variation.ID = s.nextVariationID
	variation.CreatedAt = time.Now()
	s.variations[variation.EnrollmentID] = append(s.variations[variation.EnrollmentID], variation)
	s.variationsByID[variation.ID] = variation
	s.nextVariationID++

	enrollment.VariationCount++
	if enrollment.VariationCount >= 3 {
		enrollment.SuggestExit = true
	}

	return nil
}

func (s *Store) AddVariation(variation *models.VariationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	enrollment, ok := s.enrollments[variation.EnrollmentID]
	if !ok {
		return fmt.Errorf("入径记录不存在")
	}

	variation.ID = s.nextVariationID
	variation.CreatedAt = time.Now()
	s.variations[variation.EnrollmentID] = append(s.variations[variation.EnrollmentID], variation)
	s.variationsByID[variation.ID] = variation
	s.nextVariationID++

	enrollment.VariationCount++
	if enrollment.VariationCount >= 3 {
		enrollment.SuggestExit = true
	}

	return nil
}

func (s *Store) ListVariations(enrollmentID int64) []*models.VariationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	variations := make([]*models.VariationRecord, len(s.variations[enrollmentID]))
	copy(variations, s.variations[enrollmentID])
	return variations
}

func (s *Store) ExitEnrollment(enrollmentID int64, exitReason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return fmt.Errorf("入径记录不存在")
	}

	now := time.Now()
	enrollment.Status = "exited"
	enrollment.ExitDate = &now
	enrollment.ExitReason = exitReason

	return nil
}

func (s *Store) CompleteEnrollment(enrollmentID int64, actualDays int, actualCost float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	enrollment, ok := s.enrollments[enrollmentID]
	if !ok {
		return fmt.Errorf("入径记录不存在")
	}

	now := time.Now()
	enrollment.Status = "completed"
	enrollment.ActualEndDate = &now
	enrollment.ActualDays = actualDays
	enrollment.ActualCost = actualCost

	return nil
}

func (s *Store) ListEnrollmentsByPathAndMonth(pathID int64, year int, month time.Month) []*models.PatientEnrollment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.PatientEnrollment, 0)
	for _, e := range s.enrollmentsByPath[pathID] {
		eYear, eMonth, _ := e.EnrollDate.Date()
		if eYear == year && eMonth == month {
			result = append(result, e)
		}
	}
	return result
}

func (s *Store) ListAllPatients() []*models.Patient {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patients := make([]*models.Patient, 0, len(s.patients))
	for _, p := range s.patients {
		patients = append(patients, p)
	}
	return patients
}

func (s *Store) ListAllPaths() []*models.ClinicalPath {
	s.mu.RLock()
	defer s.mu.RUnlock()

	paths := make([]*models.ClinicalPath, 0, len(s.paths))
	for _, p := range s.paths {
		paths = append(paths, p)
	}
	return paths
}

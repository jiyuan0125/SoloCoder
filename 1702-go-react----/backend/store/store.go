package store

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"dental-clinic/models"
)

type internalAppointment struct {
	ID         string
	PatientID  string
	DoctorID   string
	Date       time.Time
	StartTime  time.Time
	EndTime    time.Time
	Treatments []string
}

type internalStep struct {
	ID           string
	PlanID       string
	Index        int
	ExpectedDate time.Time
	ActualDate   *time.Time
	DoctorID     string
	Description  string
	Fee          int
	Status       models.StepStatus
	IsOverdue    bool
}

type internalPlan struct {
	ID              string
	PatientID       string
	DiscountPercent int
	Steps           []*internalStep
	Status          models.PlanStatus
	IsSurgical      bool
	CreatedAt       time.Time
}

type internalFollowUp struct {
	ID          string
	PatientID   string
	DoctorID    string
	PlanID      string
	StepID      string
	DueDate     time.Time
	Status      models.FollowUpStatus
	Method      models.FollowUpMethod
	Feedback    models.FeedbackType
	Notes       string
	CompletedAt *time.Time
}

type internalTodo struct {
	ID         string
	DoctorID   string
	Title      string
	Priority   models.TodoPriority
	PatientID  string
	FollowUpID string
	CreatedAt  time.Time
	Done       bool
}

type Store struct {
	mu             sync.RWMutex
	doctors        map[string]*models.Doctor
	patients       map[string]*models.Patient
	appointments   map[string]*internalAppointment
	treatmentPlans map[string]*internalPlan
	followUps      map[string]*internalFollowUp
	todos          map[string]*internalTodo
}

func NewStore() *Store {
	return &Store{
		doctors:        make(map[string]*models.Doctor),
		patients:       make(map[string]*models.Patient),
		appointments:   make(map[string]*internalAppointment),
		treatmentPlans: make(map[string]*internalPlan),
		followUps:      make(map[string]*internalFollowUp),
		todos:          make(map[string]*internalTodo),
	}
}

func (s *Store) SeedData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.doctors = map[string]*models.Doctor{
		"d1": {
			ID:         "d1",
			Name:       "张医生",
			Department: models.DeptImplant,
			WorkShifts: []models.WorkShift{
				{StartHour: 8, StartMinute: 0, EndHour: 12, EndMinute: 0},
				{StartHour: 14, StartMinute: 0, EndHour: 17, EndMinute: 0},
			},
		},
		"d2": {
			ID:         "d2",
			Name:       "李医生",
			Department: models.DeptOrthodontics,
			WorkShifts: []models.WorkShift{
				{StartHour: 8, StartMinute: 0, EndHour: 12, EndMinute: 0},
				{StartHour: 14, StartMinute: 0, EndHour: 17, EndMinute: 0},
			},
		},
		"d3": {
			ID:         "d3",
			Name:       "王医生",
			Department: models.DeptEndodontics,
			WorkShifts: []models.WorkShift{
				{StartHour: 8, StartMinute: 0, EndHour: 12, EndMinute: 0},
				{StartHour: 14, StartMinute: 0, EndHour: 17, EndMinute: 0},
			},
		},
		"d4": {
			ID:         "d4",
			Name:       "赵医生",
			Department: models.DeptPeriodontics,
			WorkShifts: []models.WorkShift{
				{StartHour: 8, StartMinute: 0, EndHour: 12, EndMinute: 0},
				{StartHour: 14, StartMinute: 0, EndHour: 17, EndMinute: 0},
			},
		},
		"d5": {
			ID:         "d5",
			Name:       "孙医生",
			Department: models.DeptPediatric,
			WorkShifts: []models.WorkShift{
				{StartHour: 8, StartMinute: 0, EndHour: 12, EndMinute: 0},
				{StartHour: 14, StartMinute: 0, EndHour: 17, EndMinute: 0},
			},
		},
	}

	s.patients = map[string]*models.Patient{
		"p1": {ID: "p1", Name: "患者甲", Phone: "13800000001"},
		"p2": {ID: "p2", Name: "患者乙", Phone: "13800000002"},
		"p3": {ID: "p3", Name: "患者丙", Phone: "13800000003"},
	}
}

func (s *Store) ListDoctors() []*models.Doctor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Doctor, 0, len(s.doctors))
	for _, d := range s.doctors {
		result = append(result, d)
	}
	return result
}

func (s *Store) GetDoctor(id string) (*models.Doctor, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.doctors[id]
	if !ok {
		return nil, errors.New("doctor not found")
	}
	return d, nil
}

func (s *Store) GetPatient(id string) (*models.Patient, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.patients[id]
	if !ok {
		return nil, errors.New("patient not found")
	}
	return p, nil
}

func toPublicAppointment(ia *internalAppointment) *models.Appointment {
	dateStr := ia.Date.Format("2006-01-02")
	return &models.Appointment{
		ID:         ia.ID,
		PatientID:  ia.PatientID,
		DoctorID:   ia.DoctorID,
		Date:       dateStr,
		StartTime:  ia.StartTime.Format(time.RFC3339),
		EndTime:    ia.EndTime.Format(time.RFC3339),
		Treatments: ia.Treatments,
	}
}

func (s *Store) ListAppointments() []*models.Appointment {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Appointment, 0, len(s.appointments))
	for _, a := range s.appointments {
		result = append(result, toPublicAppointment(a))
	}
	return result
}

func (s *Store) CheckAppointmentConflict(doctorID string, startTime time.Time, patientID string) (conflict bool, samePatient bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, a := range s.appointments {
		if a.DoctorID == doctorID && a.StartTime.Equal(startTime) {
			if a.PatientID == patientID {
				return true, true
			}
			return true, false
		}
	}
	return false, false
}

func (s *Store) MergeOrCreateAppointment(patientID, doctorID string, startTime, endTime, date time.Time, treatments []string) (*models.Appointment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.appointments {
		if a.DoctorID == doctorID && a.StartTime.Equal(startTime) && a.PatientID == patientID {
			existingTreatments := make(map[string]bool)
			for _, t := range a.Treatments {
				existingTreatments[t] = true
			}
			for _, t := range treatments {
				if !existingTreatments[t] {
					a.Treatments = append(a.Treatments, t)
				}
			}
			return toPublicAppointment(a), nil
		}
	}

	id := fmt.Sprintf("a%d", time.Now().UnixNano())
	appointment := &internalAppointment{
		ID:         id,
		PatientID:  patientID,
		DoctorID:   doctorID,
		Date:       date,
		StartTime:  startTime,
		EndTime:    endTime,
		Treatments: treatments,
	}
	s.appointments[id] = appointment
	return toPublicAppointment(appointment), nil
}

func toPublicStep(is *internalStep) *models.Step {
	var actualDatePtr *string
	if is.ActualDate != nil {
		str := is.ActualDate.Format("2006-01-02")
		actualDatePtr = &str
	}
	return &models.Step{
		ID:           is.ID,
		PlanID:       is.PlanID,
		Index:        is.Index,
		ExpectedDate: is.ExpectedDate.Format("2006-01-02"),
		ActualDate:   actualDatePtr,
		DoctorID:     is.DoctorID,
		Description:  is.Description,
		Fee:          is.Fee,
		Status:       is.Status,
		IsOverdue:    is.IsOverdue,
	}
}

func toPublicPlan(ip *internalPlan) *models.TreatmentPlan {
	steps := make([]*models.Step, len(ip.Steps))
	for i, step := range ip.Steps {
		steps[i] = toPublicStep(step)
	}
	return &models.TreatmentPlan{
		ID:              ip.ID,
		PatientID:       ip.PatientID,
		DiscountPercent: ip.DiscountPercent,
		Steps:           steps,
		Status:          ip.Status,
		IsSurgical:      ip.IsSurgical,
		CreatedAt:       ip.CreatedAt.Format("2006-01-02"),
	}
}

func (s *Store) CreateTreatmentPlan(patientID string, discountPercent int, isSurgical bool, steps []*models.Step) (*models.TreatmentPlan, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expectedDates := make(map[string]bool)
	internalSteps := make([]*internalStep, len(steps))

	for i, step := range steps {
		expectedDate, err := time.Parse("2006-01-02", step.ExpectedDate)
		if err != nil {
			return nil, errors.New("invalid expected date format")
		}

		dateKey := expectedDate.Format("2006-01-02")
		if expectedDates[dateKey] {
			return nil, errors.New("duplicate expected date in steps")
		}
		expectedDates[dateKey] = true

		internalSteps[i] = &internalStep{
			ExpectedDate: expectedDate,
			DoctorID:     step.DoctorID,
			Description:  step.Description,
			Fee:          step.Fee,
		}
	}

	id := fmt.Sprintf("tp%d", time.Now().UnixNano())
	plan := &internalPlan{
		ID:              id,
		PatientID:       patientID,
		DiscountPercent: discountPercent,
		Steps:           internalSteps,
		Status:          models.PlanStatusNotStarted,
		IsSurgical:      isSurgical,
		CreatedAt:       time.Now(),
	}

	for i, step := range internalSteps {
		step.ID = fmt.Sprintf("s%d_%d", time.Now().UnixNano(), i)
		step.PlanID = id
		step.Index = i
		step.Status = models.StepStatusNotStarted
		step.IsOverdue = false
	}

	s.treatmentPlans[id] = plan
	return toPublicPlan(plan), nil
}

func (s *Store) ListTreatmentPlans() []*models.TreatmentPlan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.TreatmentPlan, 0, len(s.treatmentPlans))
	for _, p := range s.treatmentPlans {
		result = append(result, toPublicPlan(p))
	}
	return result
}

func (s *Store) GetTreatmentPlan(id string) (*models.TreatmentPlan, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.treatmentPlans[id]
	if !ok {
		return nil, errors.New("treatment plan not found")
	}
	return toPublicPlan(p), nil
}

func (s *Store) UpdateStepStatus(planID, stepID string, newStatus models.StepStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan, ok := s.treatmentPlans[planID]
	if !ok {
		return errors.New("treatment plan not found")
	}

	var targetStep *internalStep
	for _, step := range plan.Steps {
		if step.ID == stepID {
			targetStep = step
			break
		}
	}
	if targetStep == nil {
		return errors.New("step not found")
	}

	currentStatus := targetStep.Status
	if currentStatus == models.StepStatusCompleted && newStatus != models.StepStatusCompleted {
		return errors.New("cannot revert completed step")
	}
	if currentStatus == models.StepStatusNotStarted && newStatus == models.StepStatusCompleted {
		return errors.New("must go through in_progress first")
	}
	if currentStatus == newStatus {
		return nil
	}

	targetStep.Status = newStatus
	if newStatus == models.StepStatusInProgress && plan.Status == models.PlanStatusNotStarted {
		plan.Status = models.PlanStatusInProgress
	}

	if newStatus == models.StepStatusCompleted {
		now := time.Now()
		targetStep.ActualDate = &now
		s.createFollowUpsForStep(plan, targetStep)
	}

	allCompleted := true
	for _, step := range plan.Steps {
		if step.Status != models.StepStatusCompleted {
			allCompleted = false
			break
		}
	}
	if allCompleted && plan.Status == models.PlanStatusInProgress {
		plan.Status = models.PlanStatusCompleted
	}

	return nil
}

func (s *Store) createFollowUpsForStep(plan *internalPlan, step *internalStep) {
	var daysToAdd []int
	if plan.IsSurgical {
		daysToAdd = []int{1, 7}
	} else {
		daysToAdd = []int{3}
	}

	now := time.Now()
	for i, days := range daysToAdd {
		dueDate := now.AddDate(0, 0, days)
		id := fmt.Sprintf("f%d_%d", time.Now().UnixNano(), i)
		followUp := &internalFollowUp{
			ID:        id,
			PatientID: plan.PatientID,
			DoctorID:  step.DoctorID,
			PlanID:    plan.ID,
			StepID:    step.ID,
			DueDate:   dueDate,
			Status:    models.FollowUpPending,
		}
		s.followUps[id] = followUp
	}
}

func toPublicFollowUp(ifup *internalFollowUp) *models.FollowUp {
	var completedAtPtr *string
	if ifup.CompletedAt != nil {
		str := ifup.CompletedAt.Format("2006-01-02")
		completedAtPtr = &str
	}
	return &models.FollowUp{
		ID:          ifup.ID,
		PatientID:   ifup.PatientID,
		DoctorID:    ifup.DoctorID,
		PlanID:      ifup.PlanID,
		StepID:      ifup.StepID,
		DueDate:     ifup.DueDate.Format("2006-01-02"),
		Status:      ifup.Status,
		Method:      ifup.Method,
		Feedback:    ifup.Feedback,
		Notes:       ifup.Notes,
		CompletedAt: completedAtPtr,
	}
}

func (s *Store) ListFollowUps() []*models.FollowUp {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.FollowUp, 0, len(s.followUps))
	for _, f := range s.followUps {
		result = append(result, toPublicFollowUp(f))
	}
	return result
}

func (s *Store) CompleteFollowUp(followUpID string, method models.FollowUpMethod, feedback models.FeedbackType, notes string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, ok := s.followUps[followUpID]
	if !ok {
		return errors.New("follow-up not found")
	}

	now := time.Now()
	f.Status = models.FollowUpCompleted
	f.Method = method
	f.Feedback = feedback
	f.Notes = notes
	f.CompletedAt = &now

	if feedback == models.FeedbackDissatisfied {
		todoID := fmt.Sprintf("t%d", time.Now().UnixNano())
		todo := &internalTodo{
			ID:         todoID,
			DoctorID:   f.DoctorID,
			Title:      "跟进处理 - 患者反馈不满意",
			Priority:   models.TodoPriorityHigh,
			PatientID:  f.PatientID,
			FollowUpID: f.ID,
			CreatedAt:  now,
			Done:       false,
		}
		s.todos[todoID] = todo
	}

	return nil
}

func toPublicTodo(it *internalTodo) *models.Todo {
	return &models.Todo{
		ID:         it.ID,
		DoctorID:   it.DoctorID,
		Title:      it.Title,
		Priority:   it.Priority,
		PatientID:  it.PatientID,
		FollowUpID: it.FollowUpID,
		CreatedAt:  it.CreatedAt.Format("2006-01-02"),
		Done:       it.Done,
	}
}

func (s *Store) ListTodos() []*models.Todo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Todo, 0, len(s.todos))
	for _, t := range s.todos {
		result = append(result, toPublicTodo(t))
	}
	return result
}

func (s *Store) CheckOverduePlans() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for _, plan := range s.treatmentPlans {
		if plan.Status != models.PlanStatusInProgress {
			continue
		}
		for _, step := range plan.Steps {
			if step.Status != models.StepStatusCompleted {
				daysSinceExpected := now.Sub(step.ExpectedDate).Hours() / 24
				step.IsOverdue = daysSinceExpected > 30
			} else {
				step.IsOverdue = false
			}
		}
	}
}

func (s *Store) CalculateTotalFee(plan *models.TreatmentPlan) (int, error) {
	totalStepsFee := 0
	for _, step := range plan.Steps {
		totalStepsFee += step.Fee
	}

	discount := float64(plan.DiscountPercent) / 100.0
	totalFee := int(float64(totalStepsFee) * discount)

	if totalFee <= 0 {
		return 0, errors.New("total fee must be positive")
	}

	return totalFee, nil
}

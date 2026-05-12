package service

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"organdonation/internal/models"
	"organdonation/internal/repository"
	apperrors "organdonation/pkg/errors"
)

type Service struct {
	repo *repository.Repository
}

func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) isBloodTypeCompatible(donorBlood, recipientBlood models.BloodType) bool {
	switch donorBlood {
	case models.BloodTypeO:
		return true
	case models.BloodTypeA:
		return recipientBlood == models.BloodTypeA || recipientBlood == models.BloodTypeAB
	case models.BloodTypeB:
		return recipientBlood == models.BloodTypeB || recipientBlood == models.BloodTypeAB
	case models.BloodTypeAB:
		return recipientBlood == models.BloodTypeAB
	default:
		return false
	}
}

func (s *Service) isValidOrganStatusTransition(current, next models.OrganStatus) bool {
	transitions := map[models.OrganStatus][]models.OrganStatus{
		models.OrganStatusPending:  {models.OrganStatusMatched, models.OrganStatusDiscarded},
		models.OrganStatusMatched:  {models.OrganStatusAcquired, models.OrganStatusPending},
		models.OrganStatusAcquired: {models.OrganStatusTransplanted, models.OrganStatusDiscarded},
	}
	valid, ok := transitions[current]
	if !ok {
		return false
	}
	for _, v := range valid {
		if v == next {
			return true
		}
	}
	return false
}

func (s *Service) isValidPostOpStatusTransition(current, next models.PostOpStatus) bool {
	transitions := map[models.PostOpStatus][]models.PostOpStatus{
		models.PostOpInSurgery:   {models.PostOpObservation},
		models.PostOpObservation: {models.PostOpDischarged},
	}
	valid, ok := transitions[current]
	if !ok {
		return false
	}
	for _, v := range valid {
		if v == next {
			return true
		}
	}
	return false
}

func (s *Service) urgencyWeight(level models.UrgencyLevel) int {
	switch level {
	case models.UrgencyEmergency:
		return 3
	case models.UrgencyModerate:
		return 2
	case models.UrgencyNormal:
		return 1
	default:
		return 0
	}
}

func (s *Service) createTodo(ctx context.Context, todoType models.TodoType, relatedID, relatedType, assignee, locationID string, dueDate time.Time, description string) *models.Todo {
	todo := &models.Todo{
		ID:          uuid.New().String(),
		Type:        todoType,
		RelatedID:   relatedID,
		RelatedType: relatedType,
		Assignee:    assignee,
		DueDate:     dueDate,
		Status:      models.TodoStatusPending,
		Description: description,
		LocationID:  locationID,
	}
	s.repo.CreateTodo(ctx, todo)
	return todo
}

func (s *Service) RegisterDonor(ctx context.Context, donor *models.Donor) error {
	if _, exists := s.repo.GetDonorByNo(ctx, donor.DonorNo); exists {
		return apperrors.ErrDonorNoDuplicate
	}

	donor.ID = uuid.New().String()
	for _, organ := range donor.Organs {
		organ.ID = uuid.New().String()
		organ.DonorID = donor.ID
		organ.Status = models.OrganStatusPending
		s.repo.CreateOrgan(ctx, organ)
	}

	err := s.repo.CreateDonor(ctx, donor)
	if err != nil {
		return err
	}

	for _, organ := range donor.Organs {
		s.createTodo(ctx, models.TodoTypeOrganAssessment, organ.ID, "organ",
			"协调员", donor.LocationID, time.Now().Add(24*time.Hour),
			"请对捐献器官进行评估")
	}

	return nil
}

func (s *Service) AssessOrgan(ctx context.Context, organID string, assessment *models.OrganAssessment) error {
	organ, exists := s.repo.GetOrgan(ctx, organID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if organ.Status != models.OrganStatusPending {
		return apperrors.ErrInvalidOrganStatus
	}

	if assessment.FunctionScore < 0 || assessment.FunctionScore > 100 {
		return apperrors.ErrInvalidRequest
	}

	assessment.AssessmentTime = time.Now()
	organ.Assessment = assessment

	if assessment.FunctionScore < 60 {
		organ.Status = models.OrganStatusDiscarded
	}

	s.repo.UpdateOrgan(ctx, organ)

	if assessment.FunctionScore >= 60 {
		donor, _ := s.repo.GetDonor(ctx, organ.DonorID)
		s.createTodo(ctx, models.TodoTypeWaitingMatching, organ.ID, "organ",
			"系统", donor.LocationID, time.Now().Add(72*time.Hour),
			"等待系统匹配合适的受体")
	}

	return nil
}

func (s *Service) RegisterRecipient(ctx context.Context, recipient *models.Recipient) error {
	if recipient.PRA < 0 || recipient.PRA > 100 {
		return apperrors.ErrInvalidPRA
	}

	recipient.ID = uuid.New().String()
	recipient.IsMatching = false
	s.repo.CreateRecipient(ctx, recipient)
	return nil
}

func (s *Service) FindBestMatch(ctx context.Context, organID string) ([]*models.Recipient, error) {
	organ, exists := s.repo.GetOrgan(ctx, organID)
	if !exists {
		return nil, apperrors.ErrNotFound
	}

	if organ.Assessment == nil || organ.Assessment.FunctionScore < 60 {
		return nil, apperrors.ErrOrganScoreTooLow
	}

	if organ.Status == models.OrganStatusDiscarded {
		return nil, apperrors.ErrOrganColdIschemiaTimeout
	}

	donor, _ := s.repo.GetDonor(ctx, organ.DonorID)
	recipients := s.repo.ListWaitingRecipients(ctx)

	filtered := make([]*models.Recipient, 0)
	for _, r := range recipients {
		if r.OrganNeeded != organ.OrganType {
			continue
		}

		if organ.OrganType != models.OrganTypeCornea {
			if !s.isBloodTypeCompatible(donor.BloodType, r.BloodType) {
				continue
			}
		}

		filtered = append(filtered, r)
	}

	sort.Slice(filtered, func(i, j int) bool {
		a := filtered[i]
		b := filtered[j]

		if s.urgencyWeight(a.UrgencyLevel) != s.urgencyWeight(b.UrgencyLevel) {
			return s.urgencyWeight(a.UrgencyLevel) > s.urgencyWeight(b.UrgencyLevel)
		}

		if a.PRA != b.PRA {
			return a.PRA > b.PRA
		}

		if !a.RegistrationDate.Equal(b.RegistrationDate) {
			return a.RegistrationDate.Before(b.RegistrationDate)
		}

		aIsChild := a.Age < 12
		bIsChild := b.Age < 12
		if aIsChild != bIsChild {
			return aIsChild
		}

		return false
	})

	return filtered, nil
}

func (s *Service) StartMatching(ctx context.Context, organID string) error {
	organ, exists := s.repo.GetOrgan(ctx, organID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if organ.Status != models.OrganStatusPending {
		return apperrors.ErrInvalidOrganStatus
	}

	if organ.Assessment == nil || organ.Assessment.FunctionScore < 60 {
		return apperrors.ErrOrganScoreTooLow
	}

	candidates, err := s.FindBestMatch(ctx, organID)
	if err != nil {
		return err
	}

	if len(candidates) == 0 {
		return nil
	}

	for i, candidate := range candidates {
		candidate.IsMatching = true
		s.repo.UpdateRecipient(ctx, candidate)

		notification := &models.MatchNotification{
			ID:          uuid.New().String(),
			OrganID:     organID,
			RecipientID: candidate.ID,
			NotifiedAt:  time.Now(),
			Confirmed:   false,
			ExpiresAt:   time.Now().Add(2 * time.Hour),
			Priority:    i,
		}
		s.repo.CreateNotification(ctx, notification)
	}

	donor, _ := s.repo.GetDonor(ctx, organ.DonorID)
	s.createTodo(ctx, models.TodoTypeNotifyRecipient, candidates[0].ID, "recipient",
		"协调员", donor.LocationID, time.Now().Add(2*time.Hour),
		"请通知并确认受体是否接受器官")

	return nil
}

func (s *Service) ConfirmMatch(ctx context.Context, organID, recipientID string, confirmed bool) error {
	organ, exists := s.repo.GetOrgan(ctx, organID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if organ.Status != models.OrganStatusPending {
		return apperrors.ErrOrganAlreadyMatched
	}

	recipient, exists := s.repo.GetRecipient(ctx, recipientID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if confirmed {
		if !s.isValidOrganStatusTransition(organ.Status, models.OrganStatusMatched) {
			return apperrors.ErrInvalidOrganStatus
		}

		organ.Status = models.OrganStatusMatched
		s.repo.UpdateOrgan(ctx, organ)

		recipient.MatchedOrganID = &organID
		recipient.IsMatching = false
		s.repo.UpdateRecipient(ctx, recipient)

		donor, _ := s.repo.GetDonor(ctx, organ.DonorID)
		s.createTodo(ctx, models.TodoTypeScheduleSurgery, organID, "organ",
			"手术协调员", donor.LocationID, time.Now().Add(24*time.Hour),
			"请安排移植手术")

		return nil
	}

	recipient.IsMatching = false
	s.repo.UpdateRecipient(ctx, recipient)

	return nil
}

func (s *Service) AcquireOrgan(ctx context.Context, organID string, coldIschemiaHours int) error {
	organ, exists := s.repo.GetOrgan(ctx, organID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if !s.isValidOrganStatusTransition(organ.Status, models.OrganStatusAcquired) {
		return apperrors.ErrInvalidOrganStatus
	}

	organ.Status = models.OrganStatusAcquired
	organ.AcquiredTime = time.Now()
	organ.ColdIschemiaDeadline = time.Now().Add(time.Duration(coldIschemiaHours) * time.Hour)
	s.repo.UpdateOrgan(ctx, organ)

	return nil
}

func (s *Service) CheckColdIschemiaTimeout(ctx context.Context) {
	organs := s.repo.ListPendingOrgans(ctx)
	now := time.Now()

	for _, organ := range organs {
		if organ.Status == models.OrganStatusAcquired && !organ.ColdIschemiaDeadline.IsZero() && now.After(organ.ColdIschemiaDeadline) {
			organ.Status = models.OrganStatusDiscarded
			s.repo.UpdateOrgan(ctx, organ)
		}
	}
}

func (s *Service) CreateTransplant(ctx context.Context, transplant *models.TransplantRecord) error {
	organ, exists := s.repo.GetOrgan(ctx, transplant.OrganID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if organ.Status != models.OrganStatusAcquired {
		return apperrors.ErrInvalidOrganStatus
	}

	recipient, exists := s.repo.GetRecipient(ctx, transplant.RecipientID)
	if !exists {
		return apperrors.ErrNotFound
	}

	if recipient.MatchedOrganID == nil || *recipient.MatchedOrganID != transplant.OrganID {
		return apperrors.ErrBloodTypeMismatch
	}

	donor, _ := s.repo.GetDonor(ctx, organ.DonorID)
	if organ.OrganType != models.OrganTypeCornea {
		if !s.isBloodTypeCompatible(donor.BloodType, recipient.BloodType) {
			return apperrors.ErrBloodTypeMismatch
		}
	}

	transplant.ID = uuid.New().String()
	transplant.PostOpStatus = models.PostOpInSurgery
	s.repo.CreateTransplant(ctx, transplant)

	organ.Status = models.OrganStatusTransplanted
	s.repo.UpdateOrgan(ctx, organ)

	locationID := ""
	if donor != nil {
		locationID = donor.LocationID
	}
	s.createTodo(ctx, models.TodoTypePostOpFollowUp, transplant.ID, "transplant",
		"随访护士", locationID, transplant.SurgeryDate.AddDate(0, 1, 0),
		"术后1个月随访")

	return nil
}

func (s *Service) UpdatePostOpStatus(ctx context.Context, transplantID string, newStatus models.PostOpStatus) error {
	transplant, exists := s.repo.GetTransplant(ctx, transplantID)
	if !exists {
		return apperrors.ErrTransplantNotFound
	}

	if !s.isValidPostOpStatusTransition(transplant.PostOpStatus, newStatus) {
		return apperrors.ErrInvalidPostOpStatus
	}

	transplant.PostOpStatus = newStatus
	s.repo.UpdateTransplant(ctx, transplant)

	return nil
}

func (s *Service) AddFollowUp(ctx context.Context, followUp *models.FollowUp) error {
	_, exists := s.repo.GetTransplant(ctx, followUp.TransplantID)
	if !exists {
		return apperrors.ErrTransplantNotFound
	}

	followUp.ID = uuid.New().String()
	s.repo.CreateFollowUp(ctx, followUp)

	return nil
}

func (s *Service) UpgradeUrgencyLevel(ctx context.Context) {
	recipients := s.repo.GetAllRecipients(ctx)
	now := time.Now()

	for _, r := range recipients {
		if r.MatchedOrganID != nil {
			continue
		}

		waitDays := now.Sub(r.RegistrationDate).Hours() / 24
		if waitDays >= 365 {
			switch r.UrgencyLevel {
			case models.UrgencyNormal:
				r.UrgencyLevel = models.UrgencyModerate
			case models.UrgencyModerate:
				r.UrgencyLevel = models.UrgencyEmergency
			}
			s.repo.UpdateRecipient(ctx, r)
		}
	}
}

func (s *Service) GenerateWeeklyReport(ctx context.Context) *models.StatisticsReport {
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	newDonors := s.repo.CountNewDonorsSince(ctx, oneMonthAgo)
	successfulMatches := s.repo.CountSuccessfulMatchesSince(ctx, oneMonthAgo)

	recipients := s.repo.GetAllRecipients(ctx)
	var totalWaitDays float64
	var matchedCount int
	for _, r := range recipients {
		if r.MatchedOrganID != nil {
			waitDays := now.Sub(r.RegistrationDate).Hours() / 24
			totalWaitDays += waitDays
			matchedCount++
		}
	}

	avgWaitTime := 0.0
	if matchedCount > 0 {
		avgWaitTime = totalWaitDays / float64(matchedCount)
	}

	organSupply := s.repo.CountOrgansByType(ctx)
	organDemand := s.repo.CountRecipientsByOrganNeeded(ctx)

	supplyDemand := make(map[models.OrganType]float64)
	for organType, supply := range organSupply {
		demand := organDemand[organType]
		if demand > 0 {
			supplyDemand[organType] = float64(supply) / float64(demand)
		} else {
			supplyDemand[organType] = 0
		}
	}

	report := &models.StatisticsReport{
		ID:               uuid.New().String(),
		ReportDate:       now,
		NewDonors:        newDonors,
		SuccessfulMatches: successfulMatches,
		AverageWaitTime:  avgWaitTime,
		OrganSupplyDemand: supplyDemand,
	}

	s.repo.CreateReport(ctx, report)
	return report
}

func (s *Service) GetDonor(ctx context.Context, id string) (*models.Donor, bool) {
	donor, ok := s.repo.GetDonor(ctx, id)
	if ok {
		donor.Organs = s.repo.ListOrgansByDonor(ctx, id)
	}
	return donor, ok
}

func (s *Service) ListDonors(ctx context.Context, page, size int) ([]*models.Donor, int) {
	return s.repo.ListDonors(ctx, page, size)
}

func (s *Service) GetRecipient(ctx context.Context, id string) (*models.Recipient, bool) {
	return s.repo.GetRecipient(ctx, id)
}

func (s *Service) ListRecipients(ctx context.Context, page, size int) ([]*models.Recipient, int) {
	return s.repo.ListRecipients(ctx, page, size)
}

func (s *Service) GetTransplant(ctx context.Context, id string) (*models.TransplantRecord, bool) {
	transplant, ok := s.repo.GetTransplant(ctx, id)
	if ok {
		transplant.FollowUps = s.repo.ListFollowUpsByTransplant(ctx, id)
	}
	return transplant, ok
}

func (s *Service) ListTransplants(ctx context.Context, page, size int) ([]*models.TransplantRecord, int) {
	return s.repo.ListTransplants(ctx, page, size)
}

func (s *Service) ListTodos(ctx context.Context, page, size int) ([]*models.Todo, int) {
	return s.repo.ListTodos(ctx, page, size)
}

func (s *Service) UpdateTodoStatus(ctx context.Context, todoID string, status models.TodoStatus) error {
	todo, exists := s.repo.GetTodo(ctx, todoID)
	if !exists {
		return apperrors.ErrNotFound
	}
	todo.Status = status
	s.repo.UpdateTodo(ctx, todo)
	return nil
}

func (s *Service) CreateLocation(ctx context.Context, location *models.Location) {
	location.ID = uuid.New().String()
	s.repo.CreateLocation(ctx, location)
}

func (s *Service) ListLocations(ctx context.Context) []*models.Location {
	return s.repo.ListLocations(ctx)
}

func (s *Service) CreatePersonnel(ctx context.Context, personnel *models.Personnel) {
	personnel.ID = uuid.New().String()
	s.repo.CreatePersonnel(ctx, personnel)
}

func (s *Service) ListPersonnel(ctx context.Context) []*models.Personnel {
	return s.repo.ListPersonnel(ctx)
}

func (s *Service) ListReports(ctx context.Context) []*models.StatisticsReport {
	return s.repo.ListReports(ctx)
}

package visitor

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"smart-park/common"
	"smart-park/core/access"
)

const DailyVisitorLimit = 5

type EmployeeService interface {
	GetEmployee(id string) *common.Employee
}

type Service struct {
	store           *Store
	employeeService EmployeeService
}

func NewService(store *Store, empSvc EmployeeService) *Service {
	return &Service{
		store:           store,
		employeeService: empSvc,
	}
}

func (s *Service) CreateReservation(
	employeeID, visitorName, visitorPhone, purpose string,
	expectedArrival time.Time,
) (*common.VisitorReservation, error) {
	if s.employeeService.GetEmployee(employeeID) == nil {
		return nil, errors.New("employee not found")
	}

	if visitorName == "" || visitorPhone == "" {
		return nil, errors.New("visitor name and phone are required")
	}

	normalizedPhone := access.NormalizePhone(visitorPhone)
	if len(normalizedPhone) == 0 || !isValidPhone(normalizedPhone) {
		return nil, errors.New("invalid phone number")
	}

	if err := s.checkDailyLimit(employeeID); err != nil {
		return nil, err
	}

	r := &common.VisitorReservation{
		ID:              uuid.New().String(),
		EmployeeID:      employeeID,
		VisitorName:     visitorName,
		VisitorPhone:    normalizedPhone,
		Purpose:         purpose,
		ExpectedArrival: expectedArrival,
		Status:          StatusPending,
		CreatedAt:       time.Now(),
	}
	s.store.AddReservation(r)
	return r, nil
}

func (s *Service) checkDailyLimit(employeeID string) error {
	today := time.Now().Format("2006-01-02")
	count := 0
	reservations := s.store.GetReservationsByEmployee(employeeID)
	for _, r := range reservations {
		if r.CreatedAt.Format("2006-01-02") == today {
			count++
		}
	}
	if count >= DailyVisitorLimit {
		return errors.New("daily visitor limit exceeded (max 5 per day)")
	}
	return nil
}

func (s *Service) ReviewReservation(reservationID string, approve bool) (*common.VisitorReservation, error) {
	r := s.store.GetReservation(reservationID)
	if r == nil {
		return nil, errors.New("reservation not found")
	}
	if r.Status != StatusPending {
		return nil, errors.New("reservation already reviewed")
	}
	if approve {
		r.Status = StatusApproved
	} else {
		r.Status = StatusRejected
	}
	r.ReviewedAt = time.Now()
	s.store.UpdateReservation(r)
	return r, nil
}

func (s *Service) CheckInVisitor(visitorPhone string) (*common.VisitorReservation, error) {
	normalizedPhone := access.NormalizePhone(visitorPhone)
	r := s.store.GetReservationByPhone(normalizedPhone)
	if r == nil {
		return nil, errors.New("reservation not found")
	}

	now := time.Now()

	if r.Status != StatusApproved {
		return nil, errors.New("reservation is not approved")
	}

	if now.Before(r.ExpectedArrival) {
		return nil, errors.New("cannot check in before expected arrival time")
	}

	if now.After(r.ExpectedArrival.Add(2 * time.Hour)) {
		r.Status = StatusExpired
		s.store.UpdateReservation(r)
		return nil, errors.New("reservation expired")
	}

	r.Status = StatusCheckedIn
	r.CheckInAt = now
	s.store.UpdateReservation(r)
	return r, nil
}

func (s *Service) GetAllReservations() []*common.VisitorReservation {
	return s.store.GetAllReservations()
}

func isValidPhone(phone string) bool {
	if len(phone) < 7 || len(phone) > 15 {
		return false
	}
	for _, c := range phone {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func NormalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	for strings.HasPrefix(phone, "0") {
		phone = phone[1:]
	}
	return phone
}

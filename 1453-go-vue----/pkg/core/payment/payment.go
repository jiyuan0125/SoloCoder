package payment

import (
	"errors"
	"math"
	"property-management/pkg/core"
	"property-management/pkg/core/models"
	"time"

	"github.com/google/uuid"
)

const (
	PenaltyRate = 0.0005
)

type Service struct {
	store *core.Store
}

func NewService(store *core.Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateBill(userID string, billType models.BillType, amount int64, dueDate time.Time, month, year int) (*models.Bill, error) {
	s.store.Lock()
	defer s.store.Unlock()

	user, ok := s.store.Users()[userID]
	if !ok {
		return nil, errors.New("user not found")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	now := time.Now()
	bill := &models.Bill{
		ID:           uuid.New().String(),
		UserID:       userID,
		UserName:     user.Name,
		Building:     user.Building,
		Unit:         user.Unit,
		BillType:     billType,
		Amount:       amount,
		PaidAmount:   0,
		UnpaidAmount: amount,
		DueDate:      dueDate,
		Status:       models.BillStatusPending,
		PenaltyAmount: 0,
		Month:        month,
		Year:         year,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	s.store.SetBill(bill.ID, bill)
	return bill, nil
}

func (s *Service) CreateRepairBill(repairID string, amount int64, userID string, month, year int) (*models.Bill, error) {
	s.store.Lock()
	defer s.store.Unlock()

	user, ok := s.store.Users()[userID]
	if !ok {
		return nil, errors.New("user not found")
	}

	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	now := time.Now()
	dueDate := now.AddDate(0, 1, 0)

	bill := &models.Bill{
		ID:              uuid.New().String(),
		UserID:          userID,
		UserName:        user.Name,
		Building:        user.Building,
		Unit:            user.Unit,
		BillType:        models.BillTypeRepair,
		Amount:          amount,
		PaidAmount:      0,
		UnpaidAmount:    amount,
		DueDate:         dueDate,
		Status:          models.BillStatusPending,
		RelatedRepairID: repairID,
		PenaltyAmount:   0,
		Month:           month,
		Year:            year,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	s.store.SetBill(bill.ID, bill)
	return bill, nil
}

func (s *Service) GenerateMonthlyBills(billType models.BillType, amount int64, dueDate time.Time, month, year int) ([]*models.Bill, error) {
	s.store.Lock()
	defer s.store.Unlock()

	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	now := time.Now()
	var bills []*models.Bill

	for _, user := range s.store.Users() {
		if user.Role != models.RoleOwner {
			continue
		}

		exists := false
		for _, bill := range s.store.Bills() {
			if bill.UserID == user.ID && bill.BillType == billType && bill.Month == month && bill.Year == year {
				exists = true
				break
			}
		}

		if exists {
			continue
		}

		bill := &models.Bill{
			ID:            uuid.New().String(),
			UserID:        user.ID,
			UserName:      user.Name,
			Building:      user.Building,
			Unit:          user.Unit,
			BillType:      billType,
			Amount:        amount,
			PaidAmount:    0,
			UnpaidAmount:  amount,
			DueDate:       dueDate,
			Status:        models.BillStatusPending,
			PenaltyAmount: 0,
			Month:         month,
			Year:          year,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		s.store.SetBill(bill.ID, bill)
		bills = append(bills, bill)
	}

	return bills, nil
}

func (s *Service) CalculatePenalty(billID string) (penaltyAmount int64, overdueDays int, err error) {
	s.store.RLock()
	defer s.store.RUnlock()

	bill, ok := s.store.Bills()[billID]
	if !ok {
		return 0, 0, errors.New("bill not found")
	}

	now := time.Now()
	if now.Before(bill.DueDate) {
		return 0, 0, nil
	}

	overdueDays = int(math.Ceil(now.Sub(bill.DueDate).Hours() / 24))
	baseAmount := bill.UnpaidAmount
	penalty := float64(baseAmount) * PenaltyRate * float64(overdueDays)
	penaltyAmount = int64(math.Round(penalty))

	return penaltyAmount, overdueDays, nil
}

func (s *Service) UpdatePenalty(billID string) error {
	s.store.Lock()
	defer s.store.Unlock()

	bill, ok := s.store.Bills()[billID]
	if !ok {
		return errors.New("bill not found")
	}

	now := time.Now()
	if now.Before(bill.DueDate) {
		bill.PenaltyAmount = 0
		bill.UpdatedAt = now
		s.store.SetBill(bill.ID, bill)
		return nil
	}

	overdueDays := int(math.Ceil(now.Sub(bill.DueDate).Hours() / 24))
	baseAmount := bill.UnpaidAmount
	penalty := float64(baseAmount) * PenaltyRate * float64(overdueDays)
	bill.PenaltyAmount = int64(math.Round(penalty))
	bill.UpdatedAt = now
	s.store.SetBill(bill.ID, bill)

	return nil
}

func (s *Service) PayBill(billID string, userID string, amount int64, paymentMethod string) (*models.Payment, error) {
	s.store.Lock()
	defer s.store.Unlock()

	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	bill, ok := s.store.Bills()[billID]
	if !ok {
		return nil, errors.New("bill not found")
	}

	if bill.UserID != userID {
		return nil, errors.New("you are not the owner of this bill")
	}

	if bill.Status == models.BillStatusPaid {
		return nil, errors.New("bill is already paid")
	}

	now := time.Now()
	if now.After(bill.DueDate) {
		overdueDays := int(math.Ceil(now.Sub(bill.DueDate).Hours() / 24))
		baseAmount := bill.UnpaidAmount
		penalty := float64(baseAmount) * PenaltyRate * float64(overdueDays)
		bill.PenaltyAmount = int64(math.Round(penalty))
	}

	totalDue := bill.UnpaidAmount + bill.PenaltyAmount
	if amount > totalDue {
		amount = totalDue
	}

	payment := &models.Payment{
		ID:            uuid.New().String(),
		BillID:        billID,
		UserID:        userID,
		Amount:        amount,
		PaidAt:        now,
		PaymentMethod: paymentMethod,
		TransactionNo: uuid.New().String(),
	}

	remainingAmount := amount
	if bill.PenaltyAmount > 0 {
		penaltyPaid := remainingAmount
		if penaltyPaid > bill.PenaltyAmount {
			penaltyPaid = bill.PenaltyAmount
		}
		bill.PenaltyAmount -= penaltyPaid
		remainingAmount -= penaltyPaid
	}

	if remainingAmount > 0 {
		bill.PaidAmount += remainingAmount
		bill.UnpaidAmount -= remainingAmount
	}

	if bill.UnpaidAmount <= 0 {
		bill.Status = models.BillStatusPaid
		bill.UnpaidAmount = 0
	} else {
		bill.Status = models.BillStatusPartial
	}

	bill.UpdatedAt = now

	s.store.SetBill(bill.ID, bill)
	s.store.SetPayment(payment.ID, payment)

	return payment, nil
}

func (s *Service) GetBillByID(id string) (*models.Bill, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	bill, ok := s.store.Bills()[id]
	if !ok {
		return nil, errors.New("bill not found")
	}
	return bill, nil
}

func (s *Service) ListBillsByUser(userID string) []*models.Bill {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Bill
	for _, bill := range s.store.Bills() {
		if bill.UserID == userID {
			result = append(result, bill)
		}
	}
	return result
}

func (s *Service) ListBillsByStatus(status models.BillStatus) []*models.Bill {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Bill
	for _, bill := range s.store.Bills() {
		if bill.Status == status {
			result = append(result, bill)
		}
	}
	return result
}

func (s *Service) ListAllBills() []*models.Bill {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Bill
	for _, bill := range s.store.Bills() {
		result = append(result, bill)
	}
	return result
}

func (s *Service) ListPaymentsByBill(billID string) []*models.Payment {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Payment
	for _, payment := range s.store.Payments() {
		if payment.BillID == billID {
			result = append(result, payment)
		}
	}
	return result
}

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type BillStatus string

const (
	BillStatusPending    BillStatus = "pending"
	BillStatusPaid       BillStatus = "paid"
	BillStatusOverdue    BillStatus = "overdue"
	BillStatusSevereOverdue BillStatus = "severe_overdue"
)

type BillDetail struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type Bill struct {
	ID              string       `json:"id"`
	CustomerID      string       `json:"customer_id"`
	BillingYear     int          `json:"billing_year"`
	BillingMonth    time.Month   `json:"billing_month"`
	PlanCost        float64      `json:"plan_cost"`
	SMSUsageCost    float64      `json:"sms_usage_cost"`
	StorageUsageCost float64     `json:"storage_usage_cost"`
	TotalAmount     float64      `json:"total_amount"`
	Status          BillStatus   `json:"status"`
	GeneratedAt     time.Time    `json:"generated_at"`
	DueDate         time.Time    `json:"due_date"`
	PaidAt          *time.Time   `json:"paid_at,omitempty"`
	PaidBy          string       `json:"paid_by,omitempty"`
	PaymentNote     string       `json:"payment_note,omitempty"`
	Details         []BillDetail `json:"details"`
}

type PaymentRecord struct {
	ID          string    `json:"id"`
	BillID      string    `json:"bill_id"`
	CustomerID  string    `json:"customer_id"`
	PaidBy      string    `json:"paid_by"`
	PaidAt      time.Time `json:"paid_at"`
	Note        string    `json:"note"`
	Amount      float64   `json:"amount"`
}

type BillingStore interface {
	Save(bill *Bill) error
	FindByID(id string) (*Bill, error)
	FindByCustomer(customerID string) ([]*Bill, error)
	FindByMonth(year int, month time.Month) ([]*Bill, error)
	FindAll() ([]*Bill, error)
	HasUnpaidBillBefore(customerID string, year int, month time.Month) (bool, error)
	SavePaymentRecord(record *PaymentRecord) error
	FindPaymentRecordsByBill(billID string) ([]*PaymentRecord, error)
}

type FileBillingStore struct {
	filePath        string
	bills           map[string]*Bill
	paymentRecords  []*PaymentRecord
	mu              sync.RWMutex
}

func (s *FileBillingStore) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bills == nil {
		s.bills = make(map[string]*Bill)
	}
	if s.paymentRecords == nil {
		s.paymentRecords = []*PaymentRecord{}
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var stored struct {
		Bills          map[string]*Bill  `json:"bills"`
		PaymentRecords []*PaymentRecord  `json:"payment_records"`
	}

	if err := json.Unmarshal(data, &stored); err != nil {
		return err
	}

	if stored.Bills != nil {
		s.bills = stored.Bills
	}
	if stored.PaymentRecords != nil {
		s.paymentRecords = stored.PaymentRecords
	}

	return nil
}

func (s *FileBillingStore) saveToFile() error {
	stored := struct {
		Bills          map[string]*Bill  `json:"bills"`
		PaymentRecords []*PaymentRecord  `json:"payment_records"`
	}{
		Bills:          s.bills,
		PaymentRecords: s.paymentRecords,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *FileBillingStore) Save(bill *Bill) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.bills[bill.ID] = bill
	return s.saveToFile()
}

func (s *FileBillingStore) FindByID(id string) (*Bill, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bill, exists := s.bills[id]
	if !exists {
		return nil, fmt.Errorf("bill not found")
	}
	return bill, nil
}

func (s *FileBillingStore) FindByCustomer(customerID string) ([]*Bill, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Bill
	for _, b := range s.bills {
		if b.CustomerID == customerID {
			result = append(result, b)
		}
	}
	return result, nil
}

func (s *FileBillingStore) FindByMonth(year int, month time.Month) ([]*Bill, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Bill
	for _, b := range s.bills {
		if b.BillingYear == year && b.BillingMonth == month {
			result = append(result, b)
		}
	}
	return result, nil
}

func (s *FileBillingStore) FindAll() ([]*Bill, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Bill, 0, len(s.bills))
	for _, b := range s.bills {
		result = append(result, b)
	}
	return result, nil
}

func (s *FileBillingStore) HasUnpaidBillBefore(customerID string, year int, month time.Month) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, b := range s.bills {
		if b.CustomerID != customerID {
			continue
		}
		if b.Status == BillStatusPaid {
			continue
		}
		if b.BillingYear < year || (b.BillingYear == year && b.BillingMonth < month) {
			return true, nil
		}
	}
	return false, nil
}

func (s *FileBillingStore) SavePaymentRecord(record *PaymentRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.paymentRecords = append(s.paymentRecords, record)
	return s.saveToFile()
}

func (s *FileBillingStore) FindPaymentRecordsByBill(billID string) ([]*PaymentRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*PaymentRecord
	for _, r := range s.paymentRecords {
		if r.BillID == billID {
			result = append(result, r)
		}
	}
	return result, nil
}

type BillingService struct {
	store         BillingStore
	customerStore CustomerStore
	usageStore    UsageStore
}

func NewBillingService(store BillingStore, customerStore CustomerStore, usageStore UsageStore) *BillingService {
	return &BillingService{
		store:         store,
		customerStore: customerStore,
		usageStore:    usageStore,
	}
}

type GenerateBillsRequest struct {
	Year  int `json:"year"`
	Month int `json:"month"`
}

type MarkBillPaidRequest struct {
	BillID   string `json:"bill_id"`
	PaidBy   string `json:"paid_by"`
	Note     string `json:"note"`
	Amount   float64 `json:"amount"`
}

func (s *BillingService) GenerateMonthlyBills(w http.ResponseWriter, r *http.Request) {
	var req GenerateBillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	now := time.Now()
	year := req.Year
	month := time.Month(req.Month)

	if year == 0 {
		year = now.Year()
	}
	if month == 0 {
		month = now.Month()
	}

	customers, err := s.customerStore.FindAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list customers: %v", err), http.StatusInternalServerError)
		return
	}

	var generatedBills []*Bill
	for _, customer := range customers {
		bill, err := s.generateBillForCustomer(customer, year, month)
		if err != nil {
			continue
		}
		generatedBills = append(generatedBills, bill)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"year":        year,
		"month":       month,
		"bills_count": len(generatedBills),
		"bills":       generatedBills,
	})
}

func (s *BillingService) generateBillForCustomer(customer *Customer, year int, month time.Month) (*Bill, error) {
	planCost, err := calculatePlanCostForMonth(customer.ID, year, month, s.customerStore)
	if err != nil {
		return nil, err
	}

	smsCost, storageCost, err := calculateUsageCost(customer.ID, year, month, s.usageStore, s.customerStore)
	if err != nil {
		return nil, err
	}

	totalAmount := planCost + smsCost + storageCost

	hasUnpaid, _ := s.store.HasUnpaidBillBefore(customer.ID, year, month)

	status := BillStatusPending
	if hasUnpaid {
		status = BillStatusOverdue
	}

	daysInMonth := getDaysInMonth(year, month)
	dueDate := time.Date(year, month, daysInMonth, 23, 59, 59, 0, time.Local)

	var details []BillDetail
	details = append(details, BillDetail{
		Description: fmt.Sprintf("套餐费用 (%d年%d月)", year, month),
		Amount:      math.Round(planCost*100) / 100,
	})

	if smsCost > 0 {
		details = append(details, BillDetail{
			Description: "短信超出套餐用量费用",
			Amount:      math.Round(smsCost*100) / 100,
		})
	}

	if storageCost > 0 {
		details = append(details, BillDetail{
			Description: "云存储超出套餐用量费用",
			Amount:      math.Round(storageCost*100) / 100,
		})
	}

	bill := &Bill{
		ID:               uuid.New().String(),
		CustomerID:       customer.ID,
		BillingYear:      year,
		BillingMonth:     month,
		PlanCost:         math.Round(planCost*100) / 100,
		SMSUsageCost:     math.Round(smsCost*100) / 100,
		StorageUsageCost: math.Round(storageCost*100) / 100,
		TotalAmount:      math.Round(totalAmount*100) / 100,
		Status:           status,
		GeneratedAt:      time.Now(),
		DueDate:          dueDate,
		Details:          details,
	}

	if err := s.store.Save(bill); err != nil {
		return nil, err
	}

	return bill, nil
}

func (s *BillingService) ListAllBills(w http.ResponseWriter, r *http.Request) {
	bills, err := s.store.FindAll()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list bills: %v", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	for _, bill := range bills {
		s.updateBillStatus(bill, now)
	}

	sort.Slice(bills, func(i, j int) bool {
		if bills[i].BillingYear != bills[j].BillingYear {
			return bills[i].BillingYear > bills[j].BillingYear
		}
		return bills[i].BillingMonth > bills[j].BillingMonth
	})

	type BillWithCustomer struct {
		*Bill
		CustomerName  string `json:"customer_name"`
		CustomerEmail string `json:"customer_email"`
		IsSevereOverdue bool `json:"is_severe_overdue"`
	}

	result := make([]BillWithCustomer, 0, len(bills))
	for _, b := range bills {
		customer, _ := s.customerStore.FindByID(b.CustomerID)
		name := ""
		email := ""
		if customer != nil {
			name = customer.Name
			email = customer.Email
		}
		result = append(result, BillWithCustomer{
			Bill:           b,
			CustomerName:   name,
			CustomerEmail:  email,
			IsSevereOverdue: b.Status == BillStatusSevereOverdue,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *BillingService) ListCustomerBills(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/bills/customer/")
	if path == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	customer, err := s.customerStore.FindByID(path)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	bills, err := s.store.FindByCustomer(customer.ID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list customer bills: %v", err), http.StatusInternalServerError)
		return
	}

	now := time.Now()
	for _, bill := range bills {
		s.updateBillStatus(bill, now)
	}

	sort.Slice(bills, func(i, j int) bool {
		if bills[i].BillingYear != bills[j].BillingYear {
			return bills[i].BillingYear > bills[j].BillingYear
		}
		return bills[i].BillingMonth > bills[j].BillingMonth
	})

	type BillWithPayments struct {
		*Bill
		Payments []*PaymentRecord `json:"payments"`
	}

	result := make([]BillWithPayments, 0, len(bills))
	for _, b := range bills {
		payments, _ := s.store.FindPaymentRecordsByBill(b.ID)
		result = append(result, BillWithPayments{
			Bill:     b,
			Payments: payments,
		})
	}

	response := map[string]interface{}{
		"customer": customer,
		"plan":     getPlan(customer.CurrentPlan),
		"bills":    result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *BillingService) MarkBillAsPaid(w http.ResponseWriter, r *http.Request) {
	var req MarkBillPaidRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.BillID == "" {
		http.Error(w, "Bill ID is required", http.StatusBadRequest)
		return
	}

	if req.PaidBy == "" {
		http.Error(w, "PaidBy is required", http.StatusBadRequest)
		return
	}

	bill, err := s.store.FindByID(req.BillID)
	if err != nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	if bill.Status == BillStatusPaid {
		http.Error(w, "Bill is already paid", http.StatusBadRequest)
		return
	}

	now := time.Now()
	bill.Status = BillStatusPaid
	bill.PaidAt = &now
	bill.PaidBy = req.PaidBy
	bill.PaymentNote = req.Note

	paymentAmount := req.Amount
	if paymentAmount <= 0 {
		paymentAmount = bill.TotalAmount
	}

	paymentRecord := &PaymentRecord{
		ID:         uuid.New().String(),
		BillID:     bill.ID,
		CustomerID: bill.CustomerID,
		PaidBy:     req.PaidBy,
		PaidAt:     now,
		Note:       req.Note,
		Amount:     paymentAmount,
	}

	if err := s.store.Save(bill); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update bill: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.store.SavePaymentRecord(paymentRecord); err != nil {
		http.Error(w, fmt.Sprintf("Failed to record payment: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"bill":          bill,
		"payment_record": paymentRecord,
	})
}

func (s *BillingService) updateBillStatus(bill *Bill, now time.Time) {
	if bill.Status == BillStatusPaid {
		return
	}

	dueDate := bill.DueDate
	if dueDate.IsZero() {
		daysInMonth := getDaysInMonth(bill.BillingYear, bill.BillingMonth)
		dueDate = time.Date(bill.BillingYear, bill.BillingMonth, daysInMonth, 23, 59, 59, 0, time.Local)
	}

	daysOverdue := int(now.Sub(dueDate).Hours() / 24)

	if daysOverdue > 60 {
		if bill.Status != BillStatusSevereOverdue {
			bill.Status = BillStatusSevereOverdue
			s.store.Save(bill)
		}
	} else if daysOverdue > 0 {
		if bill.Status != BillStatusOverdue && bill.Status != BillStatusSevereOverdue {
			bill.Status = BillStatusOverdue
			s.store.Save(bill)
		}
	}
}

func (s *BillingService) GetBill(w http.ResponseWriter, r *http.Request) {
	billID := r.URL.Query().Get("bill_id")
	if billID == "" {
		http.Error(w, "Bill ID is required", http.StatusBadRequest)
		return
	}

	bill, err := s.store.FindByID(billID)
	if err != nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	payments, _ := s.store.FindPaymentRecordsByBill(billID)

	customer, _ := s.customerStore.FindByID(bill.CustomerID)

	result := map[string]interface{}{
		"bill":      bill,
		"payments":  payments,
		"customer":  customer,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (s *BillingService) GetBillsByMonth(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	var year int
	var month time.Month
	now := time.Now()

	if yearStr != "" {
		var err error
		year, err = strconv.Atoi(yearStr)
		if err != nil {
			http.Error(w, "Invalid year", http.StatusBadRequest)
			return
		}
	} else {
		year = now.Year()
	}

	if monthStr != "" {
		m, err := strconv.Atoi(monthStr)
		if err != nil {
			http.Error(w, "Invalid month", http.StatusBadRequest)
			return
		}
		month = time.Month(m)
	} else {
		month = now.Month()
	}

	bills, err := s.store.FindByMonth(year, month)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list bills: %v", err), http.StatusInternalServerError)
		return
	}

	nowTime := time.Now()
	for _, bill := range bills {
		s.updateBillStatus(bill, nowTime)
	}

	sort.Slice(bills, func(i, j int) bool {
		return bills[i].CustomerID < bills[j].CustomerID
	})

	type BillWithCustomer struct {
		*Bill
		CustomerName  string `json:"customer_name"`
		CustomerEmail string `json:"customer_email"`
	}

	result := make([]BillWithCustomer, 0, len(bills))
	for _, b := range bills {
		customer, _ := s.customerStore.FindByID(b.CustomerID)
		name := ""
		email := ""
		if customer != nil {
			name = customer.Name
			email = customer.Email
		}
		result = append(result, BillWithCustomer{
			Bill:          b,
			CustomerName:  name,
			CustomerEmail: email,
		})
	}

	totalAmount := 0.0
	paidAmount := 0.0
	for _, b := range bills {
		totalAmount += b.TotalAmount
		if b.Status == BillStatusPaid {
			paidAmount += b.TotalAmount
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"year":          year,
		"month":         month,
		"bills_count":   len(bills),
		"total_amount":  math.Round(totalAmount*100) / 100,
		"paid_amount":   math.Round(paidAmount*100) / 100,
		"bills":         result,
	})
}

package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
	"recycling/api"
)

type RecyclingService struct {
	Categories  *CategoryStore
	Records     *RecordStore
	Settlements *SettlementManager
}

func NewRecyclingService() *RecyclingService {
	return &RecyclingService{
		Categories:  NewCategoryStore(),
		Records:     NewRecordStore(),
		Settlements: NewSettlementManager(),
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *RecyclingService) CreateRecord(req api.CreateRecordRequest) (*api.Record, error) {
	record := &api.Record{
		ID:           generateID(),
		DateTime:     time.Now(),
		CustomerID:   req.CustomerID,
		CustomerName: req.CustomerName,
		CustomerType: req.CustomerType,
		Settlement:   req.Settlement,
		Items:        make([]api.RecordItemDetail, 0, len(req.Items)),
		TotalAmount:  0,
	}

	var totalAmount float64
	for _, item := range req.Items {
		category, exists := s.Categories.GetByID(item.CategoryID)
		if !exists {
			return nil, &CategoryNotFoundError{CategoryID: item.CategoryID}
		}

		price := category.Price
		amount := CalculateAmount(item.Weight, price)

		detail := api.RecordItemDetail{
			CategoryID:   item.CategoryID,
			CategoryName: category.Name,
			Weight:       item.Weight,
			Price:        price,
			Amount:       amount,
		}
		record.Items = append(record.Items, detail)
		totalAmount += amount
	}

	record.TotalAmount = roundToCents(totalAmount)
	s.Records.Create(record)

	if req.Settlement == "monthly" {
		year, month := record.DateTime.Year(), int(record.DateTime.Month())
		s.Settlements.AddToMonthlyTotal(req.CustomerID, year, month, record.TotalAmount)
	}

	return record, nil
}

type CategoryNotFoundError struct {
	CategoryID string
}

func (e *CategoryNotFoundError) Error() string {
	return "category not found: " + e.CategoryID
}

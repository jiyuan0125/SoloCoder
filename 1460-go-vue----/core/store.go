package core

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrInquiryNotFound     = errors.New("inquiry not found")
	ErrQuotationNotFound   = errors.New("quotation not found")
	ErrOrderNotFound       = errors.New("order not found")
	ErrDeadlineTooEarly    = errors.New("deadline must be at least 3 days from now")
	ErrInquiryExpired      = errors.New("inquiry has expired")
	ErrNotInvitedSupplier  = errors.New("supplier not invited to this inquiry")
	ErrInvalidMaterials    = errors.New("invalid materials")
	ErrAlreadyClosed       = errors.New("inquiry already closed")
)

type Store struct {
	mu               sync.RWMutex
	inquiries        map[string]*Inquiry
	quotations       map[string]*Quotation
	quotationsByInq  map[string]map[string]*Quotation
	todos            map[string][]*TodoReminder
	todosBySupplier  map[string][]*TodoReminder
	orders           map[string]*PurchaseOrder
	ordersByInquiry  map[string][]*PurchaseOrder
	ordersBySupplier map[string][]*PurchaseOrder
}

func NewStore() *Store {
	return &Store{
		inquiries:        make(map[string]*Inquiry),
		quotations:       make(map[string]*Quotation),
		quotationsByInq:  make(map[string]map[string]*Quotation),
		todos:            make(map[string][]*TodoReminder),
		todosBySupplier:  make(map[string][]*TodoReminder),
		orders:           make(map[string]*PurchaseOrder),
		ordersByInquiry:  make(map[string][]*PurchaseOrder),
		ordersBySupplier: make(map[string][]*PurchaseOrder),
	}
}

func (s *Store) CreateInquiry(inq *Inquiry) error {
	if inq.Deadline.Before(time.Now().Add(72 * time.Hour)) {
		return ErrDeadlineTooEarly
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	inq.ID = generateID("INQ")
	now := time.Now()
	inq.CreatedAt = now
	inq.UpdatedAt = now
	inq.Status = InquiryStatusOpen
	inq.MaterialsVersion = 1

	for i := range inq.Materials {
		inq.Materials[i].ID = generateID("MAT")
	}

	s.inquiries[inq.ID] = inq
	s.quotationsByInq[inq.ID] = make(map[string]*Quotation)

	for _, supplierID := range inq.SupplierIDs {
		todo := &TodoReminder{
			ID:           generateID("TODO"),
			SupplierID:   supplierID,
			InquiryID:    inq.ID,
			InquiryTitle: inq.Title,
			Deadline:     inq.Deadline,
			CreatedAt:    now,
		}
		s.todos[inq.ID] = append(s.todos[inq.ID], todo)
		s.todosBySupplier[supplierID] = append(s.todosBySupplier[supplierID], todo)
	}

	return nil
}

func (s *Store) GetInquiry(id string) (*Inquiry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inq, ok := s.inquiries[id]
	if !ok {
		return nil, ErrInquiryNotFound
	}
	return inq, nil
}

func (s *Store) ListInquiries() []*Inquiry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*Inquiry, 0, len(s.inquiries))
	for _, inq := range s.inquiries {
		list = append(list, inq)
	}
	return list
}

func (s *Store) UpdateInquiryMaterials(inquiryID string, materials []Material) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inq, ok := s.inquiries[inquiryID]
	if !ok {
		return ErrInquiryNotFound
	}

	if inq.Status == InquiryStatusClosed {
		return ErrAlreadyClosed
	}

	for i := range materials {
		if materials[i].ID == "" {
			materials[i].ID = generateID("MAT")
		}
	}
	inq.Materials = materials
	inq.MaterialsVersion++
	inq.UpdatedAt = time.Now()

	for _, q := range s.quotationsByInq[inquiryID] {
		if q.Status == QuotationStatusValid {
			q.Status = QuotationStatusInvalid
		}
	}

	return nil
}

func (s *Store) GetTodoReminders(supplierID string) []*TodoReminder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.todosBySupplier[supplierID]
}

func (s *Store) SubmitQuotation(q *Quotation, inquiry *Inquiry) error {
	if time.Now().After(inquiry.Deadline) {
		return ErrInquiryExpired
	}

	isInvited := false
	for _, sid := range inquiry.SupplierIDs {
		if sid == q.SupplierID {
			isInvited = true
			break
		}
	}
	if !isInvited {
		return ErrNotInvitedSupplier
	}

	materialIDs := make(map[string]bool)
	for _, m := range inquiry.Materials {
		materialIDs[m.ID] = true
	}

	for _, item := range q.Items {
		if !materialIDs[item.MaterialID] {
			return ErrInvalidMaterials
		}
		item.UnitPrice = roundPrice(item.UnitPrice)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	q.TotalPrice = calculateTotalPrice(inquiry.Materials, q.Items)
	q.ID = generateID("QOT")
	q.SubmittedAt = time.Now()
	q.Status = QuotationStatusValid
	q.MaterialsVersion = inquiry.MaterialsVersion

	s.quotations[q.ID] = q
	s.quotationsByInq[inquiry.ID][q.SupplierID] = q

	return nil
}

func (s *Store) GetQuotationsByInquiry(inquiryID string) ([]*Quotation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inqMap, ok := s.quotationsByInq[inquiryID]
	if !ok {
		return nil, ErrInquiryNotFound
	}

	list := make([]*Quotation, 0, len(inqMap))
	for _, q := range inqMap {
		list = append(list, q)
	}
	return list, nil
}

func (s *Store) GetQuotationBySupplier(inquiryID, supplierID string) (*Quotation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inqMap, ok := s.quotationsByInq[inquiryID]
	if !ok {
		return nil, ErrInquiryNotFound
	}

	q, ok := inqMap[supplierID]
	if !ok {
		return nil, ErrQuotationNotFound
	}
	return q, nil
}

func (s *Store) GenerateComparison(inquiry *Inquiry) (*ComparisonResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	inqMap, ok := s.quotationsByInq[inquiry.ID]
	if !ok {
		return nil, ErrInquiryNotFound
	}

	result := &ComparisonResult{
		InquiryID:   inquiry.ID,
		GeneratedAt: time.Now(),
	}

	for _, material := range inquiry.Materials {
		item := ComparisonItem{
			MaterialID:   material.ID,
			MaterialName: material.Name,
			Spec:         material.Spec,
			Quantity:     material.Quantity,
			Unit:         material.Unit,
			SupplierPrices: make(map[string]float64),
			HasNoPrice:   true,
		}

		for supplierID, q := range inqMap {
			if q.Status != QuotationStatusValid {
				continue
			}
			for _, qi := range q.Items {
				if qi.MaterialID == material.ID {
					item.SupplierPrices[supplierID] = qi.UnitPrice
					item.HasNoPrice = false
				}
			}
		}

		result.Items = append(result.Items, item)
	}

	return result, nil
}

func (s *Store) CloseInquiry(inquiryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	inq, ok := s.inquiries[inquiryID]
	if !ok {
		return ErrInquiryNotFound
	}
	inq.Status = InquiryStatusClosed
	inq.UpdatedAt = time.Now()
	return nil
}

func (s *Store) CreatePurchaseOrders(inquiry *Inquiry, awards []AwardedMaterial) ([]*PurchaseOrder, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	inq, ok := s.inquiries[inquiry.ID]
	if !ok {
		return nil, ErrInquiryNotFound
	}

	if inq.Status != InquiryStatusClosed {
		return nil, errors.New("inquiry must be closed first")
	}

	materialMap := make(map[string]Material)
	for _, m := range inq.Materials {
		materialMap[m.ID] = m
	}

	inqMap := s.quotationsByInq[inquiry.ID]

	itemsBySupplier := make(map[string][]PurchaseOrderItem)
	for _, award := range awards {
		material, ok := materialMap[award.MaterialID]
		if !ok {
			return nil, ErrInvalidMaterials
		}

		q, ok := inqMap[award.SupplierID]
		if !ok || q.Status != QuotationStatusValid {
			return nil, ErrQuotationNotFound
		}

		var unitPrice float64
		var deliveryDays int
		for _, qi := range q.Items {
			if qi.MaterialID == award.MaterialID {
				unitPrice = qi.UnitPrice
				deliveryDays = qi.DeliveryDays
				break
			}
		}

		item := PurchaseOrderItem{
			MaterialID:   material.ID,
			Name:         material.Name,
			Spec:         material.Spec,
			Quantity:     material.Quantity,
			Unit:         material.Unit,
			UnitPrice:    unitPrice,
			Amount:       roundToTwoDecimals(unitPrice * material.Quantity),
			DeliveryDays: deliveryDays,
		}

		itemsBySupplier[award.SupplierID] = append(itemsBySupplier[award.SupplierID], item)
	}

	var orders []*PurchaseOrder
	now := time.Now()
	for supplierID, items := range itemsBySupplier {
		var totalAmount float64
		for _, item := range items {
			totalAmount += item.Amount
		}
		totalAmount = roundToTwoDecimals(totalAmount)

		order := &PurchaseOrder{
			ID:         generateID("PO"),
			InquiryID:  inquiry.ID,
			SupplierID: supplierID,
			BuyerID:    inq.BuyerID,
			Items:      items,
			TotalAmount: totalAmount,
			CreatedAt:  now,
		}

		s.orders[order.ID] = order
		s.ordersByInquiry[inquiry.ID] = append(s.ordersByInquiry[inquiry.ID], order)
		s.ordersBySupplier[supplierID] = append(s.ordersBySupplier[supplierID], order)

		orders = append(orders, order)
	}

	return orders, nil
}

func (s *Store) GetPurchaseOrdersByInquiry(inquiryID string) []*PurchaseOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ordersByInquiry[inquiryID]
}

func (s *Store) GetPurchaseOrdersBySupplier(supplierID string) []*PurchaseOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ordersBySupplier[supplierID]
}

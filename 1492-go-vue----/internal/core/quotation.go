package core

import (
	"errors"
	"math"
	"time"

	"renovation-management/internal/api"
)

func calculateWorkItemTotal(quantity float64, unitPriceFen int64) int64 {
	if quantity < 0 {
		quantity = 0
	}
	amount := quantity * float64(unitPriceFen)
	return int64(math.Round(amount))
}

func (s *Store) CreateQuotation(req *api.CreateQuotationRequest) (*api.Quotation, error) {
	if req.SchemeID == "" {
		return nil, errors.New("scheme_id is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	scheme, ok := s.schemes[req.SchemeID]
	if !ok {
		return nil, errors.New("design scheme not found")
	}
	if !scheme.Confirmed {
		return nil, errors.New("design scheme must be confirmed before creating quotation")
	}

	if _, exists := s.quotationByScheme[req.SchemeID]; exists {
		return nil, errors.New("quotation already exists for this scheme")
	}

	workItems := make([]api.WorkItem, 0, len(req.WorkItems))
	var workTotalFen int64 = 0

	for _, itemInput := range req.WorkItems {
		if itemInput.Quantity < 0 {
			return nil, errors.New("work item quantity cannot be negative")
		}
		if itemInput.UnitPriceFen < 0 {
			return nil, errors.New("unit price cannot be negative")
		}

		item := api.WorkItem{
			ID:            GenerateID("WORK"),
			Category:      itemInput.Category,
			SubCategory:   itemInput.SubCategory,
			Unit:          itemInput.Unit,
			Quantity:      itemInput.Quantity,
			UnitPriceFen:  itemInput.UnitPriceFen,
			ActualQuantity: itemInput.Quantity,
			NeedsReconfirm: false,
		}

		workItems = append(workItems, item)
		s.workItemsByID[item.ID] = &workItems[len(workItems)-1]
		workTotalFen += calculateWorkItemTotal(item.Quantity, item.UnitPriceFen)
	}

	materials := make([]api.Material, 0, len(req.Materials))
	var materialTotalFen int64 = 0

	for _, matInput := range req.Materials {
		if matInput.Quantity < 0 {
			return nil, errors.New("material quantity cannot be negative")
		}
		if matInput.UnitPriceFen < 0 {
			return nil, errors.New("material unit price cannot be negative")
		}

		material := api.Material{
			ID:           GenerateID("MAT"),
			Name:         matInput.Name,
			Spec:         matInput.Spec,
			Brand:        matInput.Brand,
			UnitPriceFen: matInput.UnitPriceFen,
			Quantity:     matInput.Quantity,
		}

		materials = append(materials, material)
		materialTotalFen += calculateWorkItemTotal(material.Quantity, material.UnitPriceFen)
	}

	quotation := &api.Quotation{
		ID:               GenerateID("QUOT"),
		SchemeID:         req.SchemeID,
		WorkItems:        workItems,
		Materials:        materials,
		WorkTotalFen:     workTotalFen,
		MaterialTotalFen: materialTotalFen,
		TotalFen:         workTotalFen + materialTotalFen,
		CreatedAt:        time.Now(),
	}

	s.quotations[quotation.ID] = quotation
	s.quotationByScheme[req.SchemeID] = quotation

	return quotation, nil
}

func (s *Store) GetQuotation(id string) (*api.Quotation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quotation, ok := s.quotations[id]
	if !ok {
		return nil, errors.New("quotation not found")
	}
	return quotation, nil
}

func (s *Store) ListQuotations() []*api.Quotation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	quotations := make([]*api.Quotation, 0, len(s.quotations))
	for _, q := range s.quotations {
		quotations = append(quotations, q)
	}
	return quotations
}

func (s *Store) UpdateActualQuantity(req *api.UpdateActualQuantityRequest) (*api.WorkItem, error) {
	if req.ActualQuantity < 0 {
		return nil, errors.New("actual quantity cannot be negative")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.workItemsByID[req.WorkItemID]
	if !ok {
		return nil, errors.New("work item not found")
	}

	item.ActualQuantity = req.ActualQuantity

	if item.Quantity > 0 {
		diffRatio := math.Abs(req.ActualQuantity-item.Quantity) / item.Quantity
		item.NeedsReconfirm = diffRatio > 0.1
	} else {
		item.NeedsReconfirm = req.ActualQuantity > 0
	}

	return item, nil
}

func (s *Store) GetWorkItem(id string) (*api.WorkItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.workItemsByID[id]
	if !ok {
		return nil, errors.New("work item not found")
	}
	return item, nil
}

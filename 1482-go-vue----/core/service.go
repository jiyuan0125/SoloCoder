package core

import (
	"autorepair/common"
	"errors"
	"math"
	"sync"
	"time"
)

type Service struct {
	mu                  sync.Mutex
	orders              map[int64]*common.Order
	parts               map[int64]*common.Part
	partCodeToID        map[string]int64
	usedParts           map[int64]*common.UsedPart
	replenishments      map[int64]*common.ReplenishmentTodo
	nextOrderID         int64
	nextItemID          int64
	nextPartID          int64
	nextUsedPartID      int64
	nextReplenishmentID int64
}

func NewService() *Service {
	return &Service{
		orders:              make(map[int64]*common.Order),
		parts:               make(map[int64]*common.Part),
		partCodeToID:        make(map[string]int64),
		usedParts:           make(map[int64]*common.UsedPart),
		replenishments:      make(map[int64]*common.ReplenishmentTodo),
		nextOrderID:         1,
		nextItemID:          1,
		nextPartID:          1,
		nextUsedPartID:      1,
		nextReplenishmentID: 1,
	}
}

func (s *Service) CreateOrder(req common.CreateOrderRequest) (*common.Order, error) {
	if req.PlateNumber == "" || req.CustomerName == "" || req.Phone == "" {
		return nil, errors.New("必填字段不能为空")
	}

	valid := false
	for _, ft := range common.ValidFaultTypes() {
		if req.FaultType == ft {
			valid = true
			break
		}
	}
	if !valid {
		return nil, errors.New("无效的故障类型")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	order := &common.Order{
		ID:           s.nextOrderID,
		PlateNumber:  req.PlateNumber,
		CustomerName: req.CustomerName,
		Phone:        req.Phone,
		Description:  req.Description,
		FaultType:    req.FaultType,
		Status:       common.OrderStatusCreated,
		Items:        []common.RepairItem{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.orders[order.ID] = order
	s.nextOrderID++

	return order, nil
}

func (s *Service) AssignOrder(orderID int64) (*common.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status != common.OrderStatusCreated {
		return nil, ErrOrderInvalidStatus
	}

	order.Status = common.OrderStatusAssigned
	order.UpdatedAt = time.Now()

	return order, nil
}

func (s *Service) CreateRepairItem(req common.CreateRepairItemRequest) (*common.RepairItem, error) {
	if req.Name == "" {
		return nil, errors.New("项目名称不能为空")
	}
	if req.EstimatedHours <= 0 {
		return nil, ErrInvalidHours
	}

	valid := false
	for _, tl := range common.ValidTechnicianLevels() {
		if req.TechnicianLevel == tl {
			valid = true
			break
		}
	}
	if !valid {
		return nil, errors.New("无效的技师等级")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[req.OrderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status != common.OrderStatusAssigned && order.Status != common.OrderStatusInProgress {
		return nil, ErrOrderInvalidStatus
	}

	now := time.Now()
	item := &common.RepairItem{
		ID:              s.nextItemID,
		OrderID:         req.OrderID,
		Name:            req.Name,
		TechnicianLevel: req.TechnicianLevel,
		EstimatedHours:  req.EstimatedHours,
		Parts:           []common.UsedPart{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.nextItemID++

	order.Items = append(order.Items, *item)
	if order.Status == common.OrderStatusAssigned {
		order.Status = common.OrderStatusInProgress
	}
	order.UpdatedAt = now

	return item, nil
}

func (s *Service) CompleteRepairItem(req common.CompleteRepairItemRequest) (*common.RepairItem, error) {
	if req.ActualHours <= 0 {
		return nil, ErrInvalidHours
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	var order *common.Order
	var itemIdx int
	var found bool

	for _, o := range s.orders {
		for i, item := range o.Items {
			if item.ID == req.ItemID {
				order = o
				itemIdx = i
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return nil, ErrItemNotFound
	}

	if order.Status == common.OrderStatusCompleted || order.Status == common.OrderStatusCancelled {
		return nil, ErrOrderInvalidStatus
	}

	item := &order.Items[itemIdx]
	if item.Completed {
		return nil, errors.New("项目已完成")
	}

	isAbnormal := req.ActualHours > item.EstimatedHours*1.5
	if isAbnormal && req.OvertimeReason == "" {
		return nil, ErrOvertimeReasonRequired
	}

	rate := item.TechnicianLevel.HourlyRate()
	laborCost := calculateLaborCost(req.ActualHours, rate)

	item.ActualHours = req.ActualHours
	item.LaborCost = laborCost
	item.IsAbnormalHours = isAbnormal
	item.OvertimeReason = req.OvertimeReason
	item.Completed = true
	item.UpdatedAt = time.Now()

	recalculateOrderTotals(order)
	order.UpdatedAt = time.Now()

	return item, nil
}

func (s *Service) AddPartToItem(req common.AddPartToItemRequest) (*common.UsedPart, error) {
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	part, exists := s.parts[req.PartID]
	if !exists {
		return nil, ErrPartNotFound
	}

	if part.StockQty < req.Quantity {
		return nil, ErrInsufficientStock
	}

	var order *common.Order
	var itemIdx int
	var found bool

	for _, o := range s.orders {
		for i, item := range o.Items {
			if item.ID == req.ItemID {
				if item.Completed {
					return nil, errors.New("项目已完成，不能添加配件")
				}
				order = o
				itemIdx = i
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return nil, ErrItemNotFound
	}

	if order.Status == common.OrderStatusCompleted || order.Status == common.OrderStatusCancelled {
		return nil, ErrOrderInvalidStatus
	}

	part.StockQty -= req.Quantity
	part.UpdatedAt = time.Now()

	usedPart := &common.UsedPart{
		ID:        s.nextUsedPartID,
		PartID:    part.ID,
		PartCode:  part.Code,
		PartName:  part.Name,
		Spec:      part.Spec,
		UnitPrice: part.UnitPrice,
		Quantity:  req.Quantity,
		Returned:  false,
	}
	s.usedParts[usedPart.ID] = usedPart
	s.nextUsedPartID++

	order.Items[itemIdx].Parts = append(order.Items[itemIdx].Parts, *usedPart)
	order.UpdatedAt = time.Now()

	if part.StockQty <= part.WarningLevel {
		s.createReplenishmentLocked(part)
	}

	recalculateOrderTotals(order)

	return usedPart, nil
}

func (s *Service) ReturnPart(req common.ReturnPartRequest) error {
	if req.Quantity <= 0 {
		return ErrInvalidQuantity
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	usedPart, exists := s.usedParts[req.UsedPartID]
	if !exists {
		return ErrUsedPartNotFound
	}
	if usedPart.Returned {
		return ErrUsedPartAlreadyReturned
	}

	part, exists := s.parts[usedPart.PartID]
	if !exists {
		return ErrPartNotFound
	}

	if req.Quantity > usedPart.Quantity {
		return ErrInvalidQuantity
	}

	var order *common.Order
	var itemIdx int
	var partIdx int
	var found bool

	for _, o := range s.orders {
		if o.Status == common.OrderStatusCompleted || o.Status == common.OrderStatusCancelled {
			continue
		}
		for i, item := range o.Items {
			if item.Completed {
				continue
			}
			for j, up := range item.Parts {
				if up.ID == req.UsedPartID {
					order = o
					itemIdx = i
					partIdx = j
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return errors.New("配件已不能退回")
	}

	part.StockQty += req.Quantity
	part.UpdatedAt = time.Now()

	if req.Quantity == usedPart.Quantity {
		order.Items[itemIdx].Parts[partIdx].Returned = true
		usedPart.Returned = true
	} else {
		order.Items[itemIdx].Parts[partIdx].Quantity -= req.Quantity
		usedPart.Quantity -= req.Quantity
	}

	order.UpdatedAt = time.Now()
	recalculateOrderTotals(order)

	return nil
}

func (s *Service) CompleteOrder(orderID int64) (*common.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status == common.OrderStatusCompleted {
		return nil, ErrOrderAlreadyCompleted
	}
	if order.Status == common.OrderStatusCancelled {
		return nil, ErrOrderAlreadyCancelled
	}

	for _, item := range order.Items {
		if !item.Completed {
			return nil, errors.New("存在未完成的维修项目")
		}
	}

	order.Status = common.OrderStatusCompleted
	order.UpdatedAt = time.Now()

	return order, nil
}

func (s *Service) CancelOrder(orderID int64) (*common.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}

	if order.Status == common.OrderStatusCompleted {
		return nil, ErrOrderAlreadyCompleted
	}
	if order.Status == common.OrderStatusCancelled {
		return nil, ErrOrderAlreadyCancelled
	}

	for _, item := range order.Items {
		for _, usedPart := range item.Parts {
			if usedPart.Returned {
				continue
			}
			part, exists := s.parts[usedPart.PartID]
			if exists {
				part.StockQty += usedPart.Quantity
				part.UpdatedAt = time.Now()
			}

			up, exists := s.usedParts[usedPart.ID]
			if exists {
				up.Returned = true
			}
		}
	}

	order.Status = common.OrderStatusCancelled
	order.UpdatedAt = time.Now()

	return order, nil
}

func (s *Service) GetOrders() []common.Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders := make([]common.Order, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, *order)
	}
	return orders
}

func (s *Service) GetOrder(orderID int64) (*common.Order, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, exists := s.orders[orderID]
	if !exists {
		return nil, ErrOrderNotFound
	}
	copy := *order
	return &copy, nil
}

func (s *Service) CreatePart(req common.CreatePartRequest) (*common.Part, error) {
	if req.Code == "" || req.Name == "" {
		return nil, errors.New("必填字段不能为空")
	}
	if req.UnitPrice < 0 {
		return nil, ErrInvalidPrice
	}
	if req.StockQty < 0 {
		return nil, ErrInvalidQuantity
	}
	if req.WarningLevel < 0 {
		req.WarningLevel = 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.partCodeToID[req.Code]; exists {
		return nil, ErrPartCodeDuplicate
	}

	now := time.Now()
	part := &common.Part{
		ID:           s.nextPartID,
		Code:         req.Code,
		Name:         req.Name,
		Spec:         req.Spec,
		UnitPrice:    req.UnitPrice,
		StockQty:     req.StockQty,
		WarningLevel: req.WarningLevel,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.parts[part.ID] = part
	s.partCodeToID[part.Code] = part.ID
	s.nextPartID++

	if part.StockQty <= part.WarningLevel {
		s.createReplenishmentLocked(part)
	}

	return part, nil
}

func (s *Service) UpdatePartPrice(partID int64, newPrice int64) (*common.Part, error) {
	if newPrice < 0 {
		return nil, ErrInvalidPrice
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	part, exists := s.parts[partID]
	if !exists {
		return nil, ErrPartNotFound
	}

	part.UnitPrice = newPrice
	part.UpdatedAt = time.Now()

	return part, nil
}

func (s *Service) UpdatePartStock(partID int64, newStock int64) (*common.Part, error) {
	if newStock < 0 {
		return nil, ErrInvalidQuantity
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	part, exists := s.parts[partID]
	if !exists {
		return nil, ErrPartNotFound
	}

	part.StockQty = newStock
	part.UpdatedAt = time.Now()

	if part.StockQty <= part.WarningLevel {
		s.createReplenishmentLocked(part)
	}

	return part, nil
}

func (s *Service) GetParts() []common.Part {
	s.mu.Lock()
	defer s.mu.Unlock()

	parts := make([]common.Part, 0, len(s.parts))
	for _, part := range s.parts {
		parts = append(parts, *part)
	}
	return parts
}

func (s *Service) GetPart(partID int64) (*common.Part, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	part, exists := s.parts[partID]
	if !exists {
		return nil, ErrPartNotFound
	}
	copy := *part
	return &copy, nil
}

func (s *Service) GetReplenishments() []common.ReplenishmentTodo {
	s.mu.Lock()
	defer s.mu.Unlock()

	todos := make([]common.ReplenishmentTodo, 0, len(s.replenishments))
	for _, todo := range s.replenishments {
		todos = append(todos, *todo)
	}
	return todos
}

func (s *Service) ResolveReplenishment(todoID int64) (*common.ReplenishmentTodo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	todo, exists := s.replenishments[todoID]
	if !exists {
		return nil, ErrTodoNotFound
	}

	if todo.Resolved {
		return nil, ErrTodoAlreadyResolved
	}

	todo.Resolved = true
	return todo, nil
}

func (s *Service) createReplenishmentLocked(part *common.Part) {
	for _, existing := range s.replenishments {
		if existing.PartID == part.ID && !existing.Resolved {
			existing.CurrentQty = part.StockQty
			return
		}
	}

	todo := &common.ReplenishmentTodo{
		ID:           s.nextReplenishmentID,
		PartID:       part.ID,
		PartCode:     part.Code,
		PartName:     part.Name,
		CurrentQty:   part.StockQty,
		WarningLevel: part.WarningLevel,
		CreatedAt:    time.Now(),
		Resolved:     false,
	}
	s.replenishments[todo.ID] = todo
	s.nextReplenishmentID++
}

func calculateLaborCost(hours float64, hourlyRate float64) int64 {
	total := hours * hourlyRate
	totalInJiao := math.Floor(total * 10)
	return int64(totalInJiao * 10)
}

func recalculateOrderTotals(order *common.Order) {
	var laborTotal int64
	var partsTotal int64
	hasAbnormal := false

	for _, item := range order.Items {
		laborTotal += item.LaborCost
		for _, up := range item.Parts {
			if !up.Returned {
				partsTotal += up.UnitPrice * up.Quantity
			}
		}
		if item.IsAbnormalHours {
			hasAbnormal = true
		}
	}

	order.LaborTotal = laborTotal
	order.PartsTotal = partsTotal
	order.GrandTotal = laborTotal + partsTotal
	order.HasAbnormal = hasAbnormal
}

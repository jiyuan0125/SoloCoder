package core

import (
	"fmt"
	"sync"
	"time"

	"repair-platform/common"
)

type Storage struct {
	mu sync.RWMutex

	repairRequests map[string]*common.RepairRequest
	spareParts    map[string]*common.SparePart
	sparePartMu   map[string]*sync.Mutex
	technicians   map[string]*common.Technician
	purchaseOrders map[string]*common.PurchaseOrder
	repairReports map[string]*common.RepairReport
	outboundRecords map[string][]*common.SparePartOutboundRecord
	waitQueue     []string

	nextReqID     int
	nextOrderID   int
}

func NewStorage() *Storage {
	s := &Storage{
		repairRequests: make(map[string]*common.RepairRequest),
		spareParts:     make(map[string]*common.SparePart),
		sparePartMu:    make(map[string]*sync.Mutex),
		technicians:    make(map[string]*common.Technician),
		purchaseOrders: make(map[string]*common.PurchaseOrder),
		repairReports:  make(map[string]*common.RepairReport),
		outboundRecords: make(map[string][]*common.SparePartOutboundRecord),
		waitQueue:      []string{},
	}
	return s
}

func (s *Storage) CreateRepairRequest(req common.RepairRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.repairRequests[req.ID]; exists {
		return fmt.Errorf("维修请求已存在: %s", req.ID)
	}
	copyReq := req
	s.repairRequests[req.ID] = &copyReq
	return nil
}

func (s *Storage) GetRepairRequest(id string) (*common.RepairRequest, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.repairRequests[id]
	if !ok {
		return nil, false
	}
	reqCopy := *req
	return &reqCopy, true
}

func (s *Storage) UpdateRepairRequest(id string, updateFn func(*common.RepairRequest)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	req, ok := s.repairRequests[id]
	if !ok {
		return
	}
	updateFn(req)
}

func (s *Storage) ListRepairRequests() []common.RepairRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.RepairRequest, 0, len(s.repairRequests))
	for _, req := range s.repairRequests {
		result = append(result, *req)
	}
	return result
}

func (s *Storage) AddSparePart(part common.SparePart) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.spareParts[part.Code]; exists {
		return fmt.Errorf("备件已存在: %s", part.Code)
	}
	copyPart := part
	s.spareParts[part.Code] = &copyPart
	s.sparePartMu[part.Code] = &sync.Mutex{}
	return nil
}

func (s *Storage) GetSparePart(code string) (*common.SparePart, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	part, ok := s.spareParts[code]
	if !ok {
		return nil, false
	}
	copyPart := *part
	return &copyPart, true
}

func (s *Storage) ListSpareParts() []common.SparePart {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.SparePart, 0, len(s.spareParts))
	for _, p := range s.spareParts {
		result = append(result, *p)
	}
	return result
}

func (s *Storage) UpdateSparePrice(code string, newPrice float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	part, ok := s.spareParts[code]
	if !ok {
		return fmt.Errorf("备件不存在: %s", code)
	}
	part.UnitPrice = newPrice
	return nil
}

func (s *Storage) LockSparePart(code string) func() {
	s.mu.RLock()
	mu, ok := s.sparePartMu[code]
	s.mu.RUnlock()

	if !ok {
		return func() {}
	}

	mu.Lock()
	return func() { mu.Unlock() }
}

func (s *Storage) DeductSpareStock(code string, qty int) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	part, ok := s.spareParts[code]
	if !ok {
		return 0, fmt.Errorf("备件不存在: %s", code)
	}

	if part.StockQuantity < qty {
		return 0, fmt.Errorf("库存不足: %s 剩余 %d", code, part.StockQuantity)
	}

	price := part.UnitPrice
	part.StockQuantity -= qty
	return price, nil
}

func (s *Storage) ReturnSpareStock(code string, qty int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	part, ok := s.spareParts[code]
	if !ok {
		return fmt.Errorf("备件不存在: %s", code)
	}
	part.StockQuantity += qty
	return nil
}

func (s *Storage) AddTechnician(tech common.Technician) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.technicians[tech.ID]; exists {
		return fmt.Errorf("师傅已存在: %s", tech.ID)
	}
	copyTech := tech
	s.technicians[tech.ID] = &copyTech
	return nil
}

func (s *Storage) GetTechnician(id string) (*common.Technician, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tech, ok := s.technicians[id]
	if !ok {
		return nil, false
	}
	copyTech := *tech
	return &copyTech, true
}

func (s *Storage) ListTechnicians() []common.Technician {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.Technician, 0, len(s.technicians))
	for _, t := range s.technicians {
		result = append(result, *t)
	}
	return result
}

func (s *Storage) UpdateTechnician(id string, updateFn func(*common.Technician)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	tech, ok := s.technicians[id]
	if !ok {
		return
	}
	updateFn(tech)
}

func (s *Storage) AddPurchaseOrder(po common.PurchaseOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.purchaseOrders[po.ID] = &po
}

func (s *Storage) ListPurchaseOrders() []common.PurchaseOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]common.PurchaseOrder, 0, len(s.purchaseOrders))
	for _, o := range s.purchaseOrders {
		result = append(result, *o)
	}
	return result
}

func (s *Storage) AddRepairReport(report common.RepairReport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.repairReports[report.RequestID] = &report
}

func (s *Storage) GetRepairReport(requestID string) (*common.RepairReport, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	report, ok := s.repairReports[requestID]
	if !ok {
		return nil, false
	}
	copyReport := *report
	return &copyReport, true
}

func (s *Storage) AddOutboundRecord(reqID string, record *common.SparePartOutboundRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.outboundRecords[reqID] = append(s.outboundRecords[reqID], record)
}

func (s *Storage) GetOutboundRecords(reqID string) []*common.SparePartOutboundRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records, ok := s.outboundRecords[reqID]
	if !ok {
		return nil
	}
	return records
}

func (s *Storage) PushWaitQueue(reqID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.waitQueue = append(s.waitQueue, reqID)
}

func (s *Storage) PopWaitQueue() (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.waitQueue) == 0 {
		return "", false
	}
	id := s.waitQueue[0]
	s.waitQueue = s.waitQueue[1:]
	return id, true
}

func (s *Storage) NextReqID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextReqID++
	return fmt.Sprintf("REQ%05d", s.nextReqID)
}

func (s *Storage) NextOrderID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.nextOrderID++
	return fmt.Sprintf("PO%05d", s.nextOrderID)
}

func isSameDay(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

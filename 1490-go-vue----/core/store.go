package core

import (
	"piperepair/api"
	"sync"
	"time"
)

type Store struct {
	mu             sync.RWMutex
	orders         map[string]*api.RepairOrder
	masters        map[string]*api.Master
	sites          map[string]*api.RepairSite
	repairRecords  map[string]*api.RepairRecord
	acceptanceRecords map[string]*api.AcceptanceRecord
	areaToSiteMap  map[string]string
	masterOrderMap map[string]string
}

func NewStore() *Store {
	return &Store{
		orders:           make(map[string]*api.RepairOrder),
		masters:          make(map[string]*api.Master),
		sites:            make(map[string]*api.RepairSite),
		repairRecords:    make(map[string]*api.RepairRecord),
		acceptanceRecords: make(map[string]*api.AcceptanceRecord),
		areaToSiteMap:    make(map[string]string),
		masterOrderMap:   make(map[string]string),
	}
}

func (s *Store) InitSampleData() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sites["site001"] = &api.RepairSite{
		ID:           "site001",
		Name:         "朝阳区维修站点",
		AreaCodes:    []string{"01001", "01002", "01003"},
		NeighborSiteIDs: []string{"site002", "site003"},
		MasterIDs:    []string{"m001", "m002"},
	}

	s.sites["site002"] = &api.RepairSite{
		ID:           "site002",
		Name:         "海淀区维修站点",
		AreaCodes:    []string{"01004", "01005", "01006"},
		NeighborSiteIDs: []string{"site001", "site003"},
		MasterIDs:    []string{"m003"},
	}

	s.sites["site003"] = &api.RepairSite{
		ID:           "site003",
		Name:         "丰台区维修站点",
		AreaCodes:    []string{"01007", "01008", "01009"},
		NeighborSiteIDs: []string{"site001", "site002"},
		MasterIDs:    []string{"m004", "m005"},
	}

	for _, site := range s.sites {
		for _, areaCode := range site.AreaCodes {
			s.areaToSiteMap[areaCode] = site.ID
		}
	}

	s.masters["m001"] = &api.Master{
		ID:          "m001",
		Name:        "张师傅",
		SiteID:      "site001",
		Skills:      []api.SkillTag{api.SkillGeneral, api.SkillHighPressure},
		Status:      api.MasterStatusIdle,
		DailyOrders: 0,
	}

	s.masters["m002"] = &api.Master{
		ID:          "m002",
		Name:        "李师傅",
		SiteID:      "site001",
		Skills:      []api.SkillTag{api.SkillGeneral, api.SkillInspection},
		Status:      api.MasterStatusIdle,
		DailyOrders: 0,
	}

	s.masters["m003"] = &api.Master{
		ID:          "m003",
		Name:        "王师傅",
		SiteID:      "site002",
		Skills:      []api.SkillTag{api.SkillGeneral, api.SkillHighPressure, api.SkillInspection},
		Status:      api.MasterStatusIdle,
		DailyOrders: 0,
	}

	s.masters["m004"] = &api.Master{
		ID:          "m004",
		Name:        "赵师傅",
		SiteID:      "site003",
		Skills:      []api.SkillTag{api.SkillGeneral, api.SkillReplacement},
		Status:      api.MasterStatusIdle,
		DailyOrders: 0,
	}

	s.masters["m005"] = &api.Master{
		ID:          "m005",
		Name:        "刘师傅",
		SiteID:      "site003",
		Skills:      []api.SkillTag{api.SkillHighPressure, api.SkillInspection},
		Status:      api.MasterStatusIdle,
		DailyOrders: 0,
	}
}

func (s *Store) SaveOrder(order *api.RepairOrder) {
	s.mu.Lock()
	defer s.mu.Unlock()
	order.UpdatedAt = time.Now()
	s.orders[order.ID] = order
}

func (s *Store) GetOrder(orderID string) (*api.RepairOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[orderID]
	return order, ok
}

func (s *Store) GetMaster(masterID string) (*api.Master, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	master, ok := s.masters[masterID]
	return master, ok
}

func (s *Store) GetSite(siteID string) (*api.RepairSite, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site, ok := s.sites[siteID]
	return site, ok
}

func (s *Store) GetSiteByAreaCode(areaCode string) (*api.RepairSite, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	siteID, ok := s.areaToSiteMap[areaCode]
	if !ok {
		return nil, false
	}
	site, ok := s.sites[siteID]
	return site, ok
}

func (s *Store) SaveMaster(master *api.Master) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.masters[master.ID] = master
}

func (s *Store) SaveRepairRecord(record *api.RepairRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.repairRecords[record.OrderID] = record
}

func (s *Store) GetRepairRecord(orderID string) (*api.RepairRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.repairRecords[orderID]
	return record, ok
}

func (s *Store) SaveAcceptanceRecord(record *api.AcceptanceRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.acceptanceRecords[record.OrderID] = record
}

func (s *Store) GetAcceptanceRecord(orderID string) (*api.AcceptanceRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	record, ok := s.acceptanceRecords[orderID]
	return record, ok
}

func (s *Store) SetMasterCurrentOrder(masterID string, orderID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if orderID == "" {
		delete(s.masterOrderMap, masterID)
	} else {
		s.masterOrderMap[masterID] = orderID
	}
}

func (s *Store) GetMasterCurrentOrder(masterID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	orderID, ok := s.masterOrderMap[masterID]
	return orderID, ok
}

func (s *Store) ListAllOrders() []*api.RepairOrder {
	s.mu.RLock()
	defer s.mu.RUnlock()
	orders := make([]*api.RepairOrder, 0, len(s.orders))
	for _, order := range s.orders {
		orders = append(orders, order)
	}
	return orders
}

package core

import (
	"sync"

	"renovation-management/internal/api"
)

type Store struct {
	mu sync.RWMutex

	houses         map[string]*api.HouseInfo
	schemes        map[string]*api.DesignScheme
	schemesByHouse map[string][]*api.DesignScheme

	quotations      map[string]*api.Quotation
	quotationByScheme map[string]*api.Quotation
	workItemsByID   map[string]*api.WorkItem

	changeOrders     map[string]*api.ChangeOrder
	changeOrdersByQuotation map[string][]*api.ChangeOrder

	phases           map[string]*api.ConstructionPhase
	phasesByQuotation map[string][]*api.ConstructionPhase

	settlements map[string]*api.Settlement
	settlementByQuotation map[string]*api.Settlement
}

func NewStore() *Store {
	return &Store{
		houses:         make(map[string]*api.HouseInfo),
		schemes:        make(map[string]*api.DesignScheme),
		schemesByHouse: make(map[string][]*api.DesignScheme),

		quotations:        make(map[string]*api.Quotation),
		quotationByScheme: make(map[string]*api.Quotation),
		workItemsByID:     make(map[string]*api.WorkItem),

		changeOrders:           make(map[string]*api.ChangeOrder),
		changeOrdersByQuotation: make(map[string][]*api.ChangeOrder),

		phases:                make(map[string]*api.ConstructionPhase),
		phasesByQuotation:     make(map[string][]*api.ConstructionPhase),

		settlements:            make(map[string]*api.Settlement),
		settlementByQuotation:  make(map[string]*api.Settlement),
	}
}

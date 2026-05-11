package lockservice

import "sync"

var pricingMap = map[LockType]Pricing{
	LockTypeSecurity:    {BaseFee: 150, InstallationFee: 50, UrgentRate: 1.5},
	LockTypeInterior:    {BaseFee: 80, InstallationFee: 50, UrgentRate: 1.5},
	LockTypePassword:    {BaseFee: 200, InstallationFee: 50, UrgentRate: 1.5},
	LockTypeFingerprint: {BaseFee: 250, InstallationFee: 50, UrgentRate: 1.5},
	LockTypeCar:         {BaseFee: 300, InstallationFee: 50, UrgentRate: 1.5},
}

type PaymentManager struct {
	mu         sync.RWMutex
	paidOrders map[string]bool
	complaints []string
}

func NewPaymentManager() *PaymentManager {
	return &PaymentManager{
		paidOrders: make(map[string]bool),
		complaints: []string{},
	}
}

func (pm *PaymentManager) CalculateCost(order *Order) {
	pricing := pricingMap[order.LockType]
	order.BaseCost = pricing.BaseFee
	order.UrgentFee = 0
	if order.Urgency == UrgencyUrgent {
		order.UrgentFee = order.BaseCost * (pricing.UrgentRate - 1)
	}
	order.TotalCost = order.BaseCost + order.UrgentFee
}

func (pm *PaymentManager) CalculateFinalCost(order *Order) {
	pricing := pricingMap[order.LockType]
	base := order.BaseCost
	urgent := order.UrgentFee
	additional := 0.0
	if order.Detail.NeedReplaceLock {
		additional = order.Detail.PartsCost + pricing.InstallationFee
		order.Detail.LaborCost = pricing.InstallationFee
	}
	order.TotalCost = base + urgent + additional
}

func (pm *PaymentManager) PayOrder(orderID string) bool {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.paidOrders[orderID] = true
	return true
}

func (pm *PaymentManager) IsPaid(orderID string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.paidOrders[orderID]
}

func (pm *PaymentManager) AddComplaint(orderID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.complaints = append(pm.complaints, orderID)
}

func (pm *PaymentManager) GetComplaints() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	result := make([]string, len(pm.complaints))
	copy(result, pm.complaints)
	return result
}

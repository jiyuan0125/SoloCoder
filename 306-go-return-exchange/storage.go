package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

const (
	StorageFile = "data.json"
)

type FileStorage struct {
	mu              sync.RWMutex
	Applications    map[string]*ReturnExchangeApplication `json:"applications"`
	Vouchers        map[string]*Voucher                   `json:"vouchers"`
	RefundRecords   map[string]*RefundRecord              `json:"refund_records"`
	ShippingOrders  map[string]*ShippingOrder             `json:"shipping_orders"`
	Orders          map[string]*Order                     `json:"orders"`
	SKUPrices       map[string]*SKUPrice                  `json:"sku_prices"`
	applicationSeq  int
	voucherSeq      int
	refundSeq       int
	shippingSeq     int
}

func NewFileStorage() *FileStorage {
	return &FileStorage{
		Applications:   make(map[string]*ReturnExchangeApplication),
		Vouchers:       make(map[string]*Voucher),
		RefundRecords:  make(map[string]*RefundRecord),
		ShippingOrders: make(map[string]*ShippingOrder),
		Orders:         make(map[string]*Order),
		SKUPrices:      make(map[string]*SKUPrice),
	}
}

func (s *FileStorage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(StorageFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var storageData struct {
		Applications   map[string]*ReturnExchangeApplication `json:"applications"`
		Vouchers       map[string]*Voucher                   `json:"vouchers"`
		RefundRecords  map[string]*RefundRecord              `json:"refund_records"`
		ShippingOrders map[string]*ShippingOrder             `json:"shipping_orders"`
		Orders         map[string]*Order                     `json:"orders"`
		SKUPrices      map[string]*SKUPrice                  `json:"sku_prices"`
	}

	if err := json.Unmarshal(data, &storageData); err != nil {
		return err
	}

	if storageData.Applications != nil {
		s.Applications = storageData.Applications
	}
	if storageData.Vouchers != nil {
		s.Vouchers = storageData.Vouchers
	}
	if storageData.RefundRecords != nil {
		s.RefundRecords = storageData.RefundRecords
	}
	if storageData.ShippingOrders != nil {
		s.ShippingOrders = storageData.ShippingOrders
	}
	if storageData.Orders != nil {
		s.Orders = storageData.Orders
	}
	if storageData.SKUPrices != nil {
		s.SKUPrices = storageData.SKUPrices
	}

	s.initializeSequences()
	return nil
}

func (s *FileStorage) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storageData := struct {
		Applications   map[string]*ReturnExchangeApplication `json:"applications"`
		Vouchers       map[string]*Voucher                   `json:"vouchers"`
		RefundRecords  map[string]*RefundRecord              `json:"refund_records"`
		ShippingOrders map[string]*ShippingOrder             `json:"shipping_orders"`
		Orders         map[string]*Order                     `json:"orders"`
		SKUPrices      map[string]*SKUPrice                  `json:"sku_prices"`
	}{
		Applications:   s.Applications,
		Vouchers:       s.Vouchers,
		RefundRecords:  s.RefundRecords,
		ShippingOrders: s.ShippingOrders,
		Orders:         s.Orders,
		SKUPrices:      s.SKUPrices,
	}

	data, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(StorageFile, data, 0644)
}

func (s *FileStorage) initializeSequences() {
	s.applicationSeq = len(s.Applications)
	s.voucherSeq = len(s.Vouchers)
	s.refundSeq = len(s.RefundRecords)
	s.shippingSeq = len(s.ShippingOrders)
}

func (s *FileStorage) GenerateApplicationID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applicationSeq++
	return fmt.Sprintf("APP%06d", s.applicationSeq)
}

func (s *FileStorage) GenerateVoucherID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.voucherSeq++
	return fmt.Sprintf("VOU%06d", s.voucherSeq)
}

func (s *FileStorage) GenerateRefundID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refundSeq++
	return fmt.Sprintf("REF%06d", s.refundSeq)
}

func (s *FileStorage) GenerateShippingID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shippingSeq++
	return fmt.Sprintf("SHP%06d", s.shippingSeq)
}

func (s *FileStorage) CreateApplication(app *ReturnExchangeApplication) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Applications[app.ApplicationID] = app
	return s.Save()
}

func (s *FileStorage) UpdateApplication(app *ReturnExchangeApplication) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	app.UpdatedAt = GetNowTime()
	s.Applications[app.ApplicationID] = app
	return s.Save()
}

func (s *FileStorage) GetApplicationByID(id string) (*ReturnExchangeApplication, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, exists := s.Applications[id]
	return app, exists
}

func (s *FileStorage) GetApplicationsByOrderID(orderID string) []*ReturnExchangeApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*ReturnExchangeApplication
	for _, app := range s.Applications {
		if app.OrderID == orderID {
			result = append(result, app)
		}
	}
	return result
}

func (s *FileStorage) GetApplicationsByUserID(userID string) []*ReturnExchangeApplication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*ReturnExchangeApplication
	for _, app := range s.Applications {
		if app.UserID == userID {
			result = append(result, app)
		}
	}
	return result
}

func (s *FileStorage) CreateVoucher(voucher *Voucher) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Vouchers[voucher.VoucherID] = voucher
	return s.Save()
}

func (s *FileStorage) GetVoucherByID(id string) (*Voucher, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	voucher, exists := s.Vouchers[id]
	return voucher, exists
}

func (s *FileStorage) GetVouchersByApplicationID(appID string) []*Voucher {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*Voucher
	for _, v := range s.Vouchers {
		if v.ApplicationID == appID {
			result = append(result, v)
		}
	}
	return result
}

func (s *FileStorage) CreateRefundRecord(refund *RefundRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RefundRecords[refund.RefundID] = refund
	return s.Save()
}

func (s *FileStorage) GetRefundRecordByID(id string) (*RefundRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	refund, exists := s.RefundRecords[id]
	return refund, exists
}

func (s *FileStorage) GetRefundRecordByApplicationID(appID string) (*RefundRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.RefundRecords {
		if r.ApplicationID == appID {
			return r, true
		}
	}
	return nil, false
}

func (s *FileStorage) CreateShippingOrder(shipping *ShippingOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ShippingOrders[shipping.ShippingID] = shipping
	return s.Save()
}

func (s *FileStorage) GetShippingOrderByID(id string) (*ShippingOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shipping, exists := s.ShippingOrders[id]
	return shipping, exists
}

func (s *FileStorage) GetShippingOrderByApplicationID(appID string) (*ShippingOrder, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, sh := range s.ShippingOrders {
		if sh.ApplicationID == appID {
			return sh, true
		}
	}
	return nil, false
}

func (s *FileStorage) GetOrderByID(orderID string) (*Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, exists := s.Orders[orderID]
	return order, exists
}

func (s *FileStorage) GetSKUPrice(sku, spec string) (*SKUPrice, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := fmt.Sprintf("%s_%s", sku, spec)
	price, exists := s.SKUPrices[key]
	return price, exists
}

func GetNowTime() time.Time {
	return time.Now().Truncate(time.Second)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

package core

import (
	"sync"
	"time"
)

type Product struct {
	ID        string
	Name      string
	Barcode   string
	Price     int64
	StockQty  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MemberLevel struct {
	ID           string
	Name         string
	DiscountRate float64
	CreatedAt    time.Time
}

type Member struct {
	ID            string
	Name          string
	Phone         string
	MemberLevelID string
	CreatedAt     time.Time
}

type OrderItem struct {
	ProductID   string
	ProductName string
	Price       int64
	Quantity    int
	SubTotal    int64
}

type StockStatus string

const (
	StockStatusNormal    StockStatus = "normal"
	StockStatusPending   StockStatus = "pending"
)

type Order struct {
	ID             string
	CashierID      string
	MemberID       string
	Items          []OrderItem
	TotalAmount    int64
	DiscountAmount int64
	PayableAmount  int64
	PaymentMethod  string
	PaymentTime    time.Time
	StockStatus    StockStatus
	CreatedAt      time.Time
}

type DailyClosingPaymentStat struct {
	PaymentMethod string
	TotalAmount   int64
	OrderCount    int
}

type DailyClosing struct {
	ID          string
	Date        string
	TotalAmount int64
	PaymentStats []DailyClosingPaymentStat
	OrderCount  int
	ClosedAt    time.Time
}

type StocktakingItem struct {
	ProductID   string
	ProductName string
	Barcode     string
	Price       int64
	SnapshotQty int
	ActualQty   int
	DiffQty     int
	DiffAmount  int64
}

type StocktakingStatus string

const (
	StocktakingStatusPending  StocktakingStatus = "pending"
	StocktakingStatusSubmitted StocktakingStatus = "submitted"
	StocktakingStatusCompleted StocktakingStatus = "completed"
)

type Stocktaking struct {
	ID          string
	Status      StocktakingStatus
	OperatorID  string
	CreatedAt   time.Time
	CompletedAt time.Time
	Items       []StocktakingItem
}

type PendingOperationType string

const (
	PendingOpTypeSale   PendingOperationType = "sale"
	PendingOpTypeStockIn PendingOperationType = "stock_in"
)

type PendingOperation struct {
	ID            string
	StocktakingID string
	OperationType PendingOperationType
	ProductID     string
	Quantity      int
	OperationTime time.Time
}

type StockOperation struct {
	ID            string
	ProductID     string
	OperationType string
	Quantity      int
	OperatorID    string
	CreatedAt     time.Time
}

type Store struct {
	mu sync.RWMutex

	products       map[string]*Product
	productsByBarcode map[string]string
	memberLevels   map[string]*MemberLevel
	members        map[string]*Member
	orders         map[string]*Order
	dailyClosings  map[string]*DailyClosing
	stocktakings   map[string]*Stocktaking
	pendingOps     map[string][]*PendingOperation
	stockOps       []*StockOperation

	idGen *IDGenerator
}

func NewStore() *Store {
	return &Store{
		products:       make(map[string]*Product),
		productsByBarcode: make(map[string]string),
		memberLevels:   make(map[string]*MemberLevel),
		members:        make(map[string]*Member),
		orders:         make(map[string]*Order),
		dailyClosings:  make(map[string]*DailyClosing),
		stocktakings:   make(map[string]*Stocktaking),
		pendingOps:     make(map[string][]*PendingOperation),
		stockOps:       make([]*StockOperation, 0),
		idGen:          NewIDGenerator(),
	}
}

func (s *Store) Lock() {
	s.mu.Lock()
}

func (s *Store) Unlock() {
	s.mu.Unlock()
}

func (s *Store) RLock() {
	s.mu.RLock()
}

func (s *Store) RUnlock() {
	s.mu.RUnlock()
}

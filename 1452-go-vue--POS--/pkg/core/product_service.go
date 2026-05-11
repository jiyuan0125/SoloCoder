package core

import (
	"errors"
	"time"
)

var (
	ErrProductNotFound   = errors.New("product not found")
	ErrProductExists     = errors.New("product with this barcode already exists")
	ErrInvalidPrice      = errors.New("invalid price")
	ErrInvalidStockQty   = errors.New("invalid stock quantity")
)

func (s *Store) CreateProduct(name, barcode string, price int64, stockQty int) (string, error) {
	if price <= 0 {
		return "", ErrInvalidPrice
	}
	if stockQty < 0 {
		return "", ErrInvalidStockQty
	}

	s.Lock()
	defer s.Unlock()

	if _, exists := s.productsByBarcode[barcode]; exists {
		return "", ErrProductExists
	}

	id := s.idGen.Generate()
	now := time.Now()
	product := &Product{
		ID:        id,
		Name:      name,
		Barcode:   barcode,
		Price:     price,
		StockQty:  stockQty,
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.products[id] = product
	s.productsByBarcode[barcode] = id

	return id, nil
}

func (s *Store) GetProductByID(id string) (*Product, error) {
	s.RLock()
	defer s.RUnlock()

	product, exists := s.products[id]
	if !exists {
		return nil, ErrProductNotFound
	}

	copy := *product
	return &copy, nil
}

func (s *Store) GetProductByBarcode(barcode string) (*Product, error) {
	s.RLock()
	defer s.RUnlock()

	id, exists := s.productsByBarcode[barcode]
	if !exists {
		return nil, ErrProductNotFound
	}

	product := s.products[id]
	copy := *product
	return &copy, nil
}

func (s *Store) GetAllProducts() []Product {
	s.RLock()
	defer s.RUnlock()

	products := make([]Product, 0, len(s.products))
	for _, p := range s.products {
		products = append(products, *p)
	}
	return products
}

package inventory

import (
	"errors"

	"warehouse-mgmt/internal/models"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) AddProduct(db *models.Database, id, name string, safetyStock int) error {
	if id == "" || name == "" {
		return errors.New("product id and name are required")
	}
	if safetyStock < 0 {
		return errors.New("safety stock cannot be negative")
	}
	if db.GetProductByID(id) != nil {
		return errors.New("product already exists")
	}
	db.Products = append(db.Products, models.Product{
		ID:          id,
		Name:        name,
		SafetyStock: safetyStock,
	})
	return nil
}

func (s *Service) CheckStockAlert(db *models.Database, productID string) (bool, int) {
	product := db.GetProductByID(productID)
	if product == nil {
		return false, 0
	}
	total := db.GetTotalStock(productID)
	return total < product.SafetyStock, total
}

func (s *Service) GetAlerts(db *models.Database) []string {
	var alerts []string
	for _, product := range db.Products {
		if alert, _ := s.CheckStockAlert(db, product.ID); alert {
			alerts = append(alerts, product.ID)
		}
	}
	return alerts
}

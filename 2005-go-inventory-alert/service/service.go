package service

import (
	"errors"
	"time"

	"inventory-alert/database"
	"inventory-alert/models"
)

type DuplicateAlertError struct {
	LastAlertTime time.Time
}

func (e *DuplicateAlertError) Error() string {
	return "duplicate alert within 24 hours"
}

func CreateProduct(product *models.Product) (int64, error) {
	if product.Name == "" {
		return 0, errors.New("product name is required")
	}
	if product.Category == "" {
		return 0, errors.New("product category is required")
	}
	if product.SafetyStock <= 0 {
		return 0, errors.New("safety stock must be positive")
	}
	if product.PurchaseLeadTime <= 0 {
		return 0, errors.New("purchase lead time must be positive")
	}

	return database.CreateProduct(product)
}

func GetProduct(id int64) (*models.Product, error) {
	return database.GetProductByID(id)
}

func GetProducts() ([]*models.Product, error) {
	return database.GetAllProducts()
}

func UpdateProduct(product *models.Product) error {
	existing, err := database.GetProductByID(product.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("product not found")
	}

	if product.Name == "" {
		return errors.New("product name is required")
	}
	if product.Category == "" {
		return errors.New("product category is required")
	}
	if product.SafetyStock <= 0 {
		return errors.New("safety stock must be positive")
	}
	if product.PurchaseLeadTime <= 0 {
		return errors.New("purchase lead time must be positive")
	}

	return database.UpdateProduct(product)
}

func DeleteProduct(id int64) error {
	existing, err := database.GetProductByID(id)
	if err != nil {
		return err
	}
	if existing == nil {
		return errors.New("product not found")
	}

	return database.DeleteProduct(id)
}

func CheckAndCreateAlert(productID int64) (*models.Alert, error) {
	product, err := database.GetProductByID(productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if product.CurrentStock >= product.SafetyStock {
		return nil, nil
	}

	lastAlertTime, err := database.GetLastAlertTime(productID)
	if err != nil {
		return nil, err
	}

	if !lastAlertTime.IsZero() {
		duration := time.Since(lastAlertTime)
		if duration < 24*time.Hour {
			return nil, &DuplicateAlertError{LastAlertTime: lastAlertTime}
		}
	}

	level := models.CalculateAlertLevel(product.CurrentStock, product.SafetyStock)
	suggestedQty := models.CalculateSuggestedQty(product.CurrentStock, product.SafetyStock)

	alert := &models.Alert{
		ProductID:    product.ID,
		ProductName:  product.Name,
		Category:     product.Category,
		CurrentStock: product.CurrentStock,
		SafetyStock:  product.SafetyStock,
		Level:        level,
		Status:       models.AlertStatusPending,
		AssignedTo:   "",
		SuggestedQty: suggestedQty,
	}

	id, err := database.CreateAlert(alert)
	if err != nil {
		return nil, err
	}

	alert.ID = id
	return alert, nil
}

func ProcessPurchase(productID int64, quantity int) error {
	if quantity <= 0 {
		return errors.New("purchase quantity must be positive")
	}

	product, err := database.GetProductByID(productID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}

	err = database.UpdateStock(productID, quantity)
	if err != nil {
		return err
	}

	database.CloseProductAlerts(productID)
	return nil
}

func GetAlert(id int64) (*models.Alert, error) {
	return database.GetAlertByID(id)
}

func GetAlerts(level string) ([]*models.Alert, error) {
	if level != "" && level != string(models.AlertLevelYellow) && 
	   level != string(models.AlertLevelOrange) && 
	   level != string(models.AlertLevelRed) {
		return nil, errors.New("invalid alert level")
	}
	return database.GetAllAlerts(level)
}

func UpdateAlertStatus(alertID int64, status models.AlertStatus) error {
	err := models.ValidateAlertStatus(status)
	if err != nil {
		return err
	}

	alert, err := database.GetAlertByID(alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return errors.New("alert not found")
	}

	err = models.ValidateAlertStatusTransition(alert.Status, status)
	if err != nil {
		return err
	}

	alert.Status = status
	return database.UpdateAlert(alert)
}

func AssignAlert(alertID int64, assignedTo string) error {
	if assignedTo == "" {
		return errors.New("assignee is required")
	}

	alert, err := database.GetAlertByID(alertID)
	if err != nil {
		return err
	}
	if alert == nil {
		return errors.New("alert not found")
	}

	alert.AssignedTo = assignedTo
	return database.UpdateAlert(alert)
}

func GetCategoryStatistics() ([]*models.CategoryStats, error) {
	return database.GetCategoryStats()
}

package repositories

import (
	"hospital-pharmacy/pkg/database"
	"hospital-pharmacy/pkg/models"
	"time"

	"gorm.io/gorm"
)

func GetStockItemsByDrug(drugID uint) ([]models.StockItem, error) {
	var items []models.StockItem
	err := database.DB.Where("drug_id = ? AND quantity > 0", drugID).
		Order("expiry_date ASC").Find(&items).Error
	return items, err
}

func CreateStockItem(item *models.StockItem) error {
	return database.DB.Create(item).Error
}

func UpdateStockItem(item *models.StockItem) error {
	return database.DB.Save(item).Error
}

func GetStockItemByID(id uint) (*models.StockItem, error) {
	var item models.StockItem
	err := database.DB.First(&item, id).Error
	return &item, err
}

func GetAvailableStock(drugID uint) (int, error) {
	var total int
	var items []models.StockItem
	now := time.Now()
	err := database.DB.Where("drug_id = ? AND expiry_date > ? AND quantity > 0", drugID, now).
		Find(&items).Error
	if err != nil {
		return 0, err
	}
	for _, item := range items {
		total += item.Quantity
	}
	return total, nil
}

func GetUsableStockItems(drugID uint) ([]models.StockItem, error) {
	var items []models.StockItem
	now := time.Now()
	err := database.DB.Where("drug_id = ? AND expiry_date > ? AND quantity > 0", drugID, now).
		Order("expiry_date ASC").Find(&items).Error
	return items, err
}

func UpdateStockWithTx(tx *gorm.DB, item *models.StockItem) error {
	return tx.Save(item).Error
}

func CreateTransaction(tx *gorm.DB, trans *models.StockTransaction) error {
	if tx != nil {
		return tx.Create(trans).Error
	}
	return database.DB.Create(trans).Error
}

func ListTransactions(start, end time.Time) ([]models.StockTransaction, error) {
	var transactions []models.StockTransaction
	query := database.DB.Preload("Drug")
	if !start.IsZero() {
		query = query.Where("created_at >= ?", start)
	}
	if !end.IsZero() {
		query = query.Where("created_at <= ?", end)
	}
	err := query.Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

func CreateAlert(alert *models.StockAlert) error {
	return database.DB.Create(alert).Error
}

func ListUnreadAlerts() ([]models.StockAlert, error) {
	var alerts []models.StockAlert
	err := database.DB.Preload("Drug").Where("is_read = ?", false).
		Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

func MarkAlertRead(id uint) error {
	return database.DB.Model(&models.StockAlert{}).Where("id = ?", id).Update("is_read", true).Error
}

func GetNearExpiryStockItems() ([]models.StockItem, error) {
	var items []models.StockItem
	now := time.Now()
	nearDate := now.AddDate(0, 0, 180)
	err := database.DB.Preload("Drug").Where("expiry_date <= ? AND quantity > 0", nearDate).
		Order("expiry_date ASC").Find(&items).Error
	return items, err
}

func BeginTransaction() *gorm.DB {
	return database.DB.Begin()
}

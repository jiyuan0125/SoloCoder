package services

import (
	"errors"
	"fmt"
	"hospital-pharmacy/pkg/models"
	"hospital-pharmacy/pkg/repositories"
	"sync"
	"time"
)

var stockMutex sync.Mutex

type StockInRequest struct {
	DrugID         uint
	Quantity       int
	ProductionDate time.Time
	ExpiryDate     time.Time
	Supplier       string
	SupplierCode   string
	Operator1      string
	Operator2      string
}

type StockOutRequest struct {
	DrugID    uint
	Quantity  int
	Operator1 string
	Operator2 string
	Remark    string
}

func StockIn(req StockInRequest) error {
	stockMutex.Lock()
	defer stockMutex.Unlock()

	if req.Quantity <= 0 {
		return errors.New("入库数量必须大于0")
	}
	if req.ExpiryDate.Before(time.Now()) {
		return errors.New("有效期不能早于当前日期")
	}

	drug, err := repositories.GetDrugByID(req.DrugID)
	if err != nil {
		return errors.New("药品不存在")
	}

	seq := time.Now().Unix() % 1000000
	batchNumber := req.SupplierCode + fmt.Sprintf("%06d", seq)

	stockItem := &models.StockItem{
		DrugID:         req.DrugID,
		BatchNumber:    batchNumber,
		Quantity:       req.Quantity,
		ProductionDate: req.ProductionDate,
		ExpiryDate:     req.ExpiryDate,
		Supplier:       req.Supplier,
		SupplierCode:   req.SupplierCode,
	}

	tx := repositories.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := repositories.CreateStockItemWithTx(tx, stockItem); err != nil {
		tx.Rollback()
		return err
	}

	drug.CurrentStock += req.Quantity
	if err := repositories.UpdateDrugWithTx(tx, drug); err != nil {
		tx.Rollback()
		return err
	}

	transaction := &models.StockTransaction{
		TransactionType: "in",
		DrugID:          req.DrugID,
		StockItemID:     &stockItem.ID,
		Quantity:        req.Quantity,
		BatchNumber:     batchNumber,
		ProductionDate:  req.ProductionDate,
		ExpiryDate:      req.ExpiryDate,
		Supplier:        req.Supplier,
		SupplierCode:    req.SupplierCode,
		Operator1:       req.Operator1,
		Operator2:       req.Operator2,
	}

	if err := repositories.CreateTransaction(tx, transaction); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	checkStockAlert(drug)
	return nil
}

func StockOut(req StockOutRequest) error {
	stockMutex.Lock()
	defer stockMutex.Unlock()

	if req.Quantity <= 0 {
		return errors.New("出库数量必须大于0")
	}

	drug, err := repositories.GetDrugByID(req.DrugID)
	if err != nil {
		return errors.New("药品不存在")
	}

	if drug.IsSpecialDrug && (req.Operator1 == "" || req.Operator2 == "") {
		return errors.New("特殊药品出库需要双人复核")
	}

	items, err := repositories.GetUsableStockItems(req.DrugID)
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return errors.New("没有可用库存")
	}

	totalAvailable := 0
	for _, item := range items {
		totalAvailable += item.Quantity
	}

	if totalAvailable < req.Quantity {
		return errors.New(fmt.Sprintf("库存不足，当前可用: %d，需要: %d", totalAvailable, req.Quantity))
	}

	tx := repositories.BeginTransaction()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	remaining := req.Quantity
	for i := 0; i < len(items) && remaining > 0; i++ {
		item := items[i]
		if item.Quantity <= 0 {
			continue
		}

		var takeQty int
		if item.Quantity >= remaining {
			takeQty = remaining
			item.Quantity -= remaining
			remaining = 0
		} else {
			takeQty = item.Quantity
			remaining -= item.Quantity
			item.Quantity = 0
		}

		if err := repositories.UpdateStockWithTx(tx, &item); err != nil {
			tx.Rollback()
			return err
		}

		transaction := &models.StockTransaction{
			TransactionType: "out",
			DrugID:          req.DrugID,
			StockItemID:     &item.ID,
			Quantity:        takeQty,
			BatchNumber:     item.BatchNumber,
			ProductionDate:  item.ProductionDate,
			ExpiryDate:      item.ExpiryDate,
			Supplier:        item.Supplier,
			SupplierCode:    item.SupplierCode,
			Operator1:       req.Operator1,
			Operator2:       req.Operator2,
			Remark:          req.Remark,
		}

		if err := repositories.CreateTransaction(tx, transaction); err != nil {
			tx.Rollback()
			return err
		}
	}

	drug.CurrentStock -= req.Quantity
	if err := repositories.UpdateDrugWithTx(tx, drug); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	checkStockAlert(drug)
	return nil
}

func checkStockAlert(drug *models.Drug) {
	if drug.CurrentStock < drug.MinStock {
		alert := &models.StockAlert{
			DrugID:    drug.ID,
			AlertType: "low_stock",
			Message:   fmt.Sprintf("药品【%s】库存不足，当前库存: %d，预警下限: %d", drug.GenericName, drug.CurrentStock, drug.MinStock),
		}
		repositories.CreateAlert(alert)
	} else if drug.CurrentStock > drug.MaxStock {
		alert := &models.StockAlert{
			DrugID:    drug.ID,
			AlertType: "over_stock",
			Message:   fmt.Sprintf("药品【%s】库存积压，当前库存: %d，预警上限: %d", drug.GenericName, drug.CurrentStock, drug.MaxStock),
		}
		repositories.CreateAlert(alert)
	}
}

func GetStockItems(drugID uint) ([]models.StockItem, error) {
	return repositories.GetStockItemsByDrug(drugID)
}

func GetAlerts() ([]models.StockAlert, error) {
	return repositories.ListUnreadAlerts()
}

func MarkAlertRead(id uint) error {
	return repositories.MarkAlertRead(id)
}

func GetNearExpiryItems() ([]models.StockItem, error) {
	return repositories.GetNearExpiryStockItems()
}

func GetStockValue() (float64, float64, error) {
	drugs, err := repositories.ListDrugs("")
	if err != nil {
		return 0, 0, err
	}

	var costValue, retailValue float64
	for _, drug := range drugs {
		costValue += float64(drug.CurrentStock) * drug.PurchasePrice
		retailValue += float64(drug.CurrentStock) * drug.RetailPrice
	}
	return costValue, retailValue, nil
}

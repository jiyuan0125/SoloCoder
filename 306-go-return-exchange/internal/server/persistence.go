package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"return-exchange/pkg/models"
)

const (
	ordersFile          = "orders.json"
	applicationsFile    = "applications.json"
	refundRecordsFile   = "refund_records.json"
	shippingOrdersFile  = "shipping_orders.json"
)

func SaveData(store *Store, dataDir string) error {
	if err := saveOrders(store, dataDir); err != nil {
		return err
	}
	if err := saveApplications(store, dataDir); err != nil {
		return err
	}
	if err := saveRefundRecords(store, dataDir); err != nil {
		return err
	}
	if err := saveShippingOrders(store, dataDir); err != nil {
		return err
	}
	return nil
}

func LoadData(store *Store, dataDir string) error {
	if err := loadOrders(store, dataDir); err != nil {
		fmt.Printf("Warning: could not load orders: %v\n", err)
	}
	if err := loadApplications(store, dataDir); err != nil {
		fmt.Printf("Warning: could not load applications: %v\n", err)
	}
	if err := loadRefundRecords(store, dataDir); err != nil {
		fmt.Printf("Warning: could not load refund records: %v\n", err)
	}
	if err := loadShippingOrders(store, dataDir); err != nil {
		fmt.Printf("Warning: could not load shipping orders: %v\n", err)
	}
	return nil
}

func saveOrders(store *Store, dataDir string) error {
	orders := store.GetAllOrders()
	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, ordersFile), data, 0644)
}

func saveApplications(store *Store, dataDir string) error {
	applications := store.GetAllApplications()
	data, err := json.MarshalIndent(applications, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, applicationsFile), data, 0644)
}

func saveRefundRecords(store *Store, dataDir string) error {
	records := store.GetAllRefundRecords()
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, refundRecordsFile), data, 0644)
}

func saveShippingOrders(store *Store, dataDir string) error {
	orders := store.GetAllShippingOrders()
	data, err := json.MarshalIndent(orders, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, shippingOrdersFile), data, 0644)
}

func loadOrders(store *Store, dataDir string) error {
	filePath := filepath.Join(dataDir, ordersFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var orders map[string]*models.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return err
	}

	store.LoadOrders(orders)
	return nil
}

func loadApplications(store *Store, dataDir string) error {
	filePath := filepath.Join(dataDir, applicationsFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var applications map[string]*models.ReturnExchangeApplication
	if err := json.Unmarshal(data, &applications); err != nil {
		return err
	}

	store.LoadApplications(applications)
	return nil
}

func loadRefundRecords(store *Store, dataDir string) error {
	filePath := filepath.Join(dataDir, refundRecordsFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var records map[string]*models.RefundRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return err
	}

	store.LoadRefundRecords(records)
	return nil
}

func loadShippingOrders(store *Store, dataDir string) error {
	filePath := filepath.Join(dataDir, shippingOrdersFile)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var orders map[string]*models.ShippingOrder
	if err := json.Unmarshal(data, &orders); err != nil {
		return err
	}

	store.LoadShippingOrders(orders)
	return nil
}

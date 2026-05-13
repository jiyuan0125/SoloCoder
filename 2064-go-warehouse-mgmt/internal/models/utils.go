package models

import (
	"fmt"
	"strings"
	"time"
)

func GenerateID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

func (db *Database) GetProductByID(id string) *Product {
	for i := range db.Products {
		if db.Products[i].ID == id {
			return &db.Products[i]
		}
	}
	return nil
}

func (db *Database) GetLocationByCode(code string) *Location {
	for i := range db.Locations {
		if db.Locations[i].Code == code {
			return &db.Locations[i]
		}
	}
	return nil
}

func (db *Database) GetBatchByID(id string) *Batch {
	for i := range db.Batches {
		if db.Batches[i].ID == id {
			return &db.Batches[i]
		}
	}
	return nil
}

func (db *Database) GetStockRecord(productID, locationCode string) *StockRecord {
	for i := range db.StockRecords {
		if db.StockRecords[i].ProductID == productID &&
			db.StockRecords[i].LocationCode == locationCode {
			return &db.StockRecords[i]
		}
	}
	return nil
}

func (db *Database) GetTotalStock(productID string) int {
	total := 0
	for _, sr := range db.StockRecords {
		if sr.ProductID == productID {
			total += sr.Quantity
		}
	}
	return total
}

func (db *Database) GetProductLocations(productID string) []StockRecord {
	var records []StockRecord
	for _, sr := range db.StockRecords {
		if sr.ProductID == productID {
			records = append(records, sr)
		}
	}
	return records
}

func (db *Database) GetLocationProducts(locationCode string) []StockRecord {
	var records []StockRecord
	for _, sr := range db.StockRecords {
		if sr.LocationCode == locationCode {
			records = append(records, sr)
		}
	}
	return records
}

func (db *Database) GetProductBatches(productID string) []Batch {
	var batches []Batch
	for _, b := range db.Batches {
		if b.ProductID == productID && b.Quantity > 0 {
			batches = append(batches, b)
		}
	}
	return batches
}

func (db *Database) GetLocationsWithProduct(productID string) []string {
	locations := make(map[string]bool)
	for _, b := range db.Batches {
		if b.ProductID == productID && b.Quantity > 0 {
			locations[b.LocationCode] = true
		}
	}
	var codes []string
	for code := range locations {
		codes = append(codes, code)
	}
	return codes
}

func (db *Database) GetAvailableCapacity(locationCode string) int {
	loc := db.GetLocationByCode(locationCode)
	if loc == nil {
		return 0
	}
	return loc.Capacity - loc.Used
}

func LocationPriority(code1, code2 string) bool {
	parts1 := strings.Split(code1, "-")
	parts2 := strings.Split(code2, "-")
	for i := 0; i < len(parts1) && i < len(parts2); i++ {
		if parts1[i] != parts2[i] {
			return code1 < code2
		}
	}
	return len(parts1) < len(parts2)
}

func (db *Database) UpdateStockRecord(productID, locationCode string, delta int) {
	sr := db.GetStockRecord(productID, locationCode)
	if sr == nil {
		if delta > 0 {
			db.StockRecords = append(db.StockRecords, StockRecord{
				ProductID:    productID,
				LocationCode: locationCode,
				Quantity:     delta,
			})
		}
		return
	}
	sr.Quantity += delta
	if sr.Quantity == 0 {
		for i, record := range db.StockRecords {
			if record.ProductID == productID && record.LocationCode == locationCode {
				db.StockRecords = append(db.StockRecords[:i], db.StockRecords[i+1:]...)
				return
			}
		}
	}
}

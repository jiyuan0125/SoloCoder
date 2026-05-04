package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

type DeliveryStatus struct {
	OrderID     string  `json:"order_id"`
	StatusName  string  `json:"status_name"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Timestamp   int64   `json:"timestamp"`
	ReceiveSeq  int     `json:"receive_seq"`
}

type DataStore struct {
	filePath string
	data     map[string][]DeliveryStatus
	seq      int
	mu       sync.RWMutex
}

func NewDataStore(filePath string) (*DataStore, error) {
	ds := &DataStore{
		filePath: filePath,
		data:     make(map[string][]DeliveryStatus),
		seq:      0,
	}

	if err := ds.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return ds, nil
}

func (ds *DataStore) Load() error {
	file, err := os.Open(ds.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&ds.data); err != nil {
		return err
	}

	maxSeq := 0
	for _, statuses := range ds.data {
		for _, s := range statuses {
			if s.ReceiveSeq > maxSeq {
				maxSeq = s.ReceiveSeq
			}
		}
	}
	ds.seq = maxSeq

	return nil
}

func (ds *DataStore) Save() error {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	file, err := os.Create(ds.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(ds.data)
}

func (ds *DataStore) AddStatus(status DeliveryStatus) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	orderID := status.OrderID
	existing := ds.data[orderID]

	for _, s := range existing {
		if s.StatusName == status.StatusName && s.Timestamp == status.Timestamp {
			return false
		}
	}

	ds.seq++
	status.ReceiveSeq = ds.seq
	ds.data[orderID] = append(ds.data[orderID], status)

	return true
}

func (ds *DataStore) GetStatuses(orderID string) []DeliveryStatus {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	statuses, ok := ds.data[orderID]
	if !ok {
		return nil
	}

	result := make([]DeliveryStatus, len(statuses))
	copy(result, statuses)

	sort.Slice(result, func(i, j int) bool {
		if result[i].Timestamp == result[j].Timestamp {
			return result[i].ReceiveSeq < result[j].ReceiveSeq
		}
		return result[i].Timestamp < result[j].Timestamp
	})

	return result
}

var validStatusNames = map[string]bool{
	"取餐确认": true,
	"商家出发": true,
	"到达小区": true,
	"已送达":   true,
}

func IsValidStatusName(name string) bool {
	return validStatusNames[name]
}

func IsValidCoordinate(longitude, latitude float64) bool {
	return longitude >= -180 && longitude <= 180 && latitude >= -90 && latitude <= 90
}

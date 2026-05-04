package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

const (
	venuesFile = "venues.json"
	eventsFile = "events.json"
	ordersFile = "orders.json"
)

var (
	storageMutex   sync.RWMutex
	storageTicker  *time.Ticker
)

type DataSnapshot struct {
	Venues map[string]*Venue  `json:"venues"`
	Events map[string]*Event  `json:"events"`
	Orders map[string]*Order  `json:"orders"`
}

func LoadData() error {
	storageMutex.Lock()
	defer storageMutex.Unlock()

	if _, err := os.Stat(venuesFile); err == nil {
		data, err := os.ReadFile(venuesFile)
		if err == nil {
			venueMutex.Lock()
			json.Unmarshal(data, &venues)
			venueMutex.Unlock()
		}
	}

	if _, err := os.Stat(eventsFile); err == nil {
		data, err := os.ReadFile(eventsFile)
		if err == nil {
			eventMutex.Lock()
			json.Unmarshal(data, &events)
			eventMutex.Unlock()

			if len(events) > 0 {
				startLockCleanupIfNeeded()
			}
		}
	}

	if _, err := os.Stat(ordersFile); err == nil {
		data, err := os.ReadFile(ordersFile)
		if err == nil {
			orderMutex.Lock()
			json.Unmarshal(data, &orders)
			orderMutex.Unlock()
		}
	}

	return nil
}

func SaveData() error {
	storageMutex.RLock()
	defer storageMutex.RUnlock()

	venueMutex.RLock()
	venuesData, err := json.MarshalIndent(venues, "", "  ")
	venueMutex.RUnlock()
	if err != nil {
		return err
	}
	err = os.WriteFile(venuesFile+".tmp", venuesData, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(venuesFile+".tmp", venuesFile)
	if err != nil {
		return err
	}

	eventMutex.RLock()
	eventsData, err := json.MarshalIndent(events, "", "  ")
	eventMutex.RUnlock()
	if err != nil {
		return err
	}
	err = os.WriteFile(eventsFile+".tmp", eventsData, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(eventsFile+".tmp", eventsFile)
	if err != nil {
		return err
	}

	orderMutex.RLock()
	ordersData, err := json.MarshalIndent(orders, "", "  ")
	orderMutex.RUnlock()
	if err != nil {
		return err
	}
	err = os.WriteFile(ordersFile+".tmp", ordersData, 0644)
	if err != nil {
		return err
	}
	err = os.Rename(ordersFile+".tmp", ordersFile)
	if err != nil {
		return err
	}

	return nil
}

func StartAutoSave() {
	if storageTicker != nil {
		return
	}

	storageTicker = time.NewTicker(30 * time.Second)
	go func() {
		for range storageTicker.C {
			SaveData()
		}
	}()
}

func StopAutoSave() {
	if storageTicker != nil {
		storageTicker.Stop()
		storageTicker = nil
	}
}

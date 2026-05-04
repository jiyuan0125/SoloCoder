package server

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

type Persistence struct {
	store     *Store
	filePath  string
	mu        sync.Mutex
	saveTimer *time.Ticker
}

func NewPersistence(store *Store, filePath string) *Persistence {
	return &Persistence{
		store:    store,
		filePath: filePath,
	}
}

func (p *Persistence) Save() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	data := p.store.GetData()

	file, err := os.Create(p.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (p *Persistence) Load() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Open(p.filePath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer file.Close()

	var data StoreData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	p.store.LoadData(&data)
	return nil
}

func (p *Persistence) StartAutoSave(interval time.Duration) {
	p.saveTimer = time.NewTicker(interval)
	go func() {
		for range p.saveTimer.C {
			p.Save()
		}
	}()
}

func (p *Persistence) StopAutoSave() {
	if p.saveTimer != nil {
		p.saveTimer.Stop()
	}
}

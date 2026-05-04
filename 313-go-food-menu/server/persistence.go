package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/example/food-menu/common"
)

const (
	DataDir       = "data"
	DishesFile    = "dishes.json"
	RecommendFile = "recommend.json"
)

type Persistence struct {
	store *Store
}

func NewPersistence(store *Store) *Persistence {
	return &Persistence{store: store}
}

type dishesData struct {
	UpdatedAt time.Time              `json:"updated_at"`
	Dishes    map[string]*common.Dish `json:"dishes"`
}

type recommendData struct {
	UpdatedAt time.Time            `json:"updated_at"`
	Recommend *common.RecommendInfo `json:"recommend"`
}

func (p *Persistence) ensureDataDir() error {
	return os.MkdirAll(DataDir, 0755)
}

func (p *Persistence) SaveDishes() error {
	if err := p.ensureDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	data := dishesData{
		UpdatedAt: time.Now(),
		Dishes:    p.store.GetAllDishes(),
	}

	filePath := filepath.Join(DataDir, DishesFile)
	tempPath := filePath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		file.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to encode dishes: %w", err)
	}

	file.Close()

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (p *Persistence) LoadDishes() error {
	filePath := filepath.Join(DataDir, DishesFile)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open dishes file: %w", err)
	}
	defer file.Close()

	var data dishesData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode dishes: %w", err)
	}

	p.store.LoadDishes(data.Dishes)
	return nil
}

func (p *Persistence) SaveRecommend() error {
	if err := p.ensureDataDir(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	data := recommendData{
		UpdatedAt: time.Now(),
		Recommend: p.store.GetRecommendInfo(),
	}

	filePath := filepath.Join(DataDir, RecommendFile)
	tempPath := filePath + ".tmp"

	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		file.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to encode recommend: %w", err)
	}

	file.Close()

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

func (p *Persistence) LoadRecommend() error {
	filePath := filepath.Join(DataDir, RecommendFile)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open recommend file: %w", err)
	}
	defer file.Close()

	var data recommendData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return fmt.Errorf("failed to decode recommend: %w", err)
	}

	if data.Recommend != nil {
		p.store.SetRecommendInfo(data.Recommend)
	}

	return nil
}

func (p *Persistence) LoadAll() error {
	if err := p.LoadDishes(); err != nil {
		return err
	}
	if err := p.LoadRecommend(); err != nil {
		return err
	}
	return nil
}

func (p *Persistence) SaveAll() error {
	if err := p.SaveDishes(); err != nil {
		return err
	}
	if err := p.SaveRecommend(); err != nil {
		return err
	}
	return nil
}

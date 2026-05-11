package core

import (
	"errors"
	"fmt"
	"sync"
)

type LayerStats struct {
	Index           int     `json:"index"`
	Count           int     `json:"count"`
	Capacity        int     `json:"capacity"`
	HashFunctions   int     `json:"hash_functions"`
	BitCount        int     `json:"bit_count"`
	CurrentFPR      float64 `json:"current_fpr"`
	TargetFPR       float64 `json:"target_fpr"`
	HasWarning      bool    `json:"has_warning"`
	IsFull          bool    `json:"is_full"`
}

type FilterStats struct {
	TotalCount   int          `json:"total_count"`
	TotalCapacity int          `json:"total_capacity"`
	LayerCount    int          `json:"layer_count"`
	Layers        []LayerStats `json:"layers"`
}

type InsertResult struct {
	Inserted    bool
	LayerIndex  int
	WasExisting bool
}

type LayeredBloomFilter struct {
	mu     sync.RWMutex
	layers []*Layer
	config *FilterConfig
}

func NewLayeredBloomFilter(cfg *FilterConfig) (*LayeredBloomFilter, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	if err := ValidateConfig(cfg); err != nil {
		return nil, err
	}
	layers := make([]*Layer, len(cfg.Layers))
	for i, layerCfg := range cfg.Layers {
		layers[i] = NewLayer(i, layerCfg, cfg.WarningFPRFactor)
	}
	return &LayeredBloomFilter{
		layers: layers,
		config: cfg,
	}, nil
}

func (f *LayeredBloomFilter) Contains(data []byte) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if len(f.layers) == 0 {
		return false
	}
	for _, layer := range f.layers {
		if layer.Contains(data) {
			return true
		}
	}
	return false
}

func (f *LayeredBloomFilter) Insert(data []byte) (*InsertResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, layer := range f.layers {
		if layer.Contains(data) {
			return &InsertResult{
				Inserted:    false,
				LayerIndex:  layer.Index(),
				WasExisting: true,
			}, nil
		}
	}
	for i, layer := range f.layers {
		if !layer.IsFull() {
			layer.Add(data)
			return &InsertResult{
				Inserted:    true,
				LayerIndex:  i,
				WasExisting: false,
			}, nil
		}
	}
	if f.config.AutoExpandOnFull {
		if len(f.layers) >= f.config.MaxAutoLayers {
			return nil, errors.New("maximum auto-expand layers reached")
		}
		lastConfig := f.config.Layers[len(f.config.Layers)-1]
		newConfig := LayerConfig{
			Capacity:      lastConfig.Capacity * 2,
			HashFunctions: lastConfig.HashFunctions,
			TargetFPR:     lastConfig.TargetFPR,
		}
		if newConfig.HashFunctions > 2 {
			newConfig.HashFunctions = lastConfig.HashFunctions - 1
		}
		f.config.Layers = append(f.config.Layers, newConfig)
		newLayer := NewLayer(len(f.layers), newConfig, f.config.WarningFPRFactor)
		f.layers = append(f.layers, newLayer)
		newLayer.Add(data)
		return &InsertResult{
			Inserted:    true,
			LayerIndex:  len(f.layers) - 1,
			WasExisting: false,
		}, nil
	}
	return nil, errors.New("all layers are full")
}

func (f *LayeredBloomFilter) Remove(data []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, layer := range f.layers {
		if layer.Remove(data) {
			return true
		}
	}
	return false
}

func (f *LayeredBloomFilter) Stats() *FilterStats {
	f.mu.RLock()
	defer f.mu.RUnlock()
	totalCount := 0
	totalCapacity := 0
	layerStats := make([]LayerStats, len(f.layers))
	for i, layer := range f.layers {
		totalCount += layer.Count()
		totalCapacity += layer.Capacity()
		layerStats[i] = LayerStats{
			Index:         layer.Index(),
			Count:         layer.Count(),
			Capacity:      layer.Capacity(),
			HashFunctions: layer.HashFunctions(),
			BitCount:      layer.BitCount(),
			CurrentFPR:    layer.CurrentFPR(),
			TargetFPR:     layer.TargetFPR(),
			HasWarning:    layer.HasWarning(),
			IsFull:        layer.IsFull(),
		}
	}
	return &FilterStats{
		TotalCount:    totalCount,
		TotalCapacity: totalCapacity,
		LayerCount:    len(f.layers),
		Layers:        layerStats,
	}
}

func (f *LayeredBloomFilter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, layer := range f.layers {
		layer.Reset()
	}
}

func (f *LayeredBloomFilter) Reconfigure(cfg *FilterConfig) error {
	if err := ValidateConfig(cfg); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	layers := make([]*Layer, len(cfg.Layers))
	for i, layerCfg := range cfg.Layers {
		layers[i] = NewLayer(i, layerCfg, cfg.WarningFPRFactor)
	}
	f.layers = layers
	f.config = cfg
	return nil
}

func (f *LayeredBloomFilter) String() string {
	stats := f.Stats()
	result := fmt.Sprintf("Layered Bloom Filter: %d layers, %d/%d elements\n",
		stats.LayerCount, stats.TotalCount, stats.TotalCapacity)
	for _, ls := range stats.Layers {
		warn := ""
		if ls.HasWarning {
			warn = " [FPR WARNING]"
		}
		result += fmt.Sprintf("  Layer %d: %d/%d, hash=%d, fpr=%.6f%s\n",
			ls.Index, ls.Count, ls.Capacity, ls.HashFunctions, ls.CurrentFPR, warn)
	}
	return result
}

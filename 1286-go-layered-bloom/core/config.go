package core

import (
	"fmt"
	"math"
)

type LayerConfig struct {
	Capacity       int
	HashFunctions  int
	TargetFPR      float64
}

type FilterConfig struct {
	Layers            []LayerConfig
	AutoExpandOnFull  bool
	WarningFPRFactor  float64
	MaxAutoLayers     int
}

func DefaultConfig() *FilterConfig {
	return &FilterConfig{
		Layers: []LayerConfig{
			{Capacity: 1000, HashFunctions: 8, TargetFPR: 0.001},
			{Capacity: 5000, HashFunctions: 6, TargetFPR: 0.005},
			{Capacity: 20000, HashFunctions: 4, TargetFPR: 0.01},
		},
		AutoExpandOnFull: true,
		WarningFPRFactor: 2.0,
		MaxAutoLayers:    10,
	}
}

func CalculateOptimalBits(capacity int, targetFPR float64) int {
	if capacity <= 0 || targetFPR <= 0 || targetFPR >= 1 {
		return 1024
	}
	return int(float64(capacity) * math.Abs(math.Log(targetFPR)) / (math.Ln2 * math.Ln2))
}

func CalculateOptimalHashFunctions(capacity int, bits int) int {
	if capacity <= 0 || bits <= 0 {
		return 4
	}
	return int(float64(bits) / float64(capacity) * math.Ln2)
}

func CalculateTheoreticalFPR(count int, capacity int, hashFunctions int, bits int) float64 {
	if count <= 0 || capacity <= 0 || hashFunctions <= 0 || bits <= 0 {
		return 0.0
	}
	loadFactor := float64(count) / float64(bits)
	probability := 1.0 - math.Pow(1.0-loadFactor, float64(hashFunctions))
	return probability
}

func ValidateConfig(cfg *FilterConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if len(cfg.Layers) == 0 {
		return fmt.Errorf("at least one layer is required")
	}
	if cfg.WarningFPRFactor <= 1.0 {
		return fmt.Errorf("warning FPR factor must be greater than 1")
	}
	for i, layer := range cfg.Layers {
		if layer.Capacity <= 0 {
			return fmt.Errorf("layer %d capacity must be positive", i)
		}
		if layer.HashFunctions <= 0 {
			return fmt.Errorf("layer %d hash functions must be positive", i)
		}
		if layer.TargetFPR <= 0 || layer.TargetFPR >= 1 {
			return fmt.Errorf("layer %d target FPR must be between 0 and 1", i)
		}
	}
	return nil
}

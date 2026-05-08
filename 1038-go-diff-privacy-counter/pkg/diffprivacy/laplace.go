package diffprivacy

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

type LaplaceNoise struct {
	mu    float64
	b     float64
	mu1   float64
	mu2   float64
	once1 sync.Once
	once2 sync.Once
}

func NewLaplaceNoise(location, scale float64) *LaplaceNoise {
	return &LaplaceNoise{
		mu: location,
		b:  scale,
	}
}

func (l *LaplaceNoise) Sample() float64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	u := r.Float64() - 0.5
	return l.mu - l.b*math.Copysign(1.0, u)*math.Log(1.0-2.0*math.Abs(u))
}

func AddLaplaceNoise(value int, epsilon float64, sensitivity int) float64 {
	if epsilon <= 0 {
		return float64(value)
	}
	scale := float64(sensitivity) / epsilon
	noise := NewLaplaceNoise(0, scale)
	return float64(value) + noise.Sample()
}

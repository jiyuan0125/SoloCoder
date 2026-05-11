package montecarlo

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

var (
	rngMutex     sync.Mutex
	hasSpare     bool
	spareNormal  float64
	globalRand   *rand.Rand
)

func init() {
	globalRand = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func RandomUniform() float64 {
	rngMutex.Lock()
	defer rngMutex.Unlock()
	return globalRand.Float64()
}

func RandomNormal() float64 {
	rngMutex.Lock()
	defer rngMutex.Unlock()
	if hasSpare {
		hasSpare = false
		return spareNormal
	}
	u1 := globalRand.Float64()
	u2 := globalRand.Float64()
	r := math.Sqrt(-2.0 * math.Log(u1))
	theta := 2.0 * math.Pi * u2
	z0 := r * math.Cos(theta)
	z1 := r * math.Sin(theta)
	spareNormal = z1
	hasSpare = true
	return z0
}

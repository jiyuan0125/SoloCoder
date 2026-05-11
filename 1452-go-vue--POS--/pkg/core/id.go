package core

import (
	"strconv"
	"sync"
	"time"
)

type IDGenerator struct {
	mu       sync.Mutex
	counter  int
	lastTime int64
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{}
}

func (g *IDGenerator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixNano()
	if now == g.lastTime {
		g.counter++
	} else {
		g.counter = 0
		g.lastTime = now
	}

	return strconv.FormatInt(now, 10) + "-" + strconv.Itoa(g.counter)
}

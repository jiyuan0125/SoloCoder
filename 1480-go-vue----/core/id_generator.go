package core

import (
	"sync/atomic"
	"time"
)

type IDGenerator struct {
	counter uint64
}

func NewIDGenerator() *IDGenerator {
	return &IDGenerator{
		counter: 0,
	}
}

func (g *IDGenerator) Generate() string {
	ts := time.Now().UnixNano()
	seq := atomic.AddUint64(&g.counter, 1)
	return string(rune(ts%1000000)) + "-" + string(rune(seq))
}

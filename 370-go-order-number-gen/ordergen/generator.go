package ordergen

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type OrderGenerator struct {
	config     OrderGeneratorConfig
	storage    *Storage
	lastSecond string
	serialNum  int
	mu         sync.Mutex
	randSource *rand.Rand
}

func NewOrderGenerator(config OrderGeneratorConfig) (*OrderGenerator, error) {
	storage, err := NewStorage(config.StoragePath)
	if err != nil {
		return nil, err
	}

	gen := &OrderGenerator{
		config:     config,
		storage:    storage,
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	if err := gen.loadState(); err != nil {
		return nil, err
	}

	return gen, nil
}

func (g *OrderGenerator) loadState() error {
	state, err := g.storage.Load()
	if err != nil {
		return err
	}

	if state != nil {
		nowSecond := time.Now().Format(TimeFormat)
		if state.LastSecond == nowSecond {
			g.lastSecond = state.LastSecond
			g.serialNum = state.LastSerialNum
		}
		if state.Prefix != "" {
			g.config.Prefix = state.Prefix
		}
	}

	return nil
}

func (g *OrderGenerator) saveState() error {
	state := &PersistedState{
		LastSecond:    g.lastSecond,
		LastSerialNum: g.serialNum,
		Prefix:        g.config.Prefix,
	}
	return g.storage.Save(state)
}

func (g *OrderGenerator) Generate() (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	currentSecond := now.Format(TimeFormat)

	if currentSecond != g.lastSecond {
		g.lastSecond = currentSecond
		g.serialNum = 0
	}

	if g.serialNum >= SerialNumMax {
		return "", ErrSerialNumExhausted
	}

	g.serialNum++

	randomNum := g.randSource.Intn(RandomNumMax + 1)

	orderNo := fmt.Sprintf("%s%06d%04d",
		currentSecond,
		g.serialNum,
		randomNum,
	)

	if g.config.Prefix != "" {
		orderNo = g.config.Prefix + orderNo
	}

	if err := g.saveState(); err != nil {
		return "", err
	}

	return orderNo, nil
}

func (g *OrderGenerator) GenerateWithPrefix(prefix string) (string, error) {
	g.mu.Lock()
	oldPrefix := g.config.Prefix
	g.config.Prefix = prefix
	g.mu.Unlock()

	defer func() {
		g.mu.Lock()
		g.config.Prefix = oldPrefix
		g.mu.Unlock()
	}()

	return g.Generate()
}

func (g *OrderGenerator) SetPrefix(prefix string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.config.Prefix = prefix
	g.saveState()
}

func (g *OrderGenerator) GetPrefix() string {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.config.Prefix
}

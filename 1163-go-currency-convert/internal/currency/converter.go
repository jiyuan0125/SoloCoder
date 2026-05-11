package currency

import (
	"fmt"
	"math"
	"strings"
	"sync"
)

type edge struct {
	to   Currency
	rate float64
}

type Converter struct {
	mu         sync.RWMutex
	graph      map[Currency][]edge
	rates      map[string]float64
	currencies map[Currency]bool

	cacheMu      sync.RWMutex
	rateCache    map[string]float64
	hasCache     bool
}

func NewConverter() *Converter {
	return &Converter{
		graph:      make(map[Currency][]edge),
		rates:      make(map[string]float64),
		currencies: make(map[Currency]bool),
		rateCache:  make(map[string]float64),
		hasCache:   false,
	}
}

func (c *Converter) SetRates(rates []Rate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.graph = make(map[Currency][]edge)
	c.rates = make(map[string]float64)
	c.currencies = make(map[Currency]bool)

	for _, r := range rates {
		c.addRateLocked(r)
	}

	c.clearCacheLocked()
}

func (c *Converter) AddRate(r Rate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.addRateLocked(r)
	c.clearCacheLocked()
}

func (c *Converter) AddRates(rates []Rate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, r := range rates {
		c.addRateLocked(r)
	}
	c.clearCacheLocked()
}

func (c *Converter) addRateLocked(r Rate) {
	key := fmt.Sprintf("%s->%s", r.From, r.To)
	c.rates[key] = r.Rate

	c.graph[r.From] = removeEdgeTo(c.graph[r.From], r.To)
	c.graph[r.From] = append(c.graph[r.From], edge{to: r.To, rate: r.Rate})

	reverseKey := fmt.Sprintf("%s->%s", r.To, r.From)
	reverseRate := 1.0 / r.Rate
	c.rates[reverseKey] = reverseRate

	c.graph[r.To] = removeEdgeTo(c.graph[r.To], r.From)
	c.graph[r.To] = append(c.graph[r.To], edge{to: r.From, rate: reverseRate})

	c.currencies[r.From] = true
	c.currencies[r.To] = true
}

func removeEdgeTo(edges []edge, to Currency) []edge {
	result := make([]edge, 0, len(edges))
	for _, e := range edges {
		if e.to != to {
			result = append(result, e)
		}
	}
	return result
}

func (c *Converter) GetCurrencies() []Currency {
	c.mu.RLock()
	defer c.mu.RUnlock()

	currencies := make([]Currency, 0, len(c.currencies))
	for curr := range c.currencies {
		currencies = append(currencies, curr)
	}
	return currencies
}

func (c *Converter) GetRates() []Rate {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rates := make([]Rate, 0, len(c.rates)/2)
	seen := make(map[string]bool)

	for k, rate := range c.rates {
		from, to := parseRateKey(k)
		if seen[fmt.Sprintf("%s->%s", from, to)] || seen[fmt.Sprintf("%s->%s", to, from)] {
			continue
		}
		seen[k] = true
		rates = append(rates, Rate{
			From: from,
			To:   to,
			Rate: rate,
		})
	}
	return rates
}

func parseRateKey(key string) (Currency, Currency) {
	parts := strings.Split(key, "->")
	if len(parts) != 2 {
		return "", ""
	}
	return Currency(parts[0]), Currency(parts[1])
}

func (c *Converter) Convert(amount float64, from, to Currency) (float64, error) {
	if from == to {
		return amount, nil
	}

	c.mu.RLock()
	if _, ok := c.currencies[from]; !ok {
		c.mu.RUnlock()
		return 0, fmt.Errorf("unsupported currency: %s", from)
	}
	if _, ok := c.currencies[to]; !ok {
		c.mu.RUnlock()
		return 0, fmt.Errorf("unsupported currency: %s", to)
	}
	c.mu.RUnlock()

	combinedRate, err := c.getCombinedRate(from, to)
	if err != nil {
		return 0, err
	}

	result := amount * combinedRate
	return c.roundToDecimals(result, to)
}

func (c *Converter) getCombinedRate(from, to Currency) (float64, error) {
	key := fmt.Sprintf("%s->%s", from, to)

	c.cacheMu.RLock()
	if c.hasCache {
		if rate, ok := c.rateCache[key]; ok {
			c.cacheMu.RUnlock()
			return rate, nil
		}
	}
	c.cacheMu.RUnlock()

	c.mu.RLock()
	graph := c.graph
	c.mu.RUnlock()

	rate, err := c.bfsFindRate(graph, from, to)
	if err != nil {
		return 0, err
	}

	c.cacheMu.Lock()
	c.rateCache[key] = rate
	c.cacheMu.Unlock()

	return rate, nil
}

func (c *Converter) bfsFindRate(graph map[Currency][]edge, from, to Currency) (float64, error) {
	type node struct {
		currency Currency
		rate     float64
	}

	queue := []node{{currency: from, rate: 1.0}}
	visited := make(map[Currency]bool)
	visited[from] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.currency == to {
			return current.rate, nil
		}

		for _, e := range graph[current.currency] {
			if !visited[e.to] {
				visited[e.to] = true
				queue = append(queue, node{
					currency: e.to,
					rate:     current.rate * e.rate,
				})
			}
		}
	}

	return 0, fmt.Errorf("no conversion path found from %s to %s", from, to)
}

func (c *Converter) roundToDecimals(amount float64, currency Currency) (float64, error) {
	decimals, err := GetDecimals(currency)
	if err != nil {
		return 0, err
	}

	shift := math.Pow(10, float64(decimals))
	rounded := math.Round(amount*shift) / shift
	return rounded, nil
}

func (c *Converter) clearCacheLocked() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.rateCache = make(map[string]float64)
	c.hasCache = false
}

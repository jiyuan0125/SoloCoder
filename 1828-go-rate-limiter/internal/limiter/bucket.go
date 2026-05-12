package limiter

import (
	"log"
	"sync"
	"time"
)

type Bucket struct {
	mu           sync.Mutex
	capacity     float64
	rate         float64
	waitTimeout  time.Duration
	tokens       float64
	lastRefill   time.Time
	hasTokens    bool
	queueCount   int
	totalQueued  int64
	totalRejected int64
	notifyCh     chan struct{}
	deleted      bool
}

type BucketConfig struct {
	Path        string
	Capacity    float64
	Rate        float64
	WaitTimeout time.Duration
}

type BucketStats struct {
	Path           string
	Capacity       float64
	Rate           float64
	WaitTimeout    time.Duration
	Remaining      float64
	LastRefill     time.Time
	CurrentQueue   int
	TotalQueued    int64
	TotalRejected  int64
}

func NewBucket(cfg *BucketConfig) *Bucket {
	now := time.Now()
	b := &Bucket{
		capacity:     cfg.Capacity,
		rate:         cfg.Rate,
		waitTimeout:  cfg.WaitTimeout,
		tokens:       cfg.Capacity,
		lastRefill:   now,
		hasTokens:    true,
		notifyCh:     make(chan struct{}, 1),
	}
	log.Printf("[%s] Bucket created: capacity=%.2f, rate=%.2f/s, timeout=%s",
		cfg.Path, cfg.Capacity, cfg.Rate, cfg.WaitTimeout)
	return b
}

func (b *Bucket) refill(now time.Time) {
	if b.tokens >= b.capacity {
		return
	}
	delta := now.Sub(b.lastRefill).Seconds()
	if delta > 0 {
		b.tokens += delta * b.rate
		if b.tokens > b.capacity {
			b.tokens = b.capacity
		}
	}
	b.lastRefill = now
}

func (b *Bucket) tryConsume(now time.Time) bool {
	b.refill(now)
	if b.tokens >= 1 {
		b.tokens -= 1
		if !b.hasTokens && b.tokens >= 1 {
			b.hasTokens = true
			log.Printf("[bucket] Bucket recovered, has tokens again, remaining=%.2f", b.tokens)
		}
		if b.tokens < 1 {
			if b.hasTokens {
				b.hasTokens = false
				log.Printf("[bucket] Bucket exhausted, no tokens left")
			}
		}
		return true
	}
	return false
}

func (b *Bucket) Consume() (bool, time.Duration) {
	start := time.Now()
	b.mu.Lock()

	if b.deleted {
		b.mu.Unlock()
		return true, 0
	}

	if b.tryConsume(start) {
		b.mu.Unlock()
		return true, 0
	}

	b.queueCount++
	b.totalQueued++
	queueNum := b.queueCount
	log.Printf("[bucket] Queueing request, current queue=%d", queueNum)
	notify := b.notifyCh
	b.mu.Unlock()

	timeout := time.After(b.waitTimeout)

	for {
		select {
		case <-notify:
			b.mu.Lock()
			if b.deleted {
				b.mu.Unlock()
				waitTime := time.Since(start)
				log.Printf("[bucket] Request passed after bucket deletion, waited=%s", waitTime)
				return true, waitTime
			}
			if b.tryConsume(time.Now()) {
				b.queueCount--
				waitTime := time.Since(start)
				log.Printf("[bucket] Request passed after wait, queue=%d, waited=%s",
					b.queueCount, waitTime)
				b.mu.Unlock()
				return true, waitTime
			}
			b.mu.Unlock()
		case <-timeout:
			b.mu.Lock()
			b.queueCount--
			b.totalRejected++
			waitTime := time.Since(start)
			log.Printf("[bucket] Request rejected by timeout, queue=%d, waited=%s",
				b.queueCount, waitTime)
			b.mu.Unlock()
			return false, waitTime
		}
	}
}

func (b *Bucket) Notify() {
	select {
	case b.notifyCh <- struct{}{}:
	default:
	}
}

func (b *Bucket) Reset(cfg *BucketConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.capacity = cfg.Capacity
	b.rate = cfg.Rate
	b.waitTimeout = cfg.WaitTimeout
	b.tokens = cfg.Capacity
	b.lastRefill = time.Now()
	b.hasTokens = true
	log.Printf("[%s] Bucket reset: capacity=%.2f, rate=%.2f/s, timeout=%s",
		cfg.Path, cfg.Capacity, cfg.Rate, cfg.WaitTimeout)
}

func (b *Bucket) MarkDeleted() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.deleted = true
	log.Printf("[bucket] Bucket deleted, queue=%d will be released", b.queueCount)
}

func (b *Bucket) Stats(path string) *BucketStats {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	b.refill(now)

	return &BucketStats{
		Path:          path,
		Capacity:      b.capacity,
		Rate:          b.rate,
		WaitTimeout:   b.waitTimeout,
		Remaining:     b.tokens,
		LastRefill:    b.lastRefill,
		CurrentQueue:  b.queueCount,
		TotalQueued:   b.totalQueued,
		TotalRejected: b.totalRejected,
	}
}

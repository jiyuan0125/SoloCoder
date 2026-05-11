package limiter

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type LimitMode string

const (
	ModeWait  LimitMode = "wait"
	ModeFail  LimitMode = "fail"
)

type keySemaphore struct {
	sem     *Semaphore
	lastUse int64
	refs    int32
	deleted bool
	mu      sync.Mutex
}

type Config struct {
	GlobalLimit int
	KeyLimit    int
	Mode        LimitMode
	QueueSize   int
	IdleTimeout time.Duration
	CleanupInterval time.Duration
}

type KeyStats struct {
	Key       string
	InUse     int
	Capacity  int
	Available int
	Queued    int64
	Rejected  int64
}

type Stats struct {
	GlobalInUse     int
	GlobalCapacity  int
	GlobalAvailable int
	GlobalQueued    int64
	GlobalRejected  int64
	KeyStats        map[string]KeyStats
	Mode            LimitMode
}

type Limiter struct {
	globalSem   *Semaphore
	keySems     sync.Map
	config      atomic.Value
	queued      atomic.Int64
	rejected    atomic.Int64
	keyRejected sync.Map
	stopped     atomic.Bool
	stopChan    chan struct{}
}

func DefaultConfig() Config {
	return Config{
		GlobalLimit:     100,
		KeyLimit:        10,
		Mode:            ModeFail,
		QueueSize:       100,
		IdleTimeout:     5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
	}
}

func NewLimiter(config Config) *Limiter {
	if config.GlobalLimit < 1 {
		config.GlobalLimit = 1
	}
	if config.KeyLimit < 1 {
		config.KeyLimit = 1
	}
	if config.QueueSize < 0 {
		config.QueueSize = 0
	}
	if config.IdleTimeout < 1*time.Second {
		config.IdleTimeout = 1 * time.Second
	}
	if config.CleanupInterval < 1*time.Second {
		config.CleanupInterval = 1 * time.Second
	}
	
	l := &Limiter{
		globalSem: NewSemaphore(config.GlobalLimit),
		stopChan:  make(chan struct{}),
	}
	l.config.Store(config)
	
	go l.cleanupLoop()
	
	return l
}

func (l *Limiter) getConfig() Config {
	return l.config.Load().(Config)
}

func (l *Limiter) getOrCreateKeySem(key string) (*keySemaphore, func()) {
	actual, _ := l.keySems.LoadOrStore(key, &keySemaphore{
		sem:     NewSemaphore(l.getConfig().KeyLimit),
		lastUse: time.Now().UnixNano(),
	})
	ks := actual.(*keySemaphore)
	
	ks.mu.Lock()
	if ks.deleted {
		ks.mu.Unlock()
		return l.getOrCreateKeySem(key)
	}
	atomic.AddInt32(&ks.refs, 1)
	atomic.StoreInt64(&ks.lastUse, time.Now().UnixNano())
	ks.mu.Unlock()
	
	return ks, func() {
		atomic.AddInt32(&ks.refs, -1)
		atomic.StoreInt64(&ks.lastUse, time.Now().UnixNano())
	}
}

func (l *Limiter) tryDeleteKey(key string, ks *keySemaphore) bool {
	ks.mu.Lock()
	defer ks.mu.Unlock()
	
	if ks.deleted {
		return false
	}
	
	refs := atomic.LoadInt32(&ks.refs)
	if refs != 0 {
		return false
	}
	
	now := time.Now().UnixNano()
	lastUse := atomic.LoadInt64(&ks.lastUse)
	config := l.getConfig()
	
	if time.Duration(now-lastUse) < config.IdleTimeout {
		return false
	}
	
	ks.deleted = true
	l.keySems.Delete(key)
	return true
}

func (l *Limiter) cleanupLoop() {
	ticker := time.NewTicker(l.getConfig().CleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			l.keySems.Range(func(key, value interface{}) bool {
				k := key.(string)
				ks := value.(*keySemaphore)
				l.tryDeleteKey(k, ks)
				return true
			})
		case <-l.stopChan:
			return
		}
	}
}

func (l *Limiter) Acquire(ctx context.Context, key string) (bool, func()) {
	config := l.getConfig()
	
	blocking := config.Mode == ModeWait
	if blocking {
		l.queued.Add(1)
		defer l.queued.Add(-1)
	}
	
	ks, releaseKey := l.getOrCreateKeySem(key)
	defer releaseKey()
	
	if blocking {
		globalAcquired := make(chan bool, 1)
		keyAcquired := make(chan bool, 1)
		globalDone := make(chan struct{})
		keyDone := make(chan struct{})
		
		go func() {
			defer close(globalDone)
			select {
			case <-ctx.Done():
				return
			case <-l.globalSem.tokens:
				globalAcquired <- true
			}
		}()
		
		go func() {
			defer close(keyDone)
			select {
			case <-ctx.Done():
				return
			case <-ks.sem.tokens:
				keyAcquired <- true
			}
		}()
		
		var gotGlobal, gotKey bool
		
		for !gotGlobal || !gotKey {
			select {
			case <-ctx.Done():
				if gotGlobal {
					l.globalSem.Release()
				}
				if gotKey {
					ks.sem.Release()
				}
				return false, func() {}
			case <-globalAcquired:
				gotGlobal = true
			case <-keyAcquired:
				gotKey = true
			}
		}
		
		return true, func() {
			l.globalSem.Release()
			ks.sem.Release()
		}
	} else {
		if !l.globalSem.Acquire(ctx, false) {
			l.rejected.Add(1)
			val, _ := l.keyRejected.LoadOrStore(key, &atomic.Int64{})
			val.(*atomic.Int64).Add(1)
			return false, func() {}
		}
		
		if !ks.sem.Acquire(ctx, false) {
			l.globalSem.Release()
			l.rejected.Add(1)
			val, _ := l.keyRejected.LoadOrStore(key, &atomic.Int64{})
			val.(*atomic.Int64).Add(1)
			return false, func() {}
		}
		
		return true, func() {
			l.globalSem.Release()
			ks.sem.Release()
		}
	}
}

func (l *Limiter) SetGlobalLimit(limit int) {
	if limit < 1 {
		limit = 1
	}
	config := l.getConfig()
	config.GlobalLimit = limit
	l.config.Store(config)
	l.globalSem.SetCapacity(limit)
}

func (l *Limiter) SetKeyLimit(key string, limit int) {
	if limit < 1 {
		limit = 1
	}
	config := l.getConfig()
	config.KeyLimit = limit
	l.config.Store(config)
	
	if value, ok := l.keySems.Load(key); ok {
		value.(*keySemaphore).sem.SetCapacity(limit)
	}
}

func (l *Limiter) SetDefaultKeyLimit(limit int) {
	if limit < 1 {
		limit = 1
	}
	config := l.getConfig()
	config.KeyLimit = limit
	l.config.Store(config)
}

func (l *Limiter) SetMode(mode LimitMode) {
	config := l.getConfig()
	config.Mode = mode
	l.config.Store(config)
}

func (l *Limiter) Stats() Stats {
	config := l.getConfig()
	keyStats := make(map[string]KeyStats)
	
	l.keySems.Range(func(key, value interface{}) bool {
		k := key.(string)
		ks := value.(*keySemaphore)
		
		rejected := int64(0)
		if val, ok := l.keyRejected.Load(k); ok {
			rejected = val.(*atomic.Int64).Load()
		}
		
		keyStats[k] = KeyStats{
			Key:       k,
			InUse:     ks.sem.InUse(),
			Capacity:  ks.sem.Capacity(),
			Available: ks.sem.Available(),
			Queued:    0,
			Rejected:  rejected,
		}
		return true
	})
	
	return Stats{
		GlobalInUse:     l.globalSem.InUse(),
		GlobalCapacity:  l.globalSem.Capacity(),
		GlobalAvailable: l.globalSem.Available(),
		GlobalQueued:    l.queued.Load(),
		GlobalRejected:  l.rejected.Load(),
		KeyStats:        keyStats,
		Mode:            config.Mode,
	}
}

func (l *Limiter) Config() Config {
	return l.getConfig()
}

func (l *Limiter) Stop() {
	if l.stopped.CompareAndSwap(false, true) {
		close(l.stopChan)
	}
}

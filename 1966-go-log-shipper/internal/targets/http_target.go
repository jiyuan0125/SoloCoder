package targets

import (
	"bytes"
	"container/list"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"log-shipper/internal/types"
)

type pendingBatch struct {
	entries []*types.LogEntry
	retries int
}

type HTTPTarget struct {
	id     string
	config types.HTTPTargetConfig

	buffer      []*types.LogEntry
	bufferMu    sync.Mutex
	flushTimer  *time.Timer

	retryQueue   *list.List
	retryQueueMu sync.Mutex

	lastWriteTime time.Time
	successCount  int64
	failureCount  int64
	writeCount    int64
	retryCount    int64
	startTime     time.Time

	stopCh    chan struct{}
	running   bool
	runningMu sync.Mutex
	client    *http.Client
}

func NewHTTPTarget(id string, config types.HTTPTargetConfig) *HTTPTarget {
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}
	if config.FlushTimeout <= 0 {
		config.FlushTimeout = 5 * time.Second
	}
	if config.Retries <= 0 {
		config.Retries = 3
	}
	if config.RetryDelay <= 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.Timeout <= 0 {
		config.Timeout = 10 * time.Second
	}

	return &HTTPTarget{
		id:         id,
		config:     config,
		buffer:     make([]*types.LogEntry, 0, config.BatchSize),
		retryQueue: list.New(),
		stopCh:     make(chan struct{}),
		client: &http.Client{
			Timeout: config.Timeout,
		},
		startTime: time.Now(),
	}
}

func (t *HTTPTarget) ID() string { return t.id }
func (t *HTTPTarget) Type() string { return "http" }
func (t *HTTPTarget) Config() interface{} { return t.config }

func (t *HTTPTarget) Status() types.TargetStatus {
	elapsed := time.Since(t.startTime).Seconds()
	var rate float64
	if elapsed > 0 {
		rate = float64(atomic.LoadInt64(&t.writeCount)) / elapsed
	}

	t.retryQueueMu.Lock()
	pending := int64(t.retryQueue.Len())
	t.retryQueueMu.Unlock()

	return types.TargetStatus{
		ID:                t.id,
		Type:              "http",
		Config:            t.config,
		LastWriteTime:     t.lastWriteTime,
		SuccessCount:      atomic.LoadInt64(&t.successCount),
		FailureCount:      atomic.LoadInt64(&t.failureCount),
		PendingRetryCount: pending,
		WriteRate:         rate,
	}
}

func (t *HTTPTarget) Start() error {
	t.runningMu.Lock()
	defer t.runningMu.Unlock()
	if t.running {
		return nil
	}
	t.running = true

	t.flushTimer = time.NewTimer(t.config.FlushTimeout)

	go t.flushLoop()
	go t.retryLoop()

	return nil
}

func (t *HTTPTarget) Stop() error {
	t.runningMu.Lock()
	if !t.running {
		t.runningMu.Unlock()
		return nil
	}
	t.running = false
	t.runningMu.Unlock()

	close(t.stopCh)

	if t.flushTimer != nil {
		t.flushTimer.Stop()
	}

	t.flush()
	return nil
}

func (t *HTTPTarget) Write(entries []*types.LogEntry) error {
	t.bufferMu.Lock()
	defer t.bufferMu.Unlock()

	for _, entry := range entries {
		t.buffer = append(t.buffer, entry)
		atomic.AddInt64(&t.writeCount, 1)
	}

	if len(t.buffer) >= t.config.BatchSize {
		if t.flushTimer != nil {
			t.flushTimer.Reset(t.config.FlushTimeout)
		}
		go t.flush()
	}

	return nil
}

func (t *HTTPTarget) flushLoop() {
	for {
		select {
		case <-t.stopCh:
			return
		case <-t.flushTimer.C:
			t.flush()
			t.flushTimer.Reset(t.config.FlushTimeout)
		}
	}
}

func (t *HTTPTarget) flush() {
	t.bufferMu.Lock()
	if len(t.buffer) == 0 {
		t.bufferMu.Unlock()
		return
	}

	batch := t.buffer
	t.buffer = make([]*types.LogEntry, 0, t.config.BatchSize)
	t.bufferMu.Unlock()

	if err := t.sendBatchWithRetry(batch, 0); err != nil {
		t.retryQueueMu.Lock()
		t.retryQueue.PushBack(&pendingBatch{entries: batch, retries: t.config.Retries})
		t.retryQueueMu.Unlock()
	}
}

func (t *HTTPTarget) sendBatchWithRetry(entries []*types.LogEntry, additionalRetries int) error {
	maxRetries := t.config.Retries + additionalRetries
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := t.sendBatch(entries); err != nil {
			lastErr = err
			if attempt < maxRetries-1 {
				delay := t.config.RetryDelay * time.Duration(1<<uint(attempt))
				log.Printf("[WARN] http target %s batch send attempt %d/%d failed: %v, retrying in %v",
					t.id, attempt+1, maxRetries, err, delay)
				time.Sleep(delay)
			}
			continue
		}
		t.lastWriteTime = time.Now()
		atomic.AddInt64(&t.successCount, 1)
		return nil
	}

	log.Printf("[ERROR] http target %s batch send failed after %d retries: %v, queued for retry",
		t.id, maxRetries, lastErr)
	atomic.AddInt64(&t.failureCount, 1)
	return lastErr
}

func (t *HTTPTarget) sendBatch(entries []*types.LogEntry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", t.config.URL, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return &HTTPError{StatusCode: resp.StatusCode, URL: t.config.URL}
}

func (t *HTTPTarget) retryLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.processRetryQueue()
		}
	}
}

func (t *HTTPTarget) processRetryQueue() {
	t.retryQueueMu.Lock()
	if t.retryQueue.Len() == 0 {
		t.retryQueueMu.Unlock()
		return
	}

	var items []*pendingBatch
	for e := t.retryQueue.Front(); e != nil; e = e.Next() {
		items = append(items, e.Value.(*pendingBatch))
	}
	t.retryQueue.Init()
	t.retryQueueMu.Unlock()

	for _, batch := range items {
		atomic.AddInt64(&t.retryCount, 1)
		if err := t.sendBatchWithRetry(batch.entries, 0); err != nil {
			t.retryQueueMu.Lock()
			t.retryQueue.PushBack(batch)
			t.retryQueueMu.Unlock()
		}
	}
}

type HTTPError struct {
	StatusCode int
	URL        string
}

func (e *HTTPError) Error() string {
	return e.URL + ": HTTP error " + http.StatusText(e.StatusCode)
}

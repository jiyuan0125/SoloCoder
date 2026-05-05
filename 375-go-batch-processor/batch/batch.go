package batch

import (
	"log"
	"sync"
	"time"
)

type BatchHandler[T any] interface {
	ProcessBatch(items []T) error
}

type BatchProcessorConfig struct {
	BatchSize    int
	Timeout      time.Duration
	CloseTimeout time.Duration
}

type BatchProcessor[T any] struct {
	handler      BatchHandler[T]
	buffer       []T
	mu           sync.Mutex
	batchSize    int
	timeout      time.Duration
	closeTimeout time.Duration

	timerMu      sync.Mutex
	timer        *time.Timer
	timerRunning bool

	stopChan    chan struct{}
	processChan chan struct{}

	stopped bool
	wg      sync.WaitGroup
}

const (
	DefaultBatchSize    = 1
	DefaultTimeout      = 5 * time.Second
	DefaultCloseTimeout = 10 * time.Second
)

func NewBatchProcessor[T any](handler BatchHandler[T], config BatchProcessorConfig) *BatchProcessor[T] {
	batchSize := config.BatchSize
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	timeout := config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	closeTimeout := config.CloseTimeout
	if closeTimeout <= 0 {
		closeTimeout = DefaultCloseTimeout
	}

	bp := &BatchProcessor[T]{
		handler:      handler,
		buffer:       make([]T, 0, batchSize),
		batchSize:    batchSize,
		timeout:      timeout,
		closeTimeout: closeTimeout,
		stopChan:     make(chan struct{}),
		processChan:  make(chan struct{}, 1),
		stopped:      false,
	}

	bp.wg.Add(1)
	go bp.run()

	return bp
}

func (bp *BatchProcessor[T]) Add(item T) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.stopped {
		return
	}

	bp.buffer = append(bp.buffer, item)

	if len(bp.buffer) >= bp.batchSize {
		bp.signalProcess()
	}
}

func (bp *BatchProcessor[T]) AddAll(items []T) {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.stopped {
		return
	}

	bp.buffer = append(bp.buffer, items...)

	if len(bp.buffer) >= bp.batchSize {
		bp.signalProcess()
	}
}

func (bp *BatchProcessor[T]) Flush() {
	bp.mu.Lock()
	defer bp.mu.Unlock()

	if bp.stopped {
		return
	}

	if len(bp.buffer) > 0 {
		bp.signalProcess()
	}
}

func (bp *BatchProcessor[T]) Size() int {
	bp.mu.Lock()
	defer bp.mu.Unlock()
	return len(bp.buffer)
}

func (bp *BatchProcessor[T]) Close() {
	bp.mu.Lock()
	if bp.stopped {
		bp.mu.Unlock()
		return
	}
	bp.stopped = true
	bp.mu.Unlock()

	close(bp.stopChan)

	done := make(chan struct{})
	go func() {
		bp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(bp.closeTimeout):
		log.Printf("BatchProcessor.Close() timed out after %v", bp.closeTimeout)
	}
}

func (bp *BatchProcessor[T]) signalProcess() {
	select {
	case bp.processChan <- struct{}{}:
	default:
	}
}

func (bp *BatchProcessor[T]) stopTimer() {
	bp.timerMu.Lock()
	defer bp.timerMu.Unlock()

	if bp.timerRunning && bp.timer != nil {
		bp.timer.Stop()
		bp.timerRunning = false
	}
}

func (bp *BatchProcessor[T]) startTimer() {
	bp.timerMu.Lock()
	defer bp.timerMu.Unlock()

	if bp.timerRunning {
		return
	}

	if bp.timer == nil {
		bp.timer = time.NewTimer(bp.timeout)
	} else {
		bp.timer.Reset(bp.timeout)
	}
	bp.timerRunning = true
}

func (bp *BatchProcessor[T]) run() {
	defer bp.wg.Done()

	bp.startTimer()

	for {
		select {
		case <-bp.stopChan:
			bp.stopTimer()
			bp.processRemaining()
			return

		case <-bp.processChan:
			bp.processBatch()

		case <-bp.timerChannel():
			bp.timerMu.Lock()
			bp.timerRunning = false
			bp.timerMu.Unlock()

			bp.processOnTimeout()

			bp.mu.Lock()
			hasData := len(bp.buffer) > 0
			bp.mu.Unlock()

			if hasData {
				bp.startTimer()
			}
		}
	}
}

func (bp *BatchProcessor[T]) timerChannel() <-chan time.Time {
	bp.timerMu.Lock()
	defer bp.timerMu.Unlock()

	if bp.timer != nil {
		return bp.timer.C
	}
	return nil
}

func (bp *BatchProcessor[T]) processIfNeeded() {
	bp.mu.Lock()
	if len(bp.buffer) < bp.batchSize {
		bp.mu.Unlock()
		return
	}
	bp.mu.Unlock()

	bp.processBatch()
}

func (bp *BatchProcessor[T]) processOnTimeout() {
	bp.mu.Lock()
	if len(bp.buffer) == 0 {
		bp.mu.Unlock()
		return
	}
	bp.mu.Unlock()

	bp.processBatch()
}

func (bp *BatchProcessor[T]) processRemaining() {
	for {
		bp.mu.Lock()
		hasData := len(bp.buffer) > 0
		bp.mu.Unlock()

		if !hasData {
			return
		}

		bp.processBatch()
	}
}

func (bp *BatchProcessor[T]) processBatch() {
	bp.mu.Lock()
	if len(bp.buffer) == 0 {
		bp.mu.Unlock()
		return
	}

	itemsToProcess := make([]T, len(bp.buffer))
	copy(itemsToProcess, bp.buffer)
	bp.mu.Unlock()

	err := bp.safeProcess(itemsToProcess)

	bp.mu.Lock()
	if err != nil {
		log.Printf("Batch processing failed, keeping %d items for retry: %v", len(itemsToProcess), err)
		bp.mu.Unlock()
	} else {
		bp.buffer = bp.buffer[len(itemsToProcess):]
		if len(bp.buffer) > 0 {
			newBuffer := make([]T, len(bp.buffer))
			copy(newBuffer, bp.buffer)
			bp.buffer = newBuffer
		}
		bp.mu.Unlock()

		bp.stopTimer()
		bp.mu.Lock()
		hasData := len(bp.buffer) > 0
		bp.mu.Unlock()
		if hasData {
			bp.startTimer()
		}
	}
}

func (bp *BatchProcessor[T]) safeProcess(items []T) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("BatchHandler.ProcessBatch panicked: %v", r)
			err = recoverError(r)
		}
	}()

	return bp.handler.ProcessBatch(items)
}

func recoverError(r interface{}) error {
	if err, ok := r.(error); ok {
		return err
	}
	return &panicError{r}
}

type panicError struct {
	v interface{}
}

func (e *panicError) Error() string {
	return "panic occurred"
}

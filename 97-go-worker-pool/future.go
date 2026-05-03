package workerpool

import (
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"
)

var (
	ErrPoolClosed  = errors.New("pool is closed")
	ErrDropped     = errors.New("task dropped")
	ErrFutureRead  = errors.New("future result already read")
)

type TimeoutError struct {
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("task timeout after %v", e.Timeout)
}

type PanicError struct {
	Value interface{}
	Stack []byte
}

func (e *PanicError) Error() string {
	return fmt.Sprintf("task panicked: %v\n%s", e.Value, e.Stack)
}

type Future interface {
	Result() (error, error)
	ResultWithTimeout(timeout time.Duration) (error, error)
}

type future struct {
	mu        sync.Mutex
	done      chan struct{}
	resultErr error
	panicErr  error
	read      bool
}

func newFuture() *future {
	return &future{
		done: make(chan struct{}),
	}
}

func (f *future) complete(resultErr error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-f.done:
		return
	default:
		f.resultErr = resultErr
		close(f.done)
	}
}

func (f *future) completeWithPanic(panicVal interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-f.done:
		return
	default:
		f.panicErr = &PanicError{
			Value: panicVal,
			Stack: debug.Stack(),
		}
		close(f.done)
	}
}

func (f *future) Result() (error, error) {
	f.mu.Lock()
	if f.read {
		f.mu.Unlock()
		return nil, ErrFutureRead
	}
	done := f.done
	f.mu.Unlock()

	<-done

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.read {
		return nil, ErrFutureRead
	}
	f.read = true
	return f.resultErr, f.panicErr
}

func (f *future) ResultWithTimeout(timeout time.Duration) (error, error) {
	f.mu.Lock()
	if f.read {
		f.mu.Unlock()
		return nil, ErrFutureRead
	}
	done := f.done
	f.mu.Unlock()

	select {
	case <-done:
		f.mu.Lock()
		defer f.mu.Unlock()
		if f.read {
			return nil, ErrFutureRead
		}
		f.read = true
		return f.resultErr, f.panicErr
	case <-time.After(timeout):
		return nil, &TimeoutError{Timeout: timeout}
	}
}

type task struct {
	fn      func() error
	future  *future
	timeout time.Duration
}

func newTask(fn func() error, timeout time.Duration) *task {
	return &task{
		fn:      fn,
		future:  newFuture(),
		timeout: timeout,
	}
}

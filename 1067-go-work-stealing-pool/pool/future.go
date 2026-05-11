package pool

import (
	"sync"
	"time"
)

type Future struct {
	result interface{}
	err    error
	done   chan struct{}
	once   sync.Once
}

func newFuture() *Future {
	return &Future{
		done: make(chan struct{}),
	}
}

func (f *Future) complete(res interface{}, err error) {
	f.once.Do(func() {
		f.result = res
		f.err = err
		close(f.done)
	})
}

func (f *Future) Get() (interface{}, error) {
	<-f.done
	return f.result, f.err
}

func (f *Future) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	select {
	case <-f.done:
		return f.result, f.err
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (f *Future) Done() <-chan struct{} {
	return f.done
}

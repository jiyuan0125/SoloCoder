package promise

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type State int32

const (
	Pending State = iota
	Resolved
	Rejected
	Cancelled
)

type Result[T any] struct {
	Value T
	Err   error
}

type Future[T any] struct {
	mu        sync.Mutex
	state     atomic.Int32
	result    Result[T]
	done      chan struct{}
	callbacks []func(Result[T])
	errCbs    []func(error)
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewFuture[T any]() *Future[T] {
	ctx, cancel := context.WithCancel(context.Background())
	f := &Future[T]{
		done:   make(chan struct{}),
		ctx:    ctx,
		cancel: cancel,
	}
	f.state.Store(int32(Pending))
	return f
}

func (f *Future[T]) Resolve(value T) bool {
	if f.state.CompareAndSwap(int32(Pending), int32(Resolved)) {
		f.mu.Lock()
		f.result = Result[T]{Value: value}
		callbacks := f.callbacks
		f.callbacks = nil
		f.mu.Unlock()
		close(f.done)
		for _, cb := range callbacks {
			go cb(f.result)
		}
		return true
	}
	return false
}

func (f *Future[T]) Reject(err error) bool {
	if f.state.CompareAndSwap(int32(Pending), int32(Rejected)) {
		f.mu.Lock()
		f.result = Result[T]{Err: err}
		errCbs := f.errCbs
		f.errCbs = nil
		callbacks := f.callbacks
		f.callbacks = nil
		f.mu.Unlock()
		close(f.done)
		for _, cb := range errCbs {
			go cb(err)
		}
		for _, cb := range callbacks {
			go cb(f.result)
		}
		return true
	}
	return false
}

func (f *Future[T]) Cancel() bool {
	if f.state.CompareAndSwap(int32(Pending), int32(Cancelled)) {
		f.cancel()
		f.mu.Lock()
		f.result = Result[T]{Err: errors.New("cancelled")}
		errCbs := f.errCbs
		f.errCbs = nil
		callbacks := f.callbacks
		f.callbacks = nil
		f.mu.Unlock()
		close(f.done)
		for _, cb := range errCbs {
			go cb(errors.New("cancelled"))
		}
		for _, cb := range callbacks {
			go cb(f.result)
		}
		return true
	}
	return false
}

func (f *Future[T]) State() State {
	return State(f.state.Load())
}

func (f *Future[T]) IsDone() bool {
	s := f.State()
	return s == Resolved || s == Rejected || s == Cancelled
}

func (f *Future[T]) Context() context.Context {
	return f.ctx
}

func (f *Future[T]) Then(fn func(T)) *Future[T] {
	if f.IsDone() {
		if f.State() == Resolved {
			go fn(f.result.Value)
		}
		return f
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.callbacks = append(f.callbacks, func(r Result[T]) {
		if r.Err == nil {
			fn(r.Value)
		}
	})
	return f
}

func (f *Future[T]) Catch(fn func(error)) *Future[T] {
	if f.IsDone() {
		s := f.State()
		if s == Rejected || s == Cancelled {
			go fn(f.result.Err)
		}
		return f
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.errCbs = append(f.errCbs, fn)
	return f
}

func (f *Future[T]) Await() (T, error) {
	<-f.done
	return f.result.Value, f.result.Err
}

func (f *Future[T]) AwaitContext(ctx context.Context) (T, error) {
	select {
	case <-f.done:
		return f.result.Value, f.result.Err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}

func (f *Future[T]) Timeout(duration time.Duration) *Future[T] {
	if f.IsDone() {
		return f
	}
	go func() {
		timer := time.NewTimer(duration)
		defer timer.Stop()
		select {
		case <-f.done:
			return
		case <-timer.C:
			f.Reject(errors.New("timeout"))
		}
	}()
	return f
}

type Promise[T any] struct {
	future *Future[T]
}

func New[T any](executor func(resolve func(T), reject func(error), cancel func())) *Promise[T] {
	f := NewFuture[T]()
	p := &Promise[T]{future: f}
	resolve := func(v T) { f.Resolve(v) }
	reject := func(err error) { f.Reject(err) }
	cancel := func() { f.Cancel() }
	go executor(resolve, reject, cancel)
	return p
}

func NewWithContext[T any](ctx context.Context, executor func(resolve func(T), reject func(error), cancel func())) *Promise[T] {
	f := NewFuture[T]()
	p := &Promise[T]{future: f}
	resolve := func(v T) { f.Resolve(v) }
	reject := func(err error) { f.Reject(err) }
	cancel := func() { f.Cancel() }
	go func() {
		select {
		case <-ctx.Done():
			f.Reject(ctx.Err())
			return
		default:
			executor(resolve, reject, cancel)
		}
	}()
	return p
}

func (p *Promise[T]) Future() *Future[T] {
	return p.future
}

func (p *Promise[T]) Resolve(value T) bool {
	return p.future.Resolve(value)
}

func (p *Promise[T]) Reject(err error) bool {
	return p.future.Reject(err)
}

func (p *Promise[T]) Cancel() bool {
	return p.future.Cancel()
}

func (p *Promise[T]) State() State {
	return p.future.State()
}

func (p *Promise[T]) IsDone() bool {
	return p.future.IsDone()
}

func (p *Promise[T]) Await() (T, error) {
	return p.future.Await()
}

func (p *Promise[T]) AwaitContext(ctx context.Context) (T, error) {
	return p.future.AwaitContext(ctx)
}

func (p *Promise[T]) Timeout(duration time.Duration) *Future[T] {
	return p.future.Timeout(duration)
}

func (p *Promise[T]) Then(fn func(T)) *Future[T] {
	return p.future.Then(fn)
}

func (p *Promise[T]) Catch(fn func(error)) *Future[T] {
	return p.future.Catch(fn)
}

func Then[T, U any](f *Future[T], fn func(T) U) *Future[U] {
	result := NewFuture[U]()
	f.Then(func(v T) {
		result.Resolve(fn(v))
	})
	f.Catch(func(err error) {
		result.Reject(err)
	})
	return result
}

func ThenAsync[T, U any](f *Future[T], fn func(T) *Future[U]) *Future[U] {
	result := NewFuture[U]()
	f.Then(func(v T) {
		inner := fn(v)
		inner.Then(func(u U) {
			result.Resolve(u)
		})
		inner.Catch(func(err error) {
			result.Reject(err)
		})
	})
	f.Catch(func(err error) {
		result.Reject(err)
	})
	return result
}

func Chain[T any](first func() *Future[T], rest ...func(T) *Future[T]) *Future[T] {
	if len(rest) == 0 {
		return first()
	}
	current := first()
	for _, step := range rest {
		current = ThenAsync(current, step)
	}
	return current
}

package pool

import (
	"time"
)

type Future struct {
	result chan interface{}
	err    chan error
	done   chan struct{}
}

func newFuture() *Future {
	return &Future{
		result: make(chan interface{}, 1),
		err:    make(chan error, 1),
		done:   make(chan struct{}, 1),
	}
}

func (f *Future) complete(res interface{}, err error) {
	select {
	case f.result <- res:
	default:
	}
	select {
	case f.err <- err:
	default:
	}
	select {
	case f.done <- struct{}{}:
	default:
	}
}

func (f *Future) Get() (interface{}, error) {
	<-f.done
	select {
	case res := <-f.result:
		var err error
		select {
		case err = <-f.err:
		default:
		}
		return res, err
	case err := <-f.err:
		return nil, err
	default:
		return nil, nil
	}
}

func (f *Future) GetWithTimeout(timeout time.Duration) (interface{}, error) {
	select {
	case <-f.done:
		select {
		case res := <-f.result:
			var err error
			select {
			case err = <-f.err:
			default:
			}
			return res, err
		case err := <-f.err:
			return nil, err
		default:
			return nil, nil
		}
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (f *Future) Done() <-chan struct{} {
	return f.done
}

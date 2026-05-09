package pool

import (
	"fmt"
	"time"
)

type Task struct {
	id        string
	fn        func()
	future    *Future
	timeout   time.Duration
	startTime time.Time
}

func (t *Task) run(wg *syncWaitGroup) {
	if wg != nil {
		wg.Add(1)
		defer wg.Done()
	}

	done := make(chan struct{}, 1)
	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				select {
				case errChan <- fmt.Errorf("panic: %v", r):
				default:
				}
			}
			close(done)
		}()

		t.fn()
		select {
		case resultChan <- nil:
		default:
		}
	}()

	timeout := t.timeout
	if timeout <= 0 {
		<-done
		select {
		case res := <-resultChan:
			t.future.complete(res, nil)
		case err := <-errChan:
			t.future.complete(nil, err)
		default:
			t.future.complete(nil, nil)
		}
		return
	}

	select {
	case <-done:
		select {
		case res := <-resultChan:
			t.future.complete(res, nil)
		case err := <-errChan:
			t.future.complete(nil, err)
		default:
			t.future.complete(nil, nil)
		}
	case <-time.After(timeout):
		t.future.complete(nil, ErrTimeout)
	}
}

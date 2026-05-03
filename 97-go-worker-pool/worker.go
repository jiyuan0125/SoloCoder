package workerpool

import (
	"fmt"
	"os"
	"sync/atomic"
	"time"
)

type worker struct {
	id         int
	taskChan   <-chan *task
	stopChan   chan struct{}
	pool       *Pool
	isExtra    bool
	idleTimer  *time.Timer
	running    int32
}

func newWorker(id int, taskChan <-chan *task, pool *Pool, isExtra bool) *worker {
	w := &worker{
		id:       id,
		taskChan: taskChan,
		stopChan: make(chan struct{}, 1),
		pool:     pool,
		isExtra:  isExtra,
		running:  1,
	}
	return w
}

func (w *worker) start() {
	go w.run()
}

func (w *worker) run() {
	defer func() {
		atomic.StoreInt32(&w.running, 0)
	}()

	for {
		select {
		case <-w.stopChan:
			return
		default:
		}

		var t *task
		var ok bool

		if w.isExtra {
			if w.idleTimer == nil {
				w.idleTimer = time.NewTimer(w.pool.config.IdleTimeout)
			} else {
				w.idleTimer.Reset(w.pool.config.IdleTimeout)
			}

			select {
			case t, ok = <-w.taskChan:
				if !w.idleTimer.Stop() {
					<-w.idleTimer.C
				}
			case <-w.idleTimer.C:
				w.pool.workerExit(w)
				return
			case <-w.stopChan:
				if w.idleTimer != nil && !w.idleTimer.Stop() {
					select {
					case <-w.idleTimer.C:
					default:
					}
				}
				return
			}
		} else {
			select {
			case t, ok = <-w.taskChan:
			case <-w.stopChan:
				return
			}
		}

		if !ok {
			return
		}

		w.executeTask(t)
	}
}

func (w *worker) executeTask(t *task) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "worker %d panicked: %v\n", w.id, r)
			t.future.completeWithPanic(r)
			w.pool.replaceWorker(w)
		}
	}()

	var resultErr error
	done := make(chan struct{})

	go func() {
		defer close(done)
		resultErr = t.fn()
	}()

	if t.timeout > 0 {
		select {
		case <-done:
			t.future.complete(resultErr)
		case <-time.After(t.timeout):
			t.future.completeWithTimeout(t.timeout)
			go func() {
				<-done
			}()
			return
		}
	} else {
		<-done
		t.future.complete(resultErr)
	}
}

func (w *worker) stop() {
	select {
	case w.stopChan <- struct{}{}:
	default:
	}
}

func (w *worker) isRunning() bool {
	return atomic.LoadInt32(&w.running) == 1
}

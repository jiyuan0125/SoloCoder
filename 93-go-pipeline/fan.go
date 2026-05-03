package pipeline

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
)

func fanOutFanIn(
	ctx context.Context,
	in chan interface{},
	out chan interface{},
	fn TransformFunc,
	workerCount int,
	errCh chan<- error,
) {
	var workerWg sync.WaitGroup
	workerOuts := make([]chan interface{}, workerCount)

	for i := 0; i < workerCount; i++ {
		workerOuts[i] = make(chan interface{}, 1)
	}

	for i := 0; i < workerCount; i++ {
		workerWg.Add(1)
		go func(workerID int, workerOut chan<- interface{}) {
			defer workerWg.Done()
			defer close(workerOut)
			workerLoop(ctx, in, workerOut, fn, workerID, errCh)
		}(i, workerOuts[i])
	}

	var fanInWg sync.WaitGroup
	fanInWg.Add(1)
	go func() {
		defer fanInWg.Done()
		defer close(out)
		fanIn(ctx, workerOuts, out)
	}()

	workerWg.Wait()
	fanInWg.Wait()
}

func workerLoop(
	ctx context.Context,
	in <-chan interface{},
	out chan<- interface{},
	fn TransformFunc,
	workerID int,
	errCh chan<- error,
) {
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			processItemWithRecovery(ctx, item, out, fn, workerID, errCh)
		}
	}
}

func processItemWithRecovery(
	ctx context.Context,
	item interface{},
	out chan<- interface{},
	fn TransformFunc,
	workerID int,
	errCh chan<- error,
) {
	defer func() {
		if r := recover(); r != nil {
			logError("Worker %d panic: %v\n%s", workerID, r, debug.Stack())
			select {
			case errCh <- fmt.Errorf("worker %d panic: %v", workerID, r):
			default:
			}
		}
	}()

	if fn == nil {
		select {
		case <-ctx.Done():
			return
		case out <- item:
		}
		return
	}

	result, err := fn(ctx, item)
	if err != nil {
		logError("Worker %d error: %v", workerID, err)
		select {
		case errCh <- err:
		default:
		}
		return
	}

	select {
	case <-ctx.Done():
		return
	case out <- result:
	}
}

func fanIn(ctx context.Context, workerOuts []chan interface{}, out chan<- interface{}) {
	var wg sync.WaitGroup

	for _, ch := range workerOuts {
		wg.Add(1)
		go func(in <-chan interface{}) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-in:
					if !ok {
						return
					}
					select {
					case <-ctx.Done():
						return
					case out <- item:
					}
				}
			}
		}(ch)
	}

	wg.Wait()
}

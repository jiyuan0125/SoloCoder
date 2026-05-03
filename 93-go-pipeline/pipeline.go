package pipeline

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

type Pipeline struct {
	stages []Stage
	wg     sync.WaitGroup
	errCh  chan error
}

func New() *Pipeline {
	return &Pipeline{
		stages: make([]Stage, 0),
		errCh:  make(chan error, 1),
	}
}

func (p *Pipeline) AddStage(s Stage) {
	p.stages = append(p.stages, s)
}

func (p *Pipeline) Run(ctx context.Context, timeout time.Duration) error {
	if len(p.stages) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var chans []chan interface{}
	chans = make([]chan interface{}, len(p.stages)+1)
	for i := range chans {
		chans[i] = make(chan interface{}, 1)
	}

	for i, stage := range p.stages {
		inCh := chans[i]
		outCh := chans[i+1]

		p.wg.Add(1)
		go func(s Stage, in, out chan interface{}, idx int) {
			defer p.wg.Done()
			s.Run(ctx, in, out, p.errCh)
		}(stage, inCh, outCh, i)
	}

	go func() {
		<-ctx.Done()
		for _, ch := range chans {
			drainChannel(ch)
		}
	}()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	var runErr error
	select {
	case <-ctx.Done():
		runErr = ctx.Err()
	case <-done:
		runErr = nil
	}

	select {
	case err := <-p.errCh:
		if runErr == nil {
			runErr = err
		}
	default:
	}

	return runErr
}

func drainChannel(ch chan interface{}) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}

func logError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

package pipeline

import (
	"context"
	"fmt"
	"runtime/debug"
)

type Stage interface {
	Run(ctx context.Context, in, out chan interface{}, errCh chan<- error)
}

type SourceFunc func(ctx context.Context, out chan<- interface{}) error

type TransformFunc func(ctx context.Context, in interface{}) (interface{}, error)

type SinkFunc func(ctx context.Context, in interface{}) error

type SourceStage struct {
	fn SourceFunc
}

func NewSource(fn SourceFunc) *SourceStage {
	return &SourceStage{fn: fn}
}

func (s *SourceStage) Run(ctx context.Context, in, out chan interface{}, errCh chan<- error) {
	defer close(out)
	defer func() {
		if r := recover(); r != nil {
			logError("SourceStage panic: %v\n%s", r, debug.Stack())
			select {
			case errCh <- fmt.Errorf("source panic: %v", r):
			default:
			}
		}
	}()

	if s.fn != nil {
		if err := s.fn(ctx, out); err != nil {
			logError("SourceStage error: %v", err)
			select {
			case errCh <- err:
			default:
			}
		}
	}
}

type TransformStage struct {
	fn     TransformFunc
	workers int
}

func NewTransform(fn TransformFunc, workers int) *TransformStage {
	if workers < 1 {
		workers = 1
	}
	return &TransformStage{fn: fn, workers: workers}
}

func (t *TransformStage) Run(ctx context.Context, in, out chan interface{}, errCh chan<- error) {
	if t.workers > 1 {
		fanOutFanIn(ctx, in, out, t.fn, t.workers, errCh)
		return
	}

	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			t.processOne(ctx, item, out, errCh)
		}
	}
}

func (t *TransformStage) processOne(ctx context.Context, item interface{}, out chan<- interface{}, errCh chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			logError("TransformStage panic: %v\n%s", r, debug.Stack())
			select {
			case errCh <- fmt.Errorf("transform panic: %v", r):
			default:
			}
		}
	}()

	if t.fn == nil {
		select {
		case <-ctx.Done():
			return
		case out <- item:
		}
		return
	}

	result, err := t.fn(ctx, item)
	if err != nil {
		logError("TransformStage error: %v", err)
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

type SinkStage struct {
	fn SinkFunc
}

func NewSink(fn SinkFunc) *SinkStage {
	return &SinkStage{fn: fn}
}

func (s *SinkStage) Run(ctx context.Context, in, out chan interface{}, errCh chan<- error) {
	defer close(out)
	for {
		select {
		case <-ctx.Done():
			return
		case item, ok := <-in:
			if !ok {
				return
			}
			s.processOne(ctx, item, errCh)
		}
	}
}

func (s *SinkStage) processOne(ctx context.Context, item interface{}, errCh chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			logError("SinkStage panic: %v\n%s", r, debug.Stack())
			select {
			case errCh <- fmt.Errorf("sink panic: %v", r):
			default:
			}
		}
	}()

	if s.fn == nil {
		return
	}

	if err := s.fn(ctx, item); err != nil {
		logError("SinkStage error: %v", err)
		select {
		case errCh <- err:
		default:
		}
	}
}

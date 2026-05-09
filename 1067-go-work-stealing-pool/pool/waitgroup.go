package pool

import "sync"

type syncWaitGroup struct {
	wg sync.WaitGroup
}

func (s *syncWaitGroup) Add(delta int) {
	s.wg.Add(delta)
}

func (s *syncWaitGroup) Done() {
	s.wg.Done()
}

func (s *syncWaitGroup) Wait() {
	s.wg.Wait()
}

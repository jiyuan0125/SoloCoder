package timingwheel

import (
	"sync"
)

type slot struct {
	mu    sync.Mutex
	tasks []*Task
}

func newSlot() *slot {
	return &slot{
		tasks: make([]*Task, 0),
	}
}

func (s *slot) add(t *Task) {
	s.mu.Lock()
	s.tasks = append(s.tasks, t)
	s.mu.Unlock()
}

func (s *slot) removeByID(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, t := range s.tasks {
		if t.ID == id {
			s.tasks = append(s.tasks[:i], s.tasks[i+1:]...)
			return true
		}
	}
	return false
}

func (s *slot) drain() []*Task {
	s.mu.Lock()
	result := s.tasks
	s.tasks = make([]*Task, 0)
	s.mu.Unlock()
	return result
}

func (s *slot) size() int {
	s.mu.Lock()
	n := len(s.tasks)
	s.mu.Unlock()
	return n
}

type Wheel struct {
	name      string
	slots     []*slot
	tick      int
	totalTicks int
	mu        sync.RWMutex
}

func NewWheel(name string, size int) *Wheel {
	slots := make([]*slot, size)
	for i := range slots {
		slots[i] = newSlot()
	}
	return &Wheel{
		name:      name,
		slots:     slots,
		tick:      0,
		totalTicks: size,
	}
}

func (w *Wheel) advance() int {
	w.mu.Lock()
	prev := w.tick
	w.tick = (w.tick + 1) % w.totalTicks
	w.mu.Unlock()
	return prev
}

func (w *Wheel) currentTick() int {
	w.mu.RLock()
	n := w.tick
	w.mu.RUnlock()
	return n
}

func (w *Wheel) size() int {
	return w.totalTicks
}

func (w *Wheel) slot(idx int) *slot {
	return w.slots[idx%w.totalTicks]
}

func (w *Wheel) addTask(t *Task, offset int) {
	idx := (w.currentTick() + offset) % w.totalTicks
	t.wheelLevel = int(wheelLevelByName(w.name))
	t.slotIndex = idx
	w.slot(idx).add(t)
}

func (w *Wheel) drainCurrentTick() []*Task {
	prev := w.advance()
	return w.slot(prev).drain()
}

func (w *Wheel) totalTasks() int {
	total := 0
	for _, s := range w.slots {
		total += s.size()
	}
	return total
}

func (w *Wheel) distribution() map[int]int {
	dist := make(map[int]int)
	for i, s := range w.slots {
		if n := s.size(); n > 0 {
			dist[i] = n
		}
	}
	return dist
}

func wheelLevelByName(name string) WheelLevel {
	switch name {
	case "hour":
		return WheelLevelHour
	case "minute":
		return WheelLevelMinute
	case "second":
		return WheelLevelSecond
	default:
		return WheelLevelSecond
	}
}

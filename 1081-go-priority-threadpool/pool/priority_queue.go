package pool

import (
	"container/heap"
	"time"
)

type Priority int

const (
	Low Priority = iota
	Medium
	High
)

type Task struct {
	id         uint64
	priority   Priority
	submitTime time.Time
	isStarved  bool
	fn         func()
}

type priorityQueue []*Task

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	if pq[i].isStarved {
		return true
	}
	if pq[j].isStarved {
		return false
	}
	return pq[i].priority > pq[j].priority
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	item := x.(*Task)
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*pq = old[0 : n-1]
	return item
}

func (pq *priorityQueue) Peek() *Task {
	if pq.Len() == 0 {
		return nil
	}
	return (*pq)[0]
}

func (pq *priorityQueue) checkAndPromote(starvationThreshold time.Duration) int {
	count := 0
	now := time.Now()
	for _, task := range *pq {
		if !task.isStarved && task.priority == Low {
			if now.Sub(task.submitTime) >= starvationThreshold {
				task.isStarved = true
				count++
			}
		}
	}
	if count > 0 {
		heap.Init(pq)
	}
	return count
}

func (pq *priorityQueue) countByPriority() (high, medium, low int) {
	for _, task := range *pq {
		if task.isStarved {
			high++
		} else {
			switch task.priority {
			case High:
				high++
			case Medium:
				medium++
			case Low:
				low++
			}
		}
	}
	return
}

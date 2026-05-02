package taskqueue

type worker struct {
	id     int
	queue  *TaskQueue
	stopCh chan struct{}
}

func newWorker(queue *TaskQueue, id int) *worker {
	return &worker{
		id:     id,
		queue:  queue,
		stopCh: make(chan struct{}),
	}
}

func (w *worker) start() {
	for {
		task, ok := w.queue.getTask()
		if !ok {
			w.queue.workerStopped()
			return
		}

		w.executeTask(task)
	}
}

func (w *worker) executeTask(task *Task) {
	if task.Handler == nil {
		w.queue.taskCompleted(task)
		return
	}

	err := task.Handler()
	if err != nil {
		w.queue.taskFailed(task, err)
	} else {
		w.queue.taskCompleted(task)
	}
}

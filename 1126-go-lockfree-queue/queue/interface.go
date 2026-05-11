package queue

type Queue interface {
	Enqueue(v interface{})
	Dequeue() (interface{}, bool)
	Depth() int64
}

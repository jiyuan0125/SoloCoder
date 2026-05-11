package limiter

type ringBuffer struct {
	buffer []int64
	cap    int
	head   int
	tail   int
	count  int
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity <= 0 {
		capacity = 64
	}
	return &ringBuffer{
		buffer: make([]int64, capacity),
		cap:    capacity,
		head:   0,
		tail:   0,
		count:  0,
	}
}

func (r *ringBuffer) add(ts int64) {
	if r.count == r.cap {
		r.head = (r.head + 1) % r.cap
		r.count--
	}
	r.buffer[r.tail] = ts
	r.tail = (r.tail + 1) % r.cap
	r.count++
}

func (r *ringBuffer) size() int {
	return r.count
}

func (r *ringBuffer) clearBefore(threshold int64) {
	for r.count > 0 && r.buffer[r.head] < threshold {
		r.head = (r.head + 1) % r.cap
		r.count--
	}
}

func (r *ringBuffer) forEach(fn func(ts int64)) {
	for i := 0; i < r.count; i++ {
		idx := (r.head + i) % r.cap
		fn(r.buffer[idx])
	}
}

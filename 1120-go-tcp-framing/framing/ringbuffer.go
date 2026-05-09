package framing

type RingBuffer struct {
	buf    []byte
	rIndex int
	wIndex int
	count  int
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		buf: make([]byte, capacity),
	}
}

func (rb *RingBuffer) Capacity() int {
	return len(rb.buf)
}

func (rb *RingBuffer) Len() int {
	return rb.count
}

func (rb *RingBuffer) Free() int {
	return len(rb.buf) - rb.count
}

func (rb *RingBuffer) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p)
	if n > rb.Free() {
		rb.growTo(rb.count + n)
	}
	for i := 0; i < n; i++ {
		rb.buf[rb.wIndex] = p[i]
		rb.wIndex = (rb.wIndex + 1) % len(rb.buf)
		rb.count++
	}
	return n, nil
}

func (rb *RingBuffer) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p)
	if n > rb.count {
		n = rb.count
	}
	for i := 0; i < n; i++ {
		p[i] = rb.buf[rb.rIndex]
		rb.rIndex = (rb.rIndex + 1) % len(rb.buf)
		rb.count--
	}
	return n, nil
}

func (rb *RingBuffer) Peek(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	n := len(p)
	if n > rb.count {
		n = rb.count
	}
	idx := rb.rIndex
	for i := 0; i < n; i++ {
		p[i] = rb.buf[idx]
		idx = (idx + 1) % len(rb.buf)
	}
	return n, nil
}

func (rb *RingBuffer) Discard(n int) int {
	if n <= 0 {
		return 0
	}
	if n > rb.count {
		n = rb.count
	}
	rb.rIndex = (rb.rIndex + n) % len(rb.buf)
	rb.count -= n
	return n
}

func (rb *RingBuffer) Bytes() []byte {
	if rb.count == 0 {
		return nil
	}
	result := make([]byte, rb.count)
	idx := rb.rIndex
	for i := 0; i < rb.count; i++ {
		result[i] = rb.buf[idx]
		idx = (idx + 1) % len(rb.buf)
	}
	return result
}

func (rb *RingBuffer) growTo(size int) {
	newCap := len(rb.buf) * 2
	if newCap < size {
		newCap = size
	}
	newBuf := make([]byte, newCap)
	if rb.count > 0 {
		if rb.rIndex < rb.wIndex {
			copy(newBuf, rb.buf[rb.rIndex:rb.wIndex])
		} else {
			n1 := len(rb.buf) - rb.rIndex
			copy(newBuf, rb.buf[rb.rIndex:])
			copy(newBuf[n1:], rb.buf[:rb.wIndex])
		}
	}
	rb.buf = newBuf
	rb.rIndex = 0
	rb.wIndex = rb.count
}

func (rb *RingBuffer) Find(needle []byte) int {
	if len(needle) == 0 || rb.count < len(needle) {
		return -1
	}
	n := len(needle)
	maxStart := rb.count - n
	for start := 0; start <= maxStart; start++ {
		match := true
		for i := 0; i < n; i++ {
			idx := (rb.rIndex + start + i) % len(rb.buf)
			if rb.buf[idx] != needle[i] {
				match = false
				break
			}
		}
		if match {
			return start
		}
	}
	return -1
}

func (rb *RingBuffer) ReadAt(offset int, p []byte) int {
	if offset < 0 || offset >= rb.count {
		return 0
	}
	n := len(p)
	if n > rb.count-offset {
		n = rb.count - offset
	}
	for i := 0; i < n; i++ {
		idx := (rb.rIndex + offset + i) % len(rb.buf)
		p[i] = rb.buf[idx]
	}
	return n
}

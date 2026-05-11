package framing

type RingBuffer struct {
	buffer []byte
	head   int
	tail   int
	count  int
}

func NewRingBuffer(capacity int) *RingBuffer {
	if capacity <= 0 {
		capacity = 4096
	}
	return &RingBuffer{
		buffer: make([]byte, capacity),
		head:   0,
		tail:   0,
		count:  0,
	}
}

func (rb *RingBuffer) Capacity() int {
	return len(rb.buffer)
}

func (rb *RingBuffer) Length() int {
	return rb.count
}

func (rb *RingBuffer) Free() int {
	return len(rb.buffer) - rb.count
}

func (rb *RingBuffer) ensureCapacity(need int) {
	if rb.Free() >= need {
		return
	}
	newCap := len(rb.buffer) * 2
	for newCap-rb.count < need {
		newCap *= 2
	}
	newBuf := make([]byte, newCap)
	if rb.count > 0 {
		if rb.head < rb.tail {
			copy(newBuf, rb.buffer[rb.head:rb.tail])
		} else {
			n1 := copy(newBuf, rb.buffer[rb.head:])
			copy(newBuf[n1:], rb.buffer[:rb.tail])
		}
	}
	rb.buffer = newBuf
	rb.head = 0
	rb.tail = rb.count
}

func (rb *RingBuffer) Write(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	rb.ensureCapacity(len(data))
	n := len(data)
	end := rb.tail + n
	if end <= len(rb.buffer) {
		copy(rb.buffer[rb.tail:], data)
		rb.tail = end
	} else {
		first := len(rb.buffer) - rb.tail
		copy(rb.buffer[rb.tail:], data[:first])
		remaining := n - first
		copy(rb.buffer[:remaining], data[first:])
		rb.tail = remaining
	}
	rb.count += n
	return n
}

func (rb *RingBuffer) Peek(n int) ([]byte, bool) {
	if n <= 0 || n > rb.count {
		return nil, false
	}
	result := make([]byte, n)
	if rb.head+n <= len(rb.buffer) {
		copy(result, rb.buffer[rb.head:rb.head+n])
	} else {
		first := len(rb.buffer) - rb.head
		copy(result, rb.buffer[rb.head:])
		copy(result[first:], rb.buffer[:n-first])
	}
	return result, true
}

func (rb *RingBuffer) Read(n int) ([]byte, bool) {
	data, ok := rb.Peek(n)
	if !ok {
		return nil, false
	}
	rb.head = (rb.head + n) % len(rb.buffer)
	rb.count -= n
	return data, true
}

func (rb *RingBuffer) Discard(n int) bool {
	if n < 0 || n > rb.count {
		return false
	}
	rb.head = (rb.head + n) % len(rb.buffer)
	rb.count -= n
	return true
}

func (rb *RingBuffer) Index(sep []byte) int {
	if len(sep) == 0 || rb.count < len(sep) {
		return -1
	}
	for i := 0; i <= rb.count-len(sep); i++ {
		match := true
		for j := 0; j < len(sep); j++ {
			pos := (rb.head + i + j) % len(rb.buffer)
			if rb.buffer[pos] != sep[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func (rb *RingBuffer) IndexWithEscape(sep []byte, escapeByte byte) int {
	if len(sep) == 0 || rb.count < len(sep) {
		return -1
	}
	for i := 0; i <= rb.count-len(sep); i++ {
		if rb.isEscaped(i, escapeByte) {
			continue
		}
		match := true
		for j := 0; j < len(sep); j++ {
			if i+j >= rb.count {
				match = false
				break
			}
			if rb.isEscaped(i+j, escapeByte) {
				match = false
				break
			}
			pos := (rb.head + i + j) % len(rb.buffer)
			if rb.buffer[pos] != sep[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func (rb *RingBuffer) isEscaped(idx int, escapeByte byte) bool {
	if idx <= 0 {
		return false
	}
	escapeCount := 0
	for i := idx - 1; i >= 0; i-- {
		pos := (rb.head + i) % len(rb.buffer)
		if rb.buffer[pos] == escapeByte {
			escapeCount++
		} else {
			break
		}
	}
	return escapeCount%2 == 1
}

func (rb *RingBuffer) ByteAt(idx int) (byte, bool) {
	if idx < 0 || idx >= rb.count {
		return 0, false
	}
	pos := (rb.head + idx) % len(rb.buffer)
	return rb.buffer[pos], true
}

func (rb *RingBuffer) Reset() {
	rb.head = 0
	rb.tail = 0
	rb.count = 0
}

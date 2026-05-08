package ringbuffer

import (
	"errors"
	"io"
	"sync"
	"time"
)

var (
	ErrClosed   = errors.New("ringbuffer: buffer is closed")
	ErrTimeout  = errors.New("ringbuffer: operation timeout")
)

type RingBuffer struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	buf      []byte
	readPos  int
	writePos int
	count    int
	capacity int
	closed   bool

	zeroCapWrite []byte
	zeroCapRead  []byte
	zeroCapWait  int
}

func New(capacity int) *RingBuffer {
	if capacity < 0 {
		panic("ringbuffer: capacity must be non-negative")
	}
	rb := &RingBuffer{
		capacity: capacity,
	}
	if capacity > 0 {
		rb.buf = make([]byte, capacity)
	}
	rb.notEmpty = sync.NewCond(&rb.mu)
	rb.notFull = sync.NewCond(&rb.mu)
	return rb
}

func (rb *RingBuffer) Capacity() int {
	return rb.capacity
}

func (rb *RingBuffer) Size() int {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	return rb.count
}

func (rb *RingBuffer) Write(p []byte, timeout time.Duration) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed {
		return 0, ErrClosed
	}

	if rb.capacity == 0 {
		return rb.writeZeroCap(p, timeout)
	}

	return rb.writeNormal(p, timeout)
}

func (rb *RingBuffer) Read(p []byte, timeout time.Duration) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	rb.mu.Lock()
	defer rb.mu.Unlock()

	if rb.closed && rb.count == 0 {
		return 0, io.EOF
	}

	if rb.capacity == 0 {
		return rb.readZeroCap(p, timeout)
	}

	return rb.readNormal(p, timeout)
}

func (rb *RingBuffer) Close() {
	rb.mu.Lock()
	if !rb.closed {
		rb.closed = true
		rb.notEmpty.Broadcast()
		rb.notFull.Broadcast()
	}
	rb.mu.Unlock()
}

func (rb *RingBuffer) writeNormal(p []byte, timeout time.Duration) (int, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	written := 0
	for written < len(p) {
		for rb.count == rb.capacity {
			if rb.closed {
				return written, ErrClosed
			}
			if timeout == 0 {
				if written > 0 {
					return written, nil
				}
				return 0, ErrTimeout
			}
			if !rb.waitNotFull(deadline) {
				if written > 0 {
					return written, nil
				}
				return 0, ErrTimeout
			}
		}

		n := rb.copyIntoBuffer(p[written:])
		written += n
		rb.notEmpty.Signal()
	}

	return written, nil
}

func (rb *RingBuffer) readNormal(p []byte, timeout time.Duration) (int, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	read := 0
	for read < len(p) {
		for rb.count == 0 {
			if rb.closed {
				if read > 0 {
					return read, nil
				}
				return 0, io.EOF
			}
			if timeout == 0 {
				if read > 0 {
					return read, nil
				}
				return 0, ErrTimeout
			}
			if !rb.waitNotEmpty(deadline) {
				if read > 0 {
					return read, nil
				}
				return 0, ErrTimeout
			}
		}

		n := rb.copyFromBuffer(p[read:])
		read += n
		rb.notFull.Signal()
	}

	return read, nil
}

func (rb *RingBuffer) writeZeroCap(p []byte, timeout time.Duration) (int, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	rb.zeroCapWrite = p
	rb.zeroCapWait++
	rb.notEmpty.Signal()

	for rb.zeroCapWrite != nil {
		if rb.closed {
			rb.zeroCapWrite = nil
			return 0, ErrClosed
		}
		if timeout == 0 {
			rb.zeroCapWrite = nil
			rb.zeroCapWait--
			return 0, ErrTimeout
		}
		if !rb.waitNotFull(deadline) {
			rb.zeroCapWrite = nil
			rb.zeroCapWait--
			return 0, ErrTimeout
		}
	}

	rb.zeroCapWait--
	return len(p), nil
}

func (rb *RingBuffer) readZeroCap(p []byte, timeout time.Duration) (int, error) {
	var deadline time.Time
	if timeout > 0 {
		deadline = time.Now().Add(timeout)
	}

	for rb.zeroCapWrite == nil {
		if rb.closed {
			return 0, io.EOF
		}
		if timeout == 0 {
			return 0, ErrTimeout
		}
		if !rb.waitNotEmpty(deadline) {
			return 0, ErrTimeout
		}
	}

	data := rb.zeroCapWrite
	rb.zeroCapWrite = nil
	rb.notFull.Signal()

	n := copy(p, data)
	return n, nil
}

func (rb *RingBuffer) copyIntoBuffer(p []byte) int {
	available := rb.capacity - rb.count
	toWrite := len(p)
	if toWrite > available {
		toWrite = available
	}

	end := rb.writePos + toWrite
	if end <= rb.capacity {
		copy(rb.buf[rb.writePos:], p[:toWrite])
	} else {
		part1 := rb.capacity - rb.writePos
		copy(rb.buf[rb.writePos:], p[:part1])
		copy(rb.buf, p[part1:toWrite])
	}

	rb.writePos = (rb.writePos + toWrite) % rb.capacity
	rb.count += toWrite
	return toWrite
}

func (rb *RingBuffer) copyFromBuffer(p []byte) int {
	toRead := len(p)
	if toRead > rb.count {
		toRead = rb.count
	}

	end := rb.readPos + toRead
	if end <= rb.capacity {
		copy(p, rb.buf[rb.readPos:rb.readPos+toRead])
	} else {
		part1 := rb.capacity - rb.readPos
		copy(p[:part1], rb.buf[rb.readPos:])
		copy(p[part1:toRead], rb.buf[:toRead-part1])
	}

	rb.readPos = (rb.readPos + toRead) % rb.capacity
	rb.count -= toRead
	return toRead
}

func (rb *RingBuffer) waitNotFull(deadline time.Time) bool {
	if deadline.IsZero() {
		rb.notFull.Wait()
		return true
	}

	done := make(chan struct{})
	go func() {
		rb.mu.Lock()
		rb.notFull.Wait()
		rb.mu.Unlock()
		close(done)
	}()

	rb.mu.Unlock()
	select {
	case <-done:
		rb.mu.Lock()
		return true
	case <-time.After(time.Until(deadline)):
		rb.mu.Lock()
		rb.notFull.Broadcast()
		return false
	}
}

func (rb *RingBuffer) waitNotEmpty(deadline time.Time) bool {
	if deadline.IsZero() {
		rb.notEmpty.Wait()
		return true
	}

	done := make(chan struct{})
	go func() {
		rb.mu.Lock()
		rb.notEmpty.Wait()
		rb.mu.Unlock()
		close(done)
	}()

	rb.mu.Unlock()
	select {
	case <-done:
		rb.mu.Lock()
		return true
	case <-time.After(time.Until(deadline)):
		rb.mu.Lock()
		rb.notEmpty.Broadcast()
		return false
	}
}

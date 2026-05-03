package mux

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sync"
)

const (
	WindowSize = 64 * 1024
)

var (
	ErrStreamClosed      = errors.New("stream closed")
	ErrWindowFull        = errors.New("send window full")
	ErrStreamReset       = errors.New("stream reset")
	ErrWriteAfterFIN     = errors.New("write after FIN")
)

type Stream interface {
	Read(p []byte) (n int, err error)
	Write(p []byte) (n int, err error)
	Close() error
	StreamID() uint32
}

type stream struct {
	id        uint32
	mux       *Mux
	recvBuf   *bytes.Buffer
	recvCond  *sync.Cond
	recvLock  sync.Mutex
	recvWindow uint32

	sendLock   sync.Mutex
	sendWindow uint32
	sendCond   *sync.Cond

	finSent   bool
	finRecv   bool
	rstRecv   bool
	closed    bool
	localDone bool
}

func newStream(id uint32, m *Mux) *stream {
	s := &stream{
		id:          id,
		mux:         m,
		recvBuf:     bytes.NewBuffer(nil),
		recvWindow:  WindowSize,
		sendWindow:  WindowSize,
	}
	s.recvCond = sync.NewCond(&s.recvLock)
	s.sendCond = sync.NewCond(&s.sendLock)
	return s
}

func (s *stream) StreamID() uint32 {
	return s.id
}

func (s *stream) Read(p []byte) (n int, err error) {
	s.recvLock.Lock()

	for s.recvBuf.Len() == 0 && !s.finRecv && !s.rstRecv && !s.closed {
		s.recvCond.Wait()
	}

	if s.rstRecv || s.closed {
		s.recvLock.Unlock()
		return 0, ErrStreamReset
	}

	if s.recvBuf.Len() == 0 && s.finRecv {
		s.recvLock.Unlock()
		return 0, io.EOF
	}

	n, _ = s.recvBuf.Read(p)

	var window uint32
	if n > 0 {
		s.recvWindow += uint32(n)
		window = s.recvWindow
	}

	s.recvLock.Unlock()

	if n > 0 {
		s.sendWindowUpdateImpl(window)
	}

	return n, nil
}

func (s *stream) Write(p []byte) (n int, err error) {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()

	if s.finSent || s.closed || s.rstRecv {
		if s.rstRecv {
			return 0, ErrStreamReset
		}
		return 0, ErrWriteAfterFIN
	}

	totalWritten := 0
	remaining := p

	for len(remaining) > 0 {
		for s.sendWindow == 0 && !s.closed && !s.rstRecv {
			s.sendCond.Wait()
		}

		if s.closed || s.rstRecv {
			return totalWritten, ErrStreamReset
		}

		sendSize := min(uint32(len(remaining)), s.sendWindow)
		if sendSize == 0 {
			return totalWritten, nil
		}

		toSend := remaining[:sendSize]
		frame := &Frame{
			StreamID: s.id,
			Flags:    0,
			Payload:  toSend,
		}

		s.mux.writeLock.Lock()
		err = WriteFrame(s.mux.conn, frame)
		s.mux.writeLock.Unlock()

		if err != nil {
			return totalWritten, err
		}

		s.sendWindow -= sendSize
		totalWritten += int(sendSize)
		remaining = remaining[sendSize:]
	}

	return totalWritten, nil
}

func (s *stream) Close() error {
	s.sendLock.Lock()
	alreadySent := s.finSent
	s.finSent = true
	s.sendLock.Unlock()

	if !alreadySent {
		frame := &Frame{
			StreamID: s.id,
			Flags:    FlagFIN,
			Payload:  nil,
		}

		s.mux.writeLock.Lock()
		err := WriteFrame(s.mux.conn, frame)
		s.mux.writeLock.Unlock()

		if err != nil {
			return err
		}
	}

	s.recvLock.Lock()
	localFinished := s.finSent && s.finRecv
	s.recvLock.Unlock()

	if localFinished {
		s.mux.removeStream(s.id)
	}

	return nil
}

func (s *stream) handleData(data []byte) {
	s.recvLock.Lock()
	defer s.recvLock.Unlock()

	if s.rstRecv || s.closed {
		return
	}

	dataLen := uint32(len(data))
	if s.recvWindow >= dataLen {
		s.recvWindow -= dataLen
		s.recvBuf.Write(data)
		s.recvCond.Broadcast()
	}
}

func (s *stream) handleFIN() {
	s.recvLock.Lock()
	s.finRecv = true
	s.recvCond.Broadcast()
	s.recvLock.Unlock()

	s.sendLock.Lock()
	bothFinished := s.finSent && s.finRecv
	s.sendLock.Unlock()

	if bothFinished {
		s.mux.removeStream(s.id)
	}
}

func (s *stream) handleRST() {
	s.recvLock.Lock()
	s.rstRecv = true
	s.recvCond.Broadcast()
	s.recvLock.Unlock()

	s.sendLock.Lock()
	s.sendCond.Broadcast()
	s.sendLock.Unlock()

	s.mux.removeStream(s.id)
}

func (s *stream) handleWindowUpdate(newWindow uint32) {
	s.sendLock.Lock()
	defer s.sendLock.Unlock()
	s.sendWindow = newWindow
	s.sendCond.Broadcast()
}

func (s *stream) sendWindowUpdate() {
	s.recvLock.Lock()
	window := s.recvWindow
	s.recvLock.Unlock()

	s.sendWindowUpdateImpl(window)
}

func (s *stream) sendWindowUpdateImpl(window uint32) {
	windowBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(windowBytes, window)

	frame := &Frame{
		StreamID: s.id,
		Flags:    FlagACK,
		Payload:  windowBytes,
	}

	s.mux.writeLock.Lock()
	WriteFrame(s.mux.conn, frame)
	s.mux.writeLock.Unlock()
}

func min(a, b uint32) uint32 {
	if a < b {
		return a
	}
	return b
}

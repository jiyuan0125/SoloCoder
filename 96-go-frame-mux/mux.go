package mux

import (
	"encoding/binary"
	"io"
	"net"
	"sync"
)

type Mux struct {
	conn       net.Conn
	isClient   bool
	nextStreamID uint32

	streams    map[uint32]*stream
	streamsLock sync.RWMutex

	acceptChan  chan *stream
	acceptCond  *sync.Cond
	acceptLock  sync.Mutex

	writeLock   sync.Mutex

	closed      bool
	closedLock  sync.RWMutex
}

func Client(conn net.Conn) *Mux {
	return newMux(conn, true)
}

func Server(conn net.Conn) *Mux {
	return newMux(conn, false)
}

func newMux(conn net.Conn, isClient bool) *Mux {
	m := &Mux{
		conn:        conn,
		isClient:    isClient,
		streams:     make(map[uint32]*stream),
		acceptChan:  make(chan *stream, 64),
	}
	if isClient {
		m.nextStreamID = 1
	} else {
		m.nextStreamID = 2
	}
	m.acceptCond = sync.NewCond(&m.acceptLock)
	go m.readLoop()
	return m
}

func (m *Mux) OpenStream() (Stream, error) {
	m.closedLock.RLock()
	if m.closed {
		m.closedLock.RUnlock()
		return nil, io.EOF
	}
	m.closedLock.RUnlock()

	m.streamsLock.Lock()
	id := m.nextStreamID
	m.nextStreamID += 2
	s := newStream(id, m)
	m.streams[id] = s
	m.streamsLock.Unlock()

	frame := &Frame{
		StreamID: id,
		Flags:    FlagSYN,
		Payload:  nil,
	}

	m.writeLock.Lock()
	err := WriteFrame(m.conn, frame)
	m.writeLock.Unlock()

	if err != nil {
		m.removeStream(id)
		return nil, err
	}

	return s, nil
}

func (m *Mux) AcceptStream() (Stream, error) {
	for {
		m.acceptLock.Lock()
		for len(m.acceptChan) == 0 {
			m.closedLock.RLock()
			isClosed := m.closed
			m.closedLock.RUnlock()
			if isClosed {
				m.acceptLock.Unlock()
				return nil, io.EOF
			}
			m.acceptCond.Wait()
		}
		select {
		case s := <-m.acceptChan:
			m.acceptLock.Unlock()
			return s, nil
		default:
			m.acceptLock.Unlock()
		}
	}
}

func (m *Mux) readLoop() {
	defer m.closeAll()

	for {
		frame, err := ReadFrame(m.conn)
		if err != nil {
			return
		}

		m.handleFrame(frame)
	}
}

func (m *Mux) handleFrame(frame *Frame) {
	streamID := frame.StreamID

	if streamID == 0 {
		return
	}

	if (frame.Flags & FlagSYN) != 0 {
		m.handleSYN(streamID)
		return
	}

	m.streamsLock.RLock()
	s, exists := m.streams[streamID]
	m.streamsLock.RUnlock()

	if !exists {
		return
	}

	if (frame.Flags & FlagRST) != 0 {
		s.handleRST()
		return
	}

	if (frame.Flags & FlagACK) != 0 && len(frame.Payload) == 4 {
		window := binary.BigEndian.Uint32(frame.Payload)
		s.handleWindowUpdate(window)
		return
	}

	if len(frame.Payload) > 0 {
		s.handleData(frame.Payload)
	}

	if (frame.Flags & FlagFIN) != 0 {
		s.handleFIN()
	}
}

func (m *Mux) handleSYN(streamID uint32) {
	m.streamsLock.Lock()
	defer m.streamsLock.Unlock()

	if _, exists := m.streams[streamID]; exists {
		return
	}

	s := newStream(streamID, m)
	m.streams[streamID] = s

	m.acceptLock.Lock()
	select {
	case m.acceptChan <- s:
	default:
	}
	m.acceptCond.Broadcast()
	m.acceptLock.Unlock()
}

func (m *Mux) removeStream(id uint32) {
	m.streamsLock.Lock()
	delete(m.streams, id)
	m.streamsLock.Unlock()
}

func (m *Mux) closeAll() {
	m.closedLock.Lock()
	if m.closed {
		m.closedLock.Unlock()
		return
	}
	m.closed = true
	m.closedLock.Unlock()

	m.streamsLock.Lock()
	for _, s := range m.streams {
		s.handleRST()
	}
	m.streamsLock.Unlock()

	m.acceptLock.Lock()
	m.acceptCond.Broadcast()
	m.acceptLock.Unlock()

	m.conn.Close()
}

func (m *Mux) Close() error {
	m.closeAll()
	return nil
}

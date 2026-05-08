package wsframe

import (
	"bytes"
	"io"
	"net"
	"time"
)

type Connection struct {
	conn          net.Conn
	readBuffer    []byte
	defragmenter  *Defragmenter
	serverSide    bool
	closed        bool
	closeReceived bool
	closeSent     bool
}

func NewServerConnection(conn net.Conn, maxMessageSize int) *Connection {
	return &Connection{
		conn:         conn,
		readBuffer:   make([]byte, 0),
		defragmenter: NewDefragmenter(maxMessageSize),
		serverSide:   true,
	}
}

func NewClientConnection(conn net.Conn, maxMessageSize int) *Connection {
	return &Connection{
		conn:         conn,
		readBuffer:   make([]byte, 0),
		defragmenter: NewDefragmenter(maxMessageSize),
		serverSide:   false,
	}
}

func (c *Connection) ReadMessage() (*Message, error) {
	for {
		frame, err := c.readFrame()
		if err != nil {
			return nil, err
		}

		msg, err := c.defragmenter.ProcessFrame(frame)
		if err != nil {
			return nil, err
		}

		if msg != nil {
			return msg, nil
		}
	}
}

func (c *Connection) readFrame() (*Frame, error) {
	buffer := make([]byte, 4096)

	for {
		frame, consumed, err := ParseFrame(c.readBuffer)
		if err == nil {
			c.readBuffer = c.readBuffer[consumed:]
			return frame, nil
		}

		if err != ErrInsufficientData {
			return nil, err
		}

		n, err := c.conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return nil, err
			}
			return nil, err
		}

		c.readBuffer = append(c.readBuffer, buffer[:n]...)
	}
}

func (c *Connection) WriteFrame(frame *Frame) error {
	if c.closed {
		return io.ErrClosedPipe
	}

	data := SerializeFrame(frame)
	_, err := c.conn.Write(data)
	return err
}

func (c *Connection) WriteText(text string) error {
	frame := BuildTextFrame(text, !c.serverSide)
	return c.WriteFrame(frame)
}

func (c *Connection) WriteBinary(data []byte) error {
	frame := BuildBinaryFrame(data, !c.serverSide)
	return c.WriteFrame(frame)
}

func (c *Connection) WritePing(payload []byte) error {
	frame := BuildPingFrame(payload, !c.serverSide)
	return c.WriteFrame(frame)
}

func (c *Connection) WritePong(payload []byte) error {
	frame := BuildPongFrame(payload, !c.serverSide)
	return c.WriteFrame(frame)
}

func (c *Connection) WriteClose(code int, reason string) error {
	if c.closeSent {
		return nil
	}

	frame := BuildCloseFrame(code, reason, !c.serverSide)
	err := c.WriteFrame(frame)
	if err == nil {
		c.closeSent = true
	}
	return err
}

func (c *Connection) Close() error {
	if c.closed {
		return nil
	}
	c.closed = true
	return c.conn.Close()
}

func (c *Connection) IsClosed() bool {
	return c.closed
}

func (c *Connection) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Connection) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func SplitTextMessage(text string, fragmentSize int) [][]byte {
	if fragmentSize <= 0 {
		fragmentSize = 1024
	}

	data := []byte(text)
	var fragments [][]byte

	for len(data) > fragmentSize {
		fragments = append(fragments, data[:fragmentSize])
		data = data[fragmentSize:]
	}

	if len(data) > 0 {
		fragments = append(fragments, data)
	}

	return fragments
}

func SplitBinaryMessage(data []byte, fragmentSize int) [][]byte {
	if fragmentSize <= 0 {
		fragmentSize = 1024
	}

	var fragments [][]byte

	for len(data) > fragmentSize {
		fragments = append(fragments, data[:fragmentSize])
		data = data[fragmentSize:]
	}

	if len(data) > 0 {
		fragments = append(fragments, data)
	}

	return fragments
}

func NewBufferReader(buf *bytes.Buffer) *bytes.Reader {
	return bytes.NewReader(buf.Bytes())
}

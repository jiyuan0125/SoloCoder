package common

import (
	"bufio"
	"encoding/binary"
	"io"
)

const (
	ProtocolVersion = 1
	MaxMessageSize  = 1024 * 1024
)

func SendMessage(w io.Writer, data []byte) error {
	if len(data) > MaxMessageSize {
		return ErrMessageTooLarge
	}

	header := make([]byte, 8)
	binary.BigEndian.PutUint32(header[0:4], ProtocolVersion)
	binary.BigEndian.PutUint32(header[4:8], uint32(len(data)))

	if _, err := w.Write(header); err != nil {
		return err
	}

	if _, err := w.Write(data); err != nil {
		return err
	}

	return nil
}

func ReceiveMessage(r io.Reader) ([]byte, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	version := binary.BigEndian.Uint32(header[0:4])
	if version != ProtocolVersion {
		return nil, ErrVersionMismatch
	}

	msgLen := binary.BigEndian.Uint32(header[4:8])
	if msgLen > MaxMessageSize {
		return nil, ErrMessageTooLarge
	}

	data := make([]byte, msgLen)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	return data, nil
}

func SendRequest(w io.Writer, req *DeployRequest) error {
	data, err := MarshalRequest(req)
	if err != nil {
		return err
	}
	return SendMessage(w, data)
}

func ReceiveRequest(r io.Reader) (*DeployRequest, error) {
	data, err := ReceiveMessage(r)
	if err != nil {
		return nil, err
	}
	return UnmarshalRequest(data)
}

func SendResponse(w io.Writer, resp *DeployResponse) error {
	data, err := MarshalResponse(resp)
	if err != nil {
		return err
	}
	return SendMessage(w, data)
}

func ReceiveResponse(r io.Reader) (*DeployResponse, error) {
	data, err := ReceiveMessage(r)
	if err != nil {
		return nil, err
	}
	return UnmarshalResponse(data)
}

var (
	ErrVersionMismatch = &ProtocolError{Msg: "protocol version mismatch"}
	ErrMessageTooLarge = &ProtocolError{Msg: "message too large"}
)

type ProtocolError struct {
	Msg string
}

func (e *ProtocolError) Error() string {
	return e.Msg
}

type BufferedConn struct {
	*bufio.Reader
	*bufio.Writer
}

func NewBufferedConn(rw io.ReadWriter) *BufferedConn {
	return &BufferedConn{
		Reader: bufio.NewReader(rw),
		Writer: bufio.NewWriter(rw),
	}
}

func (bc *BufferedConn) SendRequest(req *DeployRequest) error {
	if err := SendRequest(bc.Writer, req); err != nil {
		return err
	}
	return bc.Writer.Flush()
}

func (bc *BufferedConn) ReceiveRequest() (*DeployRequest, error) {
	return ReceiveRequest(bc.Reader)
}

func (bc *BufferedConn) SendResponse(resp *DeployResponse) error {
	if err := SendResponse(bc.Writer, resp); err != nil {
		return err
	}
	return bc.Writer.Flush()
}

func (bc *BufferedConn) ReceiveResponse() (*DeployResponse, error) {
	return ReceiveResponse(bc.Reader)
}

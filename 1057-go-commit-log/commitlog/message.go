package commitlog

import (
	"encoding/binary"
	"errors"
	"time"
)

const (
	messageHeaderSize = 24
	maxMessageSize    = 1024 * 1024
)

type record struct {
	offset    int64
	timestamp int64
	value     []byte
}

func newRecord(offset int64, value []byte) *record {
	return &record{
		offset:    offset,
		timestamp: time.Now().UnixNano(),
		value:     value,
	}
}

func (r *record) encode() []byte {
	buf := make([]byte, messageHeaderSize+len(r.value))
	binary.BigEndian.PutUint32(buf[0:4], uint32(len(r.value)))
	binary.BigEndian.PutUint32(buf[4:8], uint32(0))
	binary.BigEndian.PutUint64(buf[8:16], uint64(r.offset))
	binary.BigEndian.PutUint64(buf[16:24], uint64(r.timestamp))
	copy(buf[messageHeaderSize:], r.value)
	return buf
}

func (r *record) decode(buf []byte) error {
	if len(buf) < messageHeaderSize {
		return errors.New("buffer too small for header")
	}
	length := binary.BigEndian.Uint32(buf[0:4])
	if length == 0 || length > maxMessageSize {
		return errors.New("invalid message length")
	}
	if len(buf) < messageHeaderSize+int(length) {
		return errors.New("buffer too small for message")
	}
	r.offset = int64(binary.BigEndian.Uint64(buf[8:16]))
	r.timestamp = int64(binary.BigEndian.Uint64(buf[16:24]))
	r.value = make([]byte, length)
	copy(r.value, buf[messageHeaderSize:messageHeaderSize+int(length)])
	return nil
}

func (r *record) size() int {
	return messageHeaderSize + len(r.value)
}

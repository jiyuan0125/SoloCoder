package fifo

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
)

const (
	headerSize     = 4
	defaultMaxMsg  = 4096 - headerSize
	pipeBufDefault = 4096
)

var (
	ErrMessageTooLarge = errors.New("message exceeds maximum allowed size")
	ErrTruncated       = errors.New("message truncated")
	ErrPipeClosed      = errors.New("pipe closed")
)

type Writer struct {
	path     string
	file     *os.File
	mu       sync.Mutex
	maxMsg   int
	pipeBuf  int
	isOpen   bool
}

type Reader struct {
	path     string
	file     *os.File
	isOpen   bool
}

type Message struct {
	Data []byte
}

func getPipeBuf(fd uintptr) int {
	return pipeBufDefault
}

func Create(path string) error {
	err := syscall.Mkfifo(path, 0666)
	if err != nil && !os.IsExist(err) {
		return fmt.Errorf("failed to create fifo: %w", err)
	}
	return nil
}

func Exists(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeNamedPipe != 0
}

func NewWriter(path string, maxMsgSize ...int) (*Writer, error) {
	if !Exists(path) {
		if err := Create(path); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open fifo for writing: %w", err)
	}

	maxMsg := defaultMaxMsg
	if len(maxMsgSize) > 0 && maxMsgSize[0] > 0 {
		maxMsg = maxMsgSize[0]
	}

	pipeBuf := getPipeBuf(file.Fd())

	return &Writer{
		path:    path,
		file:    file,
		maxMsg:  maxMsg,
		pipeBuf: pipeBuf,
		isOpen:  true,
	}, nil
}

func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isOpen {
		return nil
	}
	w.isOpen = false
	return w.file.Close()
}

func (w *Writer) WriteMessage(data []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isOpen {
		return ErrPipeClosed
	}

	if len(data) > w.maxMsg {
		return ErrMessageTooLarge
	}

	header := make([]byte, headerSize)
	binary.BigEndian.PutUint32(header, uint32(len(data)))

	buf := make([]byte, 0, headerSize+len(data))
	buf = append(buf, header...)
	buf = append(buf, data...)

	total := 0
	for total < len(buf) {
		n, err := w.file.Write(buf[total:])
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				continue
			}
			return fmt.Errorf("write failed: %w", err)
		}
		total += n
	}

	return nil
}

func (w *Writer) MaxMessageSize() int {
	return w.maxMsg
}

func (w *Writer) Path() string {
	return w.path
}

func NewReader(path string) (*Reader, error) {
	if !Exists(path) {
		if err := Create(path); err != nil {
			return nil, err
		}
	}

	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open fifo for reading: %w", err)
	}

	return &Reader{
		path:   path,
		file:   file,
		isOpen: true,
	}, nil
}

func (r *Reader) Close() error {
	if !r.isOpen {
		return nil
	}
	r.isOpen = false
	return r.file.Close()
}

func (r *Reader) ReadMessage() (*Message, bool, error) {
	if !r.isOpen {
		return nil, false, ErrPipeClosed
	}

	header := make([]byte, headerSize)
	n, err := r.file.Read(header)
	if err != nil {
		if err == io.EOF {
			return nil, false, nil
		}
		if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("read header failed: %w", err)
	}

	if n == 0 {
		return nil, false, nil
	}

	if n < headerSize {
		return nil, true, ErrTruncated
	}

	msgLen := binary.BigEndian.Uint32(header)
	data := make([]byte, msgLen)
	total := 0

	for total < int(msgLen) {
		n, err := r.file.Read(data[total:])
		if err != nil {
			if err == io.EOF {
				if total > 0 {
					return nil, true, ErrTruncated
				}
				return nil, false, nil
			}
			if err == syscall.EAGAIN || err == syscall.EWOULDBLOCK {
				continue
			}
			return nil, true, fmt.Errorf("read body failed: %w", err)
		}
		if n == 0 {
			continue
		}
		total += n
	}

	return &Message{Data: data}, false, nil
}

func (r *Reader) Path() string {
	return r.path
}

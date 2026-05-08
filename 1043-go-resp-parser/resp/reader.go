package resp

import (
	"bufio"
	"errors"
	"io"
	"strconv"
)

type Reader struct {
	r     *bufio.Reader
	pos   int64
}

func NewReader(r io.Reader) *Reader {
	return &Reader{r: bufio.NewReader(r)}
}

func (r *Reader) ReadByte() (byte, error) {
	b, err := r.r.ReadByte()
	if err == nil {
		r.pos++
	}
	return b, err
}

func (r *Reader) UnreadByte() error {
	err := r.r.UnreadByte()
	if err == nil {
		r.pos--
	}
	return err
}

func (r *Reader) ReadLine() ([]byte, error) {
	var line []byte
	for {
		part, isPrefix, err := r.r.ReadLine()
		if err != nil {
			return nil, err
		}
		line = append(line, part...)
		r.pos += int64(len(part))
		if !isPrefix {
			break
		}
	}
	if len(line) >= 2 && line[len(line)-2] == '\r' {
		line = line[:len(line)-1]
		r.pos++
	}
	return line, nil
}

func (r *Reader) ReadN(n int) ([]byte, error) {
	if n < 0 {
		return nil, errors.New("negative read length")
	}
	buf := make([]byte, n)
	read := 0
	for read < n {
		nr, err := r.r.Read(buf[read:])
		read += nr
		r.pos += int64(nr)
		if err != nil {
			if err == io.EOF && read > 0 {
				break
			}
			return nil, err
		}
	}
	if read != n {
		return nil, io.ErrUnexpectedEOF
	}
	return buf, nil
}

func (r *Reader) ReadLineAsString() (string, error) {
	line, err := r.ReadLine()
	if err != nil {
		return "", err
	}
	if len(line) == 0 {
		return "", errors.New("empty line")
	}
	return string(line), nil
}

func (r *Reader) ReadLineAsInteger() (int64, error) {
	line, err := r.ReadLine()
	if err != nil {
		return 0, err
	}
	if len(line) == 0 {
		return 0, errors.New("empty line")
	}
	n, err := strconv.ParseInt(string(line), 10, 64)
	if err != nil {
		return 0, r.errorf("invalid integer: %v", err)
	}
	return n, nil
}

func (r *Reader) Pos() int64 {
	return r.pos
}

func (r *Reader) errorf(format string, args ...interface{}) error {
	return NewProtocolError(r.pos, format, args...)
}

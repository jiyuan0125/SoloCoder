package resp

import (
	"bufio"
	"io"
	"strconv"
)

type Writer struct {
	w *bufio.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: bufio.NewWriter(w)}
}

func (w *Writer) WriteSimpleString(s string) error {
	if err := w.w.WriteByte('+'); err != nil {
		return err
	}
	if _, err := w.w.WriteString(s); err != nil {
		return err
	}
	if _, err := w.w.WriteString(CRLF); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteError(s string) error {
	if err := w.w.WriteByte('-'); err != nil {
		return err
	}
	if _, err := w.w.WriteString(s); err != nil {
		return err
	}
	if _, err := w.w.WriteString(CRLF); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteInteger(n int64) error {
	if err := w.w.WriteByte(':'); err != nil {
		return err
	}
	if _, err := w.w.WriteString(strconv.FormatInt(n, 10)); err != nil {
		return err
	}
	if _, err := w.w.WriteString(CRLF); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteBulkString(s string) error {
	if err := w.w.WriteByte('$'); err != nil {
		return err
	}
	if _, err := w.w.WriteString(strconv.Itoa(len(s))); err != nil {
		return err
	}
	if _, err := w.w.WriteString(CRLF); err != nil {
		return err
	}
	if _, err := w.w.WriteString(s); err != nil {
		return err
	}
	if _, err := w.w.WriteString(CRLF); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteNullBulkString() error {
	if _, err := w.w.WriteString("$-1\r\n"); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteNullArray() error {
	if _, err := w.w.WriteString("*-1\r\n"); err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteArrayHeader(len int) error {
	if err := w.w.WriteByte('*'); err != nil {
		return err
	}
	if _, err := w.w.WriteString(strconv.Itoa(len)); err != nil {
		return err
	}
	if _, err := w.w.WriteString(CRLF); err != nil {
		return err
	}
	return nil
}

func (w *Writer) Flush() error {
	return w.w.Flush()
}

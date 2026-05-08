package resp

import (
	"io"
)

type Encoder struct {
	w *Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: NewWriter(w)}
}

func (e *Encoder) Encode(v Value) error {
	switch v.Type {
	case TypeSimpleString:
		return e.w.WriteSimpleString(v.Str)
	case TypeError:
		return e.w.WriteError(v.Str)
	case TypeInteger:
		return e.w.WriteInteger(v.Int)
	case TypeBulkString:
		return e.w.WriteBulkString(v.Str)
	case TypeNullBulkString:
		return e.w.WriteNullBulkString()
	case TypeNullArray:
		return e.w.WriteNullArray()
	case TypeArray:
		if v.IsNull {
			return e.w.WriteNullArray()
		}
		if err := e.w.WriteArrayHeader(len(v.Array)); err != nil {
			return err
		}
		for _, elem := range v.Array {
			if err := e.Encode(elem); err != nil {
				return err
			}
		}
		return nil
	}
	return nil
}

func (e *Encoder) EncodeCommand(args []string) error {
	arr := make([]Value, len(args))
	for i, arg := range args {
		arr[i] = NewBulkString(arg)
	}
	return e.Encode(NewArray(arr))
}

func (e *Encoder) EncodeCommands(commands [][]string) error {
	for _, args := range commands {
		if err := e.EncodeCommand(args); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) Flush() error {
	return e.w.Flush()
}

func EncodeValue(v Value) ([]byte, error) {
	var buf []byte
	w := &bytesWriter{buf: &buf}
	enc := NewEncoder(w)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf, nil
}

func EncodeCommand(args []string) ([]byte, error) {
	var buf []byte
	w := &bytesWriter{buf: &buf}
	enc := NewEncoder(w)
	if err := enc.EncodeCommand(args); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf, nil
}

type bytesWriter struct {
	buf *[]byte
}

func (b *bytesWriter) Write(p []byte) (n int, err error) {
	*b.buf = append(*b.buf, p...)
	return len(p), nil
}

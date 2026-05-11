package bencode

import (
	"bytes"
	"io"
	"math/big"
	"sort"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func Encode(v Value) ([]byte, error) {
	var buf bytes.Buffer
	err := NewEncoder(&buf).Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (e *Encoder) Encode(v Value) error {
	switch val := v.(type) {
	case ByteString:
		return e.encodeByteString(val)
	case Integer:
		return e.encodeInteger(val)
	case List:
		return e.encodeList(val)
	case Dict:
		return e.encodeDict(val)
	default:
		return e.encodeFallback(v)
	}
}

func (e *Encoder) encodeFallback(v Value) error {
	switch val := v.(type) {
	case string:
		return e.encodeByteString(ByteString(val))
	case []byte:
		return e.encodeByteString(ByteString(val))
	case int:
		return e.encodeInteger(NewInteger(int64(val)))
	case int32:
		return e.encodeInteger(NewInteger(int64(val)))
	case int64:
		return e.encodeInteger(NewInteger(val))
	case uint:
		return e.encodeInteger(NewInteger(int64(val)))
	case uint32:
		return e.encodeInteger(NewInteger(int64(val)))
	case uint64:
		i := Integer{new(big.Int).SetUint64(val)}
		return e.encodeInteger(i)
	case float64:
		if float64(int64(val)) == val {
			return e.encodeInteger(NewInteger(int64(val)))
		}
		return e.encodeByteString(ByteString(formatFloat(val)))
	case []Value:
		return e.encodeList(List(val))
	case []interface{}:
		list := make(List, len(val))
		for i, item := range val {
			list[i] = item
		}
		return e.encodeList(list)

	case map[string]Value:
		dict := make(Dict)
		for k, v := range val {
			dict[k] = v
		}
		return e.encodeDict(dict)
	case map[string]interface{}:
		dict := make(Dict)
		for k, v := range val {
			dict[k] = v
		}
		return e.encodeDict(dict)
	case map[interface{}]interface{}:
		dict := make(Dict)
		for k, v := range val {
			key, ok := k.(string)
			if !ok {
				return nil
			}
			dict[key] = v
		}
		return e.encodeDict(dict)
	default:
		return nil
	}
}

func (e *Encoder) encodeByteString(s ByteString) error {
	lenStr := formatInt64(int64(len(s)))
	e.w.Write([]byte(lenStr))
	e.w.Write([]byte(":"))
	e.w.Write(s)
	return nil
}

func (e *Encoder) encodeInteger(i Integer) error {
	e.w.Write([]byte("i"))
	e.w.Write([]byte(i.String()))
	e.w.Write([]byte("e"))
	return nil
}

func (e *Encoder) encodeList(l List) error {
	e.w.Write([]byte("l"))
	for _, item := range l {
		if err := e.Encode(item); err != nil {
			return err
		}
	}
	e.w.Write([]byte("e"))
	return nil
}

func (e *Encoder) encodeDict(d Dict) error {
	e.w.Write([]byte("d"))
	keys := make([]string, 0, len(d))
	for k := range d {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if err := e.encodeByteString(ByteString(k)); err != nil {
			return err
		}
		if err := e.Encode(d[k]); err != nil {
			return err
		}
	}
	e.w.Write([]byte("e"))
	return nil
}

func formatInt64(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf []byte
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append(buf, byte('0'+(n%10)))
		n /= 10
	}
	if neg {
		buf = append(buf, '-')
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}

func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return formatInt64(int64(f))
	}
	var buf []byte
	if f < 0 {
		buf = append(buf, '-')
		f = -f
	}
	whole := int64(f)
	frac := f - float64(whole)
	buf = append(buf, []byte(formatInt64(whole))...)
	if frac > 0 {
		buf = append(buf, '.')
		for i := 0; i < 6; i++ {
			frac *= 10
			digit := int(frac)
			buf = append(buf, byte('0'+digit))
			frac -= float64(digit)
		}
	}
	return string(buf)
}

package bencode

import (
	"bytes"
	"errors"
	"io"
	"math/big"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func Decode(data []byte) (Value, error) {
	return NewDecoder(bytes.NewReader(data)).Decode()
}

func (d *Decoder) Decode() (Value, error) {
	buf := make([]byte, 1)
	_, err := d.r.Read(buf)
	if err != nil {
		return nil, err
	}

	switch buf[0] {
	case 'i':
		return d.decodeInteger()
	case 'l':
		return d.decodeList()
	case 'd':
		return d.decodeDict()
	default:
		if buf[0] >= '0' && buf[0] <= '9' {
			return d.decodeByteString(buf[0])
		}
		return nil, errors.New("invalid bencode: unexpected byte")
	}
}

func (d *Decoder) decodeByteString(firstByte byte) (ByteString, error) {
	lengthBuf := []byte{firstByte}
	for {
		buf := make([]byte, 1)
		_, err := d.r.Read(buf)
		if err != nil {
			return nil, err
		}
		if buf[0] == ':' {
			break
		}
		if buf[0] < '0' || buf[0] > '9' {
			return nil, errors.New("invalid bencode: invalid byte string length")
		}
		lengthBuf = append(lengthBuf, buf[0])
	}

	length := new(big.Int)
	_, ok := length.SetString(string(lengthBuf), 10)
	if !ok {
		return nil, errors.New("invalid bencode: invalid byte string length")
	}

	if !length.IsInt64() {
		return nil, errors.New("invalid bencode: byte string too large")
	}
	l := length.Int64()
	if l < 0 {
		return nil, errors.New("invalid bencode: negative byte string length")
	}

	data := make([]byte, l)
	if l > 0 {
		_, err := io.ReadFull(d.r, data)
		if err != nil {
			return nil, err
		}
	}
	return ByteString(data), nil
}

func (d *Decoder) decodeInteger() (Integer, error) {
	var buf bytes.Buffer
	for {
		b := make([]byte, 1)
		_, err := d.r.Read(b)
		if err != nil {
			return Integer{}, err
		}
		if b[0] == 'e' {
			break
		}
		buf.WriteByte(b[0])
	}

	s := buf.String()
	if s == "" {
		return Integer{}, errors.New("invalid bencode: empty integer")
	}

	i := Integer{new(big.Int)}
	_, ok := i.SetString(s, 10)
	if !ok {
		return Integer{}, errors.New("invalid bencode: invalid integer")
	}

	return i, nil
}

func (d *Decoder) decodeList() (List, error) {
	var list List
	for {
		buf := make([]byte, 1)
		_, err := d.r.Read(buf)
		if err != nil {
			return nil, err
		}
		if buf[0] == 'e' {
			break
		}
		switch buf[0] {
		case 'i':
			val, err := d.decodeInteger()
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		case 'l':
			val, err := d.decodeList()
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		case 'd':
			val, err := d.decodeDict()
			if err != nil {
				return nil, err
			}
			list = append(list, val)
		default:
			if buf[0] >= '0' && buf[0] <= '9' {
				val, err := d.decodeByteString(buf[0])
				if err != nil {
					return nil, err
				}
				list = append(list, val)
			} else {
				return nil, errors.New("invalid bencode: unexpected byte in list")
			}
		}
	}
	return list, nil
}

func (d *Decoder) decodeDict() (Dict, error) {
	dict := make(Dict)
	for {
		buf := make([]byte, 1)
		_, err := d.r.Read(buf)
		if err != nil {
			return nil, err
		}
		if buf[0] == 'e' {
			break
		}
		if buf[0] < '0' || buf[0] > '9' {
			return nil, errors.New("invalid bencode: dict key must be byte string")
		}
		key, err := d.decodeByteString(buf[0])
		if err != nil {
			return nil, err
		}
		val, err := d.Decode()
		if err != nil {
			return nil, err
		}
		dict[string(key)] = val
	}
	return dict, nil
}

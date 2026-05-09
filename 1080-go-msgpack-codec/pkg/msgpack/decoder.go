package msgpack

import (
	"bytes"
	"encoding/binary"
	"errors"
	"time"
)

type Decoder struct {
	r *bytes.Reader
}

func NewDecoder(data []byte) *Decoder {
	return &Decoder{r: bytes.NewReader(data)}
}

func Unmarshal(data []byte) (interface{}, error) {
	dec := NewDecoder(data)
	return dec.decode()
}

func (d *Decoder) decode() (interface{}, error) {
	type frame struct {
		container interface{}
		index     int
		count     int
		isMap     bool
		mapKey    interface{}
	}

	type task struct {
		isValue bool
		target  interface{}
		index   int
	}

	var result interface{}
	var stack []frame
	var pending []task

	for {
		b, err := d.readByte()
		if err != nil {
			return nil, err
		}

		var value interface{}

		switch {
		case b >= FixIntPosMin && b <= FixIntPosMax:
			value = int64(b)

		case b >= FixMapMin && b <= FixMapMax:
			count := int(b & 0x0f)
			m := make(map[interface{}]interface{})
			if len(stack) == 0 && len(pending) == 0 {
				result = m
			}
			if count == 0 {
				value = m
			} else {
				stack = append(stack, frame{
					container: m,
					index:     0,
					count:     count,
					isMap:     true,
				})
				continue
			}

		case b >= FixArrayMin && b <= FixArrayMax:
			count := int(b & 0x0f)
			arr := make([]interface{}, count)
			if len(stack) == 0 && len(pending) == 0 {
				result = arr
			}
			if count == 0 {
				value = arr
			} else {
				stack = append(stack, frame{
					container: arr,
					index:     0,
					count:     count,
					isMap:     false,
				})
				continue
			}

		case b >= FixStrMin && b <= FixStrMax:
			length := int(b & 0x1f)
			s, err := d.readString(length)
			if err != nil {
				return nil, err
			}
			value = s

		default:
			switch b {
			case NilFormat:
				value = nil

			case FalseFormat:
				value = false

			case TrueFormat:
				value = true

			case Bin8Format:
				length, err := d.readUint8()
				if err != nil {
					return nil, err
				}
				value, err = d.readBytes(int(length))
				if err != nil {
					return nil, err
				}

			case Bin16Format:
				length, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				value, err = d.readBytes(int(length))
				if err != nil {
					return nil, err
				}

			case Bin32Format:
				length, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				value, err = d.readBytes(int(length))
				if err != nil {
					return nil, err
				}

			case Ext8Format:
				length, err := d.readUint8()
				if err != nil {
					return nil, err
				}
				value, err = d.readExt(int(length))
				if err != nil {
					return nil, err
				}

			case Ext16Format:
				length, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				value, err = d.readExt(int(length))
				if err != nil {
					return nil, err
				}

			case Ext32Format:
				length, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				value, err = d.readExt(int(length))
				if err != nil {
					return nil, err
				}

			case Float32Format:
				var f float32
				if err := binary.Read(d.r, binary.BigEndian, &f); err != nil {
					return nil, err
				}
				value = float64(f)

			case Float64Format:
				var f float64
				if err := binary.Read(d.r, binary.BigEndian, &f); err != nil {
					return nil, err
				}
				value = f

			case Uint8Format:
				v, err := d.readUint8()
				if err != nil {
					return nil, err
				}
				value = int64(v)

			case Uint16Format:
				v, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				value = int64(v)

			case Uint32Format:
				v, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				value = int64(v)

			case Uint64Format:
				v, err := d.readUint64()
				if err != nil {
					return nil, err
				}
				value = int64(v)

			case Int8Format:
				v, err := d.readUint8()
				if err != nil {
					return nil, err
				}
				value = int64(int8(v))

			case Int16Format:
				v, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				value = int64(int16(v))

			case Int32Format:
				v, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				value = int64(int32(v))

			case Int64Format:
				v, err := d.readUint64()
				if err != nil {
					return nil, err
				}
				value = int64(v)

			case FixExt1Format:
				value, err = d.readExt(1)
				if err != nil {
					return nil, err
				}

			case FixExt2Format:
				value, err = d.readExt(2)
				if err != nil {
					return nil, err
				}

			case FixExt4Format:
				value, err = d.readExt(4)
				if err != nil {
					return nil, err
				}

			case FixExt8Format:
				value, err = d.readExt(8)
				if err != nil {
					return nil, err
				}

			case FixExt16Format:
				value, err = d.readExt(16)
				if err != nil {
					return nil, err
				}

			case Str8Format:
				length, err := d.readUint8()
				if err != nil {
					return nil, err
				}
				value, err = d.readString(int(length))
				if err != nil {
					return nil, err
				}

			case Str16Format:
				length, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				value, err = d.readString(int(length))
				if err != nil {
					return nil, err
				}

			case Str32Format:
				length, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				value, err = d.readString(int(length))
				if err != nil {
					return nil, err
				}

			case Array16Format:
				count, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				arr := make([]interface{}, count)
				if len(stack) == 0 && len(pending) == 0 {
					result = arr
				}
				if count == 0 {
					value = arr
				} else {
					stack = append(stack, frame{
						container: arr,
						index:     0,
						count:     int(count),
						isMap:     false,
					})
					continue
				}

			case Array32Format:
				count, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				arr := make([]interface{}, count)
				if len(stack) == 0 && len(pending) == 0 {
					result = arr
				}
				if count == 0 {
					value = arr
				} else {
					stack = append(stack, frame{
						container: arr,
						index:     0,
						count:     int(count),
						isMap:     false,
					})
					continue
				}

			case Map16Format:
				count, err := d.readUint16()
				if err != nil {
					return nil, err
				}
				m := make(map[interface{}]interface{})
				if len(stack) == 0 && len(pending) == 0 {
					result = m
				}
				if count == 0 {
					value = m
				} else {
					stack = append(stack, frame{
						container: m,
						index:     0,
						count:     int(count),
						isMap:     true,
					})
					continue
				}

			case Map32Format:
				count, err := d.readUint32()
				if err != nil {
					return nil, err
				}
				m := make(map[interface{}]interface{})
				if len(stack) == 0 && len(pending) == 0 {
					result = m
				}
				if count == 0 {
					value = m
				} else {
					stack = append(stack, frame{
						container: m,
						index:     0,
						count:     int(count),
						isMap:     true,
					})
					continue
				}

			default:
				if b >= 0xe0 {
					value = int64(int8(b))
				} else {
					return nil, errors.New("msgpack: unknown format")
				}
			}
		}

	assign:
		if len(stack) == 0 {
			if result == nil {
				result = value
			}
			if d.r.Len() == 0 {
				return result, nil
			}
			continue
		}

		top := &stack[len(stack)-1]
		if top.isMap {
			if top.mapKey == nil {
				top.mapKey = value
				continue
			}
			m := top.container.(map[interface{}]interface{})
			m[top.mapKey] = value
			top.mapKey = nil
			top.index++
			if top.index >= top.count {
				value = top.container
				stack = stack[:len(stack)-1]
				goto assign
			}
		} else {
			arr := top.container.([]interface{})
			arr[top.index] = value
			top.index++
			if top.index >= top.count {
				value = top.container
				stack = stack[:len(stack)-1]
				goto assign
			}
		}
	}
}

func (d *Decoder) readByte() (byte, error) {
	b, err := d.r.ReadByte()
	if err != nil {
		return 0, errors.New("msgpack: unexpected EOF")
	}
	return b, nil
}

func (d *Decoder) readUint8() (uint8, error) {
	b, err := d.readByte()
	if err != nil {
		return 0, err
	}
	return uint8(b), nil
}

func (d *Decoder) readUint16() (uint16, error) {
	var v uint16
	if err := binary.Read(d.r, binary.BigEndian, &v); err != nil {
		return 0, errors.New("msgpack: unexpected EOF")
	}
	return v, nil
}

func (d *Decoder) readUint32() (uint32, error) {
	var v uint32
	if err := binary.Read(d.r, binary.BigEndian, &v); err != nil {
		return 0, errors.New("msgpack: unexpected EOF")
	}
	return v, nil
}

func (d *Decoder) readUint64() (uint64, error) {
	var v uint64
	if err := binary.Read(d.r, binary.BigEndian, &v); err != nil {
		return 0, errors.New("msgpack: unexpected EOF")
	}
	return v, nil
}

func (d *Decoder) readBytes(length int) ([]byte, error) {
	if length < 0 {
		return nil, errors.New("msgpack: negative length")
	}
	buf := make([]byte, length)
	if length == 0 {
		return buf, nil
	}
	n, err := d.r.Read(buf)
	if err != nil || n != length {
		return nil, errors.New("msgpack: unexpected EOF")
	}
	return buf, nil
}

func (d *Decoder) readString(length int) (string, error) {
	b, err := d.readBytes(length)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (d *Decoder) readExt(length int) (interface{}, error) {
	typeCode, err := d.readByte()
	if err != nil {
		return nil, err
	}
	data, err := d.readBytes(length)
	if err != nil {
		return nil, err
	}

	if int8(typeCode) == ExtTypeTimestamp {
		return d.decodeTimestamp(data)
	}

	return Ext{
		Type: int8(typeCode),
		Data: data,
	}, nil
}

func (d *Decoder) decodeTimestamp(data []byte) (time.Time, error) {
	switch len(data) {
	case 4:
		sec := binary.BigEndian.Uint32(data)
		return time.Unix(int64(sec), 0), nil

	case 8:
		combined := binary.BigEndian.Uint64(data)
		nsec := int64(combined >> 34)
		sec := int64(combined & ((1 << 34) - 1))
		return time.Unix(sec, nsec), nil

	case 12:
		nsec := binary.BigEndian.Uint32(data[0:4])
		sec := int64(binary.BigEndian.Uint64(data[4:12]))
		return time.Unix(sec, int64(nsec)), nil

	default:
		return time.Time{}, errors.New("msgpack: invalid timestamp length")
	}
}

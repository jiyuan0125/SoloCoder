package msgpack

import (
	"bytes"
	"encoding/binary"
	"math"
	"time"
)

type Encoder struct {
	w *bytes.Buffer
}

func NewEncoder() *Encoder {
	return &Encoder{w: bytes.NewBuffer(nil)}
}

func (e *Encoder) Bytes() []byte {
	return e.w.Bytes()
}

func Marshal(v interface{}) ([]byte, error) {
	enc := NewEncoder()
	if err := enc.encode(v); err != nil {
		return nil, err
	}
	return enc.Bytes(), nil
}

func (e *Encoder) encode(v interface{}) error {
	switch val := v.(type) {
	case nil:
		return e.encodeNil()
	case bool:
		return e.encodeBool(val)
	case int:
		return e.encodeInt(int64(val))
	case int8:
		return e.encodeInt(int64(val))
	case int16:
		return e.encodeInt(int64(val))
	case int32:
		return e.encodeInt(int64(val))
	case int64:
		return e.encodeInt(val)
	case uint:
		return e.encodeUint(uint64(val))
	case uint8:
		return e.encodeUint(uint64(val))
	case uint16:
		return e.encodeUint(uint64(val))
	case uint32:
		return e.encodeUint(uint64(val))
	case uint64:
		return e.encodeUint(val)
	case float32:
		return e.encodeFloat32(val)
	case float64:
		return e.encodeFloat64(val)
	case string:
		return e.encodeString(val)
	case []byte:
		return e.encodeBytes(val)
	case []interface{}:
		return e.encodeArray(val)
	case map[string]interface{}:
		return e.encodeMap(val)
	case map[interface{}]interface{}:
		return e.encodeMapInterface(val)
	case time.Time:
		return e.encodeTime(val)
	case Ext:
		return e.encodeExt(val)
	default:
		return &UnsupportedTypeError{Type: v}
	}
}

func (e *Encoder) encodeNil() error {
	return e.w.WriteByte(NilFormat)
}

func (e *Encoder) encodeBool(b bool) error {
	if b {
		return e.w.WriteByte(TrueFormat)
	}
	return e.w.WriteByte(FalseFormat)
}

func (e *Encoder) encodeInt(v int64) error {
	if v >= 0 {
		if v <= int64(FixIntPosMax) {
			return e.w.WriteByte(uint8(v))
		}
		if v <= math.MaxUint8 {
			if err := e.w.WriteByte(Uint8Format); err != nil {
				return err
			}
			return e.w.WriteByte(uint8(v))
		}
		if v <= math.MaxUint16 {
			if err := e.w.WriteByte(Uint16Format); err != nil {
				return err
			}
			return binary.Write(e.w, binary.BigEndian, uint16(v))
		}
		if v <= math.MaxUint32 {
			if err := e.w.WriteByte(Uint32Format); err != nil {
				return err
			}
			return binary.Write(e.w, binary.BigEndian, uint32(v))
		}
		if err := e.w.WriteByte(Uint64Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint64(v))
	}

	if v >= int64(FixIntNegMin) && v <= int64(FixIntNegMax) {
		return e.w.WriteByte(FixIntNegOffset + uint8(v+32))
	}
	if v >= math.MinInt8 {
		if err := e.w.WriteByte(Int8Format); err != nil {
			return err
		}
		return e.w.WriteByte(uint8(v))
	}
	if v >= math.MinInt16 {
		if err := e.w.WriteByte(Int16Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, int16(v))
	}
	if v >= math.MinInt32 {
		if err := e.w.WriteByte(Int32Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, int32(v))
	}
	if err := e.w.WriteByte(Int64Format); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, v)
}

func (e *Encoder) encodeUint(v uint64) error {
	if v <= uint64(FixIntPosMax) {
		return e.w.WriteByte(uint8(v))
	}
	if v <= math.MaxUint8 {
		if err := e.w.WriteByte(Uint8Format); err != nil {
			return err
		}
		return e.w.WriteByte(uint8(v))
	}
	if v <= math.MaxUint16 {
		if err := e.w.WriteByte(Uint16Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint16(v))
	}
	if v <= math.MaxUint32 {
		if err := e.w.WriteByte(Uint32Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint32(v))
	}
	if err := e.w.WriteByte(Uint64Format); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, v)
}

func (e *Encoder) encodeFloat32(v float32) error {
	if err := e.w.WriteByte(Float32Format); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, v)
}

func (e *Encoder) encodeFloat64(v float64) error {
	if err := e.w.WriteByte(Float64Format); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, v)
}

func (e *Encoder) encodeString(v string) error {
	length := len(v)
	if length <= 31 {
		return e.w.WriteByte(FixStrMin + uint8(length))
	} else if length <= math.MaxUint8 {
		if err := e.w.WriteByte(Str8Format); err != nil {
			return err
		}
		if err := e.w.WriteByte(uint8(length)); err != nil {
			return err
		}
	} else if length <= math.MaxUint16 {
		if err := e.w.WriteByte(Str16Format); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.BigEndian, uint16(length)); err != nil {
			return err
		}
	} else {
		if err := e.w.WriteByte(Str32Format); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.BigEndian, uint32(length)); err != nil {
			return err
		}
	}
	_, err := e.w.WriteString(v)
	return err
}

func (e *Encoder) encodeBytes(v []byte) error {
	length := len(v)
	if length <= math.MaxUint8 {
		if err := e.w.WriteByte(Bin8Format); err != nil {
			return err
		}
		if err := e.w.WriteByte(uint8(length)); err != nil {
			return err
		}
	} else if length <= math.MaxUint16 {
		if err := e.w.WriteByte(Bin16Format); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.BigEndian, uint16(length)); err != nil {
			return err
		}
	} else {
		if err := e.w.WriteByte(Bin32Format); err != nil {
			return err
		}
		if err := binary.Write(e.w, binary.BigEndian, uint32(length)); err != nil {
			return err
		}
	}
	_, err := e.w.Write(v)
	return err
}

func (e *Encoder) encodeArray(v []interface{}) error {
	length := len(v)
	if err := e.writeArrayHeader(length); err != nil {
		return err
	}
	for _, item := range v {
		if err := e.encode(item); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeArrayHeader(length int) error {
	if length <= 15 {
		return e.w.WriteByte(FixArrayMin + uint8(length))
	} else if length <= math.MaxUint16 {
		if err := e.w.WriteByte(Array16Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint16(length))
	} else {
		if err := e.w.WriteByte(Array32Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint32(length))
	}
}

func (e *Encoder) encodeMap(v map[string]interface{}) error {
	length := len(v)
	if err := e.writeMapHeader(length); err != nil {
		return err
	}
	for k, val := range v {
		if err := e.encodeString(k); err != nil {
			return err
		}
		if err := e.encode(val); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeMapInterface(v map[interface{}]interface{}) error {
	length := len(v)
	if err := e.writeMapHeader(length); err != nil {
		return err
	}
	for k, val := range v {
		if err := e.encode(k); err != nil {
			return err
		}
		if err := e.encode(val); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) writeMapHeader(length int) error {
	if length <= 15 {
		return e.w.WriteByte(FixMapMin + uint8(length))
	} else if length <= math.MaxUint16 {
		if err := e.w.WriteByte(Map16Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint16(length))
	} else {
		if err := e.w.WriteByte(Map32Format); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint32(length))
	}
}

func (e *Encoder) encodeTime(t time.Time) error {
	sec := t.Unix()
	nsec := t.Nanosecond()

	typeCode := int8(-1)

	if nsec == 0 && sec >= 0 && sec <= math.MaxUint32 {
		if err := e.w.WriteByte(FixExt4Format); err != nil {
			return err
		}
		if err := e.w.WriteByte(byte(typeCode)); err != nil {
			return err
		}
		return binary.Write(e.w, binary.BigEndian, uint32(sec))
	}

	if sec >= 0 && sec <= (1<<34)-1 && nsec <= (1<<30)-1 {
		if err := e.w.WriteByte(FixExt8Format); err != nil {
			return err
		}
		if err := e.w.WriteByte(byte(typeCode)); err != nil {
			return err
		}
		combined := (uint64(nsec) << 34) | uint64(sec)
		return binary.Write(e.w, binary.BigEndian, combined)
	}

	if err := e.w.WriteByte(Ext8Format); err != nil {
		return err
	}
	if err := e.w.WriteByte(12); err != nil {
		return err
	}
	if err := e.w.WriteByte(byte(typeCode)); err != nil {
		return err
	}
	if err := binary.Write(e.w, binary.BigEndian, uint32(nsec)); err != nil {
		return err
	}
	return binary.Write(e.w, binary.BigEndian, sec)
}

func (e *Encoder) encodeExt(ext Ext) error {
	length := len(ext.Data)
	switch length {
	case 1:
		if err := e.w.WriteByte(FixExt1Format); err != nil {
			return err
		}
	case 2:
		if err := e.w.WriteByte(FixExt2Format); err != nil {
			return err
		}
	case 4:
		if err := e.w.WriteByte(FixExt4Format); err != nil {
			return err
		}
	case 8:
		if err := e.w.WriteByte(FixExt8Format); err != nil {
			return err
		}
	case 16:
		if err := e.w.WriteByte(FixExt16Format); err != nil {
			return err
		}
	default:
		if length <= math.MaxUint8 {
			if err := e.w.WriteByte(Ext8Format); err != nil {
				return err
			}
			if err := e.w.WriteByte(uint8(length)); err != nil {
				return err
			}
		} else if length <= math.MaxUint16 {
			if err := e.w.WriteByte(Ext16Format); err != nil {
				return err
			}
			if err := binary.Write(e.w, binary.BigEndian, uint16(length)); err != nil {
				return err
			}
		} else {
			if err := e.w.WriteByte(Ext32Format); err != nil {
				return err
			}
			if err := binary.Write(e.w, binary.BigEndian, uint32(length)); err != nil {
				return err
			}
		}
	}
	if err := e.w.WriteByte(byte(ext.Type)); err != nil {
		return err
	}
	_, err := e.w.Write(ext.Data)
	return err
}

type UnsupportedTypeError struct {
	Type interface{}
}

func (e *UnsupportedTypeError) Error() string {
	return "msgpack: unsupported type"
}

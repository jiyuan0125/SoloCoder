package cbor

import (
	"encoding/binary"
	"errors"
	"math"
	"math/big"
)

func Marshal(v Value) ([]byte, error) {
	e := &encoder{buf: make([]byte, 0, 64)}
	if err := e.encode(v); err != nil {
		return nil, err
	}
	return e.buf, nil
}

type encoder struct {
	buf []byte
}

func (e *encoder) encode(v Value) error {
	switch val := v.(type) {
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
	case float32:
		return e.encodeFloat32(val)
	case float64:
		return e.encodeFloat64(val)
	case bool:
		return e.encodeBool(val)
	case nil:
		e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|SimpleNull)
		return nil
	case SimpleValue:
		return e.encodeSimpleValue(byte(val))
	case []byte:
		return e.encodeByteString(val)
	case string:
		return e.encodeTextString(val)
	case []Value:
		return e.encodeArray(val)
	case []interface{}:
		arr := make([]Value, len(val))
		for i, vv := range val {
			arr[i] = vv
		}
		return e.encodeArray(arr)
	case *Map:
		return e.encodeMap(val)
	case *BigInt:
		return e.encodeBigInt(val)
	default:
		return errors.New("unsupported type")
	}
}

func (e *encoder) encodeUint(n uint64) error {
	e.writeHead(MajorTypeUnsignedInt, n)
	return nil
}

func (e *encoder) encodeInt(n int64) error {
	if n >= 0 {
		e.writeHead(MajorTypeUnsignedInt, uint64(n))
	} else {
		e.writeHead(MajorTypeNegativeInt, uint64(-n-1))
	}
	return nil
}

func (e *encoder) encodeBigInt(bi *BigInt) error {
	if bi.Negative {
		tag := uint64(3)
		e.writeHead(MajorTypeTag, tag)
		neg := new(big.Int).Neg(bi.Int)
		neg = neg.Add(neg, big.NewInt(1))
		return e.encodeBigIntBytes(neg.Bytes())
	} else {
		tag := uint64(2)
		e.writeHead(MajorTypeTag, tag)
		return e.encodeBigIntBytes(bi.Int.Bytes())
	}
}

func (e *encoder) encodeBigIntBytes(b []byte) error {
	e.writeHead(MajorTypeByteString, uint64(len(b)))
	e.buf = append(e.buf, b...)
	return nil
}

func (e *encoder) encodeBool(b bool) error {
	if b {
		e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|SimpleTrue)
	} else {
		e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|SimpleFalse)
	}
	return nil
}

func (e *encoder) encodeSimpleValue(v byte) error {
	if v < 24 {
		e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|v)
	} else {
		e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|AdditionalInfo8Bit)
		e.buf = append(e.buf, v)
	}
	return nil
}

func (e *encoder) encodeFloat32(f float32) error {
	e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|SimpleSingleFloat)
	bits := math.Float32bits(f)
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, bits)
	e.buf = append(e.buf, buf...)
	return nil
}

func (e *encoder) encodeFloat64(f float64) error {
	e.buf = append(e.buf, (MajorTypeSimpleFloat<<5)|SimpleDoubleFloat)
	bits := math.Float64bits(f)
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, bits)
	e.buf = append(e.buf, buf...)
	return nil
}

func (e *encoder) encodeByteString(b []byte) error {
	e.writeHead(MajorTypeByteString, uint64(len(b)))
	e.buf = append(e.buf, b...)
	return nil
}

func (e *encoder) encodeTextString(s string) error {
	e.writeHead(MajorTypeTextString, uint64(len(s)))
	e.buf = append(e.buf, []byte(s)...)
	return nil
}

func (e *encoder) encodeArray(arr []Value) error {
	e.writeHead(MajorTypeArray, uint64(len(arr)))
	for _, v := range arr {
		if err := e.encode(v); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) encodeMap(m *Map) error {
	e.writeHead(MajorTypeMap, uint64(m.Len()))
	for _, entry := range m.Entries {
		if err := e.encode(entry.Key); err != nil {
			return err
		}
		if err := e.encode(entry.Value); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) writeHead(major byte, value uint64) {
	firstByte := major << 5
	switch {
	case value < 24:
		e.buf = append(e.buf, firstByte|byte(value))
	case value <= 0xFF:
		e.buf = append(e.buf, firstByte|AdditionalInfo8Bit)
		e.buf = append(e.buf, byte(value))
	case value <= 0xFFFF:
		e.buf = append(e.buf, firstByte|AdditionalInfo16Bit)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(value))
		e.buf = append(e.buf, buf...)
	case value <= 0xFFFFFFFF:
		e.buf = append(e.buf, firstByte|AdditionalInfo32Bit)
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(value))
		e.buf = append(e.buf, buf...)
	default:
		e.buf = append(e.buf, firstByte|AdditionalInfo64Bit)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, value)
		e.buf = append(e.buf, buf...)
	}
}

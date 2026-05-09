package cbor

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"math/big"
)

func Unmarshal(data []byte) (Value, error) {
	d := &decoder{data: data, pos: 0}
	return d.decode()
}

type decoder struct {
	data []byte
	pos  int
}

func (d *decoder) decode() (Value, error) {
	if d.pos >= len(d.data) {
		return nil, io.EOF
	}
	firstByte := d.data[d.pos]
	d.pos++
	major := firstByte >> 5
	addInfo := firstByte & 0x1F

	switch major {
	case MajorTypeUnsignedInt:
		return d.decodeUint(addInfo)
	case MajorTypeNegativeInt:
		val, err := d.decodeUint(addInfo)
		if err != nil {
			return nil, err
		}
		n, ok := val.(uint64)
		if !ok {
			bi, ok := val.(*BigInt)
			if ok {
				bi.Negative = true
				return bi, nil
			}
			return nil, errors.New("invalid negative int")
		}
		if n == math.MaxUint64 {
			bi := big.NewInt(0)
			bi.SetUint64(n)
			bi.Add(bi, big.NewInt(1))
			return &BigInt{Negative: true, Int: bi}, nil
		}
		return -int64(n) - 1, nil
	case MajorTypeByteString:
		return d.decodeByteString(addInfo)
	case MajorTypeTextString:
		return d.decodeTextString(addInfo)
	case MajorTypeArray:
		return d.decodeArray(addInfo)
	case MajorTypeMap:
		return d.decodeMap(addInfo)
	case MajorTypeTag:
		return d.decodeTag(addInfo)
	case MajorTypeSimpleFloat:
		return d.decodeSimpleFloat(addInfo)
	default:
		return nil, errors.New("unknown major type")
	}
}

func (d *decoder) decodeUint(addInfo byte) (Value, error) {
	switch {
	case addInfo < 24:
		return uint64(addInfo), nil
	case addInfo == AdditionalInfo8Bit:
		if d.pos+1 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		val := uint64(d.data[d.pos])
		d.pos++
		return val, nil
	case addInfo == AdditionalInfo16Bit:
		if d.pos+2 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		val := binary.BigEndian.Uint16(d.data[d.pos : d.pos+2])
		d.pos += 2
		return uint64(val), nil
	case addInfo == AdditionalInfo32Bit:
		if d.pos+4 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		val := binary.BigEndian.Uint32(d.data[d.pos : d.pos+4])
		d.pos += 4
		return uint64(val), nil
	case addInfo == AdditionalInfo64Bit:
		if d.pos+8 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		val := binary.BigEndian.Uint64(d.data[d.pos : d.pos+8])
		d.pos += 8
		return uint64(val), nil
	default:
		return nil, errors.New("invalid additional info for uint")
	}
}

func (d *decoder) decodeLength(addInfo byte) (int, bool, error) {
	if addInfo == AdditionalInfoIndef {
		return 0, true, nil
	}
	val, err := d.decodeUint(addInfo)
	if err != nil {
		return 0, false, err
	}
	n, ok := val.(uint64)
	if !ok {
		return 0, false, errors.New("invalid length")
	}
	return int(n), false, nil
}

func (d *decoder) decodeByteString(addInfo byte) ([]byte, error) {
	length, indef, err := d.decodeLength(addInfo)
	if err != nil {
		return nil, err
	}
	if !indef {
		if d.pos+length > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		data := make([]byte, length)
		copy(data, d.data[d.pos:d.pos+length])
		d.pos += length
		return data, nil
	}
	result := make([]byte, 0)
	for {
		if d.pos >= len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		if d.data[d.pos] == 0xFF {
			d.pos++
			break
		}
		chunkFirst := d.data[d.pos]
		d.pos++
		if chunkFirst>>5 != MajorTypeByteString {
			return nil, errors.New("invalid indefinite byte string chunk")
		}
		chunkAddInfo := chunkFirst & 0x1F
		chunkLen, chunkIndef, err := d.decodeLength(chunkAddInfo)
		if err != nil {
			return nil, err
		}
		if chunkIndef {
			return nil, errors.New("nested indefinite length not allowed")
		}
		if d.pos+chunkLen > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		result = append(result, d.data[d.pos:d.pos+chunkLen]...)
		d.pos += chunkLen
	}
	return result, nil
}

func (d *decoder) decodeTextString(addInfo byte) (string, error) {
	length, indef, err := d.decodeLength(addInfo)
	if err != nil {
		return "", err
	}
	if !indef {
		if d.pos+length > len(d.data) {
			return "", io.ErrUnexpectedEOF
		}
		data := string(d.data[d.pos : d.pos+length])
		d.pos += length
		return data, nil
	}
	result := make([]byte, 0)
	for {
		if d.pos >= len(d.data) {
			return "", io.ErrUnexpectedEOF
		}
		if d.data[d.pos] == 0xFF {
			d.pos++
			break
		}
		chunkFirst := d.data[d.pos]
		d.pos++
		if chunkFirst>>5 != MajorTypeTextString {
			return "", errors.New("invalid indefinite text string chunk")
		}
		chunkAddInfo := chunkFirst & 0x1F
		chunkLen, chunkIndef, err := d.decodeLength(chunkAddInfo)
		if err != nil {
			return "", err
		}
		if chunkIndef {
			return "", errors.New("nested indefinite length not allowed")
		}
		if d.pos+chunkLen > len(d.data) {
			return "", io.ErrUnexpectedEOF
		}
		result = append(result, d.data[d.pos:d.pos+chunkLen]...)
		d.pos += chunkLen
	}
	return string(result), nil
}

func (d *decoder) decodeArray(addInfo byte) ([]Value, error) {
	length, indef, err := d.decodeLength(addInfo)
	if err != nil {
		return nil, err
	}
	if !indef {
		arr := make([]Value, 0, length)
		for i := 0; i < length; i++ {
			v, err := d.decode()
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil
	}
	arr := make([]Value, 0)
	for {
		if d.pos >= len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		if d.data[d.pos] == 0xFF {
			d.pos++
			break
		}
		v, err := d.decode()
		if err != nil {
			return nil, err
		}
		arr = append(arr, v)
	}
	return arr, nil
}

func (d *decoder) decodeMap(addInfo byte) (*Map, error) {
	length, indef, err := d.decodeLength(addInfo)
	if err != nil {
		return nil, err
	}
	m := NewMap()
	if !indef {
		for i := 0; i < length; i++ {
			key, err := d.decode()
			if err != nil {
				return nil, err
			}
			val, err := d.decode()
			if err != nil {
				return nil, err
			}
			m.Add(key, val)
		}
		return m, nil
	}
	for {
		if d.pos >= len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		if d.data[d.pos] == 0xFF {
			d.pos++
			break
		}
		key, err := d.decode()
		if err != nil {
			return nil, err
		}
		val, err := d.decode()
		if err != nil {
			return nil, err
		}
		m.Add(key, val)
	}
	return m, nil
}

func (d *decoder) decodeTag(addInfo byte) (Value, error) {
	tagNum, err := d.decodeUint(addInfo)
	if err != nil {
		return nil, err
	}
	tag, ok := tagNum.(uint64)
	if !ok {
		return nil, errors.New("invalid tag number")
	}
	value, err := d.decode()
	if err != nil {
		return nil, err
	}
	if tag == 2 || tag == 3 {
		bs, ok := value.([]byte)
		if !ok {
			return nil, errors.New("bigint tag requires byte string")
		}
		bi := new(big.Int).SetBytes(bs)
		if tag == 3 {
			bi = bi.Neg(bi)
			bi = bi.Add(bi, big.NewInt(-1))
			if bi.Sign() < 0 {
				return &BigInt{Negative: true, Int: new(big.Int).Neg(bi)}, nil
			}
			return &BigInt{Negative: true, Int: bi}, nil
		}
		return &BigInt{Negative: false, Int: bi}, nil
	}
	return value, nil
}

func (d *decoder) decodeSimpleFloat(addInfo byte) (Value, error) {
	switch {
	case addInfo < 24:
		switch addInfo {
		case SimpleFalse:
			return false, nil
		case SimpleTrue:
			return true, nil
		case SimpleNull:
			return nil, nil
		case SimpleUndefined:
			return SimpleValue(addInfo), nil
		default:
			return SimpleValue(addInfo), nil
		}
	case addInfo == AdditionalInfo8Bit:
		if d.pos+1 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		val := d.data[d.pos]
		d.pos++
		return SimpleValue(val), nil
	case addInfo == SimpleHalfFloat:
		return d.decodeHalfFloat()
	case addInfo == SimpleSingleFloat:
		if d.pos+4 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		bits := binary.BigEndian.Uint32(d.data[d.pos : d.pos+4])
		d.pos += 4
		return float64(math.Float32frombits(bits)), nil
	case addInfo == SimpleDoubleFloat:
		if d.pos+8 > len(d.data) {
			return nil, io.ErrUnexpectedEOF
		}
		bits := binary.BigEndian.Uint64(d.data[d.pos : d.pos+8])
		d.pos += 8
		return math.Float64frombits(bits), nil
	default:
		return nil, errors.New("invalid simple/float encoding")
	}
}

func (d *decoder) decodeHalfFloat() (float64, error) {
	if d.pos+2 > len(d.data) {
		return 0, io.ErrUnexpectedEOF
	}
	bits := binary.BigEndian.Uint16(d.data[d.pos : d.pos+2])
	d.pos += 2
	sign := bits >> 15
	exp := (bits >> 10) & 0x1F
	mant := bits & 0x3FF

	var f float64
	switch {
	case exp == 0 && mant == 0:
		if sign == 0 {
			return 0.0, nil
		}
		return math.Copysign(0.0, -1), nil
	case exp == 0 && mant != 0:
		f = float64(mant) / float64(1<<10)
		return math.Ldexp(math.Copysign(f, float64(1-(sign<<1))), -14), nil
	case exp == 0x1F && mant == 0:
		if sign == 0 {
			return math.Inf(1), nil
		}
		return math.Inf(-1), nil
	case exp == 0x1F && mant != 0:
		return math.NaN(), nil
	default:
		f = float64(mant)/float64(1<<10) + 1.0
		return math.Ldexp(math.Copysign(f, float64(1-(sign<<1))), int(exp)-15), nil
	}
}

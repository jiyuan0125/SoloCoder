package flatbuffers

import (
	"encoding/binary"
	"errors"
	"math"
)

var ErrBufferTooSmall = errors.New("buffer too small")

type ArrayInfo struct {
	Pos    int
	Length int
}

type FieldType int

const (
	TypeInt8 FieldType = iota
	TypeUint8
	TypeInt16
	TypeUint16
	TypeInt32
	TypeUint32
	TypeInt64
	TypeUint64
	TypeFloat32
	TypeFloat64
	TypeBool
	TypeString
	TypeTable
	TypeScalarArray
	TypeTableArray
)

func readUOffsetT(buf []byte, pos int) (uint32, error) {
	if pos+4 > len(buf) {
		return 0, ErrBufferTooSmall
	}
	return binary.LittleEndian.Uint32(buf[pos : pos+4]), nil
}

func readVOffsetT(buf []byte, pos int) (uint16, error) {
	if pos+2 > len(buf) {
		return 0, ErrBufferTooSmall
	}
	return binary.LittleEndian.Uint16(buf[pos : pos+2]), nil
}

func readSOffsetT(buf []byte, pos int) (int32, error) {
	if pos+4 > len(buf) {
		return 0, ErrBufferTooSmall
	}
	return int32(binary.LittleEndian.Uint32(buf[pos : pos+4])), nil
}

func GetRootAsTable(buf []byte) (int, error) {
	rootOffset, err := readUOffsetT(buf, 0)
	if err != nil {
		return 0, err
	}
	tablePos := int(rootOffset)
	return tablePos, nil
}

func GetTableVOffset(buf []byte, tablePos int) (int, error) {
	soffset, err := readSOffsetT(buf, tablePos)
	if err != nil {
		return 0, err
	}
	vtablePos := tablePos + int(soffset)
	return vtablePos, nil
}

func GetFieldOffset(buf []byte, tablePos int, fieldIndex int) (uint16, error) {
	vtablePos, err := GetTableVOffset(buf, tablePos)
	if err != nil {
		return 0, err
	}

	vtableLen, err := readVOffsetT(buf, vtablePos)
	if err != nil {
		return 0, err
	}

	fieldEntryPos := vtablePos + 4 + fieldIndex*2
	if fieldEntryPos+2 > vtablePos+int(vtableLen) {
		return 0, nil
	}

	offset, err := readVOffsetT(buf, fieldEntryPos)
	if err != nil {
		return 0, err
	}
	return offset, nil
}

func HasField(buf []byte, tablePos int, fieldIndex int) (bool, error) {
	offset, err := GetFieldOffset(buf, tablePos, fieldIndex)
	if err != nil {
		return false, err
	}
	return offset != 0, nil
}

func GetField(buf []byte, tablePos int, fieldIndex int, fieldType FieldType) (interface{}, error) {
	offset, err := GetFieldOffset(buf, tablePos, fieldIndex)
	if err != nil {
		return nil, err
	}
	if offset == 0 {
		return getDefaultValue(fieldType), nil
	}

	fieldPos := tablePos + int(offset)

	switch fieldType {
	case TypeInt8:
		if fieldPos+1 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int8(buf[fieldPos]), nil
	case TypeUint8:
		if fieldPos+1 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return buf[fieldPos], nil
	case TypeInt16:
		if fieldPos+2 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int16(binary.LittleEndian.Uint16(buf[fieldPos : fieldPos+2])), nil
	case TypeUint16:
		if fieldPos+2 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return binary.LittleEndian.Uint16(buf[fieldPos : fieldPos+2]), nil
	case TypeInt32:
		if fieldPos+4 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int32(binary.LittleEndian.Uint32(buf[fieldPos : fieldPos+4])), nil
	case TypeUint32:
		if fieldPos+4 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return binary.LittleEndian.Uint32(buf[fieldPos : fieldPos+4]), nil
	case TypeInt64:
		if fieldPos+8 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int64(binary.LittleEndian.Uint64(buf[fieldPos : fieldPos+8])), nil
	case TypeUint64:
		if fieldPos+8 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return binary.LittleEndian.Uint64(buf[fieldPos : fieldPos+8]), nil
	case TypeFloat32:
		if fieldPos+4 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		bits := binary.LittleEndian.Uint32(buf[fieldPos : fieldPos+4])
		return math.Float32frombits(bits), nil
	case TypeFloat64:
		if fieldPos+8 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		bits := binary.LittleEndian.Uint64(buf[fieldPos : fieldPos+8])
		return math.Float64frombits(bits), nil
	case TypeBool:
		if fieldPos+1 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return buf[fieldPos] != 0, nil
	case TypeString:
		soffset, err := readSOffsetT(buf, fieldPos)
		if err != nil {
			return nil, err
		}
		stringPos := fieldPos + int(soffset)
		length, err := readUOffsetT(buf, stringPos)
		if err != nil {
			return nil, err
		}
		stringDataPos := stringPos + 4
		if stringDataPos+int(length) > len(buf) {
			return "", ErrBufferTooSmall
		}
		return string(buf[stringDataPos : stringDataPos+int(length)]), nil
	case TypeTable:
		soffset, err := readSOffsetT(buf, fieldPos)
		if err != nil {
			return nil, err
		}
		nestedTablePos := fieldPos + int(soffset)
		return nestedTablePos, nil
	case TypeScalarArray:
		soffset, err := readSOffsetT(buf, fieldPos)
		if err != nil {
			return nil, err
		}
		arrayPos := fieldPos + int(soffset)
		length, err := readUOffsetT(buf, arrayPos)
		if err != nil {
			return nil, err
		}
		return &ArrayInfo{Pos: arrayPos, Length: int(length)}, nil
	case TypeTableArray:
		soffset, err := readSOffsetT(buf, fieldPos)
		if err != nil {
			return nil, err
		}
		arrayPos := fieldPos + int(soffset)
		length, err := readUOffsetT(buf, arrayPos)
		if err != nil {
			return nil, err
		}
		return &ArrayInfo{Pos: arrayPos, Length: int(length)}, nil
	default:
		return nil, errors.New("unsupported field type")
	}
}

func getDefaultValue(fieldType FieldType) interface{} {
	switch fieldType {
	case TypeInt8:
		return int8(0)
	case TypeUint8:
		return uint8(0)
	case TypeInt16:
		return int16(0)
	case TypeUint16:
		return uint16(0)
	case TypeInt32:
		return int32(0)
	case TypeUint32:
		return uint32(0)
	case TypeInt64:
		return int64(0)
	case TypeUint64:
		return uint64(0)
	case TypeFloat32:
		return float32(0)
	case TypeFloat64:
		return float64(0)
	case TypeBool:
		return false
	case TypeString:
		return ""
	case TypeTable:
		return 0
	case TypeScalarArray:
		return &ArrayInfo{}
	case TypeTableArray:
		return &ArrayInfo{}
	default:
		return nil
	}
}

func GetArrayLength(arrayPos int) int {
	return 0
}

func ReadScalarArrayElement(buf []byte, arrayPos int, index int, fieldType FieldType) (interface{}, error) {
	elementPos := arrayPos + 4
	switch fieldType {
	case TypeInt8:
		pos := elementPos + index*1
		if pos+1 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int8(buf[pos]), nil
	case TypeUint8:
		pos := elementPos + index*1
		if pos+1 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return buf[pos], nil
	case TypeInt16:
		pos := elementPos + index*2
		if pos+2 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int16(binary.LittleEndian.Uint16(buf[pos : pos+2])), nil
	case TypeUint16:
		pos := elementPos + index*2
		if pos+2 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return binary.LittleEndian.Uint16(buf[pos : pos+2]), nil
	case TypeInt32:
		pos := elementPos + index*4
		if pos+4 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int32(binary.LittleEndian.Uint32(buf[pos : pos+4])), nil
	case TypeUint32:
		pos := elementPos + index*4
		if pos+4 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return binary.LittleEndian.Uint32(buf[pos : pos+4]), nil
	case TypeInt64:
		pos := elementPos + index*8
		if pos+8 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return int64(binary.LittleEndian.Uint64(buf[pos : pos+8])), nil
	case TypeUint64:
		pos := elementPos + index*8
		if pos+8 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return binary.LittleEndian.Uint64(buf[pos : pos+8]), nil
	case TypeFloat32:
		pos := elementPos + index*4
		if pos+4 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		bits := binary.LittleEndian.Uint32(buf[pos : pos+4])
		return math.Float32frombits(bits), nil
	case TypeFloat64:
		pos := elementPos + index*8
		if pos+8 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		bits := binary.LittleEndian.Uint64(buf[pos : pos+8])
		return math.Float64frombits(bits), nil
	case TypeBool:
		pos := elementPos + index*1
		if pos+1 > len(buf) {
			return nil, ErrBufferTooSmall
		}
		return buf[pos] != 0, nil
	default:
		return nil, errors.New("unsupported scalar array element type")
	}
}

func ReadTableArrayElement(buf []byte, arrayPos int, index int) (int, error) {
	elementPos := arrayPos + 4 + index*4
	soffset, err := readSOffsetT(buf, elementPos)
	if err != nil {
		return 0, err
	}
	tablePos := elementPos + int(soffset)
	return tablePos, nil
}

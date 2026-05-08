package codec

type NumberType string

const (
	TypeInt32  NumberType = "int32"
	TypeInt64  NumberType = "int64"
	TypeUint32 NumberType = "uint32"
	TypeUint64 NumberType = "uint64"
)

type EncodeMode string

const (
	ModeVarint  EncodeMode = "varint"
	ModeLEB128  EncodeMode = "leb128"
	ModeZigZag  EncodeMode = "zigzag"
)

func EncodeBatchInt32(values []int32, mode EncodeMode) []byte {
	var result []byte
	for _, v := range values {
		var encoded []byte
		switch mode {
		case ModeLEB128:
			encoded = EncodeLEB128Int32(v)
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			encoded = EncodeVarintInt32(v)
		}
		result = append(result, encoded...)
	}
	return result
}

func EncodeBatchInt64(values []int64, mode EncodeMode) []byte {
	var result []byte
	for _, v := range values {
		var encoded []byte
		switch mode {
		case ModeLEB128:
			encoded = EncodeLEB128Int64(v)
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			encoded = EncodeVarintInt64(v)
		}
		result = append(result, encoded...)
	}
	return result
}

func EncodeBatchUint32(values []uint32, mode EncodeMode) []byte {
	var result []byte
	for _, v := range values {
		var encoded []byte
		switch mode {
		case ModeLEB128:
			encoded = EncodeLEB128Uint32(v)
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			encoded = EncodeVarintUint32(v)
		}
		result = append(result, encoded...)
	}
	return result
}

func EncodeBatchUint64(values []uint64, mode EncodeMode) []byte {
	var result []byte
	for _, v := range values {
		var encoded []byte
		switch mode {
		case ModeLEB128:
			encoded = EncodeLEB128Uint64(v)
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			encoded = EncodeVarintUint64(v)
		}
		result = append(result, encoded...)
	}
	return result
}

func DecodeBatchInt32(data []byte, mode EncodeMode) ([]int32, error) {
	var values []int32
	offset := 0
	for offset < len(data) {
		var val int32
		var n int
		var err error
		switch mode {
		case ModeLEB128:
			val, n, err = DecodeLEB128Int32(data[offset:])
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			val, n, err = DecodeVarintInt32(data[offset:])
		}
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		offset += n
	}
	return values, nil
}

func DecodeBatchInt64(data []byte, mode EncodeMode) ([]int64, error) {
	var values []int64
	offset := 0
	for offset < len(data) {
		var val int64
		var n int
		var err error
		switch mode {
		case ModeLEB128:
			val, n, err = DecodeLEB128Int64(data[offset:])
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			val, n, err = DecodeVarintInt64(data[offset:])
		}
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		offset += n
	}
	return values, nil
}

func DecodeBatchUint32(data []byte, mode EncodeMode) ([]uint32, error) {
	var values []uint32
	offset := 0
	for offset < len(data) {
		var val uint32
		var n int
		var err error
		switch mode {
		case ModeLEB128:
			val, n, err = DecodeLEB128Uint32(data[offset:])
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			val, n, err = DecodeVarintUint32(data[offset:])
		}
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		offset += n
	}
	return values, nil
}

func DecodeBatchUint64(data []byte, mode EncodeMode) ([]uint64, error) {
	var values []uint64
	offset := 0
	for offset < len(data) {
		var val uint64
		var n int
		var err error
		switch mode {
		case ModeLEB128:
			val, n, err = DecodeLEB128Uint64(data[offset:])
		case ModeVarint, ModeZigZag:
			fallthrough
		default:
			val, n, err = DecodeVarintUint64(data[offset:])
		}
		if err != nil {
			return nil, err
		}
		values = append(values, val)
		offset += n
	}
	return values, nil
}

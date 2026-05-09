package core

type WireType int

const (
	WireVarint          WireType = 0
	Wire64BitFixed      WireType = 1
	WireLengthDelimited WireType = 2
	WireStartGroup      WireType = 3
	WireEndGroup        WireType = 4
	Wire32BitFixed      WireType = 5
)

func (w WireType) String() string {
	switch w {
	case WireVarint:
		return "varint"
	case Wire64BitFixed:
		return "64-bit fixed"
	case WireLengthDelimited:
		return "length-delimited"
	case WireStartGroup:
		return "start group"
	case WireEndGroup:
		return "end group"
	case Wire32BitFixed:
		return "32-bit fixed"
	default:
		return "unknown"
	}
}

type ParsedField struct {
	FieldNumber  int
	WireType     WireType
	RawBytes     []byte
	RawByteSize  int
	Value        interface{}
	ValueType    string
}

type ParseOptions struct {
	MaxNestingDepth  int
	MaxLengthValue   int
}

func DefaultParseOptions() *ParseOptions {
	return &ParseOptions{
		MaxNestingDepth: 100,
		MaxLengthValue:  1024 * 1024,
	}
}

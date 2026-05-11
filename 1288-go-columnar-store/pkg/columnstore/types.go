package columnstore

import "fmt"

type ColumnType int

const (
	ColumnTypeUnknown ColumnType = iota
	ColumnTypeString
	ColumnTypeInt64
	ColumnTypeFloat64
)

func (t ColumnType) String() string {
	switch t {
	case ColumnTypeString:
		return "string"
	case ColumnTypeInt64:
		return "int64"
	case ColumnTypeFloat64:
		return "float64"
	default:
		return "unknown"
	}
}

func ParseColumnType(s string) (ColumnType, error) {
	switch s {
	case "string":
		return ColumnTypeString, nil
	case "int64":
		return ColumnTypeInt64, nil
	case "float64":
		return ColumnTypeFloat64, nil
	default:
		return ColumnTypeUnknown, fmt.Errorf("unknown column type: %s", s)
	}
}

type EncodingType int

const (
	EncodingUnknown EncodingType = iota
	EncodingDictionary
	EncodingRunLength
)

func (e EncodingType) String() string {
	switch e {
	case EncodingDictionary:
		return "dictionary"
	case EncodingRunLength:
		return "runlength"
	default:
		return "unknown"
	}
}

func ParseEncodingType(s string) (EncodingType, error) {
	switch s {
	case "dictionary":
		return EncodingDictionary, nil
	case "runlength":
		return EncodingRunLength, nil
	default:
		return EncodingUnknown, fmt.Errorf("unknown encoding type: %s", s)
	}
}

type ColumnSchema struct {
	Name     string
	Type     ColumnType
	Encoding EncodingType
}

type TableSchema struct {
	Name    string
	Columns []ColumnSchema
}

func (s *TableSchema) ColumnIndex(name string) int {
	for i, col := range s.Columns {
		if col.Name == name {
			return i
		}
	}
	return -1
}

type Value struct {
	Null   bool
	Str    string
	Int    int64
	Float  float64
}

func (v Value) Equals(other Value) bool {
	if v.Null && other.Null {
		return true
	}
	if v.Null != other.Null {
		return false
	}
	return v.Str == other.Str && v.Int == other.Int && v.Float == other.Float
}

func (v Value) Compare(other Value) int {
	if v.Null && other.Null {
		return 0
	}
	if v.Null {
		return -1
	}
	if other.Null {
		return 1
	}
	if v.Str != "" || other.Str != "" {
		if v.Str < other.Str {
			return -1
		} else if v.Str > other.Str {
			return 1
		}
		return 0
	}
	if v.Int != 0 || other.Int != 0 {
		if v.Int < other.Int {
			return -1
		} else if v.Int > other.Int {
			return 1
		}
		return 0
	}
	if v.Float < other.Float {
		return -1
	} else if v.Float > other.Float {
		return 1
	}
	return 0
}

type PredicateOp int

const (
	OpEq PredicateOp = iota
	OpNeq
	OpLt
	OpLte
	OpGt
	OpGte
	OpIsNull
	OpIsNotNull
)

type Predicate struct {
	Column string
	Op     PredicateOp
	Value  Value
}

type Row struct {
	Values []Value
}

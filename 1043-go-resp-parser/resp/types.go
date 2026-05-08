package resp

type ValueType int

const (
	TypeSimpleString ValueType = iota
	TypeError
	TypeInteger
	TypeBulkString
	TypeArray
	TypeNullBulkString
	TypeNullArray
)

type Value struct {
	Type        ValueType
	Str         string
	Int         int64
	Array       []Value
	IsNull      bool
}

const (
	CRLF           = "\r\n"
	MaxArraySize   = 1024 * 1024
	MaxStringSize  = 512 * 1024 * 1024
)

func NewSimpleString(s string) Value {
	return Value{Type: TypeSimpleString, Str: s}
}

func NewError(s string) Value {
	return Value{Type: TypeError, Str: s}
}

func NewInteger(i int64) Value {
	return Value{Type: TypeInteger, Int: i}
}

func NewBulkString(s string) Value {
	return Value{Type: TypeBulkString, Str: s}
}

func NewNullBulkString() Value {
	return Value{Type: TypeNullBulkString, IsNull: true}
}

func NewNullArray() Value {
	return Value{Type: TypeNullArray, IsNull: true}
}

func NewArray(arr []Value) Value {
	return Value{Type: TypeArray, Array: arr}
}

func (v Value) IsNullValue() bool {
	return v.Type == TypeNullBulkString || v.Type == TypeNullArray || (v.Type == TypeArray && v.IsNull)
}

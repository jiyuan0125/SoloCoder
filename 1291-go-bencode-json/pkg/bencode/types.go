package bencode

import (
	"math/big"
)

type Value interface{}

type ByteString []byte

func (b ByteString) String() string {
	return string(b)
}

type Integer struct {
	*big.Int
}

func NewInteger(n int64) Integer {
	i := Integer{big.NewInt(n)}
	return i
}

func (i Integer) String() string {
	return i.Int.String()
}

func (i Integer) MarshalJSON() ([]byte, error) {
	return []byte(i.Int.String()), nil
}

func (i *Integer) UnmarshalJSON(data []byte) error {
	if i.Int == nil {
		i.Int = new(big.Int)
	}
	s := string(data)
	if len(s) > 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	_, ok := i.SetString(s, 10)
	if !ok {
		return nil
	}
	return nil
}

type List []Value

type Dict map[string]Value

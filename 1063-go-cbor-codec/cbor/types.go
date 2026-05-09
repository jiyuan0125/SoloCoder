package cbor

import (
	"math/big"
)

type Value interface{}

type MapEntry struct {
	Key   Value
	Value Value
}

type Map struct {
	Entries []MapEntry
}

func NewMap() *Map {
	return &Map{Entries: make([]MapEntry, 0)}
}

func (m *Map) Add(key, value Value) {
	m.Entries = append(m.Entries, MapEntry{Key: key, Value: value})
}

func (m *Map) Len() int {
	return len(m.Entries)
}

type BigInt struct {
	Negative bool
	Int      *big.Int
}

func NewBigInt(negative bool, value *big.Int) *BigInt {
	return &BigInt{Negative: negative, Int: new(big.Int).Set(value)}
}

func isSimpleValue(v byte) bool {
	return v < 32 && v != SimpleHalfFloat && v != SimpleSingleFloat && v != SimpleDoubleFloat
}

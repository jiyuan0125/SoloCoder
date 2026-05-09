package msgpack

const (
	FixIntPosMin uint8 = 0x00
	FixIntPosMax uint8 = 0x7f
	FixIntNegMin int8  = -32
	FixIntNegMax int8  = -1

	FixIntNegOffset uint8 = 0xe0

	FixMapMin  uint8 = 0x80
	FixMapMax  uint8 = 0x8f
	FixArrayMin uint8 = 0x90
	FixArrayMax uint8 = 0x9f
	FixStrMin   uint8 = 0xa0
	FixStrMax   uint8 = 0xbf
)

const (
	NilFormat    uint8 = 0xc0
	NeverUsed    uint8 = 0xc1
	FalseFormat  uint8 = 0xc2
	TrueFormat   uint8 = 0xc3
	Bin8Format   uint8 = 0xc4
	Bin16Format  uint8 = 0xc5
	Bin32Format  uint8 = 0xc6
	Ext8Format   uint8 = 0xc7
	Ext16Format  uint8 = 0xc8
	Ext32Format  uint8 = 0xc9
	Float32Format uint8 = 0xca
	Float64Format uint8 = 0xcb
	Uint8Format   uint8 = 0xcc
	Uint16Format  uint8 = 0xcd
	Uint32Format  uint8 = 0xce
	Uint64Format  uint8 = 0xcf
	Int8Format    uint8 = 0xd0
	Int16Format   uint8 = 0xd1
	Int32Format   uint8 = 0xd2
	Int64Format   uint8 = 0xd3
	FixExt1Format uint8 = 0xd4
	FixExt2Format uint8 = 0xd5
	FixExt4Format uint8 = 0xd6
	FixExt8Format uint8 = 0xd7
	FixExt16Format uint8 = 0xd8
	Str8Format    uint8 = 0xd9
	Str16Format   uint8 = 0xda
	Str32Format   uint8 = 0xdb
	Array16Format uint8 = 0xdc
	Array32Format uint8 = 0xdd
	Map16Format   uint8 = 0xde
	Map32Format   uint8 = 0xdf
)

const (
	ExtTypeTimestamp int8 = -1
)

type Ext struct {
	Type int8
	Data []byte
}

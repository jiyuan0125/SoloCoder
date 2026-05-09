package flatbuffers

import (
	"encoding/binary"
	"math"
)

type Builder struct {
	buf         []byte
	vtables     []uint16
	vtableStart int
	head        int
}

func NewBuilder(initialSize int) *Builder {
	if initialSize <= 0 {
		initialSize = 1024
	}
	return &Builder{
		buf:  make([]byte, initialSize),
		head: initialSize,
	}
}

func (b *Builder) grow(needed int) {
	newSize := len(b.buf)
	for newSize < needed {
		newSize *= 2
	}
	newBuf := make([]byte, newSize)
	copy(newBuf[newSize-len(b.buf)+b.head:], b.buf[b.head:])
	b.head = newSize - len(b.buf) + b.head
	b.buf = newBuf
}

func (b *Builder) prep(size int, alignment int) {
	requiredSize := size + alignment
	availableSize := b.head
	if availableSize < requiredSize {
		b.grow(len(b.buf) + requiredSize + 128)
	}
	padding := (0 - b.head - size) & (alignment - 1)
	for i := 0; i < padding; i++ {
		b.head--
		b.buf[b.head] = 0
	}
	b.head -= size
}

func (b *Builder) pad(n int) {
	for i := 0; i < n; i++ {
		b.head--
		b.buf[b.head] = 0
	}
}

func (b *Builder) WriteByte(x byte) {
	b.prep(1, 1)
	b.buf[b.head] = x
}

func (b *Builder) WriteInt16(x int16) {
	b.prep(2, 2)
	binary.LittleEndian.PutUint16(b.buf[b.head:], uint16(x))
}

func (b *Builder) WriteUint16(x uint16) {
	b.prep(2, 2)
	binary.LittleEndian.PutUint16(b.buf[b.head:], x)
}

func (b *Builder) WriteInt32(x int32) {
	b.prep(4, 4)
	binary.LittleEndian.PutUint32(b.buf[b.head:], uint32(x))
}

func (b *Builder) WriteUint32(x uint32) {
	b.prep(4, 4)
	binary.LittleEndian.PutUint32(b.buf[b.head:], x)
}

func (b *Builder) WriteInt64(x int64) {
	b.prep(8, 8)
	binary.LittleEndian.PutUint64(b.buf[b.head:], uint64(x))
}

func (b *Builder) WriteUint64(x uint64) {
	b.prep(8, 8)
	binary.LittleEndian.PutUint64(b.buf[b.head:], x)
}

func (b *Builder) WriteFloat32(x float32) {
	b.prep(4, 4)
	binary.LittleEndian.PutUint32(b.buf[b.head:], math.Float32bits(x))
}

func (b *Builder) WriteFloat64(x float64) {
	b.prep(8, 8)
	binary.LittleEndian.PutUint64(b.buf[b.head:], math.Float64bits(x))
}

func (b *Builder) WriteBool(x bool) {
	if x {
		b.WriteByte(1)
	} else {
		b.WriteByte(0)
	}
}

func (b *Builder) StartObject(numFields int) {
	b.vtables = make([]uint16, numFields)
	b.vtableStart = b.head
}

func (b *Builder) EndObject() int {
	b.WriteInt32(0)
	objectOffset := b.Offset()
	vtableSize := 4 + len(b.vtables)*2
	b.WriteInt16(int16(vtableSize))
	b.WriteInt16(int16(b.vtableStart - objectOffset))
	for i := len(b.vtables) - 1; i >= 0; i-- {
		b.WriteUint16(b.vtables[i])
	}
	return objectOffset
}

func (b *Builder) Offset() int {
	return len(b.buf) - b.head
}

func (b *Builder) AddBool(slot int, x bool, d bool) {
	if x != d {
		b.WriteBool(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddInt8(slot int, x int8, d int8) {
	if x != d {
		b.WriteByte(byte(x))
		b.Slot(slot)
	}
}

func (b *Builder) AddUint8(slot int, x uint8, d uint8) {
	if x != d {
		b.WriteByte(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddInt16(slot int, x int16, d int16) {
	if x != d {
		b.WriteInt16(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddUint16(slot int, x uint16, d uint16) {
	if x != d {
		b.WriteUint16(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddInt32(slot int, x int32, d int32) {
	if x != d {
		b.WriteInt32(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddUint32(slot int, x uint32, d uint32) {
	if x != d {
		b.WriteUint32(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddInt64(slot int, x int64, d int64) {
	if x != d {
		b.WriteInt64(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddUint64(slot int, x uint64, d uint64) {
	if x != d {
		b.WriteUint64(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddFloat32(slot int, x float32, d float32) {
	if x != d {
		b.WriteFloat32(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddFloat64(slot int, x float64, d float64) {
	if x != d {
		b.WriteFloat64(x)
		b.Slot(slot)
	}
}

func (b *Builder) AddOffset(slot int, o int) {
	if o != 0 {
		b.WriteInt32(int32(b.Offset() - o + 4))
		b.Slot(slot)
	}
}

func (b *Builder) Slot(slot int) {
	b.vtables[slot] = uint16(b.vtableStart - b.head)
}

func (b *Builder) CreateString(s string) int {
	b.pad(1)
	b.prep(4, 4)
	for i := len(s) - 1; i >= 0; i-- {
		b.head--
		b.buf[b.head] = s[i]
	}
	b.WriteInt32(int32(len(s)))
	return b.Offset()
}

func (b *Builder) Finish(rootTable int) []byte {
	b.prep(4, 4)
	b.WriteInt32(int32(b.Offset() - rootTable + 4))
	return b.FinishedBytes()
}

func (b *Builder) FinishedBytes() []byte {
	return b.buf[b.head:]
}

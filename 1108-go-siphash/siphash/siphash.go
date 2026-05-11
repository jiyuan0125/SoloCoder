package siphash

import (
	"encoding/binary"
	"errors"
)

const (
	DefaultC = 2
	DefaultD = 4
)

var (
	ErrInvalidKey = errors.New("siphash: invalid key size, must be 16 bytes")
)

type Key struct {
	K0, K1 uint64
}

func NewKeyFromBytes(key []byte) (*Key, error) {
	if len(key) != 16 {
		return nil, ErrInvalidKey
	}
	return &Key{
		K0: binary.LittleEndian.Uint64(key[0:8]),
		K1: binary.LittleEndian.Uint64(key[8:16]),
	}, nil
}

func NewKey(k0, k1 uint64) *Key {
	return &Key{K0: k0, K1: k1}
}

func (k *Key) Bytes() []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:8], k.K0)
	binary.LittleEndian.PutUint64(buf[8:16], k.K1)
	return buf
}

type Hash struct {
	k0, k1 uint64
	v0, v1, v2, v3 uint64
	c, d           int
	buf            [8]byte
	n              int
	length         uint64
}

func New(key *Key) *Hash {
	return NewWithRounds(key, DefaultC, DefaultD)
}

func New13(key *Key) *Hash {
	return NewWithRounds(key, 1, 3)
}

func New24(key *Key) *Hash {
	return NewWithRounds(key, 2, 4)
}

func NewWithRounds(key *Key, c, d int) *Hash {
	h := &Hash{
		k0: key.K0,
		k1: key.K1,
		c:  c,
		d:  d,
	}
	h.Reset()
	return h
}

func (h *Hash) Reset() {
	h.v0 = h.k0 ^ 0x736f6d6570736575
	h.v1 = h.k1 ^ 0x646f72616e646f6d
	h.v2 = h.k0 ^ 0x6c7967656e657261
	h.v3 = h.k1 ^ 0x7465646279746573
	h.n = 0
	h.length = 0
}

func rotl(x uint64, b uint) uint64 {
	return (x << b) | (x >> (64 - b))
}

func (h *Hash) round() {
	h.v0 += h.v1
	h.v1 = rotl(h.v1, 13)
	h.v1 ^= h.v0
	h.v0 = rotl(h.v0, 32)
	h.v2 += h.v3
	h.v3 = rotl(h.v3, 16)
	h.v3 ^= h.v2
	h.v0 += h.v3
	h.v3 = rotl(h.v3, 21)
	h.v3 ^= h.v0
	h.v2 += h.v1
	h.v1 = rotl(h.v1, 17)
	h.v1 ^= h.v2
	h.v2 = rotl(h.v2, 32)
}

func (h *Hash) Write(p []byte) (int, error) {
	n := len(p)
	h.length += uint64(n)

	if h.n > 0 {
		want := 8 - h.n
		if n < want {
			copy(h.buf[h.n:], p)
			h.n += n
			return n, nil
		}
		copy(h.buf[h.n:], p[:want])
		h.processBlock(h.buf[:])
		p = p[want:]
		h.n = 0
	}

	for len(p) >= 8 {
		h.processBlock(p[:8])
		p = p[8:]
	}

	if len(p) > 0 {
		copy(h.buf[:], p)
		h.n = len(p)
	}

	return n, nil
}

func (h *Hash) processBlock(p []byte) {
	m := binary.LittleEndian.Uint64(p)
	h.v3 ^= m
	for i := 0; i < h.c; i++ {
		h.round()
	}
	h.v0 ^= m
}

func (h *Hash) Sum64() uint64 {
	v0, v1, v2, v3 := h.v0, h.v1, h.v2, h.v3

	var b uint64
	switch h.n {
	case 7:
		b |= uint64(h.buf[6]) << 48
		fallthrough
	case 6:
		b |= uint64(h.buf[5]) << 40
		fallthrough
	case 5:
		b |= uint64(h.buf[4]) << 32
		fallthrough
	case 4:
		b |= uint64(h.buf[3]) << 24
		fallthrough
	case 3:
		b |= uint64(h.buf[2]) << 16
		fallthrough
	case 2:
		b |= uint64(h.buf[1]) << 8
		fallthrough
	case 1:
		b |= uint64(h.buf[0])
	}
	b |= uint64(h.length&0xff) << 56

	v3 ^= b
	for i := 0; i < h.c; i++ {
		v0 += v1
		v1 = (v1 << 13) | (v1 >> (64 - 13))
		v1 ^= v0
		v0 = (v0 << 32) | (v0 >> (64 - 32))
		v2 += v3
		v3 = (v3 << 16) | (v3 >> (64 - 16))
		v3 ^= v2
		v0 += v3
		v3 = (v3 << 21) | (v3 >> (64 - 21))
		v3 ^= v0
		v2 += v1
		v1 = (v1 << 17) | (v1 >> (64 - 17))
		v1 ^= v2
		v2 = (v2 << 32) | (v2 >> (64 - 32))
	}

	v0 ^= b
	v2 ^= 0xff

	for i := 0; i < h.d; i++ {
		v0 += v1
		v1 = (v1 << 13) | (v1 >> (64 - 13))
		v1 ^= v0
		v0 = (v0 << 32) | (v0 >> (64 - 32))
		v2 += v3
		v3 = (v3 << 16) | (v3 >> (64 - 16))
		v3 ^= v2
		v0 += v3
		v3 = (v3 << 21) | (v3 >> (64 - 21))
		v3 ^= v0
		v2 += v1
		v1 = (v1 << 17) | (v1 >> (64 - 17))
		v1 ^= v2
		v2 = (v2 << 32) | (v2 >> (64 - 32))
	}

	return v0 ^ v1 ^ v2 ^ v3
}

func Sum64(key *Key, data []byte) uint64 {
	return Sum64WithRounds(key, data, DefaultC, DefaultD)
}

func Sum64WithRounds(key *Key, data []byte, c, d int) uint64 {
	h := NewWithRounds(key, c, d)
	h.Write(data)
	return h.Sum64()
}

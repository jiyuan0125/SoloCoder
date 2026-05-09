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
	v0, v1, v2, v3 uint64
	key            *Key
	c, d           int
	partial        []byte
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
		key:     key,
		c:       c,
		d:       d,
		partial: make([]byte, 0, 8),
	}
	h.reset()
	return h
}

func (h *Hash) reset() {
	h.v0 = h.key.K0 ^ 0x736f6d6570736575
	h.v1 = h.key.K1 ^ 0x646f72616e646f6d
	h.v2 = h.key.K0 ^ 0x6c7967656e657261
	h.v3 = h.key.K1 ^ 0x7465646279746573
	h.partial = h.partial[:0]
	h.length = 0
}

func rotl(x uint64, b uint) uint64 {
	return (x << b) | (x >> (64 - b))
}

func (h *Hash) sipRound() {
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

func (h *Hash) compressBlocks(data []byte) {
	for len(data) >= 8 {
		m := binary.LittleEndian.Uint64(data[:8])
		h.v3 ^= m
		for i := 0; i < h.c; i++ {
			h.sipRound()
		}
		h.v0 ^= m
		data = data[8:]
	}
}

func (h *Hash) Write(p []byte) (int, error) {
	n := len(p)
	h.length += uint64(n)

	if len(h.partial) > 0 {
		need := 8 - len(h.partial)
		if n < need {
			h.partial = append(h.partial, p...)
			return n, nil
		}
		h.partial = append(h.partial, p[:need]...)
		h.compressBlocks(h.partial)
		p = p[need:]
		h.partial = h.partial[:0]
	}

	h.compressBlocks(p)

	if len(p) > 0 {
		remaining := len(p) % 8
		if remaining > 0 {
			h.partial = append(h.partial, p[len(p)-remaining:]...)
		}
	}

	return n, nil
}

func (h *Hash) Sum64() uint64 {
	h.v3 ^= uint64(h.length%256) << 56
	for i, b := range h.partial {
		h.v3 ^= uint64(b) << (8 * i)
	}

	for i := 0; i < h.c; i++ {
		h.sipRound()
	}

	h.v0 ^= uint64(h.length%256) << 56
	for i, b := range h.partial {
		h.v0 ^= uint64(b) << (8 * i)
	}
	h.v2 ^= 0xff

	for i := 0; i < h.d; i++ {
		h.sipRound()
	}

	return h.v0 ^ h.v1 ^ h.v2 ^ h.v3
}

func (h *Hash) Reset() {
	h.reset()
}

func Sum64(key *Key, data []byte) uint64 {
	return Sum64WithRounds(key, data, DefaultC, DefaultD)
}

func Sum64WithRounds(key *Key, data []byte, c, d int) uint64 {
	h := NewWithRounds(key, c, d)
	h.Write(data)
	return h.Sum64()
}

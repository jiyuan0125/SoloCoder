package murmurhash3

func rotl32(x, r uint32) uint32 {
	return (x << r) | (x >> (32 - r))
}

func fmix32(h uint32) uint32 {
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h
}

func Hash32(data []byte, seed uint32) uint32 {
	const (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
	)

	h := seed
	n := len(data)
	nblocks := n / 4

	for i := 0; i < nblocks; i++ {
		idx := i * 4
		k := uint32(data[idx+0]) | uint32(data[idx+1])<<8 | uint32(data[idx+2])<<16 | uint32(data[idx+3])<<24
		k *= c1
		k = rotl32(k, 15)
		k *= c2
		h ^= k
		h = rotl32(h, 13)
		h = h*5 + 0xe6546b64
	}

	tail := nblocks * 4
	var k1 uint32
	switch n & 3 {
	case 3:
		k1 ^= uint32(data[tail+2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(data[tail+1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(data[tail+0])
		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2
		h ^= k1
	}

	h ^= uint32(n)
	h = fmix32(h)

	return h
}

func Hash32String(data string, seed uint32) uint32 {
	return Hash32([]byte(data), seed)
}

func rotl64(x, r uint64) uint64 {
	return (x << r) | (x >> (64 - r))
}

func fmix64(k uint64) uint64 {
	k ^= k >> 33
	k *= 0xff51afd7ed558ccd
	k ^= k >> 33
	k *= 0xc4ceb9fe1a85ec53
	k ^= k >> 33
	return k
}

func Hash128(data []byte, seed uint32) (uint64, uint64) {
	const (
		c1 uint64 = 0x239b961bab0e4347
		c2 uint64 = 0x84cba2b9312b34c9
	)

	var h1, h2 uint64
	h1 = uint64(seed)
	h2 = uint64(seed)
	n := len(data)
	nblocks := n / 16

	for i := 0; i < nblocks; i++ {
		idx := i * 16
		var k1, k2 uint64
		k1 = uint64(data[idx+0]) | uint64(data[idx+1])<<8 | uint64(data[idx+2])<<16 | uint64(data[idx+3])<<24 |
			uint64(data[idx+4])<<32 | uint64(data[idx+5])<<40 | uint64(data[idx+6])<<48 | uint64(data[idx+7])<<56
		k2 = uint64(data[idx+8]) | uint64(data[idx+9])<<8 | uint64(data[idx+10])<<16 | uint64(data[idx+11])<<24 |
			uint64(data[idx+12])<<32 | uint64(data[idx+13])<<40 | uint64(data[idx+14])<<48 | uint64(data[idx+15])<<56

		k1 *= c1
		k1 = rotl64(k1, 31)
		k1 *= c2
		h1 ^= k1
		h1 = rotl64(h1, 27)
		h1 += h2
		h1 = h1*5 + 0x52dce729

		k2 *= c2
		k2 = rotl64(k2, 33)
		k2 *= c1
		h2 ^= k2
		h2 = rotl64(h2, 31)
		h2 += h1
		h2 = h2*5 + 0x38495ab5
	}

	tail := nblocks * 16
	var k1, k2 uint64
	switch n & 15 {
	case 15:
		k2 ^= uint64(data[tail+14]) << 48
		fallthrough
	case 14:
		k2 ^= uint64(data[tail+13]) << 40
		fallthrough
	case 13:
		k2 ^= uint64(data[tail+12]) << 32
		fallthrough
	case 12:
		k2 ^= uint64(data[tail+11]) << 24
		fallthrough
	case 11:
		k2 ^= uint64(data[tail+10]) << 16
		fallthrough
	case 10:
		k2 ^= uint64(data[tail+9]) << 8
		fallthrough
	case 9:
		k2 ^= uint64(data[tail+8])
		k2 *= c2
		k2 = rotl64(k2, 33)
		k2 *= c1
		h2 ^= k2
		fallthrough
	case 8:
		k1 ^= uint64(data[tail+7]) << 56
		fallthrough
	case 7:
		k1 ^= uint64(data[tail+6]) << 48
		fallthrough
	case 6:
		k1 ^= uint64(data[tail+5]) << 40
		fallthrough
	case 5:
		k1 ^= uint64(data[tail+4]) << 32
		fallthrough
	case 4:
		k1 ^= uint64(data[tail+3]) << 24
		fallthrough
	case 3:
		k1 ^= uint64(data[tail+2]) << 16
		fallthrough
	case 2:
		k1 ^= uint64(data[tail+1]) << 8
		fallthrough
	case 1:
		k1 ^= uint64(data[tail+0])
		k1 *= c1
		k1 = rotl64(k1, 31)
		k1 *= c2
		h1 ^= k1
	}

	h1 ^= uint64(n)
	h2 ^= uint64(n)
	h1 += h2
	h2 += h1
	h1 = fmix64(h1)
	h2 = fmix64(h2)
	h1 += h2
	h2 += h1

	return h1, h2
}

func Hash128String(data string, seed uint32) (uint64, uint64) {
	return Hash128([]byte(data), seed)
}

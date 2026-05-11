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

func Hash128(data []byte, seed uint32) (uint64, uint64) {
	const (
		c1 uint32 = 0x239b961b
		c2 uint32 = 0xab0e9789
		c3 uint32 = 0x38b34ae5
		c4 uint32 = 0xa1e38b93
	)

	h1 := seed
	h2 := seed
	h3 := seed
	h4 := seed
	n := len(data)
	nblocks := n / 16

	for i := 0; i < nblocks; i++ {
		idx := i * 16
		k1 := uint32(data[idx+0]) | uint32(data[idx+1])<<8 | uint32(data[idx+2])<<16 | uint32(data[idx+3])<<24
		k2 := uint32(data[idx+4]) | uint32(data[idx+5])<<8 | uint32(data[idx+6])<<16 | uint32(data[idx+7])<<24
		k3 := uint32(data[idx+8]) | uint32(data[idx+9])<<8 | uint32(data[idx+10])<<16 | uint32(data[idx+11])<<24
		k4 := uint32(data[idx+12]) | uint32(data[idx+13])<<8 | uint32(data[idx+14])<<16 | uint32(data[idx+15])<<24

		k1 *= c1
		k1 = rotl32(k1, 15)
		k1 *= c2
		h1 ^= k1
		h1 = rotl32(h1, 19)
		h1 += h2
		h1 = h1*5 + 0x561ccd1b

		k2 *= c2
		k2 = rotl32(k2, 16)
		k2 *= c3
		h2 ^= k2
		h2 = rotl32(h2, 17)
		h2 += h3
		h2 = h2*5 + 0x0bcaa747

		k3 *= c3
		k3 = rotl32(k3, 17)
		k3 *= c4
		h3 ^= k3
		h3 = rotl32(h3, 15)
		h3 += h4
		h3 = h3*5 + 0x96cd1c35

		k4 *= c4
		k4 = rotl32(k4, 18)
		k4 *= c1
		h4 ^= k4
		h4 = rotl32(h4, 13)
		h4 += h1
		h4 = h4*5 + 0x32ac3b17
	}

	tail := nblocks * 16
	var k1, k2, k3, k4 uint32
	switch n & 15 {
	case 15:
		k4 ^= uint32(data[tail+14]) << 16
		fallthrough
	case 14:
		k4 ^= uint32(data[tail+13]) << 8
		fallthrough
	case 13:
		k4 ^= uint32(data[tail+12])
		k4 *= c4
		k4 = rotl32(k4, 18)
		k4 *= c1
		h4 ^= k4
		fallthrough
	case 12:
		k3 ^= uint32(data[tail+11]) << 24
		fallthrough
	case 11:
		k3 ^= uint32(data[tail+10]) << 16
		fallthrough
	case 10:
		k3 ^= uint32(data[tail+9]) << 8
		fallthrough
	case 9:
		k3 ^= uint32(data[tail+8])
		k3 *= c3
		k3 = rotl32(k3, 17)
		k3 *= c4
		h3 ^= k3
		fallthrough
	case 8:
		k2 ^= uint32(data[tail+7]) << 24
		fallthrough
	case 7:
		k2 ^= uint32(data[tail+6]) << 16
		fallthrough
	case 6:
		k2 ^= uint32(data[tail+5]) << 8
		fallthrough
	case 5:
		k2 ^= uint32(data[tail+4])
		k2 *= c2
		k2 = rotl32(k2, 16)
		k2 *= c3
		h2 ^= k2
		fallthrough
	case 4:
		k1 ^= uint32(data[tail+3]) << 24
		fallthrough
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
		h1 ^= k1
	}

	h1 ^= uint32(n)
	h2 ^= uint32(n)
	h3 ^= uint32(n)
	h4 ^= uint32(n)

	h1 += h2
	h1 += h3
	h1 += h4
	h2 += h1
	h3 += h1
	h4 += h1

	h1 = fmix32(h1)
	h2 = fmix32(h2)
	h3 = fmix32(h3)
	h4 = fmix32(h4)

	h1 += h2
	h1 += h3
	h1 += h4
	h2 += h1
	h3 += h1
	h4 += h1

	high := (uint64(h4) << 32) | uint64(h3)
	low := (uint64(h2) << 32) | uint64(h1)

	return high, low
}

func Hash128String(data string, seed uint32) (uint64, uint64) {
	return Hash128([]byte(data), seed)
}

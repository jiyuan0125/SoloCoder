package hashtable

func hash1(key string) uint64 {
	h := uint64(5381)
	for i := 0; i < len(key); i++ {
		h = ((h << 5) + h) + uint64(key[i])
	}
	return h
}

func hash2(key string, capacity int) uint64 {
	h := uint64(0)
	for i := 0; i < len(key); i++ {
		h = (h * 33) ^ uint64(key[i])
	}
	
	result := h % uint64(capacity)
	if result == 0 {
		result = 1
	}
	return result
}

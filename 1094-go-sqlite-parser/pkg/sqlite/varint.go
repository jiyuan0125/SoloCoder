package sqlite

func readVarint(data []byte, offset int) (uint64, int) {
	var result uint64 = 0
	bytesRead := 0

	for i := 0; i < 9 && offset+i < len(data); i++ {
		b := data[offset+i]
		result = (result << 7) | uint64(b&0x7F)
		bytesRead++
		if (b & 0x80) == 0 {
			break
		}
	}

	return result, bytesRead
}

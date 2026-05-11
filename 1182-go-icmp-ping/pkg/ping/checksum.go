package ping

func CalculateChecksum(data []byte) uint16 {
	var sum uint32
	n := len(data)
	
	for i := 0; i < n-1; i += 2 {
		sum += uint32(data[i])<<8 | uint32(data[i+1])
	}
	
	if n%2 == 1 {
		sum += uint32(data[n-1]) << 8
	}
	
	for sum>>16 != 0 {
		sum = (sum & 0xffff) + (sum >> 16)
	}
	
	return ^uint16(sum)
}

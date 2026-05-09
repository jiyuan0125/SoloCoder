package deflate

const (
	windowSize   = 32768
	matchMinLen  = 3
	matchMaxLen  = 258
	hashBits     = 15
	hashMask     = (1 << hashBits) - 1
)

type LZ77Token struct {
	Type     int
	Literal  byte
	Distance int
	Length   int
}

const (
	TokenLiteral = iota
	TokenMatch
)

func LZ77Encode(data []byte) []LZ77Token {
	if len(data) == 0 {
		return nil
	}

	var tokens []LZ77Token
	head := make([]int, 1<<hashBits)
	prev := make([]int, windowSize)

	for i := range head {
		head[i] = -1
	}
	for i := range prev {
		prev[i] = -1
	}

	pos := 0
	for pos < len(data) {
		if pos+matchMinLen > len(data) {
			tokens = append(tokens, LZ77Token{Type: TokenLiteral, Literal: data[pos]})
			pos++
			continue
		}

		h := hash(data[pos:])
		bestDist := 0
		bestLen := 0
		limit := max(0, pos-windowSize)

		for chain := head[h]; chain >= limit && chain < pos; chain = prev[chain%windowSize] {
			l := 0
			for pos+l < len(data) && chain+l < pos && l < matchMaxLen {
				if data[pos+l] != data[chain+l] {
					break
				}
				l++
			}
			if l > bestLen {
				bestLen = l
				bestDist = pos - chain
			}
			if bestLen >= matchMaxLen {
				break
			}
		}

		if bestLen >= matchMinLen {
			tokens = append(tokens, LZ77Token{
				Type:     TokenMatch,
				Distance: bestDist,
				Length:   bestLen,
			})
			for i := 0; i < bestLen; i++ {
				if pos+i+matchMinLen <= len(data) {
					nh := hash(data[pos+i:])
					prev[(pos+i)%windowSize] = head[nh]
					head[nh] = pos + i
				}
			}
			pos += bestLen
		} else {
			tokens = append(tokens, LZ77Token{Type: TokenLiteral, Literal: data[pos]})
			if pos+matchMinLen <= len(data) {
				prev[pos%windowSize] = head[h]
				head[h] = pos
			}
			pos++
		}
	}

	return tokens
}

func hash(data []byte) int {
	return int(((uint32(data[0])<<16)|(uint32(data[1])<<8)|uint32(data[2]))*2654435761) & hashMask
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

var lengthBase = [29]int{
	3, 4, 5, 6, 7, 8, 9, 10,
	11, 13, 15, 17, 19, 23, 27, 31,
	35, 43, 51, 59, 67, 83, 99, 115,
	131, 163, 195, 227, 258,
}

var lengthExtra = [29]int{
	0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 1, 1, 2, 2, 2, 2,
	3, 3, 3, 3, 4, 4, 4, 4,
	5, 5, 5, 5, 0,
}

var distanceBase = [30]int{
	1, 2, 3, 4, 5, 7, 9, 13,
	17, 25, 33, 49, 65, 97, 129, 193,
	257, 385, 513, 769, 1025, 1537, 2049, 3073,
	4097, 6145, 8193, 12289, 16385, 24577,
}

var distanceExtra = [30]int{
	0, 0, 0, 0, 1, 1, 2, 2,
	3, 3, 4, 4, 5, 5, 6, 6,
	7, 7, 8, 8, 9, 9, 10, 10,
	11, 11, 12, 12, 13, 13,
}

func GetLengthCode(length int) (symbol int, extra int, extraBits int) {
	for i := 0; i < len(lengthBase); i++ {
		if length >= lengthBase[i] && (i == len(lengthBase)-1 || length < lengthBase[i+1]) {
			return 257 + i, length - lengthBase[i], lengthExtra[i]
		}
	}
	return 285, 0, 0
}

func GetDistanceCode(distance int) (symbol int, extra int, extraBits int) {
	for i := 0; i < len(distanceBase); i++ {
		if distance >= distanceBase[i] && (i == len(distanceBase)-1 || distance < distanceBase[i+1]) {
			return i, distance - distanceBase[i], distanceExtra[i]
		}
	}
	return 29, 0, 0
}

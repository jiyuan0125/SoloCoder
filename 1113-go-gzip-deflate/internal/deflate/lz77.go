package deflate

const (
	MaxWindowSize    = 32768
	MinMatchLength   = 3
	MaxMatchLength   = 258
	MaxLookaheadSize = 258
)

type LZ77Token struct {
	Literal  byte
	IsMatch  bool
	Distance int
	Length   int
}

type LZ77Encoder struct {
	window      []byte
	windowPos   int
	windowFull  bool
	lookahead   []byte
	searchSize  int
}

func NewLZ77Encoder() *LZ77Encoder {
	return &LZ77Encoder{
		window:     make([]byte, MaxWindowSize),
		searchSize: MaxWindowSize,
	}
}

func (e *LZ77Encoder) Encode(input []byte) []LZ77Token {
	e.lookahead = append(e.lookahead, input...)

	var tokens []LZ77Token

	for len(e.lookahead) > 0 {
		bestDist, bestLen := e.findBestMatch()

		if bestLen >= MinMatchLength {
			tokens = append(tokens, LZ77Token{
				IsMatch:  true,
				Distance: bestDist,
				Length:   bestLen,
			})
			e.addToWindow(e.lookahead[:bestLen])
			e.lookahead = e.lookahead[bestLen:]
		} else {
			tokens = append(tokens, LZ77Token{
				IsMatch: false,
				Literal: e.lookahead[0],
			})
			e.addToWindow(e.lookahead[:1])
			e.lookahead = e.lookahead[1:]
		}
	}

	return tokens
}

func (e *LZ77Encoder) findBestMatch() (int, int) {
	if len(e.lookahead) < MinMatchLength {
		return 0, 0
	}

	availableWindow := e.windowSize()
	if availableWindow == 0 {
		return 0, 0
	}

	searchLimit := availableWindow
	if searchLimit > MaxWindowSize {
		searchLimit = MaxWindowSize
	}

	lookLimit := len(e.lookahead)
	if lookLimit > MaxMatchLength {
		lookLimit = MaxMatchLength
	}

	bestDist := 0
	bestLen := 0

	for dist := 1; dist <= searchLimit; dist++ {
		length := e.matchLengthAt(dist, lookLimit)
		if length > bestLen {
			bestDist = dist
			bestLen = length
			if bestLen == MaxMatchLength {
				break
			}
		}
	}

	return bestDist, bestLen
}

func (e *LZ77Encoder) matchLengthAt(distance int, maxLen int) int {
	length := 0

	for length < maxLen {
		var winByte byte
		if length < distance {
			winByte = e.windowByteFromEnd(distance - length)
		} else {
			lookIdx := length - distance
			if lookIdx < len(e.lookahead) {
				winByte = e.lookahead[lookIdx]
			} else {
				break
			}
		}

		if length >= len(e.lookahead) {
			break
		}
		if winByte != e.lookahead[length] {
			break
		}
		length++
	}

	return length
}

func (e *LZ77Encoder) windowByteFromEnd(offset int) byte {
	pos := e.windowPos - offset
	if pos < 0 {
		pos += len(e.window)
	}
	return e.window[pos]
}

func (e *LZ77Encoder) addToWindow(data []byte) {
	for _, b := range data {
		e.window[e.windowPos] = b
		e.windowPos = (e.windowPos + 1) % len(e.window)
		if e.windowPos == 0 {
			e.windowFull = true
		}
	}
}

func (e *LZ77Encoder) windowSize() int {
	if e.windowFull {
		return len(e.window)
	}
	return e.windowPos
}

package brotli

const (
	MinWindowBits     = 10
	MaxWindowBits     = 24
	DefaultWindowBits = 22

	MinQuality     = 1
	MaxQuality     = 11
	DefaultQuality = 6

	MaxDistanceCodes = 16
	MaxLiteralCodes  = 256
	MaxInsertCopyCodes = 704

	WordTransformCount = 121

	EmptyMetaBlockMagic    = 0x3
	LastEmptyMetaBlockMagic = 0x3f
)

var (
	windowBitsForQuality = []int{
		16,
		16,
		17,
		18,
		19,
		20,
		21,
		22,
		22,
		24,
		24,
		24,
	}

	searchDepthForQuality = []int{
		1,
		2,
		4,
		8,
		16,
		32,
		64,
		128,
		256,
		512,
		1024,
		2048,
	}

	minMatchForQuality = []int{
		4,
		3,
		3,
		2,
		2,
		2,
		2,
		2,
		2,
		2,
		2,
		2,
	}
)

func getSearchDepth(quality int) int {
	if quality < MinQuality {
		quality = MinQuality
	}
	if quality > MaxQuality {
		quality = MaxQuality
	}
	return searchDepthForQuality[quality-1]
}

func getMinMatch(quality int) int {
	if quality < MinQuality {
		quality = MinQuality
	}
	if quality > MaxQuality {
		quality = MaxQuality
	}
	return minMatchForQuality[quality-1]
}

func getWindowBits(quality int) int {
	if quality < MinQuality {
		quality = MinQuality
	}
	if quality > MaxQuality {
		quality = MaxQuality
	}
	return windowBitsForQuality[quality-1]
}

func getWindowSize(windowBits int) int {
	if windowBits == 16 {
		return 65536
	}
	return 1 << windowBits
}

var literalContextLookup = []int{
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 6, 6, 6, 6, 6,
	6, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 6, 6, 6, 6, 6,
	6, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
}

var distanceShortCodePrefixBits = []int{
	2, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 12,
}

var distanceShortCodeValues = []int{
	0, 4, 6, 8, 12, 14, 18, 20, 28, 34, 50, 66, 98, 130, 194, 322,
}

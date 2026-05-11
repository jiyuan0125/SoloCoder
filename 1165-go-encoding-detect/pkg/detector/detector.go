package detector

import (
	"unicode/utf8"
)

type Result struct {
	Encoding   string
	Confidence float64
	HasBOM     bool
}

type encodingStats struct {
	encoding       string
	isValid        bool
	validityScore  float64
	distributionScore float64
	featureScore   float64
}

var (
	bomUTF8     = []byte{0xEF, 0xBB, 0xBF}
	bomUTF16LE  = []byte{0xFF, 0xFE}
	bomUTF16BE  = []byte{0xFE, 0xFF}
	bomUTF32LE  = []byte{0xFF, 0xFE, 0x00, 0x00}
	bomUTF32BE  = []byte{0x00, 0x00, 0xFE, 0xFF}
)

var (
	gbkCommonHighBytes = []byte{
		0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9,
		0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF, 0xC0, 0xC1, 0xC2, 0xC3,
		0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9, 0xCA, 0xCB, 0xCC, 0xCD,
		0xCE, 0xCF, 0xD0, 0xD1, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7,
	}

	big5CommonHighBytes = []byte{
		0xA4, 0xA5, 0xA6, 0xA7, 0xAA, 0xAB, 0xAC, 0xAD, 0xAE, 0xAF,
		0xB0, 0xB1, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6, 0xB7, 0xB8, 0xB9,
		0xBA, 0xBB, 0xBC, 0xBD, 0xBE, 0xBF, 0xC0, 0xC1, 0xC2, 0xC3,
	}

	shiftJISCommonHighBytes = []byte{
		0x81, 0x82, 0x83, 0x84, 0x85, 0x88, 0x89, 0x8A, 0x8B, 0x8C,
		0x8D, 0x8E, 0x8F, 0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96,
		0xE0, 0xE1, 0xE2, 0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9,
	}
)

func Detect(data []byte) Result {
	if len(data) == 0 {
		return Result{Encoding: "utf-8", Confidence: 0.5, HasBOM: false}
	}

	if result, ok := detectBOM(data); ok {
		return result
	}

	return detectStatistical(data)
}

func detectBOM(data []byte) (Result, bool) {
	if len(data) >= 4 {
		if startsWith(data, bomUTF32BE) {
			return Result{Encoding: "utf-32be", Confidence: 1.0, HasBOM: true}, true
		}
		if startsWith(data, bomUTF32LE) {
			return Result{Encoding: "utf-32le", Confidence: 1.0, HasBOM: true}, true
		}
	}
	if len(data) >= 3 {
		if startsWith(data, bomUTF8) {
			return Result{Encoding: "utf-8", Confidence: 1.0, HasBOM: true}, true
		}
	}
	if len(data) >= 2 {
		if startsWith(data, bomUTF16BE) {
			return Result{Encoding: "utf-16be", Confidence: 1.0, HasBOM: true}, true
		}
		if startsWith(data, bomUTF16LE) {
			return Result{Encoding: "utf-16le", Confidence: 1.0, HasBOM: true}, true
		}
	}
	return Result{}, false
}

func startsWith(data, prefix []byte) bool {
	if len(data) < len(prefix) {
		return false
	}
	for i := range prefix {
		if data[i] != prefix[i] {
			return false
		}
	}
	return true
}

func detectStatistical(data []byte) Result {
	validEncodings := make(map[string]*encodingStats)

	asciiOnly := isASCIIOnly(data)
	hasNonASCII := !asciiOnly

	if utf8Valid := isUTF8(data); utf8Valid {
		validEncodings["utf-8"] = &encodingStats{
			encoding:          "utf-8",
			isValid:           true,
			validityScore:     0.9,
			distributionScore: calculateUTF8DistributionScore(data),
			featureScore:      0.0,
		}
	}

	stats := analyzeByteDistribution(data)

	definitiveResult := applyDefinitiveRules(stats)
	if definitiveResult != "" {
		return Result{Encoding: definitiveResult, Confidence: 0.95, HasBOM: false}
	}

	if gbkValid := isGBK(data); gbkValid {
		validEncodings["gbk"] = &encodingStats{
			encoding:          "gbk",
			isValid:           true,
			validityScore:     0.5,
			distributionScore: calculateGBKDistributionScore(stats),
			featureScore:      calculateGBKFeatureScore(stats),
		}
	}

	if big5Valid := isBig5(data); big5Valid {
		validEncodings["big5"] = &encodingStats{
			encoding:          "big5",
			isValid:           true,
			validityScore:     0.5,
			distributionScore: calculateBig5DistributionScore(stats),
			featureScore:      calculateBig5FeatureScore(stats),
		}
	}

	if sjisValid := isShiftJIS(data); sjisValid {
		validEncodings["shift_jis"] = &encodingStats{
			encoding:          "shift_jis",
			isValid:           true,
			validityScore:     0.5,
			distributionScore: calculateShiftJISDistributionScore(stats),
			featureScore:      calculateShiftJISFeatureScore(stats),
		}
	}

	if utf16LEValid := isUTF16LE(data); utf16LEValid {
		validEncodings["utf-16le"] = &encodingStats{
			encoding:          "utf-16le",
			isValid:           true,
			validityScore:     0.4,
			distributionScore: 0.0,
			featureScore:      0.0,
		}
	}

	if utf16BEValid := isUTF16BE(data); utf16BEValid {
		validEncodings["utf-16be"] = &encodingStats{
			encoding:          "utf-16be",
			isValid:           true,
			validityScore:     0.4,
			distributionScore: 0.0,
			featureScore:      0.0,
		}
	}

	if len(validEncodings) == 0 {
		return Result{Encoding: "utf-8", Confidence: 0.3, HasBOM: false}
	}

	if !hasNonASCII {
		if utf8Stats, exists := validEncodings["utf-8"]; exists {
			return Result{Encoding: "utf-8", Confidence: utf8Stats.validityScore, HasBOM: false}
		}
		return Result{Encoding: "utf-8", Confidence: 0.5, HasBOM: false}
	}

	isShortText := stats.doubleByteCount < 5
	return resolveConflicts(validEncodings, isShortText, stats)
}

type byteStats struct {
	totalBytes       int
	asciiCount       int
	highByteCount    int
	katakanaCount    int
	firstByteFreq    map[byte]int
	secondByteFreq   map[byte]int
	doubleByteCount  int
	secondByteIn0x40To0x7ECount int
	secondByteIn0xA1To0xFECount int
	secondByteIn0x80To0x9FCount int
	secondByteIn0x7FTo0xA0Count int
	firstByteInSJISRangeCount int
	firstByteInBig5RangeCount int
	firstByteInGBKRangeCount int
}

func analyzeByteDistribution(data []byte) *byteStats {
	stats := &byteStats{
		totalBytes:      len(data),
		firstByteFreq:   make(map[byte]int),
		secondByteFreq:  make(map[byte]int),
	}

	i := 0
	for i < len(data) {
		b := data[i]
		if b <= 0x7F {
			stats.asciiCount++
			i++
		} else {
			stats.highByteCount++
			stats.firstByteFreq[b]++

			if b >= 0x81 && b <= 0x9F || b >= 0xE0 && b <= 0xEF || b >= 0xA1 && b <= 0xDF {
				stats.firstByteInSJISRangeCount++
			}
			if b >= 0xA4 && b <= 0xC3 {
				stats.firstByteInBig5RangeCount++
			}
			if b >= 0xB0 && b <= 0xD7 {
				stats.firstByteInGBKRangeCount++
			}

			if i+1 < len(data) {
				second := data[i+1]
				stats.secondByteFreq[second]++
				stats.doubleByteCount++

				if b >= 0xA1 && b <= 0xDF {
					stats.katakanaCount++
				}

				if second >= 0x40 && second <= 0x7E {
					stats.secondByteIn0x40To0x7ECount++
				}
				if second >= 0xA1 && second <= 0xFE {
					stats.secondByteIn0xA1To0xFECount++
				}
				if second >= 0x80 && second <= 0x9F {
					stats.secondByteIn0x80To0x9FCount++
				}
				if second == 0x7F || (second >= 0x80 && second <= 0xA0) {
					stats.secondByteIn0x7FTo0xA0Count++
				}
			}
			i += 2
		}
	}

	return stats
}

func applyDefinitiveRules(stats *byteStats) string {
	if stats.doubleByteCount == 0 {
		return ""
	}

	if stats.secondByteIn0x80To0x9FCount > 0 {
		return "shift_jis"
	}

	if stats.secondByteIn0x7FTo0xA0Count > 0 {
		sjisFirstByteRatio := float64(stats.firstByteInSJISRangeCount) / float64(stats.doubleByteCount)
		if sjisFirstByteRatio >= 0.8 {
			return "shift_jis"
		}
		return "gbk"
	}

	big5Compatible := stats.secondByteIn0x40To0x7ECount + stats.secondByteIn0xA1To0xFECount
	if big5Compatible == stats.doubleByteCount {
		sjisFirstByteRatio := float64(stats.firstByteInSJISRangeCount) / float64(stats.doubleByteCount)
		big5FirstByteRatio := float64(stats.firstByteInBig5RangeCount) / float64(stats.doubleByteCount)
		gbkFirstByteRatio := float64(stats.firstByteInGBKRangeCount) / float64(stats.doubleByteCount)

		if sjisFirstByteRatio >= 0.8 {
			return "shift_jis"
		}

		if big5FirstByteRatio > gbkFirstByteRatio && big5FirstByteRatio > 0.3 {
			return "big5"
		}

		if gbkFirstByteRatio > big5FirstByteRatio && gbkFirstByteRatio > 0.3 {
			return "gbk"
		}

		if sjisFirstByteRatio > big5FirstByteRatio && sjisFirstByteRatio > 0.3 {
			return "shift_jis"
		}

		return "big5"
	}

	return ""
}

func isASCIIOnly(data []byte) bool {
	for _, b := range data {
		if b > 0x7F {
			return false
		}
	}
	return true
}

func calculateUTF8DistributionScore(data []byte) float64 {
	score := 0.0
	multiByteCount := 0
	validMultiByteCount := 0

	i := 0
	for i < len(data) {
		b := data[i]
		if b <= 0x7F {
			i++
			continue
		}

		multiByteCount++
		seqLen := 0

		switch {
		case b >= 0xC2 && b <= 0xDF:
			seqLen = 2
		case b >= 0xE0 && b <= 0xEF:
			seqLen = 3
		case b >= 0xF0 && b <= 0xF4:
			seqLen = 4
		default:
			i++
			continue
		}

		if i+seqLen <= len(data) {
			valid := true
			for j := 1; j < seqLen; j++ {
				if data[i+j] < 0x80 || data[i+j] > 0xBF {
					valid = false
					break
				}
			}
			if valid {
				validMultiByteCount++
			}
		}

		i += seqLen
	}

	if multiByteCount >
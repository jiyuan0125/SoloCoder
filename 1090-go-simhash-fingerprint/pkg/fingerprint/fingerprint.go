package fingerprint

import (
	"encoding/hex"
	"hash/fnv"
	"strings"
	"unicode"
)

type Config struct {
	Threshold int
	NGram     int
}

type Fingerprint uint64

func (f Fingerprint) String() string {
	b := make([]byte, 8)
	for i := 0; i < 8; i++ {
		b[7-i] = byte(f >> (i * 8))
	}
	return hex.EncodeToString(b)
}

func Parse(s string) (Fingerprint, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	var f uint64
	for i := 0; i < 8; i++ {
		f = (f << 8) | uint64(b[i])
	}
	return Fingerprint(f), nil
}

var DefaultConfig = &Config{
	Threshold: 3,
	NGram:     3,
}

func hashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

func Tokenize(text string) []string {
	f := func(r rune) bool {
		return unicode.IsSpace(r) || unicode.IsPunct(r)
	}
	parts := strings.FieldsFunc(text, f)
	var tokens []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		tokens = append(tokens, strings.ToLower(p))
	}
	return tokens
}

func NGram(text string, n int) []string {
	runes := []rune(strings.ToLower(text))
	var grams []string
	for i := 0; i <= len(runes)-n; i++ {
		grams = append(grams, string(runes[i:i+n]))
	}
	return grams
}

func getFeatures(text string, nGram int) []string {
	features := Tokenize(text)
	if nGram > 0 {
		grams := NGram(text, nGram)
		features = append(features, grams...)
	}
	return features
}

func FingerprintText(text string, cfg *Config) Fingerprint {
	if cfg == nil {
		cfg = DefaultConfig
	}
	features := getFeatures(text, cfg.NGram)
	weights := make(map[string]int)
	for _, f := range features {
		weights[f]++
	}
	var v [64]int
	for feature, weight := range weights {
		h := hashString(feature)
		w := weight
		for i := 0; i < 64; i++ {
			bit := (h >> i) & 1
			if bit == 1 {
				v[i] += w
			} else {
				v[i] -= w
			}
		}
	}
	var fp uint64
	for i := 0; i < 64; i++ {
		if v[i] > 0 {
			fp |= 1 << i
		}
	}
	return Fingerprint(fp)
}

func HammingDistance(a, b Fingerprint) int {
	xor := uint64(a ^ b)
	var count int
	for xor != 0 {
		xor &= xor - 1
		count++
	}
	return count
}

func IsSimilar(a, b Fingerprint, cfg *Config) bool {
	if cfg == nil {
		cfg = DefaultConfig
	}
	return HammingDistance(a, b) <= cfg.Threshold
}

func Dedup(docs []string, cfg *Config) []string {
	if cfg == nil {
		cfg = DefaultConfig
	}
	var result []string
	var fps []Fingerprint
	for _, doc := range docs {
		fp := FingerprintText(doc, cfg)
		duplicate := false
		for _, existing := range fps {
			if HammingDistance(fp, existing) <= cfg.Threshold {
				duplicate = true
				break
			}
		}
		if !duplicate {
			result = append(result, doc)
			fps = append(fps, fp)
		}
	}
	return result
}

func SetThreshold(t int) {
	DefaultConfig.Threshold = t
}

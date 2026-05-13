package aql

import (
	"fmt"
	"strings"
)

type AQLConfig struct {
	LevelSampleSize map[string]map[string]int `yaml:"level_sample_size"`
}

var supportedLevels = []string{"I", "II", "III", "S1", "S2", "S3", "S4"}

func (a *AQLConfig) SupportedLevels() []string {
	return supportedLevels
}

func (a *AQLConfig) IsSupportedLevel(level string) bool {
	upper := strings.ToUpper(strings.TrimSpace(level))
	for _, l := range supportedLevels {
		if l == upper {
			return true
		}
	}
	return false
}

func (a *AQLConfig) GetSupportedLevelsString() string {
	return strings.Join(supportedLevels, ", ")
}

func CalculateSampleSize(batchSize int, inspectionLevel string) (int, error) {
	if batchSize <= 0 {
		return 0, fmt.Errorf("错误：批量大小必须大于 0，当前值: %d", batchSize)
	}

	upperLevel := strings.ToUpper(strings.TrimSpace(inspectionLevel))
	if !isValidLevel(upperLevel) {
		return 0, fmt.Errorf("错误：检验级别 %q 不在标准范围内。支持的级别: %s", inspectionLevel, strings.Join(supportedLevels, ", "))
	}

	return aqlStandardLookup(batchSize, upperLevel), nil
}

func isValidLevel(level string) bool {
	for _, l := range supportedLevels {
		if l == level {
			return true
		}
	}
	return false
}

func aqlStandardLookup(batchSize int, level string) int {
	letter := sampleSizeCode(batchSize, level)
	return codeToSampleSize(letter)
}

func sampleSizeCode(batchSize int, level string) string {
	limits := []int{1, 2, 9, 16, 26, 51, 91, 151, 281, 501, 1201, 3201, 10001, 35001, 150001, 500001}
	codes := map[string][]string{
		"S1":  {"A", "A", "A", "A", "A", "A", "B", "B", "B", "B", "C", "C", "C", "D", "D", "E"},
		"S2":  {"A", "A", "A", "A", "A", "B", "B", "B", "B", "C", "C", "C", "D", "D", "E", "E"},
		"S3":  {"A", "A", "A", "B", "B", "C", "C", "C", "D", "D", "E", "E", "F", "F", "G", "G"},
		"S4":  {"A", "A", "B", "B", "C", "C", "D", "D", "E", "E", "F", "F", "G", "G", "H", "H"},
		"I":   {"A", "A", "B", "B", "C", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N"},
		"II":  {"A", "B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N", "P", "Q", "R"},
		"III": {"B", "C", "D", "E", "F", "G", "H", "J", "K", "L", "M", "N", "P", "Q", "R", "S"},
	}

	for i := len(limits) - 1; i >= 0; i-- {
		if batchSize >= limits[i] {
			levelCodes := codes[level]
			if i < len(levelCodes) {
				return levelCodes[i]
			}
			return levelCodes[len(levelCodes)-1]
		}
	}
	return codes[level][0]
}

func codeToSampleSize(code string) int {
	sampleSizeMap := map[string]int{
		"A": 2, "B": 3, "C": 5, "D": 8, "E": 13, "F": 20,
		"G": 32, "H": 50, "J": 80, "K": 125, "L": 200, "M": 315,
		"N": 500, "P": 800, "Q": 1250, "R": 2000, "S": 3150,
	}
	return sampleSizeMap[code]
}

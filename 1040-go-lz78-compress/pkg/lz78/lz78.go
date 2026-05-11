package lz78

import (
	"errors"
	"fmt"
	"strings"
)

const DefaultMaxDictSize = 4096

type LZ78 struct {
	maxDictSize int
}

func New(maxDictSize ...int) *LZ78 {
	size := DefaultMaxDictSize
	if len(maxDictSize) > 0 && maxDictSize[0] > 1 {
		size = maxDictSize[0]
	}
	return &LZ78{
		maxDictSize: size,
	}
}

func (l *LZ78) Compress(input string) ([]int, error) {
	if input == "" {
		return []int{}, nil
	}

	dict := map[string]int{
		"": 0,
	}
	nextIndex := 1

	result := []int{}
	i := 0

	for i < len(input) {
		longestMatch := ""
		matchIndex := 0

		for j := i + 1; j <= len(input); j++ {
			current := input[i:j]
			if idx, exists := dict[current]; exists {
				longestMatch = current
				matchIndex = idx
			} else {
				break
			}
		}

		if len(longestMatch) > 0 {
			result = append(result, matchIndex)
		} else {
			result = append(result, 0)
			result = append(result, int(input[i]))
		}

		if nextIndex < l.maxDictSize && i+len(longestMatch) < len(input) {
			newEntry := longestMatch + string(input[i+len(longestMatch)])
			if _, exists := dict[newEntry]; !exists {
				dict[newEntry] = nextIndex
				nextIndex++
			}
		}

		i += len(longestMatch)
		if len(longestMatch) == 0 {
			if nextIndex < l.maxDictSize {
				char := string(input[i])
				if _, exists := dict[char]; !exists {
					dict[char] = nextIndex
					nextIndex++
				}
			}
			i++
		}
	}

	return result, nil
}

func (l *LZ78) CompressWithDict(input string) ([]int, map[int]string, error) {
	if input == "" {
		return []int{}, map[int]string{0: ""}, nil
	}

	dict := map[string]int{
		"": 0,
	}
	indexToStr := map[int]string{
		0: "",
	}
	nextIndex := 1

	result := []int{}
	i := 0

	for i < len(input) {
		longestMatch := ""
		matchIndex := 0

		for j := i + 1; j <= len(input); j++ {
			current := input[i:j]
			if idx, exists := dict[current]; exists {
				longestMatch = current
				matchIndex = idx
			} else {
				break
			}
		}

		if len(longestMatch) > 0 {
			result = append(result, matchIndex)
		} else {
			result = append(result, 0)
			result = append(result, int(input[i]))
		}

		if nextIndex < l.maxDictSize && i+len(longestMatch) < len(input) {
			newEntry := longestMatch + string(input[i+len(longestMatch)])
			if _, exists := dict[newEntry]; !exists {
				dict[newEntry] = nextIndex
				indexToStr[nextIndex] = newEntry
				nextIndex++
			}
		}

		i += len(longestMatch)
		if len(longestMatch) == 0 {
			if nextIndex < l.maxDictSize {
				char := string(input[i])
				if _, exists := dict[char]; !exists {
					dict[char] = nextIndex
					indexToStr[nextIndex] = char
					nextIndex++
				}
			}
			i++
		}
	}

	return result, indexToStr, nil
}

func (l *LZ78) Decompress(indexes []int) (string, error) {
	if len(indexes) == 0 {
		return "", nil
	}

	indexToStr := []string{""}
	nextIndex := 1

	var result strings.Builder
	i := 0

	for i < len(indexes) {
		idx := indexes[i]

		if idx == 0 {
			if i+1 >= len(indexes) {
				return "", errors.New("corrupted data: index 0 without following character code")
			}
			charCode := indexes[i+1]
			if charCode < 0 || charCode > 0x10FFFF {
				return "", errors.New(fmt.Sprintf("corrupted data: invalid character code %d", charCode))
			}
			char := rune(charCode)
			charStr := string(char)
			result.WriteString(charStr)

			if nextIndex < l.maxDictSize {
				found := false
				for _, s := range indexToStr {
					if s == charStr {
						found = true
						break
					}
				}
				if !found {
					indexToStr = append(indexToStr, charStr)
					nextIndex++
				}
			}

			if i < len(indexes)-2 {
				var charToAdd string
				nextPos := i + 2

				if nextPos < len(indexes) {
					nextIdx := indexes[nextPos]
					if nextIdx == 0 {
						if nextPos+1 < len(indexes) {
							nextCharCode := indexes[nextPos+1]
							if nextCharCode >= 0 && nextCharCode <= 0x10FFFF {
								charToAdd = string(rune(nextCharCode))
							}
						}
					} else if nextIdx > 0 && nextIdx < len(indexToStr) {
						nextStr := indexToStr[nextIdx]
						if len(nextStr) > 0 {
							charToAdd = string(nextStr[0])
						}
					}
				}

				if charToAdd != "" && nextIndex < l.maxDictSize {
					newEntry := "" + charToAdd
					if newEntry != "" {
						found := false
						for _, s := range indexToStr {
							if s == newEntry {
								found = true
								break
							}
						}
						if !found {
							indexToStr = append(indexToStr, newEntry)
							nextIndex++
						}
					}
				}
			}
			i += 2
			continue
		}

		if idx < 0 || idx >= len(indexToStr) {
			return "", errors.New(fmt.Sprintf("corrupted data: invalid index %d at position %d (dict size: %d)", idx, i, len(indexToStr)))
		}

		currentStr := indexToStr[idx]
		result.WriteString(currentStr)

		if i < len(indexes)-1 {
			var charToAdd string
			nextPos := i + 1

			if nextPos < len(indexes) {
				nextIdx := indexes[nextPos]
				if nextIdx == 0 {
					if nextPos+1 < len(indexes) {
						nextCharCode := indexes[nextPos+1]
						if nextCharCode >= 0 && nextCharCode <= 0x10FFFF {
							charToAdd = string(rune(nextCharCode))
						}
					}
				} else if nextIdx > 0 && nextIdx < len(indexToStr) {
					nextStr := indexToStr[nextIdx]
					if len(nextStr) > 0 {
						charToAdd = string(nextStr[0])
					}
				}
			}

			if charToAdd == "" && len(currentStr) > 0 {
				charToAdd = string(currentStr[0])
			}

			if nextIndex < l.maxDictSize && charToAdd != "" {
				newEntry := currentStr + charToAdd
				if newEntry != "" {
					found := false
					for _, s := range indexToStr {
						if s == newEntry {
							found = true
							break
						}
					}
					if !found {
						indexToStr = append(indexToStr, newEntry)
						nextIndex++
					}
				}
			}
		}

		i++
	}

	return result.String(), nil
}

func (l *LZ78) DecompressWithDict(indexes []int) (string, map[int]string, error) {
	if len(indexes) == 0 {
		return "", map[int]string{0: ""}, nil
	}

	indexToStrSlice := []string{""}
	indexToStrMap := map[int]string{0: ""}
	nextIndex := 1

	var result strings.Builder
	i := 0

	for i < len(indexes) {
		idx := indexes[i]

		if idx == 0 {
			if i+1 >= len(indexes) {
				return "", nil, errors.New("corrupted data: index 0 without following character code")
			}
			charCode := indexes[i+1]
			if charCode < 0 || charCode > 0x10FFFF {
				return "", nil, errors.New(fmt.Sprintf("corrupted data: invalid character code %d", charCode))
			}
			char := rune(charCode)
			charStr := string(char)
			result.WriteString(charStr)

			if nextIndex < l.maxDictSize {
				found := false
				for _, s := range indexToStrSlice {
					if s == charStr {
						found = true
						break
					}
				}
				if !found {
					indexToStrSlice = append(indexToStrSlice, charStr)
					indexToStrMap[nextIndex] = charStr
					nextIndex++
				}
			}

			if i < len(indexes)-2 {
				var charToAdd string
				nextPos := i + 2

				if nextPos < len(indexes) {
					nextIdx := indexes[nextPos]
					if nextIdx == 0 {
						if nextPos+1 < len(indexes) {
							nextCharCode := indexes[nextPos+1]
							if nextCharCode >= 0 && nextCharCode <= 0x10FFFF {
								charToAdd = string(rune(nextCharCode))
							}
						}
					} else if nextIdx > 0 && nextIdx < len(indexToStrSlice) {
						nextStr := indexToStrSlice[nextIdx]
						if len(nextStr) > 0 {
							charToAdd = string(nextStr[0])
						}
					}
				}

				if charToAdd != "" && nextIndex < l.maxDictSize {
					newEntry := "" + charToAdd
					if newEntry != "" {
						found := false
						for _, s := range indexToStrSlice {
							if s == newEntry {
								found = true
								break
							}
						}
						if !found {
							indexToStrSlice = append(indexToStrSlice, newEntry)
							indexToStrMap[nextIndex] = newEntry
							nextIndex++
						}
					}
				}
			}
			i += 2
			continue
		}

		if idx < 0 || idx >= len(indexToStrSlice) {
			return "", nil, errors.New(fmt.Sprintf("corrupted data: invalid index %d at position %d (dict size: %d)", idx, i, len(indexToStrSlice)))
		}

		currentStr := indexToStrSlice[idx]
		result.WriteString(currentStr)

		if i < len(indexes)-1 {
			var charToAdd string
			nextPos := i + 1

			if nextPos < len(indexes) {
				nextIdx := indexes[nextPos]
				if nextIdx == 0 {
					if nextPos+1 < len(indexes) {
						nextCharCode := indexes[nextPos+1]
						if nextCharCode >= 0 && nextCharCode <= 0x10FFFF {
							charToAdd = string(rune(nextCharCode))
						}
					}
				} else if nextIdx > 0 && nextIdx < len(indexToStrSlice) {
					nextStr := indexToStrSlice[nextIdx]
					if len(nextStr) > 0 {
						charToAdd = string(nextStr[0])
					}
				}
			}

			if charToAdd == "" && len(currentStr) > 0 {
				charToAdd = string(currentStr[0])
			}

			if nextIndex < l.maxDictSize && charToAdd != "" {
				newEntry := currentStr + charToAdd
				if newEntry != "" {
					found := false
					for _, s := range indexToStrSlice {
						if s == newEntry {
							found = true
							break
						}
					}
					if !found {
						indexToStrSlice = append(indexToStrSlice, newEntry)
						indexToStrMap[nextIndex] = newEntry
						nextIndex++
					}
				}
			}
		}

		i++
	}

	return result.String(), indexToStrMap, nil
}

func (l *LZ78) GetMaxDictSize() int {
	return l.maxDictSize
}

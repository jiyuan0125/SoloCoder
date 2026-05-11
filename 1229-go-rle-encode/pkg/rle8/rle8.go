package rle8

import (
	"fmt"
)

func Encode(pixels [][]byte) ([]byte, error) {
	if len(pixels) == 0 {
		return []byte{}, nil
	}

	var result []byte

	for _, row := range pixels {
		if len(row) == 0 {
			continue
		}

		i := 0
		for i < len(row) {
			current := row[i]
			count := 1

			for i+count < len(row) && row[i+count] == current && count < 255 {
				count++
			}

			if count >= 3 {
				for count > 255 {
					result = append(result, 0xFF, current)
					count -= 255
				}
				result = append(result, byte(count), current)
			} else {
				absoluteStart := i
				absoluteEnd := i

				for absoluteEnd < len(row) {
					testPos := absoluteEnd
					testCount := 1
					for testPos+testCount < len(row) && row[testPos+testCount] == row[testPos] && testCount < 3 {
						testCount++
					}

					if testCount >= 3 {
						break
					}

					absoluteEnd++

					if absoluteEnd-absoluteStart >= 255 {
						break
					}
				}

				absLen := absoluteEnd - absoluteStart
				if absLen > 0 {
					result = append(result, 0x00, byte(absLen))
					result = append(result, row[absoluteStart:absoluteEnd]...)
					if absLen%2 == 1 {
						result = append(result, 0x00)
					}
					i = absoluteEnd
					continue
				}
			}

			i += count
		}

		result = append(result, 0x00, 0x00)
	}

	result = append(result, 0x00, 0x01)

	return result, nil
}

func Decode(data []byte, width int) ([][]byte, error) {
	if width <= 0 {
		return nil, fmt.Errorf("width must be positive")
	}

	var result [][]byte
	var currentRow []byte
	i := 0

	for i < len(data) {
		count := data[i]
		i++

		if i >= len(data) {
			break
		}
		escape := data[i]
		i++

		if count == 0 {
			switch escape {
			case 0x00:
				for len(currentRow) < width {
					currentRow = append(currentRow, 0x00)
				}
				result = append(result, currentRow)
				currentRow = []byte{}
			case 0x01:
				if len(currentRow) > 0 {
					for len(currentRow) < width {
						currentRow = append(currentRow, 0x00)
					}
					result = append(result, currentRow)
				}
				return result, nil
			case 0x02:
				if i+1 >= len(data) {
					return result, fmt.Errorf("insufficient data for delta offset")
				}
				dx := int(data[i])
				dy := int(data[i+1])
				i += 2

				for dy > 0 {
					for len(currentRow) < width {
						currentRow = append(currentRow, 0x00)
					}
					result = append(result, currentRow)
					currentRow = []byte{}
					dy--
				}

				for len(currentRow) < dx && len(currentRow) < width {
					currentRow = append(currentRow, 0x00)
				}
			default:
				n := int(escape)
				if i+n > len(data) {
					return result, fmt.Errorf("insufficient data for absolute mode")
				}
				for j := 0; j < n; j++ {
					currentRow = append(currentRow, data[i+j])
				}
				i += n
				if n%2 == 1 {
					i++
				}
			}
		} else {
			for j := 0; j < int(count); j++ {
				currentRow = append(currentRow, escape)
			}
		}
	}

	if len(currentRow) > 0 {
		for len(currentRow) < width {
			currentRow = append(currentRow, 0x00)
		}
		result = append(result, currentRow)
	}

	return result, nil
}

package barcode

import (
	"errors"
	"unicode"
)

type Encoder interface {
	Encode(input string) (*Result, error)
}

type Result struct {
	Pattern       string
	WidthSequence []int
	TotalModules  int
	Input         string
}

type code128BEncoder struct{}

func NewEncoder() Encoder {
	return &code128BEncoder{}
}

func (e *code128BEncoder) Encode(input string) (*Result, error) {
	if err := validateInput(input); err != nil {
		return nil, err
	}

	var codeValues []int
	codeValues = append(codeValues, startCodeB)

	for _, char := range input {
		value, _ := getCodeBValue(char)
		codeValues = append(codeValues, value)
	}

	checksum := calculateChecksum(codeValues)
	codeValues = append(codeValues, checksum)
	codeValues = append(codeValues, stopCode)

	pattern := ""
	for _, value := range codeValues {
		p, exists := getPattern(value)
		if !exists {
			return nil, errors.New("failed to get pattern for code value")
		}
		pattern += p
	}

	widthSequence := patternToWidthSequence(pattern)
	totalModules := calculateTotalModules(len(input))

	return &Result{
		Pattern:       pattern,
		WidthSequence: widthSequence,
		TotalModules:  totalModules,
		Input:         input,
	}, nil
}

func validateInput(input string) error {
	if len(input) == 0 {
		return errors.New("input string cannot be empty")
	}

	for _, char := range input {
		if char > unicode.MaxASCII {
			return errors.New("input contains non-ASCII characters")
		}
		if char < 32 || char > 126 {
			return errors.New("input contains non-printable ASCII characters")
		}
	}

	return nil
}

func calculateChecksum(codeValues []int) int {
	sum := codeValues[0]

	for i := 1; i < len(codeValues); i++ {
		sum += i * codeValues[i]
	}

	return sum % 103
}

func patternToWidthSequence(pattern string) []int {
	if len(pattern) == 0 {
		return []int{}
	}

	var sequence []int
	currentRun := 1
	currentBit := pattern[0]

	for i := 1; i < len(pattern); i++ {
		if pattern[i] == currentBit {
			currentRun++
		} else {
			sequence = append(sequence, currentRun)
			currentRun = 1
			currentBit = pattern[i]
		}
	}

	sequence = append(sequence, currentRun)

	return sequence
}

func calculateTotalModules(inputLength int) int {
	return 11 + 11*inputLength + 11 + 13
}

package search

import (
	"strings"
	"unicode"
)

func Tokenize(text string) []string {
	var tokens []string
	var current strings.Builder
	var currentRunes []rune

	runes := []rune(text)
	for i, r := range runes {
		if unicode.IsSpace(r) || isPunctuation(r) {
			if currentRunes != nil {
				tokens = append(tokens, string(currentRunes))
				currentRunes = nil
			}
			current.Reset()
			continue
		}

		if isChinese(r) {
			if currentRunes != nil {
				tokens = append(tokens, string(currentRunes))
				currentRunes = nil
			}
			tokens = append(tokens, string(r))
			continue
		}

		if isAlphaNum(r) {
			currentRunes = append(currentRunes, r)
		}

		if i == len(runes)-1 && currentRunes != nil {
			tokens = append(tokens, string(currentRunes))
		}
	}

	return tokens
}

func isChinese(r rune) bool {
	return unicode.Is(unicode.Han, r)
}

func isAlphaNum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isPunctuation(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

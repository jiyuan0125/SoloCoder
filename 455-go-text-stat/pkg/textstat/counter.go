package textstat

import (
	"regexp"
	"strings"
	"unicode"
)

func CountChars(text string, includeSpaces bool) int {
	if text == "" {
		return 0
	}
	if includeSpaces {
		return len([]rune(text))
	}
	count := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			count++
		}
	}
	return count
}

func CountWords(text string) int {
	if text == "" {
		return 0
	}
	words := SplitWords(text)
	return len(words)
}

func CountLines(text string) int {
	if text == "" {
		return 0
	}
	lines := strings.Split(text, "\n")
	return len(lines)
}

func CountParagraphs(text string) int {
	if text == "" {
		return 0
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return 0
	}
	re := regexp.MustCompile(`\n\s*\n`)
	paragraphs := re.Split(text, -1)
	count := 0
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			count++
		}
	}
	return count
}

func SplitWords(text string) []string {
	if text == "" {
		return []string{}
	}
	var words []string
	var currentWord []rune
	isEnglish := false
	isChinese := false

	flushWord := func() {
		if len(currentWord) > 0 {
			words = append(words, string(currentWord))
			currentWord = nil
			isEnglish = false
			isChinese = false
		}
	}

	for _, r := range text {
		if isPunctuation(r) || unicode.IsSpace(r) {
			flushWord()
			continue
		}

		if unicode.Is(unicode.Scripts["Han"], r) {
			if isEnglish && len(currentWord) > 0 {
				flushWord()
			}
			words = append(words, string(r))
			isChinese = true
			isEnglish = false
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' {
			if isChinese && len(currentWord) > 0 {
				flushWord()
			}
			currentWord = append(currentWord, r)
			isEnglish = true
			isChinese = false
		} else {
			flushWord()
		}
	}
	flushWord()

	return words
}

func isPunctuation(r rune) bool {
	return unicode.IsPunct(r) || unicode.IsSymbol(r)
}

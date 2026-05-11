package bayes

import (
	"regexp"
	"strings"
	"unicode"
)

type Tokenizer struct {
	lowerCase   bool
	keepURLs    bool
	keepEmails  bool
	filterFunc  func(rune) bool
}

func NewTokenizer() *Tokenizer {
	return &Tokenizer{
		lowerCase:  true,
		keepURLs:   true,
		keepEmails: true,
		filterFunc: func(r rune) bool {
			return unicode.IsPunct(r) || unicode.IsSymbol(r) || unicode.IsDigit(r)
		},
	}
}

var (
	urlRegex    = regexp.MustCompile(`https?://[^\s]+`)
	emailRegex  = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
)

func (t *Tokenizer) Tokenize(text string) []string {
	if t.lowerCase {
		text = strings.ToLower(text)
	}

	var specialTokens []string
	if t.keepURLs {
		urls := urlRegex.FindAllString(text, -1)
		specialTokens = append(specialTokens, urls...)
		text = urlRegex.ReplaceAllString(text, " ")
	}
	if t.keepEmails {
		emails := emailRegex.FindAllString(text, -1)
		specialTokens = append(specialTokens, emails...)
		text = emailRegex.ReplaceAllString(text, " ")
	}

	var tokens []string
	var currentWord strings.Builder

	for _, r := range text {
		if t.filterFunc(r) {
			if currentWord.Len() > 0 {
				tokens = append(tokens, currentWord.String())
				currentWord.Reset()
			}
			continue
		}
		if unicode.IsSpace(r) {
			if currentWord.Len() > 0 {
				tokens = append(tokens, currentWord.String())
				currentWord.Reset()
			}
			continue
		}
		currentWord.WriteRune(r)
	}

	if currentWord.Len() > 0 {
		tokens = append(tokens, currentWord.String())
	}

	tokens = append(tokens, specialTokens...)

	var result []string
	for _, token := range tokens {
		if len(token) > 0 {
			result = append(result, token)
		}
	}

	return result
}

func (t *Tokenizer) TokenizeChinese(text string) []string {
	if t.lowerCase {
		text = strings.ToLower(text)
	}

	var specialTokens []string
	if t.keepURLs {
		urls := urlRegex.FindAllString(text, -1)
		specialTokens = append(specialTokens, urls...)
		text = urlRegex.ReplaceAllString(text, " ")
	}
	if t.keepEmails {
		emails := emailRegex.FindAllString(text, -1)
		specialTokens = append(specialTokens, emails...)
		text = emailRegex.ReplaceAllString(text, " ")
	}

	var tokens []string
	var englishWord strings.Builder

	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			if englishWord.Len() > 0 {
				tokens = append(tokens, englishWord.String())
				englishWord.Reset()
			}
			tokens = append(tokens, string(r))
			continue
		}
		if unicode.IsLetter(r) {
			englishWord.WriteRune(r)
			continue
		}
		if unicode.IsDigit(r) {
			if englishWord.Len() > 0 {
				tokens = append(tokens, englishWord.String())
				englishWord.Reset()
			}
			continue
		}
		if unicode.IsSpace(r) || unicode.IsPunct(r) || unicode.IsSymbol(r) {
			if englishWord.Len() > 0 {
				tokens = append(tokens, englishWord.String())
				englishWord.Reset()
			}
		}
	}

	if englishWord.Len() > 0 {
		tokens = append(tokens, englishWord.String())
	}

	tokens = append(tokens, specialTokens...)

	return tokens
}

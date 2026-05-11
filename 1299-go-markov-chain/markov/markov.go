package markov

import (
	"errors"
	"math/rand"
	"strings"
	"unicode/utf8"
)

type Chain struct {
	order      int
	gramSize   int
	tokenizer  func(string) []string
	joiner     func([]string) string
	transitions map[string]map[string]int
	prefixes    []string
	rng        *rand.Rand
}

func NewCharChain(order int, seed int64) (*Chain, error) {
	if order < 1 {
		return nil, errors.New("order must be at least 1")
	}

	return &Chain{
		order:       order,
		gramSize:    order,
		tokenizer:   tokenizeChars,
		joiner:      joinChars,
		transitions: make(map[string]map[string]int),
		prefixes:    []string{},
		rng:         rand.New(rand.NewSource(seed)),
	}, nil
}

func NewWordChain(order int, seed int64) (*Chain, error) {
	if order < 1 {
		return nil, errors.New("order must be at least 1")
	}

	return &Chain{
		order:       order,
		gramSize:    order,
		tokenizer:   tokenizeWords,
		joiner:      joinWords,
		transitions: make(map[string]map[string]int),
		prefixes:    []string{},
		rng:         rand.New(rand.NewSource(seed)),
	}, nil
}

func (c *Chain) Train(text string) error {
	if text == "" {
		return nil
	}

	tokens := c.tokenizer(text)
	if len(tokens) == 0 {
		return nil
	}

	if len(tokens) <= c.gramSize {
		return errors.New("text too short for given order")
	}

	for i := 0; i <= len(tokens)-c.gramSize-1; i++ {
		prefixTokens := tokens[i : i+c.gramSize]
		prefix := serializePrefix(prefixTokens)
		nextToken := tokens[i+c.gramSize]

		if _, exists := c.transitions[prefix]; !exists {
			c.transitions[prefix] = make(map[string]int)
			c.prefixes = append(c.prefixes, prefix)
		}
		c.transitions[prefix][nextToken]++
	}

	return nil
}

func (c *Chain) Generate(length int) (string, error) {
	if length == 0 {
		return "", nil
	}

	if len(c.transitions) == 0 {
		return "", errors.New("model not trained")
	}

	if len(c.prefixes) == 0 {
		return "", errors.New("no valid prefixes")
	}

	startIdx := c.rng.Intn(len(c.prefixes))
	currentPrefix := c.prefixes[startIdx]
	currentTokens := deserializePrefix(currentPrefix)

	resultTokens := make([]string, len(currentTokens))
	copy(resultTokens, currentTokens)

	for len(resultTokens) < length {
		nextToken, err := c.nextToken(currentPrefix)
		if err != nil {
			newStartIdx := c.rng.Intn(len(c.prefixes))
			currentPrefix = c.prefixes[newStartIdx]
			currentTokens = deserializePrefix(currentPrefix)
			resultTokens = append(resultTokens, currentTokens...)
			continue
		}

		resultTokens = append(resultTokens, nextToken)
		currentTokens = append(currentTokens[1:], nextToken)
		currentPrefix = serializePrefix(currentTokens)
	}

	if len(resultTokens) > length {
		resultTokens = resultTokens[:length]
	}

	return c.joiner(resultTokens), nil
}

func (c *Chain) nextToken(prefix string) (string, error) {
	nextTokens, exists := c.transitions[prefix]
	if !exists || len(nextTokens) == 0 {
		return "", errors.New("no transition")
	}

	total := 0
	for _, count := range nextTokens {
		total += count
	}

	if total == 0 {
		return "", errors.New("no transitions")
	}

	randomValue := c.rng.Intn(total)
	cumulative := 0

	for token, count := range nextTokens {
		cumulative += count
		if randomValue < cumulative {
			return token, nil
		}
	}

	for token := range nextTokens {
		return token, nil
	}

	return "", errors.New("no token found")
}

func tokenizeChars(text string) []string {
	var tokens []string
	for _, r := range text {
		tokens = append(tokens, string(r))
	}
	return tokens
}

func joinChars(tokens []string) string {
	return strings.Join(tokens, "")
}

func tokenizeWords(text string) []string {
	return strings.Fields(text)
}

func joinWords(tokens []string) string {
	return strings.Join(tokens, " ")
}

const prefixSeparator = "\x00"

func serializePrefix(tokens []string) string {
	return strings.Join(tokens, prefixSeparator)
}

func deserializePrefix(prefix string) []string {
	return strings.Split(prefix, prefixSeparator)
}

func CountRunes(s string) int {
	return utf8.RuneCountInString(s)
}

package generator

import (
	"bufio"
	"crypto/rand"
	"errors"
	"math/big"
	"os"
	"strings"
)

var builtinWords = []string{
	"apple", "banana", "cherry", "dragon", "elephant",
	"flower", "garden", "happy", "island", "jungle",
	"kingdom", "lemon", "mountain", "night", "ocean",
	"piano", "queen", "river", "sunset", "tiger",
	"umbrella", "valley", "water", "xylophone", "yellow",
	"zebra", "anchor", "bridge", "castle", "dolphin",
	"eagle", "forest", "galaxy", "harbor", "iceberg",
	"jasmine", "knight", "lighthouse", "magnolia", "nebula",
	"orchid", "peacock", "quartz", "rainbow", "starlight",
	"thunder", "unicorn", "violet", "willow", "zenith",
	"amber", "blossom", "crimson", "diamond", "emerald",
	"frost", "glacier", "horizon", "iris", "jade",
	"karma", "luna", "mystic", "nova", "opal",
	"phoenix", "ruby", "sapphire", "turquoise", "viper",
	"wolf", "yacht", "zeppelin", "astral", "breeze",
	"canyon", "delta", "ember", "flame", "grove",
}

func (g *Generator) GenerateReadable(length int, separator string, dictFile string) (string, error) {
	if length < MinLength {
		return "", errors.New("密码长度不能小于 4")
	}

	words, err := g.loadDictionary(dictFile)
	if err != nil {
		return "", err
	}

	return g.buildReadablePassword(length, separator, words)
}

func (g *Generator) loadDictionary(dictFile string) ([]string, error) {
	if dictFile == "" {
		return builtinWords, nil
	}

	file, err := os.Open(dictFile)
	if err != nil {
		return builtinWords, nil
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" {
			words = append(words, word)
		}
	}

	if err := scanner.Err(); err != nil {
		return builtinWords, nil
	}

	if len(words) == 0 {
		return builtinWords, nil
	}

	return words, nil
}

func (g *Generator) buildReadablePassword(length int, separator string, words []string) (string, error) {
	if separator == "" {
		separator = "-"
	}

	if len(words) == 0 {
		words = builtinWords
	}

	wordLen := big.NewInt(int64(len(words)))
	var result []string

	attempts := 0
	for {
		idx, err := rand.Int(rand.Reader, wordLen)
		if err != nil {
			return "", err
		}
		result = append(result, words[idx.Int64()])

		currentLen := len(strings.Join(result, separator))
		if currentLen >= length {
			break
		}

		attempts++
		if attempts > 100 {
			break
		}
	}

	password := strings.Join(result, separator)

	for len(password) < length {
		idx, err := rand.Int(rand.Reader, wordLen)
		if err != nil {
			return password, nil
		}
		password = password + separator + words[idx.Int64()]
	}

	return password, nil
}

package emojinorm

import (
	"strings"
	"unicode"
)

const (
	zwj           rune = 0x200D
	variation16    rune = 0xFE0F
	regionalIndicatorA rune = 0x1F1E6
	regionalIndicatorZ rune = 0x1F1FF
)

var skinToneModifiers = []rune{
	0x1F3FB, 0x1F3FC, 0x1F3FD, 0x1F3FE, 0x1F3FF,
}

var emojiRanges = []struct {
	start rune
	end   rune
}{
	{0x1F600, 0x1F64F},
	{0x1F300, 0x1F5FF},
	{0x1F680, 0x1F6FF},
	{0x1F900, 0x1F9FF},
	{0x1F1E6, 0x1F1FF},
	{0x2600, 0x26FF},
	{0x2700, 0x27BF},
	{0x1F700, 0x1F77F},
	{0x1FA70, 0x1FAFF},
}

func isEmojiBase(r rune) bool {
	for _, rng := range emojiRanges {
		if r >= rng.start && r <= rng.end {
			return true
		}
	}
	return false
}

func isSkinToneModifier(r rune) bool {
	for _, mod := range skinToneModifiers {
		if r == mod {
			return true
		}
	}
	return false
}

func isVariationSelector(r rune) bool {
	return r == variation16
}

func isRegionalIndicator(r rune) bool {
	return r >= regionalIndicatorA && r <= regionalIndicatorZ
}

func ExtractEmojis(text string) []EmojiInfo {
	var result []EmojiInfo
	runes := []rune(text)
	bytePos := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		runeLen := len(string(r))

		if isEmojiBase(r) {
			startByte := bytePos
			endByte := bytePos + runeLen
			emojiRunes := []rune{r}

			j := i + 1
			for j < len(runes) {
				next := runes[j]
				if isSkinToneModifier(next) || isVariationSelector(next) {
					emojiRunes = append(emojiRunes, next)
					endByte += len(string(next))
					j++
				} else if next == zwj && j+1 < len(runes) {
					emojiRunes = append(emojiRunes, next)
					endByte += len(string(next))
					j++
					if j < len(runes) {
						emojiRunes = append(emojiRunes, runes[j])
						endByte += len(string(runes[j]))
						j++
					}
				} else {
					break
				}
			}

			if isRegionalIndicator(r) && j < len(runes) && isRegionalIndicator(runes[j]) {
				emojiRunes = append(emojiRunes, runes[j])
				endByte += len(string(runes[j]))
				j++
			}

			emojiStr := string(emojiRunes)
			if !isSymbolOnly(emojiStr) {
				result = append(result, EmojiInfo{
					StartByte: startByte,
					EndByte:   endByte,
					Emoji:     emojiStr,
				})
			}

			i = j - 1
			bytePos = endByte
		} else {
			bytePos += runeLen
		}
	}

	return result
}

func CountEmojis(text string) int {
	return len(ExtractEmojis(text))
}

func RemoveEmojis(text string) string {
	emojis := ExtractEmojis(text)
	if len(emojis) == 0 {
		return text
	}

	var builder strings.Builder
	lastPos := 0

	for _, emoji := range emojis {
		builder.WriteString(text[lastPos:emoji.StartByte])
		lastPos = emoji.EndByte
	}

	builder.WriteString(text[lastPos:])
	return builder.String()
}

func ReplaceEmojis(text string, replacement string) string {
	emojis := ExtractEmojis(text)
	if len(emojis) == 0 {
		return text
	}

	var builder strings.Builder
	lastPos := 0

	for _, emoji := range emojis {
		builder.WriteString(text[lastPos:emoji.StartByte])
		builder.WriteString(replacement)
		lastPos = emoji.EndByte
	}

	builder.WriteString(text[lastPos:])
	return builder.String()
}

func IsOnlyEmojis(text string) bool {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return false
	}

	emojis := ExtractEmojis(trimmed)
	if len(emojis) == 0 {
		return false
	}

	nonEmoji := RemoveEmojis(trimmed)
	nonEmojiTrimmed := strings.TrimSpace(nonEmoji)

	if nonEmojiTrimmed == "" {
		return true
	}

	for _, r := range nonEmojiTrimmed {
		if !unicode.IsSpace(r) && !unicode.IsPunct(r) {
			return false
		}
	}

	return true
}

func isSymbolOnly(text string) bool {
	notEmojiSymbols := []rune{
		0x2122,
		0x00A9,
		0x00AE,
	}

	for _, r := range []rune(text) {
		for _, sym := range notEmojiSymbols {
			if r == sym {
				return true
			}
		}
	}

	return false
}

func ContainsEmoji(text string) bool {
	for _, r := range []rune(text) {
		if isEmojiBase(r) {
			return true
		}
	}
	return false
}

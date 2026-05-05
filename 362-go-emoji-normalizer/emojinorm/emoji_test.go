package emojinorm

import (
	"testing"
)

func TestExtractEmojis(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty string", "", 0},
		{"no emojis", "Hello World", 0},
		{"single emoji", "Hello 👋", 1},
		{"multiple emojis", "👋 🌍 😊", 3},
		{"family emoji with ZWJ", "Hello 👨‍👩‍👧‍👦 World", 1},
		{"skin tone modifier", "👋🏻", 1},
		{"flag emoji", "🇨🇳", 1},
		{"mixed text and emojis", "Hello 👋 World 🌍!", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractEmojis(tt.text)
			if len(result) != tt.expected {
				t.Errorf("Expected %d emojis, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestCountEmojis(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int
	}{
		{"empty string", "", 0},
		{"no emojis", "Hello World", 0},
		{"single emoji", "Hello 👋", 1},
		{"multiple emojis", "👋 🌍 😊", 3},
		{"family emoji with ZWJ", "👨‍👩‍👧‍👦", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CountEmojis(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %d emojis, got %d", tt.expected, result)
			}
		})
	}
}

func TestRemoveEmojis(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{"empty string", "", ""},
		{"no emojis", "Hello World", "Hello World"},
		{"remove emoji", "Hello 👋 World", "Hello  World"},
		{"remove multiple emojis", "👋 Hello 🌍", " Hello "},
		{"remove family emoji", "Test 👨‍👩‍👧‍👦 End", "Test  End"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RemoveEmojis(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestReplaceEmojis(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		replacement string
		expected    string
	}{
		{"empty string", "", "[表情]", ""},
		{"no emojis", "Hello World", "[表情]", "Hello World"},
		{"replace single", "Hello 👋", "[表情]", "Hello [表情]"},
		{"replace multiple", "👋 🌍", "[EMOJI]", "[EMOJI] [EMOJI]"},
		{"replace family", "👨‍👩‍👧‍👦 Test", "[表情]", "[表情] Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceEmojis(tt.text, tt.replacement)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestIsOnlyEmojis(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"empty string", "", false},
		{"only spaces", "   ", false},
		{"only punctuation", "!!!", false},
		{"single emoji", "👋", true},
		{"multiple emojis", "👋 🌍 😊", true},
		{"family emoji", "👨‍👩‍👧‍👦", true},
		{"mixed text and emoji", "Hello 👋", false},
		{"emoji with spaces", " 👋 ", true},
		{"emoji with punctuation", "👋!", true},
		{"only special symbols", "™️", false},
		{"copyright symbol", "©️", false},
		{"registered symbol", "®️", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOnlyEmojis(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for %q", tt.expected, result, tt.text)
			}
		})
	}
}

func TestContainsEmoji(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected bool
	}{
		{"empty string", "", false},
		{"no emojis", "Hello World", false},
		{"with emoji", "Hello 👋", true},
		{"family emoji", "Test 👨‍👩‍👧‍👦", true},
		{"only special symbols", "™️ ©️", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsEmoji(tt.text)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for %q", tt.expected, result, tt.text)
			}
		})
	}
}

func TestEmojiPositions(t *testing.T) {
	text := "Hello 👋 World"
	emojis := ExtractEmojis(text)

	if len(emojis) != 1 {
		t.Fatalf("Expected 1 emoji, got %d", len(emojis))
	}

	emoji := emojis[0]
	expectedStart := 6
	expectedEnd := 10

	if emoji.StartByte != expectedStart {
		t.Errorf("Expected start byte %d, got %d", expectedStart, emoji.StartByte)
	}

	if emoji.EndByte != expectedEnd {
		t.Errorf("Expected end byte %d, got %d", expectedEnd, emoji.EndByte)
	}

	if emoji.Emoji != "👋" {
		t.Errorf("Expected emoji %q, got %q", "👋", emoji.Emoji)
	}
}

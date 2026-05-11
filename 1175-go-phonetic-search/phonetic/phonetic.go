package phonetic

type PhoneticCode struct {
	Soundex   string
	Metaphone string
}

type NameStore interface {
	Add(name string) bool
	Remove(name string) bool
	SearchBySoundex(code string) []string
	SearchByMetaphone(code string) []string
	Total() int
	UniqueSoundexCodes() int
	UniqueMetaphoneCodes() int
	TopSoundexCodes(n int) []CodeCount
	TopMetaphoneCodes(n int) []CodeCount
	Exists(name string) bool
}

type CodeCount struct {
	Code  string
	Count int
	Names []string
}

func Encode(name string) PhoneticCode {
	return PhoneticCode{
		Soundex:   Soundex(name),
		Metaphone: Metaphone(name),
	}
}

func filterLetters(name string) string {
	result := make([]rune, 0, len(name))
	for _, r := range name {
		if r >= 'a' && r <= 'z' {
			result = append(result, r)
		} else if r >= 'A' && r <= 'Z' {
			result = append(result, r)
		}
	}
	return string(result)
}

func isVowel(r rune) bool {
	switch r {
	case 'A', 'E', 'I', 'O', 'U':
		return true
	}
	return false
}

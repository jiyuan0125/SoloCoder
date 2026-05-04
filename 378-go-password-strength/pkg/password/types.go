package password

type StrengthLevel string

const (
	LevelWeak     StrengthLevel = "弱"
	LevelMedium   StrengthLevel = "中"
	LevelStrong   StrengthLevel = "强"
	LevelVeryStrong StrengthLevel = "很强"
)

type Result struct {
	Level    StrengthLevel `json:"level"`
	Score    int           `json:"score"`
	Suggestions []string   `json:"suggestions"`
}

type Evaluator struct {
	customWeakPasswords map[string]bool
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		customWeakPasswords: make(map[string]bool),
	}
}

func (e *Evaluator) AddWeakPassword(password string) {
	e.customWeakPasswords[toLower(password)] = true
}

func (e *Evaluator) AddWeakPasswords(passwords []string) {
	for _, p := range passwords {
		e.customWeakPasswords[toLower(p)] = true
	}
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

package password

var keyboardRows = []string{
	"qwertyuiop",
	"asdfghjkl",
	"zxcvbnm",
}

func (e *Evaluator) hasKeyboardSequence(password string) bool {
	lowerPwd := toLower(password)
	
	for _, row := range keyboardRows {
		for i := 0; i <= len(lowerPwd)-4; i++ {
			seq := lowerPwd[i : i+4]
			if isConsecutiveInRow(seq, row) {
				return true
			}
			if isReverseConsecutiveInRow(seq, row) {
				return true
			}
		}
	}
	
	for _, row := range keyboardRows {
		for i := 0; i <= len(lowerPwd)-3; i++ {
			seq := lowerPwd[i : i+3]
			if isConsecutiveInRow(seq, row) {
				return true
			}
			if isReverseConsecutiveInRow(seq, row) {
				return true
			}
		}
	}
	
	return false
}

func isConsecutiveInRow(seq, row string) bool {
	if len(seq) < 2 {
		return false
	}
	
	startIdx := -1
	for i := 0; i < len(row); i++ {
		if row[i] == seq[0] {
			startIdx = i
			break
		}
	}
	
	if startIdx == -1 {
		return false
	}
	
	for j := 1; j < len(seq); j++ {
		expectedIdx := startIdx + j
		if expectedIdx >= len(row) || row[expectedIdx] != seq[j] {
			return false
		}
	}
	
	return true
}

func isReverseConsecutiveInRow(seq, row string) bool {
	if len(seq) < 2 {
		return false
	}
	
	startIdx := -1
	for i := 0; i < len(row); i++ {
		if row[i] == seq[0] {
			startIdx = i
			break
		}
	}
	
	if startIdx == -1 {
		return false
	}
	
	for j := 1; j < len(seq); j++ {
		expectedIdx := startIdx - j
		if expectedIdx < 0 || row[expectedIdx] != seq[j] {
			return false
		}
	}
	
	return true
}

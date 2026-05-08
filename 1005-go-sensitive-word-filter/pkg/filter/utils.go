package filter

import "time"

func isInterference(r rune) bool {
	if r >= 0x200B && r <= 0x200D {
		return true
	}
	if r == 0xFEFF || r == 0x2060 {
		return true
	}
	if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
		return true
	}
	if (r >= 0x20 && r <= 0x2F) || (r >= 0x3A && r <= 0x40) ||
		(r >= 0x5B && r <= 0x60) || (r >= 0x7B && r <= 0x7E) {
		return true
	}
	if r >= 0x3000 && r <= 0x303F {
		return true
	}
	if r >= 0xFF00 && r <= 0xFFEF {
		if (r >= 0xFF10 && r <= 0xFF19) || (r >= 0xFF21 && r <= 0xFF3A) || (r >= 0xFF41 && r <= 0xFF5A) {
			return false
		}
		return true
	}
	if (r >= '0' && r <= '9') || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
		return false
	}
	if r >= 0x4E00 && r <= 0x9FFF {
		return false
	}
	if r >= 0x3040 && r <= 0x30FF {
		return false
	}
	if r >= 0xAC00 && r <= 0xD7AF {
		return false
	}
	return false
}

func isSignificant(r rune) bool {
	return !isInterference(r)
}

func getDateKey(t ...time.Time) string {
	var now time.Time
	if len(t) > 0 {
		now = t[0]
	} else {
		now = time.Now()
	}
	return now.Format("2006-01-02")
}

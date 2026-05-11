package formatter

import (
	"strings"
)

func processEmptyLines(result *string, cfg Config) int {
	if !cfg.RemoveEmptyLines {
		return 0
	}
	
	lines := strings.Split(*result, "\n")
	if len(lines) == 0 {
		return 0
	}
	
	newLines := make([]string, 0, len(lines))
	modified := 0
	
	packageLine := -1
	importStart := -1
	importEnd := -1
	
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "package ") && packageLine < 0 {
			packageLine = i
		}
		if strings.HasPrefix(trimmed, "import (") && importStart < 0 {
			importStart = i
		}
		if importStart >= 0 && importEnd < 0 && strings.TrimSpace(line) == ")" {
			importEnd = i
		}
	}
	
	consecutiveEmpty := 0
	for i := 0; i < len(lines); i++ {
		line := lines[i]
		isBlank := strings.TrimSpace(line) == ""
		
		if isBlank {
			consecutiveEmpty++
		} else {
			consecutiveEmpty = 0
		}
		
		if isBlank {
			if consecutiveEmpty > 1 {
				modified++
				continue
			}
		}
		
		if isBlank {
			if i == packageLine+1 && importStart > packageLine+2 {
				if consecutiveEmpty == 1 {
					newLines = append(newLines, line)
				}
				continue
			}
			
			if importStart >= 0 && i == importStart-1 && i > packageLine {
				if packageLine < 0 || i > packageLine+1 {
					modified++
				}
				newLines = append(newLines, line)
				continue
			}
			
			if importEnd >= 0 && i == importEnd+1 {
				modified++
				newLines = append(newLines, line)
				continue
			}
		}
		
		newLines = append(newLines, line)
	}
	
	*result = strings.Join(newLines, "\n")
	return modified
}

package formatter

import (
	"strings"
)

type importGroup int

const (
	groupStdlib importGroup = iota
	groupThirdParty
	groupInternal
)

type importLine struct {
	group    importGroup
	path     string
	fullLine string
	name     string
}

func classifyImportPath(path string, moduleName string) importGroup {
	cleanPath := strings.Trim(path, `"`)

	if !strings.Contains(cleanPath, ".") {
		return groupStdlib
	}

	if moduleName != "" && strings.HasPrefix(cleanPath, moduleName) {
		return groupInternal
	}

	return groupThirdParty
}

func processImportBlock(source string, cfg Config) (string, int) {
	lines := strings.Split(source, "\n")
	result := make([]string, 0, len(lines))

	inImportBlock := false
	importLines := make([]importLine, 0)
	importIndent := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if !inImportBlock {
			if strings.HasPrefix(trimmed, "import (") {
				inImportBlock = true
				importIndent = line[:strings.Index(line, "import")]
				result = append(result, line)
			} else {
				result = append(result, line)
			}
			continue
		}

		if trimmed == ")" {
			formattedImports := formatImportLines(importLines, importIndent, cfg)
			result = append(result, formattedImports...)
			result = append(result, line)
			inImportBlock = false
			importLines = importLines[:0]
			continue
		}

		if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
			continue
		}

		imp := parseImportLine(line)
		if imp != nil {
			importLines = append(importLines, *imp)
		}
	}

	return strings.Join(result, "\n"), len(importLines)
}

func parseImportLine(line string) *importLine {
	trimmed := strings.TrimSpace(line)

	commentIdx := findCommentStart(trimmed)
	codePart := trimmed
	if commentIdx >= 0 {
		codePart = strings.TrimSpace(trimmed[:commentIdx])
	}

	fields := strings.Fields(codePart)
	if len(fields) == 0 {
		return nil
	}

	name := ""
	path := ""

	if len(fields) == 1 {
		path = fields[0]
	} else if len(fields) >= 2 {
		name = fields[0]
		path = fields[1]
	}

	if path == "" {
		return nil
	}

	return &importLine{
		path:     path,
		fullLine: line,
		name:     name,
	}
}

func findCommentStart(s string) int {
	inString := false
	escaped := false

	for i := 0; i < len(s); i++ {
		ch := s[i]

		if inString {
			if escaped {
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			continue
		}

		if ch == '/' && i+1 < len(s) {
			if s[i+1] == '/' || s[i+1] == '*' {
				return i
			}
		}
	}

	return -1
}

func formatImportLines(imports []importLine, indent string, cfg Config) []string {
	if len(imports) == 0 {
		return nil
	}

	for i := range imports {
		imports[i].group = classifyImportPath(imports[i].path, cfg.ModuleName)
	}

	groups := make(map[importGroup][]importLine)
	for _, imp := range imports {
		groups[imp.group] = append(groups[imp.group], imp)
	}

	groupOrder := []importGroup{groupStdlib, groupThirdParty, groupInternal}

	result := make([]string, 0)
	firstGroup := true

	for _, group := range groupOrder {
		groupImports, exists := groups[group]
		if !exists || len(groupImports) == 0 {
			continue
		}

		sortImportsByPath(groupImports)

		if !firstGroup {
			result = append(result, "")
		}
		firstGroup = false

		for _, imp := range groupImports {
			result = append(result, formatSingleImport(imp, indent))
		}
	}

	return result
}

func sortImportsByPath(imports []importLine) {
	for i := 0; i < len(imports)-1; i++ {
		for j := 0; j < len(imports)-1-i; j++ {
			keyJ := makeSortKey(imports[j])
			keyJ1 := makeSortKey(imports[j+1])
			if keyJ > keyJ1 {
				imports[j], imports[j+1] = imports[j+1], imports[j]
			}
		}
	}
}

func makeSortKey(imp importLine) string {
	if imp.name != "" {
		return imp.name + " " + imp.path
	}
	return imp.path
}

func formatSingleImport(imp importLine, indent string) string {
	if imp.name != "" {
		return indent + "\t" + imp.name + " " + imp.path
	}
	return indent + "\t" + imp.path
}

package output

import (
	"config-diff/internal/diff"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

const (
	colorGreen  = "\033[32m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorReset  = "\033[0m"
)

func FormatTerminal(changes []diff.Change) string {
	if len(changes) == 0 {
		return "没有检测到差异"
	}

	var result string
	for _, c := range changes {
		switch c.Action {
		case diff.ActionAdd:
			result += formatAdd(c)
		case diff.ActionRemove:
			result += formatRemove(c)
		case diff.ActionModify:
			result += formatModify(c)
		}
	}
	return result
}

func formatAdd(c diff.Change) string {
	if c.Index != nil && !strings.Contains(c.Path, "[") {
		return fmt.Sprintf("%s[+] %s[%d] = %v%s\n",
			colorGreen, c.Path, *c.Index, formatValue(c.NewValue), colorReset)
	}
	return fmt.Sprintf("%s[+] %s = %v%s\n",
		colorGreen, c.Path, formatValue(c.NewValue), colorReset)
}

func formatRemove(c diff.Change) string {
	if c.Index != nil && !strings.Contains(c.Path, "[") {
		return fmt.Sprintf("%s[-] %s[%d] = %v%s\n",
			colorRed, c.Path, *c.Index, formatValue(c.OldValue), colorReset)
	}
	return fmt.Sprintf("%s[-] %s = %v%s\n",
		colorRed, c.Path, formatValue(c.OldValue), colorReset)
}

func formatModify(c diff.Change) string {
	if c.Index != nil && !strings.Contains(c.Path, "[") {
		return fmt.Sprintf("%s[~] %s[%d]: %v -> %v%s\n",
			colorYellow, c.Path, *c.Index, formatValue(c.OldValue), formatValue(c.NewValue), colorReset)
	}
	return fmt.Sprintf("%s[~] %s: %v -> %v%s\n",
		colorYellow, c.Path, formatValue(c.OldValue), formatValue(c.NewValue), colorReset)
}

func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return strconv.Quote(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func FormatJSON(changes []diff.Change) (string, error) {
	if len(changes) == 0 {
		return "[]", nil
	}

	data, err := json.MarshalIndent(changes, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

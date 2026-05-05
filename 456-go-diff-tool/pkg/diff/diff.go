package diff

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type DiffType string

const (
	DiffEqual   DiffType = "equal"
	DiffAdd     DiffType = "add"
	DiffDelete  DiffType = "delete"
	DiffChange  DiffType = "change"
	DiffMove    DiffType = "move"
)

type LineDiff struct {
	OldLine   int
	NewLine   int
	Type      DiffType
	Content   string
	WordDiffs []WordDiff
}

type WordDiff struct {
	Type    DiffType
	Content string
}

type BlockMove struct {
	OldStart   int
	OldEnd     int
	NewStart   int
	NewEnd     int
	Content    string
}

type DiffStats struct {
	Added   int
	Deleted int
	Changed int
	Moved   int
}

type DiffResult struct {
	LineDiffs    []LineDiff
	Stats        DiffStats
	MovedBlocks  []BlockMove
}

var alphanumericRegex = regexp.MustCompile(`[a-zA-Z0-9]+`)

func NormalizeContent(content string) string {
	matches := alphanumericRegex.FindAllString(content, -1)
	return strings.Join(matches, "")
}

func CompareLines(a, b string) bool {
	return a == b
}

func CompareLinesNormalized(a, b string) bool {
	return NormalizeContent(a) == NormalizeContent(b)
}

func getComparisonTokens(line string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if !unicode.IsSpace(r) {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(r))
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func getComparisonTokensString(line string) string {
	tokens := getComparisonTokens(line)
	return strings.Join(tokens, "|")
}

func areLinesWhitespaceOnlyChange(oldLine, newLine string) bool {
	if oldLine == newLine {
		return true
	}

	oldNorm := NormalizeContent(oldLine)
	newNorm := NormalizeContent(newLine)

	if oldNorm == newNorm && oldNorm != "" {
		oldTokens := getComparisonTokensString(oldLine)
		newTokens := getComparisonTokensString(newLine)
		return oldTokens == newTokens
	}

	return false
}

func areLinesPotentialModification(oldLine, newLine string) bool {
	if oldLine == newLine {
		return true
	}

	oldCompare := getComparisonTokens(oldLine)
	newCompare := getComparisonTokens(newLine)

	if len(oldCompare) != len(newCompare) {
		return false
	}

	diffCount := 0
	for i := range oldCompare {
		if oldCompare[i] != newCompare[i] {
			diffCount++
		}
	}

	if diffCount == 0 {
		return true
	}

	if diffCount <= 2 && len(oldCompare) >= 3 {
		return true
	}

	return false
}

func LCSInts(a, b []int) []int {
	if len(a) == 0 || len(b) == 0 {
		return nil
	}

	lenA := len(a)
	lenB := len(b)

	dp := make([][]int, lenA+1)
	for i := range dp {
		dp[i] = make([]int, lenB+1)
	}

	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				if dp[i-1][j] > dp[i][j-1] {
					dp[i][j] = dp[i-1][j]
				} else {
					dp[i][j] = dp[i][j-1]
				}
			}
		}
	}

	var lcs []int
	i, j := lenA, lenB
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			lcs = append(lcs, a[i-1])
			i--
			j--
		} else {
			if dp[i-1][j] > dp[i][j-1] {
				i--
			} else {
				j--
			}
		}
	}

	for i, j := 0, len(lcs)-1; i < j; i, j = i+1, j-1 {
		lcs[i], lcs[j] = lcs[j], lcs[i]
	}

	return lcs
}

func basicLineDiff(oldLines, newLines []string) []LineDiff {
	lineMap := make(map[string]int)
	var uniqueID int
	var oldInts, newInts []int

	for _, line := range oldLines {
		normKey := getComparisonTokensString(line)
		if id, exists := lineMap[normKey]; exists {
			oldInts = append(oldInts, id)
		} else {
			lineMap[normKey] = uniqueID
			oldInts = append(oldInts, uniqueID)
			uniqueID++
		}
	}

	for _, line := range newLines {
		normKey := getComparisonTokensString(line)
		if id, exists := lineMap[normKey]; exists {
			newInts = append(newInts, id)
		} else {
			lineMap[normKey] = uniqueID
			newInts = append(newInts, uniqueID)
			uniqueID++
		}
	}

	lcs := LCSInts(oldInts, newInts)

	var lineDiffs []LineDiff
	oldPtr, newPtr, lcsPtr := 0, 0, 0

	for oldPtr < len(oldLines) || newPtr < len(newLines) {
		if lcsPtr < len(lcs) {
			currentLCS := lcs[lcsPtr]

			var addBuffer, deleteBuffer []string
			var addStartIdx, delStartIdx int

			for oldPtr < len(oldLines) && newPtr < len(newLines) {
				oldVal := oldInts[oldPtr]
				newVal := newInts[newPtr]

				if oldVal == currentLCS && newVal == currentLCS {
					break
				}

				if oldVal != currentLCS {
					if len(deleteBuffer) == 0 {
						delStartIdx = oldPtr
					}
					deleteBuffer = append(deleteBuffer, oldLines[oldPtr])
					oldPtr++
				}

				if newVal != currentLCS {
					if len(addBuffer) == 0 {
						addStartIdx = newPtr
					}
					addBuffer = append(addBuffer, newLines[newPtr])
					newPtr++
				}

				if oldPtr >= len(oldLines) || newPtr >= len(newLines) {
					break
				}
			}

			for i, line := range deleteBuffer {
				lineDiffs = append(lineDiffs, LineDiff{
					OldLine: delStartIdx + i,
					NewLine: -1,
					Type:    DiffDelete,
					Content: line,
				})
			}

			for i, line := range addBuffer {
				lineDiffs = append(lineDiffs, LineDiff{
					OldLine: -1,
					NewLine: addStartIdx + i,
					Type:    DiffAdd,
					Content: line,
				})
			}

			if oldPtr < len(oldLines) && newPtr < len(newLines) {
				lineDiffs = append(lineDiffs, LineDiff{
					OldLine: oldPtr,
					NewLine: newPtr,
					Type:    DiffEqual,
					Content: oldLines[oldPtr],
				})
				oldPtr++
				newPtr++
				lcsPtr++
			}
		} else {
			for oldPtr < len(oldLines) {
				lineDiffs = append(lineDiffs, LineDiff{
					OldLine: oldPtr,
					NewLine: -1,
					Type:    DiffDelete,
					Content: oldLines[oldPtr],
				})
				oldPtr++
			}

			for newPtr < len(newLines) {
				lineDiffs = append(lineDiffs, LineDiff{
					OldLine: -1,
					NewLine: newPtr,
					Type:    DiffAdd,
					Content: newLines[newPtr],
				})
				newPtr++
			}
		}
	}

	return lineDiffs
}

func tokenizeWords(line string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range line {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else {
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
			tokens = append(tokens, string(r))
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

func ComputeWordDiff(oldLine, newLine string) []WordDiff {
	oldTokens := tokenizeWords(oldLine)
	newTokens := tokenizeWords(newLine)

	oldCompare := getComparisonTokens(oldLine)
	newCompare := getComparisonTokens(newLine)

	if len(oldCompare) == 0 && len(newCompare) == 0 {
		return []WordDiff{{Type: DiffEqual, Content: newLine}}
	}

	if len(oldCompare) == 0 {
		return []WordDiff{{Type: DiffAdd, Content: newLine}}
	}

	if len(newCompare) == 0 {
		return []WordDiff{{Type: DiffDelete, Content: oldLine}}
	}

	allEqual := true
	for i := range oldCompare {
		if i >= len(newCompare) || oldCompare[i] != newCompare[i] {
			allEqual = false
			break
		}
	}

	if allEqual && len(oldCompare) == len(newCompare) {
		return []WordDiff{{Type: DiffEqual, Content: newLine}}
	}

	if len(oldTokens) == len(newTokens) {
		var wordDiffs []WordDiff
		for i, oldTok := range oldTokens {
			newTok := newTokens[i]
			if oldTok == newTok {
				wordDiffs = append(wordDiffs, WordDiff{Type: DiffEqual, Content: oldTok})
			} else {
				wordDiffs = append(wordDiffs, WordDiff{Type: DiffDelete, Content: oldTok})
				wordDiffs = append(wordDiffs, WordDiff{Type: DiffAdd, Content: newTok})
			}
		}
		return wordDiffs
	}

	return []WordDiff{
		{Type: DiffDelete, Content: oldLine},
		{Type: DiffAdd, Content: newLine},
	}
}

func DetectMovedBlocks(oldLines, newLines []string, threshold int) []BlockMove {
	oldLineMap := make(map[string][]int)
	for i, line := range oldLines {
		oldLineMap[line] = append(oldLineMap[line], i)
	}

	newLineMap := make(map[string][]int)
	for i, line := range newLines {
		newLineMap[line] = append(newLineMap[line], i)
	}

	commonLines := make(map[string]bool)
	for line := range oldLineMap {
		if _, exists := newLineMap[line]; exists {
			commonLines[line] = true
		}
	}

	var movedBlocks []BlockMove
	usedOld := make(map[int]bool)
	usedNew := make(map[int]bool)

	for line := range commonLines {
		oldIdxs := oldLineMap[line]
		newIdxs := newLineMap[line]

		for _, oi := range oldIdxs {
			for _, ni := range newIdxs {
				if usedOld[oi] || usedNew[ni] {
					continue
				}

				blockContent := strings.Builder{}
				startOld := oi
				startNew := ni
				endOld := oi
				endNew := ni

				for endOld < len(oldLines) && endNew < len(newLines) && oldLines[endOld] == newLines[endNew] {
					if !usedOld[endOld] && !usedNew[endNew] {
						blockContent.WriteString(oldLines[endOld])
						endOld++
						endNew++
					} else {
						break
					}
				}

				endOld--
				endNew--

				if blockContent.Len() >= threshold && startOld != startNew {
					block := BlockMove{
						OldStart: startOld,
						OldEnd:   endOld,
						NewStart: startNew,
						NewEnd:   endNew,
						Content:  oldLines[startOld : endOld+1][0],
					}

					if len(oldLines[startOld:endOld+1]) > 1 {
						var sb strings.Builder
						for _, l := range oldLines[startOld : endOld+1] {
							sb.WriteString(l)
						}
						block.Content = sb.String()
					}

					movedBlocks = append(movedBlocks, block)

					for i := startOld; i <= endOld; i++ {
						usedOld[i] = true
					}
					for i := startNew; i <= endNew; i++ {
						usedNew[i] = true
					}
				}
			}
		}
	}

	return movedBlocks
}

func processDiffs(lineDiffs []LineDiff, oldLines, newLines []string) []LineDiff {
	var processedDiffs []LineDiff

	for i := 0; i < len(lineDiffs); i++ {
		d := lineDiffs[i]

		if d.Type == DiffDelete && i+1 < len(lineDiffs) {
			nextD := lineDiffs[i+1]
			if nextD.Type == DiffAdd {
				oldLine := oldLines[d.OldLine]
				newLine := newLines[nextD.NewLine]

				if oldLine == newLine {
					processedDiffs = append(processedDiffs, LineDiff{
						OldLine: d.OldLine,
						NewLine: nextD.NewLine,
						Type:    DiffEqual,
						Content: oldLine,
					})
					i++
					continue
				}

				if areLinesWhitespaceOnlyChange(oldLine, newLine) {
					processedDiffs = append(processedDiffs, LineDiff{
						OldLine: d.OldLine,
						NewLine: nextD.NewLine,
						Type:    DiffEqual,
						Content: newLine,
					})
					i++
					continue
				}

				if areLinesPotentialModification(oldLine, newLine) {
					processedDiffs = append(processedDiffs, LineDiff{
						OldLine:   d.OldLine,
						NewLine:   nextD.NewLine,
						Type:      DiffChange,
						Content:   newLine,
						WordDiffs: ComputeWordDiff(oldLine, newLine),
					})
					i++
					continue
				}
			}
		}

		if d.Type == DiffEqual && d.OldLine >= 0 && d.NewLine >= 0 {
			oldLine := oldLines[d.OldLine]
			newLine := newLines[d.NewLine]

			if oldLine != newLine {
				if areLinesWhitespaceOnlyChange(oldLine, newLine) {
					d.Content = newLine
				} else {
					d.Type = DiffChange
					d.WordDiffs = ComputeWordDiff(oldLine, newLine)
				}
			}
		}

		processedDiffs = append(processedDiffs, d)
	}

	return processedDiffs
}

func GenerateUnifiedDiff(oldLines, newLines []string, result DiffResult) string {
	var sb strings.Builder

	if len(result.LineDiffs) == 0 {
		return ""
	}

	sb.WriteString("--- old\n")
	sb.WriteString("+++ new\n")

	hunkContext := 3

	for i := 0; i < len(result.LineDiffs); i++ {
		d := result.LineDiffs[i]

		if d.Type != DiffEqual {
			start := max(0, i-hunkContext)
			end := min(len(result.LineDiffs), i+hunkContext+1)

			var oldStart, oldCount, newStart, newCount int
			foundOld := false
			foundNew := false

			for j := start; j < end; j++ {
				hd := result.LineDiffs[j]
				if hd.OldLine >= 0 && !foundOld {
					oldStart = hd.OldLine
					foundOld = true
				}
				if hd.NewLine >= 0 && !foundNew {
					newStart = hd.NewLine
					foundNew = true
				}
			}

			oldCount = 0
			newCount = 0
			for j := start; j < end; j++ {
				hd := result.LineDiffs[j]
				if hd.Type != DiffMove {
					if hd.OldLine >= 0 {
						oldCount++
					}
					if hd.NewLine >= 0 {
						newCount++
					}
				}
			}

			sb.WriteString("@@ -")
			if foundOld {
				sb.WriteString(strconv.Itoa(oldStart + 1))
				if oldCount > 1 {
					sb.WriteString(",")
					sb.WriteString(strconv.Itoa(oldCount))
				}
			} else {
				sb.WriteString("0,0")
			}
			sb.WriteString(" +")
			if foundNew {
				sb.WriteString(strconv.Itoa(newStart + 1))
				if newCount > 1 {
					sb.WriteString(",")
					sb.WriteString(strconv.Itoa(newCount))
				}
			} else {
				sb.WriteString("0,0")
			}
			sb.WriteString(" @@\n")

			for j := start; j < end; j++ {
				hd := result.LineDiffs[j]
				switch hd.Type {
				case DiffEqual:
					sb.WriteString(" ")
					sb.WriteString(hd.Content)
					sb.WriteString("\n")
				case DiffAdd:
					sb.WriteString("+")
					sb.WriteString(hd.Content)
					sb.WriteString("\n")
				case DiffDelete:
					sb.WriteString("-")
					sb.WriteString(hd.Content)
					sb.WriteString("\n")
				case DiffMove:
					sb.WriteString("M")
					sb.WriteString(hd.Content)
					sb.WriteString("\n")
				case DiffChange:
					sb.WriteString("~")
					sb.WriteString(hd.Content)
					sb.WriteString("\n")
				}
			}

			i = end - 1
		}
	}

	return sb.String()
}

func Compare(oldText, newText string, moveThreshold int) DiffResult {
	if oldText == newText {
		return DiffResult{
			LineDiffs:   []LineDiff{},
			Stats:       DiffStats{},
			MovedBlocks: []BlockMove{},
		}
	}

	oldLines := strings.Split(oldText, "\n")
	newLines := strings.Split(newText, "\n")

	if len(oldLines) == 0 {
		oldLines = []string{""}
	}
	if len(newLines) == 0 {
		newLines = []string{""}
	}

	movedBlocks := DetectMovedBlocks(oldLines, newLines, moveThreshold)

	basicDiffs := basicLineDiff(oldLines, newLines)

	processedDiffs := processDiffs(basicDiffs, oldLines, newLines)

	stats := DiffStats{}
	for _, d := range processedDiffs {
		switch d.Type {
		case DiffAdd:
			stats.Added++
		case DiffDelete:
			stats.Deleted++
		case DiffChange:
			stats.Changed++
		case DiffMove:
			stats.Moved++
		}
	}

	return DiffResult{
		LineDiffs:   processedDiffs,
		Stats:       stats,
		MovedBlocks: movedBlocks,
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

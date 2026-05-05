package main

import (
	"kb/common"
	"strings"
)

func CompareVersions(v1, v2 *common.ArticleVersion) *common.VersionCompareResult {
	if v1 == nil || v2 == nil {
		return nil
	}

	diff := buildDiff(v1.Content, v2.Content)

	return &common.VersionCompareResult{
		ArticleID:   v1.ArticleID,
		Version1:    v1.VersionNum,
		Version2:    v2.VersionNum,
		DiffContent: diff,
		CreatedAt1:  v1.CreatedAt,
		CreatedAt2:  v2.CreatedAt,
	}
}

func buildDiff(oldStr, newStr string) string {
	oldLines := strings.Split(oldStr, "\n")
	newLines := strings.Split(newStr, "\n")

	oldMap := make(map[string][]int)
	for i, line := range oldLines {
		oldMap[line] = append(oldMap[line], i)
	}

	var result strings.Builder
	result.WriteString("=== 版本差异对比 ===\n")

	i, j := 0, 0
	for i < len(oldLines) && j < len(newLines) {
		if oldLines[i] == newLines[j] {
			result.WriteString("  " + oldLines[i] + "\n")
			i++
			j++
		} else {
			found := false
			for k := j + 1; k < len(newLines) && k < j+5; k++ {
				if oldLines[i] == newLines[k] {
					for m := j; m < k; m++ {
						result.WriteString("+ " + newLines[m] + "\n")
					}
					j = k
					found = true
					break
				}
			}
			if !found {
				for k := i + 1; k < len(oldLines) && k < i+5; k++ {
					if newLines[j] == oldLines[k] {
						for m := i; m < k; m++ {
							result.WriteString("- " + oldLines[m] + "\n")
						}
						i = k
						found = true
						break
					}
				}
			}
			if !found {
				result.WriteString("- " + oldLines[i] + "\n")
				result.WriteString("+ " + newLines[j] + "\n")
				i++
				j++
			}
		}
	}

	for i < len(oldLines) {
		result.WriteString("- " + oldLines[i] + "\n")
		i++
	}
	for j < len(newLines) {
		result.WriteString("+ " + newLines[j] + "\n")
		j++
	}

	return result.String()
}

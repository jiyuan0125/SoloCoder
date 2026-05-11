package deadcode

import (
	"strings"

	"deadcode-detector/internal/common"
)

type Analyzer struct {
	files          []*ParsedFile
	packageName    string
	declarations   []common.Declaration
	referenceMap   *ReferenceMap
	specialHandler *SpecialCaseHandler
}

type AnalysisResult struct {
	Report       common.AnalysisReport
	ReferenceMap *ReferenceMap
}

func NewAnalyzer(filesContent map[string]string, packageName string) (*Analyzer, error) {
	files := make([]*ParsedFile, 0, len(filesContent))
	for filename, content := range filesContent {
		parsed, err := ParseFile(filename, content)
		if err != nil {
			return nil, err
		}
		files = append(files, parsed)
	}

	if packageName == "" && len(files) > 0 {
		packageName = files[0].Package
	}

	declarations := make([]common.Declaration, 0)
	for _, file := range files {
		declarations = append(declarations, ExtractDeclarations(file)...)
	}

	referenceMap := TrackReferences(files, declarations)
	specialHandler := NewSpecialCaseHandler(files)

	return &Analyzer{
		files:          files,
		packageName:    packageName,
		declarations:   declarations,
		referenceMap:   referenceMap,
		specialHandler: specialHandler,
	}, nil
}

func (a *Analyzer) Analyze() *AnalysisResult {
	entries := make([]common.DeadCodeEntry, 0)

	for _, decl := range a.declarations {
		entry := a.checkDeclaration(decl)
		if entry != nil {
			entries = append(entries, *entry)
		}
	}

	summary := a.buildSummary(entries)

	report := common.AnalysisReport{
		PackageName: a.packageName,
		Entries:     entries,
		Summary:     summary,
	}

	return &AnalysisResult{
		Report:       report,
		ReferenceMap: a.referenceMap,
	}
}

func (a *Analyzer) checkDeclaration(decl common.Declaration) *common.DeadCodeEntry {
	if IsBlankIdentifier(decl.Name) {
		return nil
	}

	if a.specialHandler.IsExcluded(decl, a.files) {
		return nil
	}

	switch decl.Type {
	case common.TypeFunction:
		return a.checkFunction(decl)
	case common.TypeType:
		return a.checkType(decl)
	case common.TypeVariable:
		return a.checkVariable(decl)
	case common.TypeConstant:
		return a.checkConstant(decl)
	case common.TypeField:
		return a.checkField(decl)
	}

	return nil
}

func (a *Analyzer) checkFunction(decl common.Declaration) *common.DeadCodeEntry {
	referenced := IsReferenced(decl.Name, a.referenceMap)

	if !referenced {
		if decl.IsExported {
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      "导出的函数在当前包内没有被引用（可能被外部包使用）",
				Confidence:  common.ConfidencePossible,
			}
		}
		return &common.DeadCodeEntry{
			Declaration: decl,
			Reason:      "函数未被引用",
			Confidence:  common.ConfidenceCertain,
		}
	}

	return nil
}

func (a *Analyzer) checkType(decl common.Declaration) *common.DeadCodeEntry {
	referenced := IsReferenced(decl.Name, a.referenceMap)

	if !referenced {
		if decl.IsAlias {
			reason := "类型别名未被引用"
			if decl.IsExported {
				return &common.DeadCodeEntry{
					Declaration: decl,
					Reason:      "导出的" + reason + "（可能被外部包使用）",
					Confidence:  common.ConfidencePossible,
				}
			}
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      reason,
				Confidence:  common.ConfidenceCertain,
			}
		}

		if decl.IsInterface {
			reason := "接口类型未被引用"
			if decl.IsExported {
				return &common.DeadCodeEntry{
					Declaration: decl,
					Reason:      "导出的" + reason + "（可能被外部包使用）",
					Confidence:  common.ConfidencePossible,
				}
			}
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      reason,
				Confidence:  common.ConfidenceCertain,
			}
		}

		reason := "类型未被引用"
		if decl.IsExported {
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      "导出的" + reason + "（可能被外部包使用）",
				Confidence:  common.ConfidencePossible,
			}
		}
		return &common.DeadCodeEntry{
			Declaration: decl,
			Reason:      reason,
			Confidence:  common.ConfidenceCertain,
		}
	}

	return nil
}

func (a *Analyzer) checkVariable(decl common.Declaration) *common.DeadCodeEntry {
	referenced := IsReferenced(decl.Name, a.referenceMap)
	assigned := IsVariableAssigned(decl.Name, a.referenceMap)
	read := IsVariableRead(decl.Name, a.referenceMap)

	if !referenced {
		if decl.IsExported {
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      "导出的变量未被引用（可能被外部包使用）",
				Confidence:  common.ConfidencePossible,
			}
		}
		return &common.DeadCodeEntry{
			Declaration: decl,
			Reason:      "变量未被引用",
			Confidence:  common.ConfidenceCertain,
		}
	}

	if assigned && !read {
		if decl.IsExported {
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      "导出的变量已声明未读取（可能被外部包使用）",
				Confidence:  common.ConfidencePossible,
			}
		}
		return &common.DeadCodeEntry{
			Declaration: decl,
			Reason:      "变量已声明未读取",
			Confidence:  common.ConfidenceCertain,
		}
	}

	return nil
}

func (a *Analyzer) checkConstant(decl common.Declaration) *common.DeadCodeEntry {
	referenced := IsReferenced(decl.Name, a.referenceMap)

	if !referenced {
		if decl.IsExported {
			return &common.DeadCodeEntry{
				Declaration: decl,
				Reason:      "导出的常量未被引用（可能被外部包使用）",
				Confidence:  common.ConfidencePossible,
			}
		}
		return &common.DeadCodeEntry{
			Declaration: decl,
			Reason:      "常量未被引用",
			Confidence:  common.ConfidenceCertain,
		}
	}

	return nil
}

func (a *Analyzer) checkField(decl common.Declaration) *common.DeadCodeEntry {
	referenced := IsFieldReferenced(decl.Location.File, decl.Name, a.referenceMap)

	if !referenced {
		return &common.DeadCodeEntry{
			Declaration: decl,
			Reason:      "结构体字段未被引用",
			Confidence:  common.ConfidenceCertain,
		}
	}

	return nil
}

func (a *Analyzer) buildSummary(entries []common.DeadCodeEntry) common.Summary {
	summary := common.Summary{}

	for _, entry := range entries {
		switch entry.Declaration.Type {
		case common.TypeFunction:
			summary.UnusedFunctions++
		case common.TypeType:
			summary.UnusedTypes++
		case common.TypeVariable:
			summary.UnusedVariables++
		case common.TypeConstant:
			summary.UnusedConstants++
		case common.TypeField:
			summary.UnusedFields++
		}
	}

	return summary
}

func (a *Analyzer) GetReferenceMap() *ReferenceMap {
	return a.referenceMap
}

func (a *Analyzer) GetDeclarations() []common.Declaration {
	return a.declarations
}

func FilterReportByType(report common.AnalysisReport, filter common.DeclarationType) common.AnalysisReport {
	if filter == "" {
		return report
	}

	filtered := common.AnalysisReport{
		PackageName: report.PackageName,
		Entries:     make([]common.DeadCodeEntry, 0),
	}

	for _, entry := range report.Entries {
		if entry.Declaration.Type == filter {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}

	filtered.Summary = (&Analyzer{}).buildSummary(filtered.Entries)

	return filtered
}

func ApplyExclusions(report common.AnalysisReport, patterns []string) common.AnalysisReport {
	if len(patterns) == 0 {
		return report
	}

	filtered := common.AnalysisReport{
		PackageName: report.PackageName,
		Entries:     make([]common.DeadCodeEntry, 0),
	}

	for _, entry := range report.Entries {
		excluded := false
		for _, pattern := range patterns {
			if strings.Contains(entry.Declaration.Name, pattern) {
				excluded = true
				break
			}
			if strings.Contains(entry.Declaration.Location.File, pattern) {
				excluded = true
				break
			}
		}
		if !excluded {
			filtered.Entries = append(filtered.Entries, entry)
		}
	}

	filtered.Summary = (&Analyzer{}).buildSummary(filtered.Entries)

	return filtered
}

package coverage

import (
	"math"
	"sort"
	"strings"

	"github.com/solocoder/coverage-analyzer/pkg/api"
)

type AnalysisResult struct {
	Profile            *CoverProfile
	FileIntervals      map[string][]LineInterval
	FileCoveredLines   map[string]map[int]int
	FileTotalLines     map[int]bool
	HasSourceAnalysis  bool
}

func NewAnalysisResult(profile *CoverProfile) *AnalysisResult {
	fileIntervals := MergeIntervalsByFile(profile.Blocks)
	fileCoveredLines := make(map[string]map[int]int)

	for file, intervals := range fileIntervals {
		fileCoveredLines[file] = GetCoveredLineNumbers(intervals)
	}

	return &AnalysisResult{
		Profile:          profile,
		FileIntervals:    fileIntervals,
		FileCoveredLines: fileCoveredLines,
		HasSourceAnalysis: false,
	}
}

func (a *AnalysisResult) CalculateFunctionCoverage() *api.FunctionCoverageReport {
	report := &api.FunctionCoverageReport{}
	funcData := make(map[string]*api.FunctionCoverage)

	fileFuncs := a.extractFunctions()

	for file, funcs := range fileFuncs {
		coveredLines := a.FileCoveredLines[file]
		pkg := extractPackageName(file)

		for _, fn := range funcs {
			key := pkg + "." + fn.Name
			if _, exists := funcData[key]; !exists {
				funcData[key] = &api.FunctionCoverage{
					Package:      pkg,
					Function:     fn.Name,
					CoveredLines: 0,
					TotalLines:   0,
					Percentage:   0,
				}
			}

			for line := fn.StartLine; line <= fn.EndLine; line++ {
				funcData[key].TotalLines++
				if count, covered := coveredLines[line]; covered && count > 0 {
					funcData[key].CoveredLines++
				}
			}
		}
	}

	for file, coveredLines := range a.FileCoveredLines {
		pkg := extractPackageName(file)
		intervals := a.FileIntervals[file]

		for line, count := range coveredLines {
			found := false
			for _, fn := range fileFuncs[file] {
				if line >= fn.StartLine && line <= fn.EndLine {
					found = true
					break
				}
			}

			if !found {
				key := pkg + ".(global)"
				if _, exists := funcData[key]; !exists {
					funcData[key] = &api.FunctionCoverage{
						Package:      pkg,
						Function:     "(global)",
						CoveredLines: 0,
						TotalLines:   0,
						Percentage:   0,
					}
				}

				isInAny := false
				for _, interval := range intervals {
					if line >= interval.Start && line <= interval.End {
						isInAny = true
						break
					}
				}

				if isInAny {
					funcData[key].TotalLines++
					if count > 0 {
						funcData[key].CoveredLines++
					}
				}
			}
		}
	}

	for _, fd := range funcData {
		if fd.TotalLines > 0 {
			fd.Percentage = float64(fd.CoveredLines) / float64(fd.TotalLines) * 100
		}
		report.Functions = append(report.Functions, *fd)
	}

	sort.Slice(report.Functions, func(i, j int) bool {
		if report.Functions[i].Package != report.Functions[j].Package {
			return report.Functions[i].Package < report.Functions[j].Package
		}
		return report.Functions[i].Function < report.Functions[j].Function
	})

	return report
}

func (a *AnalysisResult) CalculatePackageCoverage() *api.PackageCoverageReport {
	report := &api.PackageCoverageReport{}
	pkgData := make(map[string]*api.PackageCoverage)

	funcReport := a.CalculateFunctionCoverage()

	for _, fn := range funcReport.Functions {
		if _, exists := pkgData[fn.Package]; !exists {
			pkgData[fn.Package] = &api.PackageCoverage{
				Package:      fn.Package,
				CoveredLines: 0,
				TotalLines:   0,
				Percentage:   0,
			}
		}
		pkgData[fn.Package].CoveredLines += fn.CoveredLines
		pkgData[fn.Package].TotalLines += fn.TotalLines
	}

	for _, pd := range pkgData {
		if pd.TotalLines > 0 {
			pd.Percentage = float64(pd.CoveredLines) / float64(pd.TotalLines) * 100
		}
		report.Packages = append(report.Packages, *pd)
	}

	sort.Slice(report.Packages, func(i, j int) bool {
		return report.Packages[i].Package < report.Packages[j].Package
	})

	return report
}

func (a *AnalysisResult) CalculateLineCoverage() *api.LineCoverageReport {
	report := &api.LineCoverageReport{}

	files := make([]string, 0, len(a.FileCoveredLines))
	for file := range a.FileCoveredLines {
		files = append(files, file)
	}
	sort.Strings(files)

	for _, file := range files {
		coveredLines := a.FileCoveredLines[file]
		intervals := a.FileIntervals[file]

		allLines := make(map[int]struct{})
		for _, interval := range intervals {
			for line := interval.Start; line <= interval.End; line++ {
				allLines[line] = struct{}{}
			}
		}

		lineNums := make([]int, 0, len(allLines))
		for line := range allLines {
			lineNums = append(lineNums, line)
		}
		sort.Ints(lineNums)

		fileCoverage := api.FileLineCoverage{
			File:  file,
			Lines: make([]api.LineCoverage, 0, len(lineNums)),
		}

		for _, lineNum := range lineNums {
			count, hasCount := coveredLines[lineNum]
			status := "uncovered"
			execCount := 0

			if hasCount {
				execCount = count
				if count > 0 {
					if a.Profile.Mode == api.ModeCount || a.Profile.Mode == api.ModeAtomic {
						if count == 1 {
							status = "partially-covered"
						} else {
							status = "covered"
						}
					} else {
						status = "covered"
					}
				}
			}

			fileCoverage.Lines = append(fileCoverage.Lines, api.LineCoverage{
				LineNumber:     lineNum,
				Status:         status,
				ExecutionCount: execCount,
			})
		}

		report.Files = append(report.Files, fileCoverage)
	}

	return report
}

func (a *AnalysisResult) CalculateSummary() *api.SummaryResponse {
	summary := &api.SummaryResponse{}

	pkgReport := a.CalculatePackageCoverage()

	for _, pkg := range pkgReport.Packages {
		summary.TotalLines += pkg.TotalLines
		summary.CoveredLines += pkg.CoveredLines
	}

	summary.UncoveredLines = summary.TotalLines - summary.CoveredLines

	if summary.TotalLines > 0 {
		summary.Percentage = float64(summary.CoveredLines) / float64(summary.TotalLines) * 100
	}

	if a.Profile.Mode == api.ModeCount || a.Profile.Mode == api.ModeAtomic {
		summary.ConfidenceLower, summary.ConfidenceUpper = calculateConfidenceInterval(a)
	}

	sortedPackages := make([]api.PackageCoverage, len(pkgReport.Packages))
	copy(sortedPackages, pkgReport.Packages)

	sort.Slice(sortedPackages, func(i, j int) bool {
		return sortedPackages[i].Percentage < sortedPackages[j].Percentage
	})

	lowestCount := 5
	if len(sortedPackages) < lowestCount {
		lowestCount = len(sortedPackages)
	}
	summary.LowestPackages = sortedPackages[:lowestCount]

	highestCount := 5
	if len(sortedPackages) < highestCount {
		highestCount = len(sortedPackages)
	}
	summary.HighestPackages = make([]api.PackageCoverage, highestCount)
	for i := 0; i < highestCount; i++ {
		summary.HighestPackages[i] = sortedPackages[len(sortedPackages)-1-i]
	}

	return summary
}

func (a *AnalysisResult) extractFunctions() map[string][]FunctionInfo {
	result := make(map[string][]FunctionInfo)

	for file := range a.FileCoveredLines {
		funcs := a.extractFunctionsFromFile(file)
		if len(funcs) > 0 {
			result[file] = funcs
		}
	}

	return result
}

func (a *AnalysisResult) extractFunctionsFromFile(file string) []FunctionInfo {
	var funcs []FunctionInfo

	blocksByFile := make([]CoverBlock, 0)
	for _, block := range a.Profile.Blocks {
		if block.File == file {
			blocksByFile = append(blocksByFile, block)
		}
	}

	if len(blocksByFile) == 0 {
		return funcs
	}

	sort.Slice(blocksByFile, func(i, j int) bool {
		return blocksByFile[i].StartLine < blocksByFile[j].StartLine
	})

	pkg := extractPackageName(file)
	funcIndex := 0
	currentFuncStart := blocksByFile[0].StartLine
	currentFuncEnd := blocksByFile[0].EndLine

	for i := 1; i < len(blocksByFile); i++ {
		block := blocksByFile[i]
		if block.StartLine <= currentFuncEnd+1 {
			if block.EndLine > currentFuncEnd {
				currentFuncEnd = block.EndLine
			}
		} else {
			funcIndex++
			funcs = append(funcs, FunctionInfo{
				Name:      generateFuncName(pkg, funcIndex),
				StartLine: currentFuncStart,
				EndLine:   currentFuncEnd,
			})
			currentFuncStart = block.StartLine
			currentFuncEnd = block.EndLine
		}
	}

	funcIndex++
	funcs = append(funcs, FunctionInfo{
		Name:      generateFuncName(pkg, funcIndex),
		StartLine: currentFuncStart,
		EndLine:   currentFuncEnd,
	})

	return funcs
}

func generateFuncName(pkg string, index int) string {
	return "Func" + string(rune('A'+index-1))
}

func extractPackageName(filePath string) string {
	parts := strings.Split(filePath, "/")
	if len(parts) <= 1 {
		return "main"
	}
	return strings.Join(parts[:len(parts)-1], "/")
}

func calculateConfidenceInterval(a *AnalysisResult) (float64, float64) {
	var totalCounts []int

	for _, coveredLines := range a.FileCoveredLines {
		for _, count := range coveredLines {
			if count > 0 {
				totalCounts = append(totalCounts, count)
			}
		}
	}

	if len(totalCounts) == 0 {
		return 0, 0
	}

	mean := 0.0
	for _, c := range totalCounts {
		mean += float64(c)
	}
	mean /= float64(len(totalCounts))

	variance := 0.0
	for _, c := range totalCounts {
		diff := float64(c) - mean
		variance += diff * diff
	}
	variance /= float64(len(totalCounts))

	stdDev := math.Sqrt(variance)
	standardError := stdDev / math.Sqrt(float64(len(totalCounts)))

	marginOfError := 1.96 * standardError
	return mean - marginOfError, mean + marginOfError
}

func DiffReports(oldReport, newReport *api.FunctionCoverageReport) *api.DiffReport {
	diff := &api.DiffReport{
		Changed:   []api.DiffFunction{},
		Decreased: []api.DiffFunction{},
	}

	oldMap := make(map[string]api.FunctionCoverage)
	for _, fn := range oldReport.Functions {
		key := fn.Package + "." + fn.Function
		oldMap[key] = fn
	}

	newMap := make(map[string]api.FunctionCoverage)
	for _, fn := range newReport.Functions {
		key := fn.Package + "." + fn.Function
		newMap[key] = fn
	}

	for key, newFn := range newMap {
		if oldFn, exists := oldMap[key]; exists {
			change := newFn.Percentage - oldFn.Percentage
			if math.Abs(change) > 0.001 {
				diffFn := api.DiffFunction{
					Package:       newFn.Package,
					Function:      newFn.Function,
					OldPercentage: oldFn.Percentage,
					NewPercentage: newFn.Percentage,
					Change:        change,
					IsDecrease:    change < 0,
				}
				diff.Changed = append(diff.Changed, diffFn)
				if change < 0 {
					diff.Decreased = append(diff.Decreased, diffFn)
				}
			}
		}
	}

	sort.Slice(diff.Changed, func(i, j int) bool {
		return diff.Changed[i].Change < diff.Changed[j].Change
	})

	sort.Slice(diff.Decreased, func(i, j int) bool {
		return diff.Decreased[i].Change < diff.Decreased[j].Change
	})

	return diff
}

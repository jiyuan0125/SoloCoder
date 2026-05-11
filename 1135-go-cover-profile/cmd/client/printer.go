package main

import (
	"fmt"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/solocoder/coverage-analyzer/pkg/api"
)

type ReportPrinter struct {
	w *tabwriter.Writer
}

func NewReportPrinter(output func(format string, a ...interface{})) *ReportPrinter {
	w := tabwriter.NewWriter(nil, 0, 0, 2, ' ', tabwriter.Debug)
	return &ReportPrinter{w: w}
}

func (p *ReportPrinter) PrintSummary(summary *api.SummaryResponse) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    COVERAGE SUMMARY                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	fmt.Printf("\nTotal Lines:        %d\n", summary.TotalLines)
	fmt.Printf("Covered Lines:      %d\n", summary.CoveredLines)
	fmt.Printf("Uncovered Lines:    %d\n", summary.UncoveredLines)
	fmt.Printf("Overall Coverage:   %.2f%%\n", summary.Percentage)

	if summary.ConfidenceLower != 0 || summary.ConfidenceUpper != 0 {
		fmt.Printf("Confidence Interval: [%.2f, %.2f] (95%%)\n", 
			summary.ConfidenceLower, summary.ConfidenceUpper)
	}

	if len(summary.LowestPackages) > 0 {
		fmt.Println("\n━━━━━━━━━━━━━━━━━ LOWEST COVERAGE PACKAGES ━━━━━━━━━━━━━━━━━")
		for _, pkg := range summary.LowestPackages {
			fmt.Printf("  %s: %.2f%% (%d/%d lines)\n", 
				pkg.Package, pkg.Percentage, pkg.CoveredLines, pkg.TotalLines)
		}
	}

	if len(summary.HighestPackages) > 0 {
		fmt.Println("\n━━━━━━━━━━━━━━━━━ HIGHEST COVERAGE PACKAGES ━━━━━━━━━━━━━━━━━")
		for _, pkg := range summary.HighestPackages {
			fmt.Printf("  %s: %.2f%% (%d/%d lines)\n", 
				pkg.Package, pkg.Percentage, pkg.CoveredLines, pkg.TotalLines)
		}
	}
}

func (p *ReportPrinter) PrintFunctionCoverage(report *api.FunctionCoverageReport) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  FUNCTION-LEVEL COVERAGE                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	if len(report.Functions) == 0 {
		fmt.Println("\nNo functions found.")
		return
	}

	w := tabwriter.NewWriter(nil, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "\nPACKAGE\tFUNCTION\tCOVERED\tTOTAL\tCOVERAGE%")
	fmt.Fprintln(w, "───────\t────────\t───────\t─────\t─────────")

	sorted := make([]api.FunctionCoverage, len(report.Functions))
	copy(sorted, report.Functions)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Percentage < sorted[j].Percentage
	})

	for _, fn := range sorted {
		bar := p.createPercentageBar(fn.Percentage, 20)
		fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%.2f%% %s\n",
			fn.Package, fn.Function, fn.CoveredLines, fn.TotalLines, fn.Percentage, bar)
	}

	w.Flush()
}

func (p *ReportPrinter) PrintPackageCoverage(report *api.PackageCoverageReport) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   PACKAGE-LEVEL COVERAGE                   ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	if len(report.Packages) == 0 {
		fmt.Println("\nNo packages found.")
		return
	}

	sorted := make([]api.PackageCoverage, len(report.Packages))
	copy(sorted, report.Packages)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Percentage < sorted[j].Percentage
	})

	maxPkgLen := 0
	for _, pkg := range sorted {
		if len(pkg.Package) > maxPkgLen {
			maxPkgLen = len(pkg.Package)
		}
	}

	fmt.Println("")
	for _, pkg := range sorted {
		bar := p.createPercentageBar(pkg.Percentage, 40)
		paddedPkg := pkg.Package + strings.Repeat(" ", maxPkgLen-len(pkg.Package))
		fmt.Printf("  %s │%s│ %.2f%% (%d/%d)\n",
			paddedPkg, bar, pkg.Percentage, pkg.CoveredLines, pkg.TotalLines)
	}
}

func (p *ReportPrinter) PrintLineCoverage(report *api.LineCoverageReport) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                     LINE-LEVEL COVERAGE                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	if len(report.Files) == 0 {
		fmt.Println("\nNo files found.")
		return
	}

	for _, file := range report.Files {
		fmt.Printf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("\nFile: %s\n", file.File)
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		covered := 0
		uncovered := 0
		partial := 0

		for _, line := range file.Lines {
			var statusChar string
			switch line.Status {
			case "covered":
				statusChar = "[+]"
				covered++
			case "partially-covered":
				statusChar = "[~]"
				partial++
			default:
				statusChar = "[-]"
				uncovered++
			}

			countInfo := ""
			if line.ExecutionCount > 0 {
				countInfo = fmt.Sprintf(" (x%d)", line.ExecutionCount)
			}

			fmt.Printf("  %s Line %4d: %s%s\n", 
				statusChar, line.LineNumber, line.Status, countInfo)
		}

		fmt.Printf("\n  Summary: %d covered, %d partial, %d uncovered\n",
			covered, partial, uncovered)
	}
}

func (p *ReportPrinter) PrintDiff(diff *api.DiffReport) {
	fmt.Println("\n╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║                   COVERAGE DIFF REPORT                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	fmt.Printf("\nComparison: %s -> %s\n", diff.OldID[:8]+"...", diff.NewID[:8]+"...")
	
	oldCov := diff.OldSummary.Percentage
	newCov := diff.NewSummary.Percentage
	change := newCov - oldCov
	
	changeSymbol := "="
	if change > 0.001 {
		changeSymbol = "▲"
	} else if change < -0.001 {
		changeSymbol = "▼"
	}
	
	fmt.Printf("Overall Coverage: %.2f%% -> %.2f%% %s (%.2f%% change)\n",
		oldCov, newCov, changeSymbol, change)

	if len(diff.Decreased) > 0 {
		fmt.Println("\n━━━━━━━━━━━━━━━━━ DECREASED FUNCTIONS ━━━━━━━━━━━━━━━━━")
		for _, fn := range diff.Decreased {
			fmt.Printf("  ⚠️  %s.%s: %.2f%% -> %.2f%% (%.2f%%)\n",
				fn.Package, fn.Function, fn.OldPercentage, fn.NewPercentage, fn.Change)
		}
	} else {
		fmt.Println("\n✓ No functions with decreased coverage")
	}

	if len(diff.Changed) > 0 {
		fmt.Println("\n━━━━━━━━━━━━━━━━━ ALL CHANGED FUNCTIONS ━━━━━━━━━━━━━━━━━")
		for _, fn := range diff.Changed {
			symbol := "▲"
			if fn.Change < 0 {
				symbol = "▼"
			}
			fmt.Printf("  %s %s.%s: %.2f%% -> %.2f%% (%+.2f%%)\n",
				symbol, fn.Package, fn.Function, fn.OldPercentage, fn.NewPercentage, fn.Change)
		}
	} else {
		fmt.Println("\nNo functions with coverage changes")
	}
}

func (p *ReportPrinter) createPercentageBar(percentage float64, width int) string {
	filled := int((percentage / 100) * float64(width))
	if percentage > 0 && filled == 0 {
		filled = 1
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return bar
}

package client

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"slowquery/protocol"
)

type TablePrinter struct {
	writer io.Writer
}

func NewTablePrinter(writer io.Writer) *TablePrinter {
	return &TablePrinter{
		writer: writer,
	}
}

func (p *TablePrinter) PrintSummary(stats *protocol.Statistics) {
	fmt.Fprintf(p.writer, "\n")
	fmt.Fprintf(p.writer, "=== Summary ===\n")
	fmt.Fprintf(p.writer, "Total rows parsed:   %d\n", stats.TotalRowsParsed)
	fmt.Fprintf(p.writer, "Rows skipped:        %d\n", stats.RowsSkipped)
	fmt.Fprintf(p.writer, "Unique SQL templates: %d\n", stats.UniqueTemplates)
	fmt.Fprintf(p.writer, "\n")
}

func (p *TablePrinter) PrintTable(stats *protocol.Statistics) {
	if len(stats.TemplateStats) == 0 {
		fmt.Fprintln(p.writer, "No results to display")
		return
	}

	tw := tabwriter.NewWriter(p.writer, 0, 0, 2, ' ', 0)

	fmt.Fprintln(tw, "Rank\tSQL Template\tCount\tAvg Time(ms)\tMax Time(ms)\tTotal Time(ms)\tAvg Scan Rows")
	fmt.Fprintln(tw, "----\t------------\t-----\t------------\t------------\t--------------\t-------------")

	for _, stat := range stats.TemplateStats {
		sqlTmpl := truncateString(stat.SQLTemplate, 60)
		fmt.Fprintf(tw, "%d\t%s\t%d\t%.2f\t%.2f\t%.2f\t%.1f\n",
			stat.Rank,
			sqlTmpl,
			stat.Count,
			stat.AvgExecTimeMs,
			stat.MaxExecTimeMs,
			stat.TotalExecTimeMs,
			stat.AvgScanRows,
		)
	}

	tw.Flush()
}

func (p *TablePrinter) PrintFullTemplate(stats *protocol.Statistics, rank int) {
	for _, stat := range stats.TemplateStats {
		if stat.Rank == rank {
			fmt.Fprintf(p.writer, "\n")
			fmt.Fprintf(p.writer, "=== SQL Template #%d ===\n", rank)
			fmt.Fprintf(p.writer, "SQL:\n%s\n\n", stat.SQLTemplate)
			fmt.Fprintf(p.writer, "Occurrences:        %d\n", stat.Count)
			fmt.Fprintf(p.writer, "Avg Exec Time:      %.2f ms\n", stat.AvgExecTimeMs)
			fmt.Fprintf(p.writer, "Max Exec Time:      %.2f ms\n", stat.MaxExecTimeMs)
			fmt.Fprintf(p.writer, "Total Exec Time:    %.2f ms\n", stat.TotalExecTimeMs)
			fmt.Fprintf(p.writer, "Avg Scan Rows:      %.1f\n", stat.AvgScanRows)
			return
		}
	}
	fmt.Fprintf(p.writer, "No template found with rank %d\n", rank)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

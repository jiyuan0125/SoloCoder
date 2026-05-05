package csvproc

import (
	"encoding/csv"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"csv-merger/common"
)

type MergeOptions struct {
	DedupColumns []string
	SortColumn   string
	SortOrder    common.SortOrder
	CustomHeader []string
}

func MergeCSVFiles(files []*CSVFile, outputPath string, options MergeOptions) (int, error) {
	if len(files) == 0 {
		return 0, fmt.Errorf("no valid files to merge")
	}

	var allRows [][]string
	header := files[0].Header

	if len(options.CustomHeader) > 0 {
		header = options.CustomHeader
	}

	for _, file := range files {
		allRows = append(allRows, file.Rows...)
	}

	if len(options.DedupColumns) > 0 {
		allRows = deduplicate(allRows, files[0].Header, options.DedupColumns)
	}

	if options.SortColumn != "" {
		var err error
		allRows, err = sortRows(allRows, files[0].Header, options.SortColumn, options.SortOrder)
		if err != nil {
			return 0, err
		}
	}

	if err := writeCSV(outputPath, header, allRows); err != nil {
		return 0, err
	}

	return len(allRows), nil
}

func deduplicate(rows [][]string, header []string, dedupColumns []string) [][]string {
	colIndexMap := make(map[string]int)
	for i, col := range header {
		colIndexMap[col] = i
	}

	var dedupIndices []int
	for _, col := range dedupColumns {
		if idx, exists := colIndexMap[col]; exists {
			dedupIndices = append(dedupIndices, idx)
		}
	}

	if len(dedupIndices) == 0 {
		return rows
	}

	uniqueMap := make(map[string]int)
	var result [][]string

	for _, row := range rows {
		keyParts := make([]string, len(dedupIndices))
		for j, idx := range dedupIndices {
			if idx < len(row) {
				keyParts[j] = row[idx]
			}
		}
		key := strings.Join(keyParts, "|")

		if existingIdx, exists := uniqueMap[key]; exists {
			result[existingIdx] = row
		} else {
			uniqueMap[key] = len(result)
			result = append(result, row)
		}
	}

	return result
}

func sortRows(rows [][]string, header []string, sortColumn string, sortOrder common.SortOrder) ([][]string, error) {
	colIndex := -1
	for i, col := range header {
		if col == sortColumn {
			colIndex = i
			break
		}
	}

	if colIndex == -1 {
		return nil, fmt.Errorf("sort column '%s' not found in header", sortColumn)
	}

	allNumeric := true
	for _, row := range rows {
		if colIndex < len(row) {
			if _, err := strconv.ParseFloat(strings.TrimSpace(row[colIndex]), 64); err != nil {
				allNumeric = false
				break
			}
		}
	}

	sort.Slice(rows, func(i, j int) bool {
		rowI := rows[i]
		rowJ := rows[j]

		var valI, valJ string
		if colIndex < len(rowI) {
			valI = strings.TrimSpace(rowI[colIndex])
		}
		if colIndex < len(rowJ) {
			valJ = strings.TrimSpace(rowJ[colIndex])
		}

		if allNumeric && valI != "" && valJ != "" {
			numI, _ := strconv.ParseFloat(valI, 64)
			numJ, _ := strconv.ParseFloat(valJ, 64)

			if sortOrder == common.SortDesc {
				return numI > numJ
			}
			return numI < numJ
		}

		if sortOrder == common.SortDesc {
			return valI > valJ
		}
		return valI < valJ
	})

	return rows, nil
}

func writeCSV(outputPath string, header []string, rows [][]string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	writer.UseCRLF = false

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	if err := writer.WriteAll(rows); err != nil {
		return fmt.Errorf("failed to write rows: %w", err)
	}

	writer.Flush()
	return writer.Error()
}

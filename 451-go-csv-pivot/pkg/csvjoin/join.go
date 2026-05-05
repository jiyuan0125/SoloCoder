package csvjoin

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// JoinResult represents the result of a join operation.
type JoinResult struct {
	OutputPath string
	RowCount   int
	Columns    []string
}

// Join performs the join operation and writes the result to the output file.
// It uses a memory-efficient approach: only one file is fully loaded into memory
// (the smaller one), while the other is streamed.
func (j *Joiner) Join(outputPath string) (*JoinResult, error) {
	if len(j.options.Files) < 2 {
		return nil, &JoinError{Op: "Join", Err: ErrInsufficientFiles}
	}

	// Start with the first file, then join with each subsequent file
	currentResultPath := j.options.Files[0].Path
	currentOptions := j.options.Files[0]
	var resultColumns []string
	var finalRowCount int

	for i := 1; i < len(j.options.Files); i++ {
		rightOptions := j.options.Files[i]

		// Create a temporary output file for intermediate results
		tempOutput := outputPath
		if i < len(j.options.Files)-1 {
			tempOutput = fmt.Sprintf("%s.tmp.%d", outputPath, i)
			defer os.Remove(tempOutput)
		}

		// Perform pairwise join
		joinResult, err := j.joinTwoFiles(
			currentResultPath,
			currentOptions,
			rightOptions.Path,
			rightOptions,
			tempOutput,
			i == 1, // First join needs to read both headers
		)
		if err != nil {
			return nil, err
		}

		currentResultPath = tempOutput
		resultColumns = joinResult.Columns
		finalRowCount = joinResult.RowCount

		// Update current options for next iteration
		currentOptions = FileOptions{
			Path:         tempOutput,
			SkipHeader:   false,
			ColumnNames:  resultColumns,
			FileSuffix:   "", // Already has suffix from previous join
		}
	}

	return &JoinResult{
		OutputPath: outputPath,
		RowCount:   finalRowCount,
		Columns:    resultColumns,
	}, nil
}

// joinTwoFiles joins two CSV files with memory-efficient processing.
func (j *Joiner) joinTwoFiles(
	leftPath string, leftOpts FileOptions,
	rightPath string, rightOpts FileOptions,
	outputPath string, isFirstJoin bool,
) (*JoinResult, error) {
	// Determine which file is smaller to minimize memory usage
	leftSize, _ := getFileSize(leftPath)
	rightSize, _ := getFileSize(rightPath)

	// Load the smaller file into memory as a lookup table
	var lookupTable map[string][][]string
	var lookupColumns []string
	var lookupKeyIndex int
	var streamPath string
	var streamOpts FileOptions

	if leftSize <= rightSize {
		// Load left file, stream right file
		var err error
		lookupTable, lookupColumns, lookupKeyIndex, err = j.loadFileIntoLookup(leftPath, leftOpts, isFirstJoin)
		if err != nil {
			return nil, err
		}
		streamPath = rightPath
		streamOpts = rightOpts
	} else {
		// Load right file, stream left file
		var err error
		lookupTable, lookupColumns, lookupKeyIndex, err = j.loadFileIntoLookup(rightPath, rightOpts, isFirstJoin)
		if err != nil {
			return nil, err
		}
		streamPath = leftPath
		streamOpts = leftOpts
	}

	// Perform the join based on which file was loaded
	if leftSize <= rightSize {
		return j.joinStreamingRight(lookupTable, lookupColumns, lookupKeyIndex, streamPath, streamOpts, outputPath, isFirstJoin)
	}
	return j.joinStreamingLeft(lookupTable, lookupColumns, lookupKeyIndex, streamPath, streamOpts, outputPath, isFirstJoin)
}

// loadFileIntoLookup reads a CSV file and creates a lookup table indexed by the join key.
func (j *Joiner) loadFileIntoLookup(
	filePath string, opts FileOptions, isFirstJoin bool,
) (map[string][][]string, []string, int, error) {
	reader, err := j.openCSVReader(filePath, opts, isFirstJoin)
	if err != nil {
		return nil, nil, -1, err
	}
	defer reader.Close()

	columns := reader.Columns()
	if len(columns) == 0 {
		return nil, nil, -1, &JoinError{
			Op:  "loadFileIntoLookup",
			Err: &ColumnNotFoundError{ColumnName: j.options.JoinKey, FilePath: filePath},
		}
	}

	// Find the join key column index
	keyIndex, err := j.findJoinKeyColumn(columns, filePath)
	if err != nil {
		return nil, nil, -1, err
	}

	// Create lookup table: key -> list of rows
	lookupTable := make(map[string][][]string)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, nil, -1, err
		}

		if len(row) <= keyIndex {
			// Row doesn't have enough columns, skip or handle error
			continue
		}

		key := row[keyIndex]
		if j.options.TrimKeys {
			key = strings.TrimSpace(key)
		}

		lookupTable[key] = append(lookupTable[key], row)
	}

	return lookupTable, columns, keyIndex, nil
}

// findJoinKeyColumn finds the index of the join key column.
// It checks both the original column name and the column name with suffix.
func (j *Joiner) findJoinKeyColumn(columns []string, filePath string) (int, error) {
	joinKey := j.options.JoinKey

	// First, try exact match
	for i, col := range columns {
		if col == joinKey {
			return i, nil
		}
	}

	// Try without suffix (if the key has a suffix)
	for i, col := range columns {
		if strings.HasSuffix(joinKey, "_f0") || strings.HasSuffix(joinKey, "_f1") {
			// Remove the suffix and check
			baseKey := joinKey[:strings.LastIndex(joinKey, "_")]
			if col == baseKey || strings.HasPrefix(col, baseKey+"_") {
				return i, nil
			}
		}
		// Check if column name starts with the join key
		if strings.HasPrefix(col, joinKey+"_") {
			return i, nil
		}
	}

	return -1, &JoinError{
		Op:  "findJoinKeyColumn",
		Err: &ColumnNotFoundError{ColumnName: joinKey, FilePath: filePath},
	}
}

// joinStreamingRight joins when the left file is in memory and the right file is streamed.
func (j *Joiner) joinStreamingRight(
	lookupTable map[string][][]string, lookupColumns []string, lookupKeyIndex int,
	streamPath string, streamOpts FileOptions,
	outputPath string, isFirstJoin bool,
) (*JoinResult, error) {
	streamReader, err := j.openCSVReader(streamPath, streamOpts, isFirstJoin)
	if err != nil {
		return nil, err
	}
	defer streamReader.Close()

	streamColumns := streamReader.Columns()
	streamKeyIndex, err := j.findJoinKeyColumn(streamColumns, streamPath)
	if err != nil {
		return nil, err
	}

	// Prepare output columns with suffixes
	leftSuffix := j.getFileSuffix(0)
	rightSuffix := j.getFileSuffix(1)

	if !isFirstJoin {
		// For subsequent joins, left columns already have suffixes
		leftSuffix = ""
	}

	outputColumns := j.prepareOutputColumns(
		lookupColumns, lookupKeyIndex, streamColumns, streamKeyIndex,
		leftSuffix, rightSuffix,
	)

	// Create output writer
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, &JoinError{Op: "joinStreamingRight", Err: err}
	}
	defer outputFile.Close()

	writer := NewRFC4180Writer(outputFile)

	// Write header
	if err := writer.Write(outputColumns); err != nil {
		return nil, err
	}

	rowCount := 0
	matchedKeys := make(map[string]bool)

	// Stream the right file and join with lookup table
	for {
		streamRow, err := streamReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(streamRow) <= streamKeyIndex {
			continue
		}

		streamKey := streamRow[streamKeyIndex]
		if j.options.TrimKeys {
			streamKey = strings.TrimSpace(streamKey)
		}

		if lookupRows, ok := lookupTable[streamKey]; ok {
			// Match found
			matchedKeys[streamKey] = true

			for _, lookupRow := range lookupRows {
				outputRow := j.buildOutputRow(
					lookupRow, lookupKeyIndex, streamRow, streamKeyIndex,
					len(lookupColumns), len(streamColumns),
				)
				if err := writer.Write(outputRow); err != nil {
					return nil, err
				}
				rowCount++
			}
		} else if j.options.JoinType == LeftJoin {
			// No match, but it's a left join - wait, we need to think about this
			// Since lookup is left file and we're streaming right,
			// for left join we need to include all lookup rows
			// Let's handle this differently
		}
	}

	// For left join, add all unmatched rows from the lookup table
	if j.options.JoinType == LeftJoin {
		for key, lookupRows := range lookupTable {
			if !matchedKeys[key] {
				for _, lookupRow := range lookupRows {
					// Create empty right row
					emptyRightRow := make([]string, len(streamColumns))
					outputRow := j.buildOutputRow(
						lookupRow, lookupKeyIndex, emptyRightRow, streamKeyIndex,
						len(lookupColumns), len(streamColumns),
					)
					if err := writer.Write(outputRow); err != nil {
						return nil, err
					}
					rowCount++
				}
			}
		}
	}

	return &JoinResult{
		OutputPath: outputPath,
		RowCount:   rowCount,
		Columns:    outputColumns,
	}, nil
}

// joinStreamingLeft joins when the right file is in memory and the left file is streamed.
func (j *Joiner) joinStreamingLeft(
	lookupTable map[string][][]string, lookupColumns []string, lookupKeyIndex int,
	streamPath string, streamOpts FileOptions,
	outputPath string, isFirstJoin bool,
) (*JoinResult, error) {
	streamReader, err := j.openCSVReader(streamPath, streamOpts, isFirstJoin)
	if err != nil {
		return nil, err
	}
	defer streamReader.Close()

	streamColumns := streamReader.Columns()
	streamKeyIndex, err := j.findJoinKeyColumn(streamColumns, streamPath)
	if err != nil {
		return nil, err
	}

	// Prepare output columns with suffixes
	leftSuffix := j.getFileSuffix(0)
	rightSuffix := j.getFileSuffix(1)

	if !isFirstJoin {
		// For subsequent joins, left columns already have suffixes
		leftSuffix = ""
	}

	// Note: in this case, stream is left, lookup is right
	outputColumns := j.prepareOutputColumns(
		streamColumns, streamKeyIndex, lookupColumns, lookupKeyIndex,
		leftSuffix, rightSuffix,
	)

	// Create output writer
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return nil, &JoinError{Op: "joinStreamingLeft", Err: err}
	}
	defer outputFile.Close()

	writer := NewRFC4180Writer(outputFile)

	// Write header
	if err := writer.Write(outputColumns); err != nil {
		return nil, err
	}

	rowCount := 0

	// Stream the left file and join with lookup table
	for {
		streamRow, err := streamReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(streamRow) <= streamKeyIndex {
			continue
		}

		streamKey := streamRow[streamKeyIndex]
		if j.options.TrimKeys {
			streamKey = strings.TrimSpace(streamKey)
		}

		if lookupRows, ok := lookupTable[streamKey]; ok {
			// Match found
			for _, lookupRow := range lookupRows {
				outputRow := j.buildOutputRow(
					streamRow, streamKeyIndex, lookupRow, lookupKeyIndex,
					len(streamColumns), len(lookupColumns),
				)
				if err := writer.Write(outputRow); err != nil {
					return nil, err
				}
				rowCount++
			}
		} else if j.options.JoinType == LeftJoin {
			// No match, but left join - include with empty right row
			emptyRightRow := make([]string, len(lookupColumns))
			outputRow := j.buildOutputRow(
				streamRow, streamKeyIndex, emptyRightRow, lookupKeyIndex,
				len(streamColumns), len(lookupColumns),
			)
			if err := writer.Write(outputRow); err != nil {
				return nil, err
			}
			rowCount++
		}
		// For inner join, skip unmatched rows
	}

	return &JoinResult{
		OutputPath: outputPath,
		RowCount:   rowCount,
		Columns:    outputColumns,
	}, nil
}

// openCSVReader opens a CSV file with the given options.
func (j *Joiner) openCSVReader(filePath string, opts FileOptions, isFirstJoin bool) (CSVReader, error) {
	// Open file with encoding detection
	reader, _, err := OpenFileForReading(filePath)
	if err != nil {
		return nil, err
	}

	// Check if we need to read the header
	hasHeader := !opts.SkipHeader
	if !isFirstJoin {
		// For intermediate files, we always have a header
		hasHeader = true
	}

	csvReader, err := NewRFC4180Reader(reader, hasHeader)
	if err != nil {
		return nil, err
	}

	// If SkipHeader is true, we need to either:
	// 1. Use provided ColumnNames, or
	// 2. Generate default column names
	if opts.SkipHeader && isFirstJoin {
		// The first row was read as header, but it's actually data
		// We need to push it back and set custom column names

		// First, let's get the "header" which is actually the first data row
		firstDataRow := csvReader.Columns()

		// Determine column names
		var columnNames []string
		if len(opts.ColumnNames) > 0 {
			columnNames = opts.ColumnNames
		} else {
			// Generate default names: col0, col1, ...
			columnNames = make([]string, len(firstDataRow))
			for i := range columnNames {
				columnNames[i] = fmt.Sprintf("col%d", i)
			}
		}

		// We need to create a new reader that includes the first data row
		// This is a bit tricky - we'll read all data and prepend the first row

		// Read remaining data
		remainingData, err := csvReader.ReadAll()
		if err != nil {
			return nil, err
		}

		// Combine first data row with remaining data
		allData := make([][]string, 0, 1+len(remainingData))
		allData = append(allData, firstDataRow)
		allData = append(allData, remainingData...)

		// Create a new reader from the combined data
		var buf bytes.Buffer
		tempWriter := NewRFC4180Writer(&buf)
		if err := tempWriter.WriteAll(allData); err != nil {
			return nil, err
		}

		// Create new reader with proper header (column names)
		newReader := bytes.NewReader(buf.Bytes())

		// First, write the column names as header
		var headerBuf bytes.Buffer
		headerWriter := NewRFC4180Writer(&headerBuf)
		if err := headerWriter.Write(columnNames); err != nil {
			return nil, err
		}

		// Combine header with data
		combinedReader := io.MultiReader(&headerBuf, newReader)

		// Create new CSV reader
		newCSVReader, err := NewRFC4180Reader(combinedReader, true)
		if err != nil {
			return nil, err
		}

		return newCSVReader, nil
	}

	return csvReader, nil
}

// getFileSuffix returns the suffix for the file at the given index.
func (j *Joiner) getFileSuffix(index int) string {
	if index >= len(j.options.Files) {
		return fmt.Sprintf("_f%d", index)
	}

	suffix := j.options.Files[index].FileSuffix
	if suffix == "" {
		return fmt.Sprintf("_f%d", index)
	}
	return suffix
}

// prepareOutputColumns prepares the output column names with suffixes.
func (j *Joiner) prepareOutputColumns(
	leftColumns []string, leftKeyIndex int,
	rightColumns []string, rightKeyIndex int,
	leftSuffix, rightSuffix string,
) []string {
	outputColumns := make([]string, 0, len(leftColumns)+len(rightColumns)-1)

	// Add left columns (including key)
	for i, col := range leftColumns {
		if i == leftKeyIndex {
			// Key column - keep as is (will be first column)
			outputColumns = append(outputColumns, col)
		} else {
			// Non-key column - add suffix
			outputColumns = append(outputColumns, col+leftSuffix)
		}
	}

	// Add right columns (excluding key to avoid duplication)
	for i, col := range rightColumns {
		if i == rightKeyIndex {
			// Skip the key column from the right
			continue
		}
		outputColumns = append(outputColumns, col+rightSuffix)
	}

	return outputColumns
}

// buildOutputRow builds the output row from left and right rows.
func (j *Joiner) buildOutputRow(
	leftRow []string, leftKeyIndex int,
	rightRow []string, rightKeyIndex int,
	leftColCount, rightColCount int,
) []string {
	outputRow := make([]string, 0, leftColCount+rightColCount-1)

	// Add left row (including key)
	for i := 0; i < leftColCount; i++ {
		if i < len(leftRow) {
			outputRow = append(outputRow, leftRow[i])
		} else {
			outputRow = append(outputRow, "")
		}
	}

	// Add right row (excluding key)
	for i := 0; i < rightColCount; i++ {
		if i == rightKeyIndex {
			continue
		}
		if i < len(rightRow) {
			outputRow = append(outputRow, rightRow[i])
		} else {
			outputRow = append(outputRow, "")
		}
	}

	return outputRow
}

// getFileSize returns the size of a file in bytes.
func getFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// ReadAllRows reads all rows from a CSV file (for demo purposes).
func ReadAllRows(filePath string, skipHeader bool, columnNames []string) ([][]string, []string, error) {
	reader, _, err := OpenFileForReading(filePath)
	if err != nil {
		return nil, nil, err
	}

	hasHeader := !skipHeader
	csvReader, err := NewRFC4180Reader(reader.(io.Reader), hasHeader)
	if err != nil {
		return nil, nil, err
	}
	defer csvReader.Close()

	cols := csvReader.Columns()

	// If skipHeader is true and no columnNames provided, use the "header" as first data row
	if skipHeader {
		// The first row was read as header, but it's actually data
		firstDataRow := cols

		// Determine column names
		var actualColumnNames []string
		if len(columnNames) > 0 {
			actualColumnNames = columnNames
		} else {
			actualColumnNames = make([]string, len(firstDataRow))
			for i := range actualColumnNames {
				actualColumnNames[i] = fmt.Sprintf("col%d", i)
			}
		}

		// Read remaining data
		remainingData, err := csvReader.ReadAll()
		if err != nil {
			return nil, nil, err
		}

		// Combine first data row with remaining data
		allData := make([][]string, 0, 1+len(remainingData))
		allData = append(allData, firstDataRow)
		allData = append(allData, remainingData...)

		return allData, actualColumnNames, nil
	}

	// Read all data
	rows, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, err
	}

	return rows, cols, nil
}

// NewBufferedWriter creates a buffered writer for large files.
func NewBufferedWriter(w io.Writer) *bufio.Writer {
	return bufio.NewWriterSize(w, 64*1024) // 64KB buffer
}

package csvjoin

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"unicode/utf8"
)

// RFC4180Reader is a CSV reader that implements RFC 4180 standard.
// It correctly handles fields containing commas, quotes, and newlines.
type RFC4180Reader struct {
	scanner   *bufio.Scanner
	buffer    bytes.Buffer
	columns   []string
	hasHeader bool
}

// NewRFC4180Reader creates a new RFC 4180 compliant CSV reader.
// If hasHeader is true, the first row is read as column names.
func NewRFC4180Reader(r io.Reader, hasHeader bool) (*RFC4180Reader, error) {
	reader := &RFC4180Reader{
		scanner:   bufio.NewScanner(r),
		hasHeader: hasHeader,
	}
	reader.scanner.Split(reader.scanLines)

	if hasHeader {
		columns, err := reader.Read()
		if err != nil {
			return nil, &JoinError{Op: "NewRFC4180Reader", Err: err}
		}
		reader.columns = columns
	}

	return reader, nil
}

// Columns returns the column names.
func (r *RFC4180Reader) Columns() []string {
	return r.columns
}

// Read reads a single row from the CSV file.
// Returns io.EOF when there are no more rows.
func (r *RFC4180Reader) Read() ([]string, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return nil, &JoinError{Op: "Read", Err: err}
		}
		return nil, io.EOF
	}

	line := r.scanner.Text()
	return r.parseLine(line)
}

// ReadAll reads all remaining rows from the CSV file.
func (r *RFC4180Reader) ReadAll() ([][]string, error) {
	var records [][]string
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	return records, nil
}

// Close closes the reader (no-op for this implementation).
func (r *RFC4180Reader) Close() error {
	return nil
}

// scanLines is a custom split function for bufio.Scanner that handles
// quoted fields containing newlines according to RFC 4180.
func (r *RFC4180Reader) scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	inQuotes := false
	quoteCount := 0

	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '"':
			quoteCount++
			// RFC 4180: quotes inside quoted fields are escaped by preceding them with another quote
			// So we only toggle inQuotes when we see an odd number of quotes
			if quoteCount%2 == 1 {
				if i == 0 || data[i-1] != '"' {
					inQuotes = !inQuotes
				}
			}
		case '\n':
			if !inQuotes {
				// Return the line without the trailing newline
				if i > 0 && data[i-1] == '\r' {
					return i + 1, data[0 : i-1], nil
				}
				return i + 1, data[0:i], nil
			}
		}
	}

	if atEOF {
		// Return remaining data as the last line
		return len(data), data, nil
	}

	// Request more data
	return 0, nil, nil
}

// parseLine parses a single CSV line according to RFC 4180.
func (r *RFC4180Reader) parseLine(line string) ([]string, error) {
	var fields []string
	var currentField strings.Builder
	inQuotes := false
	i := 0

	for i < len(line) {
		c := line[i]

		if inQuotes {
			if c == '"' {
				// Check if this is an escaped quote ("")
				if i+1 < len(line) && line[i+1] == '"' {
					currentField.WriteByte('"')
					i += 2
					continue
				}
				// End of quoted field
				inQuotes = false
				i++
				continue
			}
			// Regular character inside quotes
			currentField.WriteByte(c)
			i++
			continue
		}

		// Not in quotes
		switch c {
		case '"':
			inQuotes = true
			i++
		case ',':
			// End of field
			fields = append(fields, currentField.String())
			currentField.Reset()
			i++
		default:
			currentField.WriteByte(c)
			i++
		}
	}

	// Add the last field
	fields = append(fields, currentField.String())

	// According to RFC 4180, all records should have the same number of fields
	// But we don't enforce this here - let the caller decide
	return fields, nil
}

// RFC4180Writer is a CSV writer that implements RFC 4180 standard.
type RFC4180Writer struct {
	writer io.Writer
}

// NewRFC4180Writer creates a new RFC 4180 compliant CSV writer.
func NewRFC4180Writer(w io.Writer) *RFC4180Writer {
	return &RFC4180Writer{writer: w}
}

// Write writes a single row to the CSV file.
func (w *RFC4180Writer) Write(record []string) error {
	for i, field := range record {
		if i > 0 {
			if _, err := w.writer.Write([]byte{','}); err != nil {
				return &JoinError{Op: "Write", Err: err}
			}
		}

		encodedField := w.encodeField(field)
		if _, err := w.writer.Write([]byte(encodedField)); err != nil {
			return &JoinError{Op: "Write", Err: err}
		}
	}

	if _, err := w.writer.Write([]byte("\r\n")); err != nil {
		return &JoinError{Op: "Write", Err: err}
	}

	return nil
}

// WriteAll writes all rows to the CSV file.
func (w *RFC4180Writer) WriteAll(records [][]string) error {
	for _, record := range records {
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// encodeField encodes a field according to RFC 4180.
// Fields containing commas, quotes, or newlines must be enclosed in quotes.
// Quotes inside fields must be escaped by preceding them with another quote.
func (w *RFC4180Writer) encodeField(field string) string {
	needsQuotes := false

	// Check if field needs quoting
	for i := 0; i < len(field); i++ {
		c := field[i]
		if c == ',' || c == '"' || c == '\n' || c == '\r' {
			needsQuotes = true
			break
		}
	}

	if !needsQuotes {
		return field
	}

	// Encode the field with quotes
	var encoded strings.Builder
	encoded.WriteByte('"')

	for i := 0; i < len(field); i++ {
		c := field[i]
		if c == '"' {
			// Escape quotes by doubling them
			encoded.WriteString(`""`)
		} else {
			encoded.WriteByte(c)
		}
	}

	encoded.WriteByte('"')
	return encoded.String()
}

// Flush flushes any buffered data (no-op for this implementation).
func (w *RFC4180Writer) Flush() {
	// No buffering in this implementation
}

// Error reports any error that occurred during a previous Write or Flush (no-op for this implementation).
func (w *RFC4180Writer) Error() error {
	return nil
}

// Close closes the writer (no-op for this implementation).
func (w *RFC4180Writer) Close() error {
	return nil
}

// detectBOM checks for UTF-8 BOM and removes it if present.
func detectBOM(r io.Reader) (io.Reader, error) {
	buf := bufio.NewReader(r)

	// Read the first 3 bytes to check for BOM
	bom, err := buf.Peek(3)
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Check for UTF-8 BOM (EF BB BF)
	if len(bom) >= 3 && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		// Discard the BOM
		_, _ = buf.Discard(3)
	}

	return buf, nil
}

// IsUTF8 checks if the data is valid UTF-8.
func IsUTF8(data []byte) bool {
	return utf8.Valid(data)
}

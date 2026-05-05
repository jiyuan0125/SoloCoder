package logparser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
	"unicode/utf8"
)

type LogEntry struct {
	Fields    map[string]string
	Line      string
	LineNum   int
	Timestamp any
}

type ParseError struct {
	LineNum int
	Reason  string
}

func (e ParseError) Error() string {
	return fmt.Sprintf("line %d: %s", e.LineNum, e.Reason)
}

type tokenType int

const (
	tokenLiteral tokenType = iota
	tokenField
)

type token struct {
	typ        tokenType
	value      string
	quoteStart string
	quoteEnd   string
}

type Parser struct {
	format           string
	fieldNames       []string
	tokens           []token
	startsWithField  bool
	multilineHandler MultilineHandler
	registry         *FormatRegistry
}

type MultilineHandler func(currentLine string, prevEntry *LogEntry) bool

var (
	DefaultMultilineHandler = func(currentLine string, prevEntry *LogEntry) bool {
		return prevEntry != nil && (strings.HasPrefix(currentLine, " ") ||
			strings.HasPrefix(currentLine, "\t") ||
			strings.HasPrefix(currentLine, "at "))
	}

	builtInFormats = map[string]FormatDefinition{
		"nginx_combined": {
			FormatString: `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"`,
			Fields: []string{"remote_addr", "remote_user", "time_local", "request",
				"status", "body_bytes_sent", "http_referer", "http_user_agent"},
		},
		"apache_common": {
			FormatString: `$remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent`,
			Fields: []string{"remote_addr", "remote_user", "time_local", "request",
				"status", "body_bytes_sent"},
		},
		"syslog": {
			FormatString: `<$priority>$version $timestamp $hostname $app_name $procid $msgid $message`,
			Fields: []string{"priority", "version", "timestamp", "hostname",
				"app_name", "procid", "msgid", "message"},
		},
	}
)

type FormatDefinition struct {
	FormatString string
	Fields       []string
}

type FormatRegistry struct {
	formats map[string]FormatDefinition
	mu      sync.RWMutex
}

func NewFormatRegistry() *FormatRegistry {
	r := &FormatRegistry{
		formats: make(map[string]FormatDefinition),
	}
	for name, def := range builtInFormats {
		r.Register(name, def)
	}
	return r
}

func (r *FormatRegistry) Register(name string, def FormatDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.formats[name] = def
}

func (r *FormatRegistry) Get(name string) (FormatDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	def, ok := r.formats[name]
	return def, ok
}

func (r *FormatRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.formats))
	for name := range r.formats {
		names = append(names, name)
	}
	return names
}

var DefaultRegistry = NewFormatRegistry()

func RegisterFormat(name string, def FormatDefinition) {
	DefaultRegistry.Register(name, def)
}

func NewParser(format string, fieldNames []string) (*Parser, error) {
	return newParserWithRegistry(format, fieldNames, DefaultRegistry)
}

func newParserWithRegistry(format string, fieldNames []string, registry *FormatRegistry) (*Parser, error) {
	if def, ok := registry.Get(format); ok {
		format = def.FormatString
		fieldNames = def.Fields
	}

	tokens, startsWithField, extractedFields, err := tokenizeFormat(format)
	if err != nil {
		return nil, err
	}

	if len(fieldNames) == 0 {
		fieldNames = extractedFields
	}

	if len(fieldNames) == 0 {
		return nil, fmt.Errorf("no fields defined in format")
	}

	return &Parser{
		format:           format,
		fieldNames:       fieldNames,
		tokens:           tokens,
		startsWithField:  startsWithField,
		multilineHandler: DefaultMultilineHandler,
		registry:         registry,
	}, nil
}

func tokenizeFormat(format string) ([]token, bool, []string, error) {
	var tokens []token
	var fields []string
	var builder strings.Builder
	inField := false
	startsWithField := false

	if len(format) > 0 && format[0] == '$' {
		startsWithField = true
	}

	for i := 0; i < len(format); i++ {
		c := format[i]
		if c == '$' && !inField {
			if builder.Len() > 0 {
				tokens = append(tokens, token{
					typ:   tokenLiteral,
					value: builder.String(),
				})
				builder.Reset()
			}
			inField = true
		} else if inField {
			if isFieldChar(c) {
				builder.WriteByte(c)
			} else {
				if builder.Len() > 0 {
					fieldName := builder.String()
					fields = append(fields, fieldName)

					quoteStart := ""
					if len(tokens) > 0 {
						prevToken := tokens[len(tokens)-1]
						if prevToken.typ == tokenLiteral {
							if strings.HasSuffix(prevToken.value, `"`) {
								quoteStart = `"`
							} else if strings.HasSuffix(prevToken.value, `[`) {
								quoteStart = `[`
							} else if strings.HasSuffix(prevToken.value, `'`) {
								quoteStart = `'`
							} else if strings.HasSuffix(prevToken.value, `<`) {
								quoteStart = `<`
							}
						}
					}

					quoteEnd := ""
					switch quoteStart {
					case `"`:
						quoteEnd = `"`
					case `[`:
						quoteEnd = `]`
					case `'`:
						quoteEnd = `'`
					case `<`:
						quoteEnd = `>`
					}

					tokens = append(tokens, token{
						typ:        tokenField,
						value:      fieldName,
						quoteStart: quoteStart,
						quoteEnd:   quoteEnd,
					})
					builder.Reset()
				}
				builder.WriteByte(c)
				inField = false
			}
		} else {
			builder.WriteByte(c)
		}
	}

	if inField && builder.Len() > 0 {
		fieldName := builder.String()
		fields = append(fields, fieldName)

		quoteStart := ""
		if len(tokens) > 0 {
			prevToken := tokens[len(tokens)-1]
			if prevToken.typ == tokenLiteral {
				if strings.HasSuffix(prevToken.value, `"`) {
					quoteStart = `"`
				} else if strings.HasSuffix(prevToken.value, `[`) {
					quoteStart = `[`
				} else if strings.HasSuffix(prevToken.value, `'`) {
					quoteStart = `'`
				} else if strings.HasSuffix(prevToken.value, `<`) {
					quoteStart = `<`
				}
			}
		}

		quoteEnd := ""
		switch quoteStart {
		case `"`:
			quoteEnd = `"`
		case `[`:
			quoteEnd = `]`
		case `'`:
			quoteEnd = `'`
		case `<`:
			quoteEnd = `>`
		}

		tokens = append(tokens, token{
			typ:        tokenField,
			value:      fieldName,
			quoteStart: quoteStart,
			quoteEnd:   quoteEnd,
		})
	} else if builder.Len() > 0 {
		tokens = append(tokens, token{
			typ:   tokenLiteral,
			value: builder.String(),
		})
	}

	tokens = postProcessTokens(tokens)

	return tokens, startsWithField, fields, nil
}

func postProcessTokens(tokens []token) []token {
	var result []token

	for i := 0; i < len(tokens); i++ {
		tok := tokens[i]

		if tok.typ == tokenField && tok.quoteStart != "" {
			if len(result) > 0 {
				lastIdx := len(result) - 1
				lastTok := result[lastIdx]
				if lastTok.typ == tokenLiteral && strings.HasSuffix(lastTok.value, tok.quoteStart) {
					newVal := lastTok.value[:len(lastTok.value)-len(tok.quoteStart)]
					if newVal == "" {
						result = result[:lastIdx]
					} else {
						result[lastIdx] = token{
							typ:   tokenLiteral,
							value: newVal,
						}
					}
				}
			}
		}

		if tok.typ == tokenField && tok.quoteEnd != "" {
			if i+1 < len(tokens) {
				nextTok := tokens[i+1]
				if nextTok.typ == tokenLiteral && strings.HasPrefix(nextTok.value, tok.quoteEnd) {
					newVal := nextTok.value[len(tok.quoteEnd):]
					tokens[i+1] = token{
						typ:   tokenLiteral,
						value: newVal,
					}
				}
			}
		}

		result = append(result, tok)
	}

	var finalResult []token
	for _, tok := range result {
		if tok.typ == tokenLiteral && tok.value == "" {
			continue
		}
		finalResult = append(finalResult, tok)
	}

	return finalResult
}

func isFieldChar(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_'
}

func (p *Parser) SetMultilineHandler(handler MultilineHandler) {
	p.multilineHandler = handler
}

func (p *Parser) ParseLine(line string, lineNum int) (*LogEntry, error) {
	fields, err := p.extractFields(line)
	if err != nil {
		return nil, &ParseError{LineNum: lineNum, Reason: err.Error()}
	}

	entry := &LogEntry{
		Fields:  fields,
		Line:    line,
		LineNum: lineNum,
	}

	if timeStr, ok := fields["time_local"]; ok {
		entry.Timestamp = parseTimestamp(timeStr)
	} else if timeStr, ok := fields["timestamp"]; ok {
		entry.Timestamp = parseTimestamp(timeStr)
	}

	return entry, nil
}

func (p *Parser) extractFields(line string) (map[string]string, error) {
	fields := make(map[string]string)
	remaining := line
	fieldIdx := 0

	fieldTokenCount := 0
	for _, tok := range p.tokens {
		if tok.typ == tokenField {
			fieldTokenCount++
		}
	}

	if fieldTokenCount == 0 {
		if len(p.fieldNames) == 1 {
			fields[p.fieldNames[0]] = line
			return fields, nil
		}
		return nil, fmt.Errorf("format mismatch: no fields defined")
	}

	for i, tok := range p.tokens {
		switch tok.typ {
		case tokenLiteral:
			if !strings.HasPrefix(remaining, tok.value) {
				return nil, fmt.Errorf("format mismatch: expected %q at position %d", tok.value, len(line)-len(remaining))
			}
			remaining = remaining[len(tok.value):]

		case tokenField:
			var fieldValue string
			var err error

			if tok.quoteStart != "" {
				if !strings.HasPrefix(remaining, tok.quoteStart) {
					return nil, fmt.Errorf("format mismatch: expected opening quote %q at position %d", tok.quoteStart, len(line)-len(remaining))
				}
				remaining = remaining[len(tok.quoteStart):]
			}

			if tok.quoteEnd != "" {
				fieldValue, remaining, err = p.extractQuotedField(remaining, tok.quoteEnd)
			} else {
				nextLiteral := p.findNextLiteral(i)
				if nextLiteral != "" {
					fieldValue, remaining, err = p.extractFieldToLiteral(remaining, nextLiteral)
				} else {
					fieldValue = remaining
					remaining = ""
					err = nil
				}
			}

			if err != nil {
				return nil, err
			}

			fieldName := tok.value
			if fieldIdx < len(p.fieldNames) {
				fieldName = p.fieldNames[fieldIdx]
			}
			fields[fieldName] = fieldValue
			fieldIdx++
		}
	}

	for _, name := range p.fieldNames {
		if _, ok := fields[name]; !ok {
			fields[name] = ""
		}
	}

	return fields, nil
}

func (p *Parser) findNextLiteral(currentIdx int) string {
	for i := currentIdx + 1; i < len(p.tokens); i++ {
		if p.tokens[i].typ == tokenLiteral {
			return p.tokens[i].value
		}
	}
	return ""
}

func (p *Parser) extractQuotedField(remaining, quoteEnd string) (string, string, error) {
	idx := strings.Index(remaining, quoteEnd)
	if idx == -1 {
		return "", "", fmt.Errorf("format mismatch: missing closing quote %q", quoteEnd)
	}
	return remaining[:idx], remaining[idx+len(quoteEnd):], nil
}

func (p *Parser) extractFieldToLiteral(remaining, nextLiteral string) (string, string, error) {
	idx := strings.Index(remaining, nextLiteral)
	if idx == -1 {
		return "", "", fmt.Errorf("format mismatch: expected delimiter %q not found", nextLiteral)
	}
	return remaining[:idx], remaining[idx:], nil
}

type ParseResult struct {
	Entry  *LogEntry
	Errors []ParseError
}

func (p *Parser) Parse(reader io.Reader) (<-chan ParseResult, error) {
	resultChan := make(chan ParseResult, 1000)

	go func() {
		defer close(resultChan)

		scanner := bufio.NewScanner(reader)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

		var currentEntry *LogEntry
		var errors []ParseError
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			rawLine := scanner.Bytes()
			line := sanitizeBytes(rawLine)

			if p.multilineHandler != nil && p.multilineHandler(line, currentEntry) {
				if currentEntry != nil {
					currentEntry.Line += "\n" + line
				}
				continue
			}

			if currentEntry != nil {
				resultChan <- ParseResult{Entry: currentEntry}
				currentEntry = nil
			}

			entry, err := p.ParseLine(line, lineNum)
			if err != nil {
				if parseErr, ok := err.(*ParseError); ok {
					errors = append(errors, *parseErr)
				}
				continue
			}

			currentEntry = entry
		}

		if currentEntry != nil {
			resultChan <- ParseResult{Entry: currentEntry}
		}

		if len(errors) > 0 {
			resultChan <- ParseResult{Errors: errors}
		}

		if err := scanner.Err(); err != nil {
			resultChan <- ParseResult{
				Errors: []ParseError{{LineNum: -1, Reason: err.Error()}},
			}
		}
	}()

	return resultChan, nil
}

func sanitizeBytes(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}

	var buf bytes.Buffer
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r == utf8.RuneError && size == 1 {
			buf.WriteRune('\uFFFD')
			data = data[1:]
		} else {
			buf.WriteRune(r)
			data = data[size:]
		}
	}
	return buf.String()
}

func NewParserFromFormatName(formatName string) (*Parser, error) {
	return NewParser(formatName, nil)
}

func ListBuiltInFormats() []string {
	return DefaultRegistry.List()
}

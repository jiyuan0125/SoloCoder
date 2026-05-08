package resp

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

func ParseInlineCommand(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}

	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)
	escaped := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if escaped {
			current.WriteByte(c)
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if !inQuote {
			if c == '"' || c == '\'' {
				inQuote = true
				quoteChar = c
				continue
			}
			if c == ' ' || c == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
				continue
			}
			current.WriteByte(c)
		} else {
			if c == quoteChar {
				inQuote = false
				continue
			}
			current.WriteByte(c)
		}
	}

	if inQuote {
		return nil, NewProtocolError(0, "unterminated quoted string")
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args, nil
}

func ParseInlineCommands(r io.Reader) ([][]string, error) {
	scanner := bufio.NewScanner(r)
	var commands [][]string

	for scanner.Scan() {
		line := scanner.Text()
		args, err := ParseInlineCommand(line)
		if err != nil {
			return nil, err
		}
		if len(args) > 0 {
			commands = append(commands, args)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return commands, nil
}

func DetectInlineCommand(data []byte) bool {
	if len(data) == 0 {
		return false
	}
	firstByte := data[0]
	return firstByte != '+' && firstByte != '-' && firstByte != ':' &&
		firstByte != '$' && firstByte != '*'
}

func ParseCommandsFromReader(r io.Reader) ([][]string, error) {
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if DetectInlineCommand(buf) {
		return ParseInlineCommands(bytes.NewReader(buf))
	}

	parser := NewParser(bytes.NewReader(buf))
	var commands [][]string

	for {
		v, err := parser.Parse()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if v.Type == TypeArray {
			args := make([]string, 0, len(v.Array))
			for _, elem := range v.Array {
				if elem.Type == TypeBulkString {
					args = append(args, elem.Str)
				} else if elem.Type == TypeSimpleString {
					args = append(args, elem.Str)
				} else {
					return nil, NewProtocolError(parser.Pos(), "unexpected element type in command array")
				}
			}
			commands = append(commands, args)
		}
	}

	return commands, nil
}

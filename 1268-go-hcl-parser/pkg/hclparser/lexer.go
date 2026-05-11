package hclparser

import (
	"fmt"
	"strconv"
	"unicode"
)

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 1,
	}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
	l.position = l.readPosition
	l.readPosition++
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekChars(n int) string {
	if l.readPosition+n > len(l.input) {
		return l.input[l.readPosition:]
	}
	return l.input[l.readPosition : l.readPosition+n]
}

func (l *Lexer) NextToken() (Token, error) {
	var tok Token
	l.skipWhitespaceAndComments()

	line := l.line
	column := l.column

	switch l.ch {
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenEqualEqual, "==", line, column, nil)
		} else {
			tok = NewToken(TokenEquals, "=", line, column, nil)
		}
	case ',':
		tok = NewToken(TokenComma, ",", line, column, nil)
	case '(':
		tok = NewToken(TokenLParen, "(", line, column, nil)
	case ')':
		tok = NewToken(TokenRParen, ")", line, column, nil)
	case '{':
		tok = NewToken(TokenLBrace, "{", line, column, nil)
	case '}':
		tok = NewToken(TokenRBrace, "}", line, column, nil)
	case '[':
		tok = NewToken(TokenLBracket, "[", line, column, nil)
	case ']':
		tok = NewToken(TokenRBracket, "]", line, column, nil)
	case '.':
		tok = NewToken(TokenDot, ".", line, column, nil)
	case '?':
		tok = NewToken(TokenQuestion, "?", line, column, nil)
	case ':':
		tok = NewToken(TokenColon, ":", line, column, nil)
	case '+':
		tok = NewToken(TokenPlus, "+", line, column, nil)
	case '-':
		tok = NewToken(TokenMinus, "-", line, column, nil)
	case '*':
		tok = NewToken(TokenStar, "*", line, column, nil)
	case '/':
		tok = NewToken(TokenSlash, "/", line, column, nil)
	case '%':
		tok = NewToken(TokenPercent, "%", line, column, nil)
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = NewToken(TokenDoubleAmp, "&&", line, column, nil)
		} else {
			tok = NewToken(TokenAmpersand, "&", line, column, nil)
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = NewToken(TokenDoublePipe, "||", line, column, nil)
		} else {
			tok = NewToken(TokenPipe, "|", line, column, nil)
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenNotEqual, "!=", line, column, nil)
		} else {
			tok = NewToken(TokenExclamation, "!", line, column, nil)
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenLessEqual, "<=", line, column, nil)
		} else if l.peekChar() == '<' {
			return l.readHeredoc(line, column)
		} else {
			tok = NewToken(TokenLess, "<", line, column, nil)
		}
	case '>':
		if l.peekChar() == '=' {
			l.readChar()
			tok = NewToken(TokenGreaterEqual, ">=", line, column, nil)
		} else {
			tok = NewToken(TokenGreater, ">", line, column, nil)
		}
	case '"':
		return l.readString(line, column)
	case 0:
		tok = NewToken(TokenEOF, "", line, column, nil)
	default:
		if isLetter(l.ch) {
			ident := l.readIdentifier()
			switch ident {
			case "true", "false":
				val, _ := strconv.ParseBool(ident)
				tok = NewToken(TokenBoolean, ident, line, column, val)
			case "null":
				tok = NewToken(TokenNull, ident, line, column, nil)
			default:
				tok = NewToken(TokenIdentifier, ident, line, column, nil)
			}
			return tok, nil
		} else if isDigit(l.ch) || (l.ch == '-' && isDigit(l.peekChar())) {
			num, literal, err := l.readNumber()
			if err != nil {
				return tok, err
			}
			return NewToken(TokenNumber, literal, line, column, num), nil
		} else {
			return tok, fmt.Errorf("unexpected character: %c at line %d, column %d", l.ch, line, column)
		}
	}

	l.readChar()
	return tok, nil
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		switch l.ch {
		case ' ', '\t', '\n', '\r':
			l.readChar()
		case '#':
			l.skipSingleLineComment()
		case '/':
			if l.peekChar() == '*' {
				l.skipMultiLineComment()
			} else {
				return
			}
		default:
			return
		}
	}
}

func (l *Lexer) skipSingleLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) skipMultiLineComment() {
	l.readChar()
	l.readChar()
	for {
		if l.ch == 0 {
			return
		}
		if l.ch == '*' && l.peekChar() == '/' {
			l.readChar()
			l.readChar()
			return
		}
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' || l.ch == '-' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() (interface{}, string, error) {
	position := l.position
	isFloat := false

	if l.ch == '-' {
		l.readChar()
	}

	for isDigit(l.ch) {
		l.readChar()
	}

	if l.ch == '.' && isDigit(l.peekChar()) {
		isFloat = true
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	if l.ch == 'e' || l.ch == 'E' {
		isFloat = true
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[position:l.position]

	if isFloat {
		val, err := strconv.ParseFloat(literal, 64)
		return val, literal, err
	}

	val, err := strconv.ParseInt(literal, 10, 64)
	if err != nil {
		return nil, literal, err
	}
	return val, literal, nil
}

func (l *Lexer) readString(line, column int) (Token, error) {
	l.readChar()
	position := l.position
	var result []byte

	for {
		if l.ch == 0 {
			return Token{}, fmt.Errorf("unterminated string at line %d, column %d", line, column)
		}
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case 'r':
				result = append(result, '\r')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case '/':
				result = append(result, '/')
			default:
				result = append(result, '\\', l.ch)
			}
		} else if l.ch == '"' {
			break
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}

	l.readChar()
	return NewToken(TokenString, l.input[position:l.position-1], line, column, string(result)), nil
}

func (l *Lexer) readHeredoc(line, column int) (Token, error) {
	l.readChar()
	l.readChar()

	stripIndent := false
	if l.ch == '-' {
		stripIndent = true
		l.readChar()
	}

	delimiterStart := l.position
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	delimiter := l.input[delimiterStart:l.position]

	if l.ch == 0 {
		return Token{}, fmt.Errorf("unterminated heredoc at line %d, column %d", line, column)
	}

	l.readChar()

	contentStart := l.position
	for {
		if l.ch == 0 {
			return Token{}, fmt.Errorf("unterminated heredoc at line %d, column %d", line, column)
		}

		if l.ch == '\n' {
			nextLineStart := l.readPosition
			hasLeadingWS := false
			for l.peekChar() == ' ' || l.peekChar() == '\t' {
				hasLeadingWS = true
				l.readChar()
			}

			indentEnd := l.readPosition
			if l.peekChars(len(delimiter)) == delimiter {
				contentEnd := nextLineStart

				if stripIndent && hasLeadingWS {
					indentSize := indentEnd - nextLineStart
					contentEnd = l.position - indentSize
					rawContent := l.input[contentStart:contentEnd]
					stripped := stripHeredocIndent(rawContent, indentSize)
					return NewToken(TokenHeredoc, stripped, line, column, map[string]interface{}{
						"delimiter":    delimiter,
						"stripIndent":  true,
						"hasInterp":    hasHeredocInterpolation(stripped),
						"raw":          rawContent,
					}), nil
				}

				rawContent := l.input[contentStart:contentEnd]
				return NewToken(TokenHeredoc, rawContent, line, column, map[string]interface{}{
					"delimiter":   delimiter,
					"stripIndent": false,
					"hasInterp":   hasHeredocInterpolation(rawContent),
					"raw":         rawContent,
				}), nil
			}
		}

		l.readChar()
	}
}

func stripHeredocIndent(content string, indentSize int) string {
	var result []byte
	lines := splitLines(content)

	for i, line := range lines {
		if i > 0 {
			result = append(result, '\n')
		}
		if len(line) >= indentSize {
			allWhitespace := true
			for j := 0; j < indentSize; j++ {
				if line[j] != ' ' && line[j] != '\t' {
					allWhitespace = false
					break
				}
			}
			if allWhitespace {
				result = append(result, []byte(line[indentSize:])...)
			} else {
				result = append(result, []byte(line)...)
			}
		} else {
			result = append(result, []byte(line)...)
		}
	}

	return string(result)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	lines = append(lines, s[start:])
	return lines
}

func hasHeredocInterpolation(content string) bool {
	for i := 0; i < len(content)-1; i++ {
		if content[i] == '$' && content[i+1] == '{' {
			return true
		}
	}
	return false
}

func isLetter(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || ch == '_'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

package lexer

type Lexer struct {
	input        string
	position     int
	readPosition int
	ch           byte
	line         int
	column       int
}

func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, column: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else if l.ch != 0 {
		l.column++
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespaceAndComments()

	startLine := l.line
	startCol := l.column

	switch l.ch {
	case '{':
		tok = newToken(TokenBraceL, l.ch, startLine, startCol)
	case '}':
		tok = newToken(TokenBraceR, l.ch, startLine, startCol)
	case '(':
		tok = newToken(TokenParenL, l.ch, startLine, startCol)
	case ')':
		tok = newToken(TokenParenR, l.ch, startLine, startCol)
	case '[':
		tok = newToken(TokenBracketL, l.ch, startLine, startCol)
	case ']':
		tok = newToken(TokenBracketR, l.ch, startLine, startCol)
	case ':':
		tok = newToken(TokenColon, l.ch, startLine, startCol)
	case ',':
		tok = newToken(TokenComma, l.ch, startLine, startCol)
	case '=':
		tok = newToken(TokenEquals, l.ch, startLine, startCol)
	case '!':
		tok = newToken(TokenBang, l.ch, startLine, startCol)
	case '$':
		tok = newToken(TokenDollar, l.ch, startLine, startCol)
	case '.':
		if l.peekChar() == '.' {
			l.readChar()
			if l.peekChar() == '.' {
				l.readChar()
				tok = Token{Type: TokenEllipsis, Literal: "...", Line: startLine, Column: startCol}
			} else {
				tok = Token{Type: TokenIllegal, Literal: "..", Line: startLine, Column: startCol}
			}
		} else {
			tok = newToken(TokenDot, l.ch, startLine, startCol)
		}
	case 0:
		tok.Literal = ""
		tok.Type = TokenEOF
		tok.Line = startLine
		tok.Column = startCol
	case '"':
		tok.Type = TokenString
		tok.Literal = l.readString()
		tok.Line = startLine
		tok.Column = startCol
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			tok.Line = startLine
			tok.Column = startCol
			return tok
		} else if isDigit(l.ch) || (l.ch == '-' && isDigit(l.peekChar())) {
			tok.Literal = l.readNumber()
			if containsDot(tok.Literal) || containsExp(tok.Literal) {
				tok.Type = TokenFloat
			} else {
				tok.Type = TokenInt
			}
			tok.Line = startLine
			tok.Column = startCol
			return tok
		} else {
			tok = newToken(TokenIllegal, l.ch, startLine, startCol)
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		switch l.ch {
		case ' ', '\t', '\n', '\r':
			l.readChar()
		case '#':
			l.skipComment()
		default:
			return
		}
	}
}

func (l *Lexer) skipComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '_' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readNumber() string {
	position := l.position
	if l.ch == '-' {
		l.readChar()
	}
	for isDigit(l.ch) {
		l.readChar()
	}
	if l.ch == '.' {
		l.readChar()
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	if l.ch == 'e' || l.ch == 'E' {
		l.readChar()
		if l.ch == '+' || l.ch == '-' {
			l.readChar()
		}
		for isDigit(l.ch) {
			l.readChar()
		}
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '\\' {
			l.readChar()
			continue
		}
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	return l.input[position:l.position]
}

func newToken(tokenType TokenType, ch byte, line, col int) Token {
	return Token{Type: tokenType, Literal: string(ch), Line: line, Column: col}
}

func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func containsDot(s string) bool {
	for _, ch := range s {
		if ch == '.' {
			return true
		}
	}
	return false
}

func containsExp(s string) bool {
	for _, ch := range s {
		if ch == 'e' || ch == 'E' {
			return true
		}
	}
	return false
}

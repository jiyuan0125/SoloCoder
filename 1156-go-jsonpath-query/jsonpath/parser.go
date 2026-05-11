package jsonpath

import (
	"strconv"
)

type Parser struct {
	lexer *Lexer
	cur   Token
}

func NewParser(input string) *Parser {
	return &Parser{lexer: NewLexer(input)}
}

func (p *Parser) advance() error {
	tok, err := p.lexer.NextToken()
	if err != nil {
		return err
	}
	p.cur = tok
	return nil
}

func (p *Parser) Parse() (*Path, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.Type != TokenRoot {
		return nil, ErrInvalidPath
	}
	if err := p.advance(); err != nil {
		return nil, err
	}

	segments := []Segment{}
	for p.cur.Type != TokenEOF {
		seg, err := p.parseSegment()
		if err != nil {
			return nil, err
		}
		segments = append(segments, seg)
	}
	return &Path{Segments: segments}, nil
}

func (p *Parser) parseSegment() (Segment, error) {
	switch p.cur.Type {
	case TokenDotDot:
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenIdentifier {
			name := p.cur.Literal
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &DotDotSegment{Name: name}, nil
		}
		if p.cur.Type == TokenWildcard {
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &DotDotSegment{}, nil
		}
		return nil, ErrUnexpectedToken
	case TokenDot:
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenIdentifier {
			name := p.cur.Literal
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &DotSegment{Name: name}, nil
		}
		if p.cur.Type == TokenWildcard {
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &BracketSegment{Selectors: []Selector{&WildcardSelector{}}}, nil
		}
		return nil, ErrUnexpectedToken
	case TokenBracketOpen:
		return p.parseBracketSegment()
	default:
		return nil, ErrUnexpectedToken
	}
}

func (p *Parser) parseBracketSegment() (Segment, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}

	selectors := []Selector{}
	for p.cur.Type != TokenBracketClose {
		if len(selectors) > 0 {
			if p.cur.Type != TokenComma {
				return nil, ErrUnexpectedToken
			}
			if err := p.advance(); err != nil {
				return nil, err
			}
		}

		sel, err := p.parseSelector()
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, sel)
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return &BracketSegment{Selectors: selectors}, nil
}

func (p *Parser) parseSelector() (Selector, error) {
	switch p.cur.Type {
	case TokenWildcard:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &WildcardSelector{}, nil
	case TokenString:
		name := p.cur.Literal
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &NameSelector{Name: name}, nil
	case TokenNumber:
		idx, err := strconv.Atoi(p.cur.Literal)
		if err != nil {
			return nil, err
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenColon {
			return p.parseSliceSelector(&idx)
		}
		return &IndexSelector{Index: idx}, nil
	case TokenColon:
		return p.parseSliceSelector(nil)
	case TokenQuestion:
		return p.parseFilterSelector()
	default:
		return nil, ErrUnexpectedToken
	}
}

func (p *Parser) parseSliceSelector(start *int) (Selector, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	sel := &SliceSelector{Start: start}

	if p.cur.Type == TokenNumber {
		end, err := strconv.Atoi(p.cur.Literal)
		if err != nil {
			return nil, err
		}
		sel.End = &end
		if err := p.advance(); err != nil {
			return nil, err
		}
	}
	if p.cur.Type == TokenColon {
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.cur.Type == TokenNumber {
			step, err := strconv.Atoi(p.cur.Literal)
			if err != nil {
				return nil, err
			}
			sel.Step = &step
			if err := p.advance(); err != nil {
				return nil, err
			}
		}
	}
	return sel, nil
}

func (p *Parser) parseFilterSelector() (Selector, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.Type != TokenParenOpen {
		return nil, ErrUnexpectedToken
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	expr, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}
	if p.cur.Type != TokenParenClose {
		return nil, ErrUnexpectedToken
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return &FilterSelector{Expr: expr}, nil
}

func (p *Parser) parseOrExpr() (Expression, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TokenOr {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: BinaryOr, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseAndExpr() (Expression, error) {
	left, err := p.parseCompareExpr()
	if err != nil {
		return nil, err
	}
	for p.cur.Type == TokenAnd {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parseCompareExpr()
		if err != nil {
			return nil, err
		}
		left = &BinaryExpr{Op: BinaryAnd, Left: left, Right: right}
	}
	return left, nil
}

func (p *Parser) parseCompareExpr() (Expression, error) {
	left, err := p.parsePrimaryExpr()
	if err != nil {
		return nil, err
	}
	opMap := map[TokenType]BinaryOp{
		TokenEqual:         BinaryEqual,
		TokenNotEqual:      BinaryNotEqual,
		TokenLess:          BinaryLess,
		TokenLessEqual:     BinaryLessEqual,
		TokenGreater:       BinaryGreater,
		TokenGreaterEqual:  BinaryGreaterEqual,
	}
	if op, ok := opMap[p.cur.Type]; ok {
		if err := p.advance(); err != nil {
			return nil, err
		}
		right, err := p.parsePrimaryExpr()
		if err != nil {
			return nil, err
		}
		return &BinaryExpr{Op: op, Left: left, Right: right}, nil
	}
	return left, nil
}

func (p *Parser) parsePrimaryExpr() (Expression, error) {
	switch p.cur.Type {
	case TokenString:
		val := p.cur.Literal
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Literal{Value: val}, nil
	case TokenBool:
		val := p.cur.Literal == "true"
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Literal{Value: val}, nil
	case TokenNumber:
		val, err := strconv.ParseFloat(p.cur.Literal, 64)
		if err != nil {
			return nil, err
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &Literal{Value: val}, nil
	case TokenAt, TokenRoot:
		return p.parsePathExpr()
	default:
		return nil, ErrUnexpectedToken
	}
}

func (p *Parser) parsePathExpr() (Expression, error) {
	rel := p.cur.Type == TokenAt
	if err := p.advance(); err != nil {
		return nil, err
	}
	steps := []PathStep{}
	for {
		if p.cur.Type == TokenDotDot {
			step, err := p.parseDotDotStep()
			if err != nil {
				return nil, err
			}
			steps = append(steps, step)
		} else if p.cur.Type == TokenDot {
			step, err := p.parseDotStep()
			if err != nil {
				return nil, err
			}
			steps = append(steps, step)
		} else if p.cur.Type == TokenBracketOpen {
			step, err := p.parseBracketStep()
			if err != nil {
				return nil, err
			}
			steps = append(steps, step)
		} else {
			break
		}
	}
	return &PathExpression{Rel: rel, Steps: steps}, nil
}

func (p *Parser) parseDotDotStep() (PathStep, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.Type == TokenIdentifier {
		name := p.cur.Literal
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DotDotStep{Name: name}, nil
	}
	if p.cur.Type == TokenWildcard {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DotDotStep{}, nil
	}
	return nil, ErrUnexpectedToken
}

func (p *Parser) parseDotStep() (PathStep, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	if p.cur.Type == TokenIdentifier {
		name := p.cur.Literal
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &DotStep{Name: name}, nil
	}
	if p.cur.Type == TokenWildcard {
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &BracketStep{Selectors: []BracketStepSelector{&WildcardStepSelector{}}}, nil
	}
	return nil, ErrUnexpectedToken
}

func (p *Parser) parseBracketStep() (PathStep, error) {
	if err := p.advance(); err != nil {
		return nil, err
	}
	selectors := []BracketStepSelector{}
	for p.cur.Type != TokenBracketClose {
		if len(selectors) > 0 {
			if p.cur.Type != TokenComma {
				return nil, ErrUnexpectedToken
			}
			if err := p.advance(); err != nil {
				return nil, err
			}
		}
		sel, err := p.parseBracketStepSelector()
		if err != nil {
			return nil, err
		}
		selectors = append(selectors, sel)
	}
	if err := p.advance(); err != nil {
		return nil, err
	}
	return &BracketStep{Selectors: selectors}, nil
}

func (p *Parser) parseBracketStepSelector() (BracketStepSelector, error) {
	switch p.cur.Type {
	case TokenWildcard:
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &WildcardStepSelector{}, nil
	case TokenString:
		name := p.cur.Literal
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &NameStepSelector{Name: name}, nil
	case TokenNumber:
		idx, err := strconv.Atoi(p.cur.Literal)
		if err != nil {
			return nil, err
		}
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &IndexStepSelector{Index: idx}, nil
	default:
		return nil, ErrUnexpectedToken
	}
}

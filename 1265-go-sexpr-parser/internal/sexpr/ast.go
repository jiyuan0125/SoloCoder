package sexpr

import (
	"fmt"
	"strings"
)

type Position struct {
	Line   int
	Column int
}

func (p Position) String() string {
	return fmt.Sprintf("line %d, column %d", p.Line, p.Column)
}

type Expr interface {
	Pos() Position
	IsNil() bool
	String() string
}

type AtomType int

const (
	AtomInt AtomType = iota
	AtomFloat
	AtomString
	AtomSymbol
)

type Atom struct {
	Type AtomType
	IntValue    int64
	FloatValue  float64
	StringValue string
	SymbolValue string
	pos    Position
}

func (a *Atom) Pos() Position { return a.pos }
func (a *Atom) IsNil() bool   { return false }

func (a *Atom) String() string {
	switch a.Type {
	case AtomInt:
		return fmt.Sprintf("%d", a.IntValue)
	case AtomFloat:
		if a.FloatValue == float64(int64(a.FloatValue)) {
			return fmt.Sprintf("%.1f", a.FloatValue)
		}
		return fmt.Sprintf("%g", a.FloatValue)
	case AtomString:
		return fmt.Sprintf("%q", a.StringValue)
	case AtomSymbol:
		return a.SymbolValue
	default:
		return "unknown"
	}
}

type List struct {
	Elements []Expr
	pos     Position
}

func (l *List) Pos() Position { return l.pos }
func (l *List) IsNil() bool   { return len(l.Elements) == 0 }

func (l *List) String() string {
	if l.IsNil() {
		return "()"
	}
	parts := make([]string, len(l.Elements))
	for i, e := range l.Elements {
		parts[i] = e.String()
	}
	return "(" + strings.Join(parts, " ") + ")"
}

func NewInt(v int64, pos Position) *Atom {
	return &Atom{Type: AtomInt, IntValue: v, pos: pos}
}

func NewFloat(v float64, pos Position) *Atom {
	return &Atom{Type: AtomFloat, FloatValue: v, pos: pos}
}

func NewString(v string, pos Position) *Atom {
	return &Atom{Type: AtomString, StringValue: v, pos: pos}
}

func NewSymbol(v string, pos Position) *Atom {
	return &Atom{Type: AtomSymbol, SymbolValue: v, pos: pos}
}

func NewList(elements []Expr, pos Position) *List {
	return &List{Elements: elements, pos: pos}
}

func NewNil(pos Position) *List {
	return &List{Elements: []Expr{}, pos: pos}
}

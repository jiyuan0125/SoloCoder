package jsonpath

type Node interface{}

type Path struct {
	Segments []Segment
}

type Segment interface{}

type DotSegment struct {
	Name string
}

type DotDotSegment struct {
	Name string
}

type BracketSegment struct {
	Selectors []Selector
}

type Selector interface{}

type NameSelector struct {
	Name string
}

type IndexSelector struct {
	Index int
}

type SliceSelector struct {
	Start *int
	End   *int
	Step  *int
}

type WildcardSelector struct{}

type FilterSelector struct {
	Expr Expression
}

type Expression interface {
	Eval(current interface{}, root interface{}) (interface{}, error)
}

type PathExpression struct {
	Rel   bool
	Steps []PathStep
}

type PathStep interface{}

type DotStep struct {
	Name string
}

type DotDotStep struct {
	Name string
}

type BracketStep struct {
	Selectors []BracketStepSelector
}

type BracketStepSelector interface{}

type NameStepSelector struct {
	Name string
}

type IndexStepSelector struct {
	Index int
}

type WildcardStepSelector struct{}

type Literal struct {
	Value interface{}
}

type BinaryExpr struct {
	Op    BinaryOp
	Left  Expression
	Right Expression
}

type BinaryOp int

const (
	BinaryEqual BinaryOp = iota
	BinaryNotEqual
	BinaryLess
	BinaryLessEqual
	BinaryGreater
	BinaryGreaterEqual
	BinaryAnd
	BinaryOr
)

package template

type Node interface {
	Pos() Position
}

type Position struct {
	Line   int
	Column int
	Source string
}

type RootNode struct {
	Children []Node
}

func (r *RootNode) Pos() Position {
	if len(r.Children) > 0 {
		return r.Children[0].Pos()
	}
	return Position{Line: 1, Column: 1}
}

type TextNode struct {
	Text     string
	position Position
}

func (t *TextNode) Pos() Position {
	return t.position
}

type VariableNode struct {
	Path     []string
	position Position
}

func (v *VariableNode) Pos() Position {
	return v.position
}

type FunctionNode struct {
	Name     string
	Args     []Node
	position Position
}

func (f *FunctionNode) Pos() Position {
	return f.position
}

type IfNode struct {
	Condition Node
	Body      []Node
	Else      []Node
	position  Position
}

func (i *IfNode) Pos() Position {
	return i.position
}

type RangeNode struct {
	Variable Node
	Body     []Node
	position Position
}

func (r *RangeNode) Pos() Position {
	return r.position
}

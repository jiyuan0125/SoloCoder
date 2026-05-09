package ast

import "fmt"

type OperationType string

const (
	OperationTypeQuery    OperationType = "query"
	OperationTypeMutation OperationType = "mutation"
)

type Value interface{}

type Node interface {
	Pos() Position
	String() string
}

type Position struct {
	Line   int
	Column int
}

func (p Position) Pos() Position {
	return p
}

type Document struct {
	Position
	Definitions []Definition
}

type Definition interface {
	Node
	definitionNode()
}

type OperationDefinition struct {
	Position
	OperationType OperationType
	Name          string
	Variables     []VariableDefinition
	SelectionSet  SelectionSet
}

func (d *OperationDefinition) definitionNode() {}

func (d *OperationDefinition) String() string {
	return fmt.Sprintf("OperationDefinition{Op:%s, Name:%s}", d.OperationType, d.Name)
}

type FragmentDefinition struct {
	Position
	Name          string
	TypeCondition string
	SelectionSet  SelectionSet
}

func (d *FragmentDefinition) definitionNode() {}

func (d *FragmentDefinition) String() string {
	return fmt.Sprintf("FragmentDefinition{Name:%s, On:%s}", d.Name, d.TypeCondition)
}

type VariableDefinition struct {
	Position
	Variable     string
	Type         Type
	DefaultValue Value
}

type SelectionSet []Selection

type Selection interface {
	Node
	selectionNode()
}

type Field struct {
	Position
	Alias        string
	Name         string
	Arguments    []Argument
	Directives   []Directive
	SelectionSet SelectionSet
}

func (f *Field) selectionNode() {}

func (f *Field) String() string {
	return fmt.Sprintf("Field{Alias:%q, Name:%q}", f.Alias, f.Name)
}

type FragmentSpread struct {
	Position
	Name       string
	Directives []Directive
}

func (s *FragmentSpread) selectionNode() {}

func (s *FragmentSpread) String() string {
	return fmt.Sprintf("FragmentSpread{Name:%q}", s.Name)
}

type InlineFragment struct {
	Position
	TypeCondition string
	Directives    []Directive
	SelectionSet  SelectionSet
}

func (f *InlineFragment) selectionNode() {}

func (f *InlineFragment) String() string {
	return fmt.Sprintf("InlineFragment{On:%q}", f.TypeCondition)
}

type Argument struct {
	Position
	Name  string
	Value Value
}

type Directive struct {
	Position
	Name      string
	Arguments []Argument
}

type Type interface {
	Node
	typeNode()
	GetTypeName() string
	IsNonNull() bool
}

type NamedType struct {
	Position
	Name string
}

func (t *NamedType) typeNode()    {}
func (t *NamedType) Pos() Position { return t.Position }
func (t *NamedType) String() string {
	return t.Name
}
func (t *NamedType) GetTypeName() string { return t.Name }
func (t *NamedType) IsNonNull() bool    { return false }

type ListType struct {
	Position
	OfType Type
}

func (t *ListType) typeNode()    {}
func (t *ListType) Pos() Position { return t.Position }
func (t *ListType) String() string {
	return fmt.Sprintf("[%s]", t.OfType)
}
func (t *ListType) GetTypeName() string { return t.OfType.GetTypeName() }
func (t *ListType) IsNonNull() bool    { return false }

type NonNullType struct {
	Position
	OfType Type
}

func (t *NonNullType) typeNode()    {}
func (t *NonNullType) Pos() Position { return t.Position }
func (t *NonNullType) String() string {
	return fmt.Sprintf("%s!", t.OfType)
}
func (t *NonNullType) GetTypeName() string { return t.OfType.GetTypeName() }
func (t *NonNullType) IsNonNull() bool    { return true }

type Variable struct {
	Name string
}

func (d *Document) String() string {
	return fmt.Sprintf("Document{Definitions:%d}", len(d.Definitions))
}

func (d *VariableDefinition) Pos() Position  { return d.Position }
func (a *Argument) Pos() Position            { return a.Position }
func (d *Directive) Pos() Position           { return d.Position }
func (v *Variable) Pos() Position            { return Position{} }
func (s SelectionSet) Pos() Position         { return Position{} }

func (d *VariableDefinition) String() string {
	return fmt.Sprintf("VarDef{%s: %s}", d.Variable, d.Type)
}
func (a *Argument) String() string {
	return fmt.Sprintf("Arg{%s: %v}", a.Name, a.Value)
}

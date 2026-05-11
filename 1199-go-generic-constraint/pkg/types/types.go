package types

type Kind int

const (
	KindBasic Kind = iota
	KindNamed
	KindPointer
	KindSlice
	KindArray
	KindMap
	KindStruct
	KindInterface
	KindFunction
)

type Method struct {
	Name     string
	Params   []Type
	Results  []Type
}

type Type struct {
	Kind         Kind
	Name         string
	Underlying   *Type
	Methods      []Method
	Fields       map[string]Type
	ElementType  *Type
	KeyType      *Type
}

func (t *Type) UnderlyingType() *Type {
	if t.Kind != KindNamed {
		return t
	}
	if t.Underlying == nil {
		return t
	}
	return t.Underlying.UnderlyingType()
}

type TypeParam struct {
	Name      string
	Constraint Constraint
}

type Constraint struct {
	Kind       ConstraintKind
	Types      []Type
	UnionTerms []Constraint
	Methods    []Method
	Approximate bool
}

type ConstraintKind int

const (
	ConstraintAny ConstraintKind = iota
	ConstraintTypes
	ConstraintUnion
	ConstraintInterface
)

type FunctionSignature struct {
	Name        string
	TypeParams  []TypeParam
	Params      []Type
	Results     []Type
}

type CallSite struct {
	ID           string
	SourceFile   string
	Line         int
	ExplicitTypeArgs map[string]Type
	Args         []Type
}

type CheckResult struct {
	CallSiteID      string
	Status          CheckStatus
	InferredTypes   map[string]Type
	Errors          []CheckError
}

type CheckStatus string

const (
	StatusPassed         CheckStatus = "passed"
	StatusConstraintViolation CheckStatus = "constraint_violation"
	StatusInferenceFailed CheckStatus = "inference_failed"
)

type CheckError struct {
	TypeParamName string
	Constraint    string
	Message       string
}

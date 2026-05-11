package mockgen

import (
	"go/ast"
	"sort"
)

type Parameter struct {
	Name string
	Type string
}

type Method struct {
	Name       string
	Params     []Parameter
	Results    []Parameter
	IsVariadic bool
}

type InterfaceInfo struct {
	Package string
	Name    string
	Methods []Method
}

type CallRecord struct {
	MethodName string
	Params     []interface{}
	Results    []interface{}
}

type MethodMatcher func([]interface{}) bool

func (m Method) SignatureHash() string {
	sig := m.Name + "("
	for i, p := range m.Params {
		if i > 0 {
			sig += ","
		}
		sig += p.Type
	}
	sig += ")("
	for i, r := range m.Results {
		if i > 0 {
			sig += ","
		}
		sig += r.Type
	}
	sig += ")"
	return sig
}

func (m Method) HasResults() bool {
	return len(m.Results) > 0
}

func (i *InterfaceInfo) SortMethods() {
	sort.Slice(i.Methods, func(a, b int) bool {
		return i.Methods[a].Name < i.Methods[b].Name
	})
}

type MethodConflictError struct {
	Name      string
	Signature string
	Existing  string
}

func (e *MethodConflictError) Error() string {
	if e.Signature != e.Existing {
		return "method " + e.Name + " has conflicting signatures: " + e.Signature + " vs " + e.Existing
	}
	return "method " + e.Name + " is defined multiple times with the same signature"
}

type ParseConfig struct {
	IncludeUnexported bool
}

func NewParseConfig() *ParseConfig {
	return &ParseConfig{
		IncludeUnexported: true,
	}
}

type GeneratorConfig struct {
	Package string
}

func NewGeneratorConfig() *GeneratorConfig {
	return &GeneratorConfig{}
}

type typeInfo struct {
	expr ast.Expr
}

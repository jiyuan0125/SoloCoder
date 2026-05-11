package parser

import (
	"encoding/json"
	"strings"
)

type Location struct {
	Filename string `json:"filename"`
	Line     int    `json:"line"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Field struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Comment  string  `json:"comment"`
	Tags     []Tag   `json:"tags"`
	Exported bool    `json:"exported"`
	Location Location `json:"location"`
}

type Parameter struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	IsVariadic bool   `json:"is_variadic"`
	Comment    string `json:"comment"`
}

type Method struct {
	Name        string      `json:"name"`
	Receiver    *Parameter  `json:"receiver"`
	Parameters  []Parameter `json:"parameters"`
	ReturnTypes []string    `json:"return_types"`
	Comment     string      `json:"comment"`
	Summary     string      `json:"summary"`
	Examples    []Example   `json:"examples"`
	Exported    bool        `json:"exported"`
	Location    Location    `json:"location"`
}

type StructInfo struct {
	Fields []Field `json:"fields"`
}

type InterfaceInfo struct {
	Methods []Method `json:"methods"`
}

type AliasInfo struct {
	OriginalType string `json:"original_type"`
}

type TypeKind string

const (
	TypeKindStruct    TypeKind = "struct"
	TypeKindInterface TypeKind = "interface"
	TypeKindAlias     TypeKind = "alias"
	TypeKindOther     TypeKind = "other"
)

type Type struct {
	Name         string       `json:"name"`
	Kind         TypeKind     `json:"kind"`
	Comment      string       `json:"comment"`
	Summary      string       `json:"summary"`
	Examples     []Example    `json:"examples"`
	Exported     bool         `json:"exported"`
	Location     Location     `json:"location"`
	StructInfo   *StructInfo  `json:"struct_info,omitempty"`
	InterfaceInfo *InterfaceInfo `json:"interface_info,omitempty"`
	AliasInfo    *AliasInfo   `json:"alias_info,omitempty"`
}

type Function struct {
	Name        string      `json:"name"`
	Parameters  []Parameter `json:"parameters"`
	ReturnTypes []string    `json:"return_types"`
	Comment     string      `json:"comment"`
	Summary     string      `json:"summary"`
	Examples    []Example   `json:"examples"`
	Exported    bool        `json:"exported"`
	Location    Location    `json:"location"`
}

type Const struct {
	Name     string   `json:"name"`
	Value    string   `json:"value"`
	Comment  string   `json:"comment"`
	Summary  string   `json:"summary"`
	Exported bool     `json:"exported"`
	Location Location `json:"location"`
}

type Variable struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Value    string   `json:"value"`
	Comment  string   `json:"comment"`
	Summary  string   `json:"summary"`
	Exported bool     `json:"exported"`
	Location Location `json:"location"`
}

type Import struct {
	Path string `json:"path"`
	Name string `json:"name,omitempty"`
}

type Example struct {
	Title string `json:"title"`
	Code  string `json:"code"`
}

type PackageDoc struct {
	Name        string     `json:"name"`
	Comment     string     `json:"comment"`
	Summary     string     `json:"summary"`
	Examples    []Example  `json:"examples"`
	Imports     []Import   `json:"imports"`
	Constants   []Const    `json:"constants"`
	Variables   []Variable `json:"variables"`
	Types       []Type     `json:"types"`
	Functions   []Function `json:"functions"`
	Methods     []Method   `json:"methods"`
	Files       []string   `json:"files"`
}

func (p *PackageDoc) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

func (p *PackageDoc) AllDeclarations() []Searchable {
	var result []Searchable

	for _, c := range p.Constants {
		result = append(result, Searchable{
			Kind:     "const",
			Name:     c.Name,
			Comment:  c.Comment,
			Summary:  c.Summary,
			Exported: c.Exported,
		})
	}

	for _, v := range p.Variables {
		result = append(result, Searchable{
			Kind:     "var",
			Name:     v.Name,
			Comment:  v.Comment,
			Summary:  v.Summary,
			Exported: v.Exported,
		})
	}

	for _, t := range p.Types {
		result = append(result, Searchable{
			Kind:     "type",
			Name:     t.Name,
			Comment:  t.Comment,
			Summary:  t.Summary,
			Exported: t.Exported,
		})
	}

	for _, f := range p.Functions {
		result = append(result, Searchable{
			Kind:     "func",
			Name:     f.Name,
			Comment:  f.Comment,
			Summary:  f.Summary,
			Exported: f.Exported,
		})
	}

	for _, m := range p.Methods {
		result = append(result, Searchable{
			Kind:     "method",
			Name:     m.Name,
			Comment:  m.Comment,
			Summary:  m.Summary,
			Exported: m.Exported,
		})
	}

	return result
}

type Searchable struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Comment  string `json:"comment"`
	Summary  string `json:"summary"`
	Exported bool   `json:"exported"`
}

func (s *Searchable) Matches(query string) bool {
	q := strings.ToLower(query)
	name := strings.ToLower(s.Name)
	comment := strings.ToLower(s.Comment)
	summary := strings.ToLower(s.Summary)

	return strings.Contains(name, q) || strings.Contains(comment, q) || strings.Contains(summary, q)
}

type ParseOptions struct {
	IncludeUnexported bool
}

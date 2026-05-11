package flatten

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

type StructInfo struct {
	Name   string
	Fields []FieldDef
}

type FieldDef struct {
	Name      string
	Type      TypeInfo
	JSONTag   *JSONTagInfo
	Embedded  bool
	Exported  bool
}

type JSONTagInfo struct {
	Name      string
	Omitempty bool
	Ignored   bool
}

type TypeInfo struct {
	Kind       TypeKind
	Name       string
	Elem       *TypeInfo
	Key        *TypeInfo
	IsPointer  bool
	Package    string
}

type TypeKind int

const (
	KindBasic TypeKind = iota
	KindStruct
	KindSlice
	KindArray
	KindMap
	KindInterface
)

func ParseSource(source string) (map[string]*StructInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "input.go", source, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}

	structs := make(map[string]*StructInfo)

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			info := &StructInfo{
				Name: typeSpec.Name.Name,
			}

			for _, field := range structType.Fields.List {
				fieldDefs := parseField(field)
				info.Fields = append(info.Fields, fieldDefs...)
			}

			structs[info.Name] = info
		}
	}

	return structs, nil
}

func parseField(field *ast.Field) []FieldDef {
	var fields []FieldDef

	typeInfo := parseType(field.Type)

	var names []string
	var embedded bool

	if len(field.Names) > 0 {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	} else {
		embedded = true
		names = append(names, typeNameToLower(typeInfo))
	}

	jsonTag := parseJSONTag(field.Tag)

	for _, name := range names {
		fields = append(fields, FieldDef{
			Name:     name,
			Type:     typeInfo,
			JSONTag:  jsonTag,
			Embedded: embedded,
			Exported: token.IsExported(name),
		})
	}

	return fields
}

func typeNameToLower(t TypeInfo) string {
	if t.IsPointer {
		if t.Elem != nil {
			return typeNameToLower(*t.Elem)
		}
	}
	switch t.Kind {
	case KindStruct:
		return strings.ToLower(t.Name)
	}
	return strings.ToLower(t.Name)
}

func parseType(expr ast.Expr) TypeInfo {
	switch t := expr.(type) {
	case *ast.Ident:
		if isBuiltinType(t.Name) {
			return TypeInfo{Kind: KindBasic, Name: t.Name}
		}
		return TypeInfo{Kind: KindStruct, Name: t.Name}
	case *ast.StarExpr:
		elem := parseType(t.X)
		return TypeInfo{
			Kind:      elem.Kind,
			Name:      elem.Name,
			Elem:      elem.Elem,
			Key:       elem.Key,
			IsPointer: true,
			Package:   elem.Package,
		}
	case *ast.ArrayType:
		elem := parseType(t.Elt)
		if t.Len == nil {
			return TypeInfo{Kind: KindSlice, Elem: &elem}
		}
		return TypeInfo{Kind: KindArray, Elem: &elem}
	case *ast.MapType:
		key := parseType(t.Key)
		elem := parseType(t.Value)
		return TypeInfo{Kind: KindMap, Key: &key, Elem: &elem}
	case *ast.InterfaceType:
		return TypeInfo{Kind: KindInterface, Name: "interface{}"}
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return TypeInfo{
				Kind:    KindStruct,
				Name:    t.Sel.Name,
				Package: ident.Name,
			}
		}
		return TypeInfo{Kind: KindStruct, Name: t.Sel.Name}
	default:
		return TypeInfo{Kind: KindBasic, Name: fmt.Sprintf("%T", t)}
	}
}

func isBuiltinType(name string) bool {
	builtins := map[string]bool{
		"bool": true,
		"string": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true,
		"float32": true, "float64": true,
		"complex64": true, "complex128": true,
		"byte": true, "rune": true,
	}
	return builtins[name]
}

func parseJSONTag(tag *ast.BasicLit) *JSONTagInfo {
	if tag == nil {
		return nil
	}

	tagValue := tag.Value
	if len(tagValue) < 2 || tagValue[0] != '`' || tagValue[len(tagValue)-1] != '`' {
		return nil
	}
	tagValue = tagValue[1 : len(tagValue)-1]

	jsonTag := ""
	parts := strings.Split(tagValue, " ")
	for _, p := range parts {
		if strings.HasPrefix(p, "json:") {
			jsonTag = strings.Trim(p[5:], `"`)
			break
		}
	}

	if jsonTag == "" {
		return nil
	}

	components := strings.Split(jsonTag, ",")
	if len(components) == 0 || components[0] == "-" {
		return &JSONTagInfo{Ignored: true}
	}

	info := &JSONTagInfo{Name: components[0]}
	for i := 1; i < len(components); i++ {
		if components[i] == "omitempty" {
			info.Omitempty = true
		}
	}

	return info
}

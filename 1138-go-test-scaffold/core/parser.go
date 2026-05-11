package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func ParseSource(sourceCode string, fileName string) (*ParsedSource, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, fileName, sourceCode, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	parsed := &ParsedSource{
		PackageName: file.Name.Name,
		Structs:     make(map[string]*StructInfo),
		Interfaces:  make(map[string]*InterfaceInfo),
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if ts, ok := s.Type.(*ast.StructType); ok {
						structInfo := parseStruct(s.Name.Name, ts)
						if structInfo != nil {
							parsed.Structs[structInfo.Name] = structInfo
						}
					}
					if is, ok := s.Type.(*ast.InterfaceType); ok {
						ifaceInfo := parseInterface(s.Name.Name, is)
						if ifaceInfo != nil {
							parsed.Interfaces[ifaceInfo.Name] = ifaceInfo
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Recv != nil && d.Recv.List != nil && len(d.Recv.List) > 0 {
				methodInfo := parseMethod(d)
				if methodInfo != nil {
					receiverType := getReceiverType(d.Recv.List[0])
					if receiverType != "" {
						if structInfo, exists := parsed.Structs[receiverType]; exists {
							structInfo.Methods = append(structInfo.Methods, methodInfo)
						}
					}
				}
			}
		}
	}

	for _, structInfo := range parsed.Structs {
		checkInterfaceImplementation(structInfo, parsed.Interfaces)
	}

	return parsed, nil
}

func parseStruct(name string, structType *ast.StructType) *StructInfo {
	if structType.Fields == nil {
		return nil
	}

	structInfo := &StructInfo{
		Name:   name,
		Fields: []*FieldInfo{},
	}

	for _, field := range structType.Fields.List {
		fieldInfos := parseField(field)
		structInfo.Fields = append(structInfo.Fields, fieldInfos...)
	}

	return structInfo
}

func parseField(field *ast.Field) []*FieldInfo {
	fieldInfos := []*FieldInfo{}
	fieldType := getTypeString(field.Type)
	isExported := ast.IsExported(fieldType)

	if len(field.Names) == 0 {
		jsonTag := getJSONTag(field.Tag)
		fieldInfos = append(fieldInfos, &FieldInfo{
			Name:       getTypeBaseName(fieldType),
			Type:       fieldType,
			IsEmbedded: true,
			JSONTag:    jsonTag,
			IsExported: isExported,
		})
	} else {
		jsonTag := getJSONTag(field.Tag)
		for _, name := range field.Names {
			fieldInfos = append(fieldInfos, &FieldInfo{
				Name:       name.Name,
				Type:       fieldType,
				IsEmbedded: false,
				JSONTag:    jsonTag,
				IsExported: ast.IsExported(name.Name),
			})
		}
	}

	return fieldInfos
}

func parseInterface(name string, interfaceType *ast.InterfaceType) *InterfaceInfo {
	if interfaceType.Methods == nil {
		return nil
	}

	ifaceInfo := &InterfaceInfo{
		Name:    name,
		Methods: []*MethodInfo{},
	}

	for _, method := range interfaceType.Methods.List {
		if len(method.Names) == 0 {
			continue
		}

		if ft, ok := method.Type.(*ast.FuncType); ok {
			ifaceInfo.Methods = append(ifaceInfo.Methods, &MethodInfo{
				Name:       method.Names[0].Name,
				Params:     getParamTypes(ft.Params),
				Results:    getResultTypes(ft.Results),
				IsExported: ast.IsExported(method.Names[0].Name),
			})
		}
	}

	return ifaceInfo
}

func parseMethod(funcDecl *ast.FuncDecl) *MethodInfo {
	if len(funcDecl.Name.Name) == 0 {
		return nil
	}

	return &MethodInfo{
		Name:       funcDecl.Name.Name,
		Params:     getParamTypes(funcDecl.Type.Params),
		Results:    getResultTypes(funcDecl.Type.Results),
		IsExported: ast.IsExported(funcDecl.Name.Name),
	}
}

func getReceiverType(field *ast.Field) string {
	if field.Type == nil {
		return ""
	}

	typeStr := getTypeString(field.Type)
	typeStr = strings.TrimPrefix(typeStr, "*")
	return typeStr
}

func getTypeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + getTypeString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + getTypeString(t.Elt)
		}
		return fmt.Sprintf("[%d]%s", t.Len.(*ast.BasicLit).Value, getTypeString(t.Elt))
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", getTypeString(t.Key), getTypeString(t.Value))
	case *ast.SelectorExpr:
		return fmt.Sprintf("%s.%s", getTypeString(t.X), t.Sel.Name)
	default:
		return fmt.Sprintf("%T", expr)
	}
}

func getTypeBaseName(typeStr string) string {
	typeStr = strings.TrimPrefix(typeStr, "*")
	parts := strings.Split(typeStr, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return typeStr
}

func getJSONTag(tag *ast.BasicLit) string {
	if tag == nil {
		return ""
	}
	tagValue := strings.Trim(tag.Value, "`")
	for _, part := range strings.Split(tagValue, " ") {
		if strings.HasPrefix(part, "json:") {
			jsonTag := strings.TrimPrefix(part, "json:")
			jsonTag = strings.Trim(jsonTag, `"`)
			return strings.Split(jsonTag, ",")[0]
		}
	}
	return ""
}

func getParamTypes(params *ast.FieldList) []string {
	if params == nil {
		return []string{}
	}
	result := []string{}
	for _, param := range params.List {
		result = append(result, getTypeString(param.Type))
	}
	return result
}

func getResultTypes(results *ast.FieldList) []string {
	if results == nil {
		return []string{}
	}
	result := []string{}
	for _, res := range results.List {
		result = append(result, getTypeString(res.Type))
	}
	return result
}

func checkInterfaceImplementation(structInfo *StructInfo, interfaces map[string]*InterfaceInfo) {
	for name, iface := range interfaces {
		if implementsInterface(structInfo, iface) {
			structInfo.Interfaces = append(structInfo.Interfaces, name)
		}
	}
}

func implementsInterface(structInfo *StructInfo, iface *InterfaceInfo) bool {
	methodMap := make(map[string]*MethodInfo)
	for _, m := range structInfo.Methods {
		methodMap[m.Name] = m
	}

	for _, ifaceMethod := range iface.Methods {
		if !ifaceMethod.IsExported {
			continue
		}

		structMethod, exists := methodMap[ifaceMethod.Name]
		if !exists {
			return false
		}

		if !sameMethodSignature(structMethod, ifaceMethod) {
			return false
		}
	}

	return true
}

func sameMethodSignature(m1, m2 *MethodInfo) bool {
	if len(m1.Params) != len(m2.Params) {
		return false
	}
	if len(m1.Results) != len(m2.Results) {
		return false
	}

	for i := range m1.Params {
		if m1.Params[i] != m2.Params[i] {
			return false
		}
	}

	for i := range m1.Results {
		if m1.Results[i] != m2.Results[i] {
			return false
		}
	}

	return true
}

package wiregen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func ScanPackage(pkgPath string) (*PackageInfo, error) {
	absPath, err := filepath.Abs(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	fset := token.NewFileSet()

	pkgs, err := parser.ParseDir(fset, absPath, func(info os.FileInfo) bool {
		return !strings.HasSuffix(info.Name(), "_test.go") && !strings.HasSuffix(info.Name(), "_gen.go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory: %w", err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("no Go packages found in %s", absPath)
	}

	var pkgName string
	var files []*ast.File
	for name, pkg := range pkgs {
		pkgName = name
		for _, f := range pkg.Files {
			files = append(files, f)
		}
	}

	info := &PackageInfo{
		Name:      pkgName,
		Providers: make(map[string]*Provider),
		Types:     make(map[string]*TypeInfo),
	}

	for _, file := range files {
		ast.Inspect(file, func(n ast.Node) bool {
			if fn, ok := n.(*ast.FuncDecl); ok {
				if fn.Recv != nil {
					return true
				}
				if fn.Name.IsExported() && strings.HasPrefix(fn.Name.Name, "New") {
					provider := parseProvider(fn, file)
					if provider != nil {
						info.Providers[provider.ID] = provider
					}
				}
			}
			if ts, ok := n.(*ast.TypeSpec); ok {
				typeInfo := parseType(ts, file)
				if typeInfo != nil {
					info.Types[typeInfo.Name] = typeInfo
				}
			}
			return true
		})
	}

	return info, nil
}

func parseProvider(fn *ast.FuncDecl, file *ast.File) *Provider {
	if fn.Type.Results == nil || len(fn.Type.Results.List) == 0 {
		return nil
	}

	provider := &Provider{
		ID:   fn.Name.Name,
		Name: fn.Name.Name,
	}

	for i, param := range fn.Type.Params.List {
		typeName := getTypeName(param.Type)
		if typeName == "" {
			continue
		}
		paramName := ""
		if len(param.Names) > 0 {
			paramName = param.Names[0].Name
		} else {
			paramName = fmt.Sprintf("param%d", i)
		}
		provider.Params = append(provider.Params, ParamInfo{
			Name: paramName,
			Type: typeName,
		})
	}

	results := fn.Type.Results.List
	if len(results) == 0 {
		return nil
	}

	primaryResult := results[0]
	provider.ReturnType = getTypeName(primaryResult.Type)

	if len(results) > 1 {
		secondResult := results[1]
		if getTypeName(secondResult.Type) == "error" {
			provider.ReturnsError = true
		}
	}

	if provider.ReturnType == "" {
		return nil
	}

	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			if strings.Contains(comment.Text, "+wire:provider") {
				provider.IsExplicitProvider = true
			}
		}
	}

	return provider
}

func parseType(ts *ast.TypeSpec, file *ast.File) *TypeInfo {
	if !ts.Name.IsExported() {
		return nil
	}

	info := &TypeInfo{
		Name: ts.Name.Name,
	}

	switch t := ts.Type.(type) {
	case *ast.InterfaceType:
		info.Kind = TypeKindInterface
		for _, method := range t.Methods.List {
			if len(method.Names) > 0 {
				info.Methods = append(info.Methods, method.Names[0].Name)
			}
		}
	case *ast.StructType:
		info.Kind = TypeKindStruct
	}

	return info
}

func getTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		name := getTypeName(e.X)
		if name != "" {
			return "*" + name
		}
	case *ast.SelectorExpr:
		pkg := getTypeName(e.X)
		if pkg != "" {
			return pkg + "." + e.Sel.Name
		}
	case *ast.InterfaceType:
		return "interface{}"
	}
	return ""
}

func (p *PackageInfo) GetProvidersByReturnType(typeName string) []*Provider {
	var providers []*Provider
	for _, provider := range p.Providers {
		if p.typesMatch(provider.ReturnType, typeName) {
			providers = append(providers, provider)
		}
	}
	return providers
}

func (p *PackageInfo) typesMatch(provided, requested string) bool {
	if provided == requested {
		return true
	}

	if strings.HasPrefix(provided, "*") {
		if provided[1:] == requested {
			return true
		}
	}

	providedInfo, providedOk := p.Types[stripPointer(provided)]
	requestedInfo, requestedOk := p.Types[stripPointer(requested)]

	if requestedOk && requestedInfo.Kind == TypeKindInterface {
		if providedOk {
			if providedInfo.Kind == TypeKindStruct {
				return implementsInterface(providedInfo, requestedInfo)
			}
		}
	}

	return false
}

func stripPointer(name string) string {
	return strings.TrimPrefix(name, "*")
}

func implementsInterface(structInfo *TypeInfo, ifaceInfo *TypeInfo) bool {
	methodSet := make(map[string]bool)
	for _, m := range structInfo.Methods {
		methodSet[m] = true
	}
	for _, m := range ifaceInfo.Methods {
		if !methodSet[m] {
			return false
		}
	}
	return true
}

func IsBasicType(typeName string) bool {
	basicTypes := map[string]bool{
		"string": true,
		"int": true, "int8": true, "int16": true, "int32": true, "int64": true,
		"uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true,
		"float32": true, "float64": true,
		"bool": true,
		"byte": true,
		"rune": true,
	}
	return basicTypes[typeName]
}

func IsContextType(typeName string) bool {
	return typeName == "context.Context"
}

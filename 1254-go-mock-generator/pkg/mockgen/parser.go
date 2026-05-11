package mockgen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
)

type Parser struct {
	fset *token.FileSet
	conf *ParseConfig
}

func NewParser() *Parser {
	return NewParserWithConfig(NewParseConfig())
}

func NewParserWithConfig(conf *ParseConfig) *Parser {
	if conf == nil {
		conf = NewParseConfig()
	}
	return &Parser{
		fset: token.NewFileSet(),
		conf: conf,
	}
}

func (p *Parser) ParseFile(filename string, src []byte) (*ast.File, error) {
	return parser.ParseFile(p.fset, filename, src, parser.AllErrors|parser.ParseComments)
}

func (p *Parser) ParseString(src string) (*ast.File, error) {
	return p.ParseFile("mock.go", []byte(src))
}

func (p *Parser) ListInterfaces(src string) ([]string, error) {
	file, err := p.ParseString(src)
	if err != nil {
		return nil, err
	}

	var ifaces []string
	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, ok := ts.Type.(*ast.InterfaceType); ok {
				ifaces = append(ifaces, ts.Name.Name)
			}
		}
	}
	return ifaces, nil
}

func (p *Parser) ExtractInterface(src string, ifaceName string) (*InterfaceInfo, error) {
	file, err := p.ParseString(src)
	if err != nil {
		return nil, err
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != ifaceName {
				continue
			}
			it, ok := ts.Type.(*ast.InterfaceType)
			if !ok {
				return nil, fmt.Errorf("%s is not an interface", ifaceName)
			}

			info := &InterfaceInfo{
				Package: file.Name.Name,
				Name:    ifaceName,
			}

			methods, err := p.expandInterfaceMethods(src, it, map[string]string{})
			if err != nil {
				return nil, err
			}

			if err := p.checkMethodConflicts(methods); err != nil {
				return nil, err
			}

			info.Methods = methods
			info.SortMethods()
			return info, nil
		}
	}

	return nil, fmt.Errorf("interface %s not found", ifaceName)
}

func (p *Parser) expandInterfaceMethods(src string, it *ast.InterfaceType, seen map[string]string) ([]Method, error) {
	var methods []Method

	for _, field := range it.Methods.List {
		switch t := field.Type.(type) {
		case *ast.FuncType:
			if len(field.Names) == 0 {
				continue
			}
			methodName := field.Names[0].Name
			if !isExported(methodName) && !p.conf.IncludeUnexported {
				continue
			}

			method, err := p.parseMethod(src, methodName, t)
			if err != nil {
				return nil, err
			}

			if existingSig, exists := seen[methodName]; exists {
				if existingSig != method.SignatureHash() {
					return nil, &MethodConflictError{
						Name:      methodName,
						Signature: method.SignatureHash(),
						Existing:  existingSig,
					}
				}
				continue
			}
			seen[methodName] = method.SignatureHash()
			methods = append(methods, *method)

		case *ast.Ident:
			embedded, err := p.findEmbeddedInterface(src, t.Name)
			if err != nil {
				return nil, err
			}
			embeddedMethods, err := p.expandInterfaceMethods(src, embedded, seen)
			if err != nil {
				return nil, err
			}
			methods = append(methods, embeddedMethods...)

		case *ast.SelectorExpr:
			continue
		}
	}

	return methods, nil
}

func (p *Parser) parseMethod(src string, name string, ft *ast.FuncType) (*Method, error) {
	method := &Method{Name: name}

	if ft.Params != nil {
		for i, field := range ft.Params.List {
			typeStr, err := typeExprToString(src, field.Type)
			if err != nil {
				return nil, err
			}

			_, isVariadic := field.Type.(*ast.Ellipsis)
			if isVariadic {
				method.IsVariadic = true
			}

			if len(field.Names) == 0 {
				paramName := fmt.Sprintf("p%d", i)
				method.Params = append(method.Params, Parameter{Name: paramName, Type: typeStr})
			} else {
				for _, n := range field.Names {
					method.Params = append(method.Params, Parameter{Name: n.Name, Type: typeStr})
				}
			}
		}
	}

	if ft.Results != nil {
		for i, field := range ft.Results.List {
			typeStr, err := typeExprToString(src, field.Type)
			if err != nil {
				return nil, err
			}

			if len(field.Names) == 0 {
				resultName := fmt.Sprintf("r%d", i)
				method.Results = append(method.Results, Parameter{Name: resultName, Type: typeStr})
			} else {
				for _, n := range field.Names {
					method.Results = append(method.Results, Parameter{Name: n.Name, Type: typeStr})
				}
			}
		}
	}

	return method, nil
}

func (p *Parser) findEmbeddedInterface(src string, name string) (*ast.InterfaceType, error) {
	file, err := p.ParseString(src)
	if err != nil {
		return nil, err
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name.Name != name {
				continue
			}
			if it, ok := ts.Type.(*ast.InterfaceType); ok {
				return it, nil
			}
		}
	}

	return nil, fmt.Errorf("embedded interface %s not found", name)
}

func (p *Parser) checkMethodConflicts(methods []Method) error {
	byName := make(map[string][]Method)
	for _, m := range methods {
		byName[m.Name] = append(byName[m.Name], m)
	}

	for name, ms := range byName {
		if len(ms) <= 1 {
			continue
		}
		sig := ms[0].SignatureHash()
		for i := 1; i < len(ms); i++ {
			if ms[i].SignatureHash() != sig {
				return &MethodConflictError{
					Name:      name,
					Signature: ms[i].SignatureHash(),
					Existing:  sig,
				}
			}
		}
	}
	return nil
}

func typeExprToString(_ string, expr ast.Expr) (string, error) {
	var buf bytes.Buffer
	fset := token.NewFileSet()
	if err := printer.Fprint(&buf, fset, expr); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

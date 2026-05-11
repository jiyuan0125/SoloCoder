package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"iface-checker/pkg/models"
	"os"
	"path/filepath"
	"strings"
)

type ParseResult struct {
	Structs    []models.StructType
	Interfaces []models.InterfaceType
}

func ParseFile(filename string) (*ParseResult, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}
	
	var files []string
	if info.IsDir() {
		files, err = filepath.Glob(filepath.Join(filename, "*.go"))
		if err != nil {
			return nil, fmt.Errorf("failed to list go files: %w", err)
		}
	} else {
		files = []string{filename}
	}
	
	result := &ParseResult{}
	
	for _, f := range files {
		fileResult, err := parseSingleFile(f)
		if err != nil {
			return nil, err
		}
		result.Structs = append(result.Structs, fileResult.Structs...)
		result.Interfaces = append(result.Interfaces, fileResult.Interfaces...)
	}
	
	return result, nil
}

func parseSingleFile(filename string) (*ParseResult, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filename, err)
	}
	
	result := &ParseResult{}
	
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			
			switch t := typeSpec.Type.(type) {
			case *ast.StructType:
				strct := parseStruct(typeSpec.Name.Name, t, node)
				result.Structs = append(result.Structs, strct)
				
			case *ast.InterfaceType:
				iface := parseInterface(typeSpec.Name.Name, t)
				result.Interfaces = append(result.Interfaces, iface)
			}
		}
	}
	
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}
		
		recv := funcDecl.Recv.List[0]
		recvTypeName, isPointer := getReceiverTypeName(recv.Type)
		if recvTypeName == "" {
			continue
		}
		
		for i := range result.Structs {
			if result.Structs[i].Name == recvTypeName {
				method := parseMethod(funcDecl)
				method.IsPointerReceiver = isPointer
				result.Structs[i].Methods = append(result.Structs[i].Methods, method)
				break
			}
		}
	}
	
	return result, nil
}

func parseStruct(name string, structType *ast.StructType, file *ast.File) models.StructType {
	strct := models.StructType{
		Name: name,
	}
	
	if structType.Fields == nil {
		return strct
	}
	
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 {
			embedded := parseEmbeddedField(field.Type)
			if embedded != nil {
				strct.EmbeddedFields = append(strct.EmbeddedFields, *embedded)
			}
		}
	}
	
	return strct
}

func parseEmbeddedField(expr ast.Expr) *models.EmbeddedField {
	switch t := expr.(type) {
	case *ast.Ident:
		return &models.EmbeddedField{
			Type: t.Name,
		}
		
	case *ast.StarExpr:
		switch inner := t.X.(type) {
		case *ast.Ident:
			return &models.EmbeddedField{
				Type:      inner.Name,
				IsPointer: true,
			}
		case *ast.SelectorExpr:
			return &models.EmbeddedField{
				Type:      exprToString(inner),
				IsPointer: true,
			}
		}
		
	case *ast.SelectorExpr:
		return &models.EmbeddedField{
			Type: exprToString(t),
		}
	}
	
	return nil
}

func parseInterface(name string, interfaceType *ast.InterfaceType) models.InterfaceType {
	iface := models.InterfaceType{
		Name: name,
	}
	
	if interfaceType.Methods == nil {
		return iface
	}
	
	for _, field := range interfaceType.Methods.List {
		if len(field.Names) == 0 {
			continue
		}
		
		methodName := field.Names[0].Name
		funcType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		
		method := models.Method{
			Name: methodName,
		}
		
		if funcType.Params != nil {
			for _, param := range funcType.Params.List {
				paramType := exprToString(param.Type)
				isVariadic := false
				
				if _, ok := param.Type.(*ast.Ellipsis); ok {
					isVariadic = true
				}
				
				if len(param.Names) == 0 {
					method.Params = append(method.Params, models.MethodParam{
						Type:       paramType,
						IsVariadic: isVariadic,
					})
				} else {
					for _, pName := range param.Names {
						method.Params = append(method.Params, models.MethodParam{
							Name:       pName.Name,
							Type:       paramType,
							IsVariadic: isVariadic,
						})
					}
				}
			}
		}
		
		if funcType.Results != nil {
			for _, result := range funcType.Results.List {
				resultType := exprToString(result.Type)
				if len(result.Names) == 0 {
					method.Returns = append(method.Returns, resultType)
				} else {
					for range result.Names {
						method.Returns = append(method.Returns, resultType)
					}
				}
			}
		}
		
		iface.Methods = append(iface.Methods, method)
	}
	
	return iface
}

func parseMethod(funcDecl *ast.FuncDecl) models.Method {
	method := models.Method{
		Name: funcDecl.Name.Name,
	}
	
	funcType := funcDecl.Type
	
	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			paramType := exprToString(param.Type)
			isVariadic := false
			
			if _, ok := param.Type.(*ast.Ellipsis); ok {
				isVariadic = true
			}
			
			if len(param.Names) == 0 {
				method.Params = append(method.Params, models.MethodParam{
					Type:       paramType,
					IsVariadic: isVariadic,
				})
			} else {
				for _, pName := range param.Names {
					method.Params = append(method.Params, models.MethodParam{
						Name:       pName.Name,
						Type:       paramType,
						IsVariadic: isVariadic,
					})
				}
			}
		}
	}
	
	if funcType.Results != nil {
		for _, result := range funcType.Results.List {
			resultType := exprToString(result.Type)
			if len(result.Names) == 0 {
				method.Returns = append(method.Returns, resultType)
			} else {
				for range result.Names {
					method.Returns = append(method.Returns, resultType)
				}
			}
		}
	}
	
	return method
}

func getReceiverTypeName(expr ast.Expr) (string, bool) {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name, false
	case *ast.StarExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name, true
		}
	}
	return "", false
}

func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.FuncType:
		return funcTypeToString(t)
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.ChanType:
		return chanTypeToString(t)
	case *ast.ParenExpr:
		return "(" + exprToString(t.X) + ")"
	default:
		return fmt.Sprintf("%T", t)
	}
}

func funcTypeToString(t *ast.FuncType) string {
	var params []string
	if t.Params != nil {
		for _, p := range t.Params.List {
			params = append(params, exprToString(p.Type))
		}
	}
	
	var results []string
	if t.Results != nil {
		for _, r := range t.Results.List {
			results = append(results, exprToString(r.Type))
		}
	}
	
	paramStr := strings.Join(params, ", ")
	resultStr := strings.Join(results, ", ")
	
	if len(results) == 0 {
		return "func(" + paramStr + ")"
	}
	if len(results) == 1 {
		return "func(" + paramStr + ") " + resultStr
	}
	return "func(" + paramStr + ") (" + resultStr + ")"
}

func chanTypeToString(t *ast.ChanType) string {
	switch t.Dir {
	case ast.SEND:
		return "chan<- " + exprToString(t.Value)
	case ast.RECV:
		return "<-chan " + exprToString(t.Value)
	default:
		return "chan " + exprToString(t.Value)
	}
}

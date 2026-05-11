package deadcode

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"unicode"

	"deadcode-detector/internal/common"
)

type ParsedFile struct {
	Filename  string
	FileSet   *token.FileSet
	AstFile   *ast.File
	Content   string
	Lines     []string
	Package   string
}

type InterfaceMethod struct {
	Name       string
	Signature  string
	NumParams  int
	NumResults int
}

type InterfaceInfo struct {
	Name    string
	Methods []InterfaceMethod
}

func ParseFile(filename string, content string) (*ParsedFile, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")

	return &ParsedFile{
		Filename: filename,
		FileSet:  fset,
		AstFile:  node,
		Content:  content,
		Lines:    lines,
		Package:  node.Name.Name,
	}, nil
}

func ExtractDeclarations(file *ParsedFile) []common.Declaration {
	var decls []common.Declaration

	for _, decl := range file.AstFile.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			decls = append(decls, extractFunctionDeclaration(file, d))
		case *ast.GenDecl:
			switch d.Tok {
			case token.TYPE:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						decls = append(decls, extractTypeDeclaration(file, d, ts))
						decls = append(decls, extractStructFields(file, ts)...)
					}
				}
			case token.VAR:
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						decls = append(decls, extractVariableDeclarations(file, d, vs)...)
					}
				}
			case token.CONST:
				for _, spec := range d.Specs {
					if cs, ok := spec.(*ast.ValueSpec); ok {
						decls = append(decls, extractConstantDeclarations(file, d, cs)...)
					}
				}
			}
		}
	}

	return decls
}

func extractFunctionDeclaration(file *ParsedFile, decl *ast.FuncDecl) common.Declaration {
	pos := file.FileSet.Position(decl.Pos())
	name := decl.Name.Name
	isExported := isExported(name)

	snippet := getSnippet(file, decl.Pos(), decl.End())

	return common.Declaration{
		Name:       name,
		Type:       common.TypeFunction,
		Location:   common.Location{File: file.Filename, Line: pos.Line, Column: pos.Column},
		Snippet:    snippet,
		IsExported: isExported,
	}
}

func extractTypeDeclaration(file *ParsedFile, genDecl *ast.GenDecl, ts *ast.TypeSpec) common.Declaration {
	pos := file.FileSet.Position(ts.Pos())
	name := ts.Name.Name
	isExported := isExported(name)
	isAlias := ts.Assign != 0

	isInterface := false
	if _, ok := ts.Type.(*ast.InterfaceType); ok {
		isInterface = true
	}

	snippet := getSnippet(file, genDecl.Pos(), genDecl.End())

	return common.Declaration{
		Name:        name,
		Type:        common.TypeType,
		Location:    common.Location{File: file.Filename, Line: pos.Line, Column: pos.Column},
		Snippet:     snippet,
		IsExported:  isExported,
		IsAlias:     isAlias,
		IsInterface: isInterface,
	}
}

func extractStructFields(file *ParsedFile, ts *ast.TypeSpec) []common.Declaration {
	var fields []common.Declaration

	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		return fields
	}

	for _, fieldList := range st.Fields.List {
		for _, name := range fieldList.Names {
			if name == nil {
				continue
			}
			pos := file.FileSet.Position(name.Pos())
			snippet := getSnippet(file, fieldList.Pos(), fieldList.End())

			fields = append(fields, common.Declaration{
				Name:       name.Name,
				Type:       common.TypeField,
				Location:   common.Location{File: file.Filename, Line: pos.Line, Column: pos.Column},
				Snippet:    snippet,
				IsExported: isExported(name.Name),
			})
		}
	}

	return fields
}

func extractVariableDeclarations(file *ParsedFile, genDecl *ast.GenDecl, vs *ast.ValueSpec) []common.Declaration {
	var vars []common.Declaration

	for _, name := range vs.Names {
		if name == nil {
			continue
		}
		pos := file.FileSet.Position(name.Pos())
		snippet := getSnippet(file, genDecl.Pos(), genDecl.End())

		vars = append(vars, common.Declaration{
			Name:       name.Name,
			Type:       common.TypeVariable,
			Location:   common.Location{File: file.Filename, Line: pos.Line, Column: pos.Column},
			Snippet:    snippet,
			IsExported: isExported(name.Name),
		})
	}

	return vars
}

func extractConstantDeclarations(file *ParsedFile, genDecl *ast.GenDecl, cs *ast.ValueSpec) []common.Declaration {
	var consts []common.Declaration

	for _, name := range cs.Names {
		if name == nil {
			continue
		}
		pos := file.FileSet.Position(name.Pos())
		snippet := getSnippet(file, genDecl.Pos(), genDecl.End())

		consts = append(consts, common.Declaration{
			Name:       name.Name,
			Type:       common.TypeConstant,
			Location:   common.Location{File: file.Filename, Line: pos.Line, Column: pos.Column},
			Snippet:    snippet,
			IsExported: isExported(name.Name),
		})
	}

	return consts
}

func ExtractInterfaces(files []*ParsedFile) []InterfaceInfo {
	var interfaces []InterfaceInfo

	for _, file := range files {
		for _, decl := range file.AstFile.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						if it, ok := ts.Type.(*ast.InterfaceType); ok {
							interfaces = append(interfaces, extractInterfaceInfo(file, ts, it))
						}
					}
				}
			}
		}
	}

	return interfaces
}

func extractInterfaceInfo(file *ParsedFile, ts *ast.TypeSpec, it *ast.InterfaceType) InterfaceInfo {
	info := InterfaceInfo{
		Name:    ts.Name.Name,
		Methods: []InterfaceMethod{},
	}

	for _, method := range it.Methods.List {
		for _, name := range method.Names {
			if name == nil {
				continue
			}
			numParams := 0
			numResults := 0
			if ft, ok := method.Type.(*ast.FuncType); ok {
				numParams = countInterfaceParams(ft.Params)
				numResults = countInterfaceResults(ft.Results)
			}
			sig := formatMethodSignature(file, method)
			info.Methods = append(info.Methods, InterfaceMethod{
				Name:       name.Name,
				Signature:  sig,
				NumParams:  numParams,
				NumResults: numResults,
			})
		}
	}

	return info
}

func countInterfaceParams(params *ast.FieldList) int {
	if params == nil {
		return 0
	}
	count := 0
	for _, p := range params.List {
		if len(p.Names) > 0 {
			count += len(p.Names)
		} else {
			count += 1
		}
	}
	return count
}

func countInterfaceResults(results *ast.FieldList) int {
	if results == nil {
		return 0
	}
	count := 0
	for _, r := range results.List {
		if len(r.Names) > 0 {
			count += len(r.Names)
		} else {
			count += 1
		}
	}
	return count
}

func formatMethodSignature(file *ParsedFile, method *ast.Field) string {
	var buf bytes.Buffer
	start := file.FileSet.Position(method.Pos())
	end := file.FileSet.Position(method.End())

	if start.Line == end.Line && start.Line > 0 && start.Line <= len(file.Lines) {
		line := file.Lines[start.Line-1]
		if start.Column >= 1 && end.Column <= len(line)+1 {
			return line[start.Column-1 : end.Column-1]
		}
	}

	buf.WriteString("func")
	if ft, ok := method.Type.(*ast.FuncType); ok {
		if ft.Params != nil {
			buf.WriteString("(")
			for i, p := range ft.Params.List {
				if i > 0 {
					buf.WriteString(", ")
				}
				if len(p.Names) > 0 {
					buf.WriteString(p.Names[0].Name)
				}
			}
			buf.WriteString(")")
		}
		if ft.Results != nil {
			buf.WriteString("(")
			for i, r := range ft.Results.List {
				if i > 0 {
					buf.WriteString(", ")
				}
				if len(r.Names) > 0 {
					buf.WriteString(r.Names[0].Name)
				}
			}
			buf.WriteString(")")
		}
	}
	return buf.String()
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)[0]
	return unicode.IsUpper(r)
}

func getSnippet(file *ParsedFile, start, end token.Pos) string {
	startPos := file.FileSet.Position(start)
	endPos := file.FileSet.Position(end)

	if startPos.Line == endPos.Line && startPos.Line > 0 && startPos.Line <= len(file.Lines) {
		line := file.Lines[startPos.Line-1]
		return strings.TrimSpace(line)
	}

	var snippet strings.Builder
	for i := startPos.Line; i <= endPos.Line; i++ {
		if i > 0 && i <= len(file.Lines) {
			snippet.WriteString(file.Lines[i-1])
			if i < endPos.Line {
				snippet.WriteString("\n")
			}
		}
	}
	return snippet.String()
}

func IsMainPackage(files []*ParsedFile) bool {
	for _, file := range files {
		if file.Package == "main" {
			return true
		}
	}
	return false
}

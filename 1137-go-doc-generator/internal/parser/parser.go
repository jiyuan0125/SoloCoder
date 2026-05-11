package parser

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type Parser struct {
	fset *token.FileSet
	opts ParseOptions
}

func NewParser(opts ParseOptions) *Parser {
	return &Parser{
		fset: token.NewFileSet(),
		opts: opts,
	}
}

func (p *Parser) ParsePackage(pkgPath string) (*PackageDoc, error) {
	info, err := os.Stat(pkgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		fileDoc, err := p.ParseFile(pkgPath)
		if err != nil {
			return nil, err
		}
		return fileDoc, nil
	}

	files, err := filepath.Glob(filepath.Join(pkgPath, "*.go"))
	if err != nil {
		return nil, fmt.Errorf("failed to find go files: %w", err)
	}

	if len(files) == 0 {
		return nil, fmt.Errorf("no go files found in directory: %s", pkgPath)
	}

	var pkgDoc *PackageDoc
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			continue
		}
		if fileInfo.IsDir() {
			continue
		}

		base := filepath.Base(file)
		if strings.HasSuffix(base, "_test.go") {
			continue
		}

		doc, err := p.ParseFile(file)
		if err != nil {
			return nil, err
		}

		if pkgDoc == nil {
			pkgDoc = doc
		} else {
			mergePackageDocs(pkgDoc, doc)
		}
	}

	return pkgDoc, nil
}

func (p *Parser) ParseFile(filename string) (*PackageDoc, error) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	astFile, err := parser.ParseFile(p.fset, filename, file, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	pkgDoc := &PackageDoc{
		Name:  astFile.Name.Name,
		Files: []string{filename},
	}

	if astFile.Doc != nil {
		pkgDoc.Comment = astFile.Doc.Text()
		pkgDoc.Summary = extractSummary(pkgDoc.Comment)
		pkgDoc.Examples = extractExamples(pkgDoc.Comment)
	}

	for _, imp := range astFile.Imports {
		importDoc := p.parseImport(imp)
		pkgDoc.Imports = append(pkgDoc.Imports, importDoc)
	}

	for _, decl := range astFile.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			switch d.Tok {
			case token.CONST:
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						consts := p.parseConstSpec(vs, p.getDoc(vs, d), filename)
						for _, c := range consts {
							if p.shouldInclude(c.Exported) {
								pkgDoc.Constants = append(pkgDoc.Constants, c)
							}
						}
					}
				}
			case token.VAR:
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						vars := p.parseVarSpec(vs, p.getDoc(vs, d), filename)
						for _, v := range vars {
							if p.shouldInclude(v.Exported) {
								pkgDoc.Variables = append(pkgDoc.Variables, v)
							}
						}
					}
				}
			case token.TYPE:
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						typ := p.parseTypeSpec(ts, p.getDoc(ts, d), filename)
						if p.shouldInclude(typ.Exported) {
							pkgDoc.Types = append(pkgDoc.Types, typ)
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Recv == nil {
				fn := p.parseFuncDecl(d, filename)
				if p.shouldInclude(fn.Exported) {
					pkgDoc.Functions = append(pkgDoc.Functions, fn)
				}
			} else {
				method := p.parseMethodDecl(d, filename)
				if p.shouldInclude(method.Exported) {
					pkgDoc.Methods = append(pkgDoc.Methods, method)
				}
			}
		}
	}

	return pkgDoc, nil
}

func (p *Parser) shouldInclude(exported bool) bool {
	return exported || p.opts.IncludeUnexported
}

func (p *Parser) getDoc(spec ast.Spec, decl *ast.GenDecl) string {
	if vs, ok := spec.(*ast.ValueSpec); ok && vs.Doc != nil {
		return vs.Doc.Text()
	}
	if ts, ok := spec.(*ast.TypeSpec); ok && ts.Doc != nil {
		return ts.Doc.Text()
	}
	if decl.Doc != nil {
		return decl.Doc.Text()
	}
	return ""
}

func (p *Parser) parseImport(imp *ast.ImportSpec) Import {
	importDoc := Import{}

	if imp.Name != nil {
		importDoc.Name = imp.Name.Name
	}

	pathValue := imp.Path.Value
	if unquoted, err := strconv.Unquote(pathValue); err == nil {
		importDoc.Path = unquoted
	} else {
		importDoc.Path = pathValue
	}

	return importDoc
}

func (p *Parser) parseConstSpec(vs *ast.ValueSpec, comment string, filename string) []Const {
	var result []Const

	for i, name := range vs.Names {
		c := Const{
			Name:     name.Name,
			Comment:  comment,
			Summary:  extractSummary(comment),
			Exported: token.IsExported(name.Name),
			Location: Location{
				Filename: filename,
				Line:     p.fset.Position(name.Pos()).Line,
			},
		}

		if vs.Values != nil && i < len(vs.Values) {
			c.Value = p.nodeToString(vs.Values[i])
		} else if vs.Type != nil {
			c.Value = p.nodeToString(vs.Type)
		}

		result = append(result, c)
	}

	return result
}

func (p *Parser) parseVarSpec(vs *ast.ValueSpec, comment string, filename string) []Variable {
	var result []Variable

	for i, name := range vs.Names {
		v := Variable{
			Name:     name.Name,
			Comment:  comment,
			Summary:  extractSummary(comment),
			Exported: token.IsExported(name.Name),
			Location: Location{
				Filename: filename,
				Line:     p.fset.Position(name.Pos()).Line,
			},
		}

		if vs.Type != nil {
			v.Type = p.nodeToString(vs.Type)
		}

		if vs.Values != nil && i < len(vs.Values) {
			v.Value = p.nodeToString(vs.Values[i])
		}

		result = append(result, v)
	}

	return result
}

func (p *Parser) parseTypeSpec(ts *ast.TypeSpec, comment string, filename string) Type {
	typ := Type{
		Name:     ts.Name.Name,
		Comment:  comment,
		Summary:  extractSummary(comment),
		Examples: extractExamples(comment),
		Exported: token.IsExported(ts.Name.Name),
		Location: Location{
			Filename: filename,
			Line:     p.fset.Position(ts.Pos()).Line,
		},
	}

	switch t := ts.Type.(type) {
	case *ast.StructType:
		typ.Kind = TypeKindStruct
		typ.StructInfo = &StructInfo{
			Fields: p.parseStructFields(t, filename),
		}
	case *ast.InterfaceType:
		typ.Kind = TypeKindInterface
		typ.InterfaceInfo = &InterfaceInfo{
			Methods: p.parseInterfaceMethods(t, filename),
		}
	case *ast.Ident:
		typ.Kind = TypeKindAlias
		typ.AliasInfo = &AliasInfo{
			OriginalType: t.Name,
		}
	default:
		if ts.Assign.IsValid() {
			typ.Kind = TypeKindAlias
			typ.AliasInfo = &AliasInfo{
				OriginalType: p.nodeToString(ts.Type),
			}
		} else {
			typ.Kind = TypeKindOther
		}
	}

	return typ
}

func (p *Parser) parseStructFields(st *ast.StructType, filename string) []Field {
	var fields []Field

	if st.Fields == nil {
		return fields
	}

	for _, field := range st.Fields.List {
		fieldType := p.nodeToString(field.Type)

		if len(field.Names) == 0 {
			embeddedField := Field{
				Name:     fieldType,
				Type:     fieldType,
				Exported: token.IsExported(fieldType),
				Location: Location{
					Filename: filename,
					Line:     p.fset.Position(field.Pos()).Line,
				},
			}
			if field.Doc != nil {
				embeddedField.Comment = field.Doc.Text()
			}
			if field.Tag != nil {
				embeddedField.Tags = parseStructTag(field.Tag.Value)
			}
			if p.shouldInclude(embeddedField.Exported) {
				fields = append(fields, embeddedField)
			}
		} else {
			for _, name := range field.Names {
				f := Field{
					Name:     name.Name,
					Type:     fieldType,
					Exported: token.IsExported(name.Name),
					Location: Location{
						Filename: filename,
						Line:     p.fset.Position(name.Pos()).Line,
					},
				}
				if field.Doc != nil {
					f.Comment = field.Doc.Text()
				}
				if field.Tag != nil {
					f.Tags = parseStructTag(field.Tag.Value)
				}
				if p.shouldInclude(f.Exported) {
					fields = append(fields, f)
				}
			}
		}
	}

	return fields
}

func parseStructTag(tagValue string) []Tag {
	var tags []Tag

	unquoted, err := strconv.Unquote(tagValue)
	if err != nil {
		return tags
	}

	tag := reflect.StructTag(unquoted)

	for _, key := range []string{"json", "xml", "yaml", "db", "gorm", "form", "uri", "binding", "validate"} {
		if value, ok := tag.Lookup(key); ok {
			tags = append(tags, Tag{
				Key:   key,
				Value: value,
			})
		}
	}

	if len(tags) == 0 {
		for _, part := range strings.Fields(unquoted) {
			kv := strings.SplitN(part, ":", 2)
			if len(kv) == 2 {
				value := strings.Trim(kv[1], "\"")
				tags = append(tags, Tag{
					Key:   kv[0],
					Value: value,
				})
			}
		}
	}

	return tags
}

func (p *Parser) parseInterfaceMethods(it *ast.InterfaceType, filename string) []Method {
	var methods []Method

	if it.Methods == nil {
		return methods
	}

	for _, field := range it.Methods.List {
		if ft, ok := field.Type.(*ast.FuncType); ok {
			for _, name := range field.Names {
				method := Method{
					Name:     name.Name,
					Exported: token.IsExported(name.Name),
					Location: Location{
						Filename: filename,
						Line:     p.fset.Position(name.Pos()).Line,
					},
				}
				if field.Doc != nil {
					method.Comment = field.Doc.Text()
					method.Summary = extractSummary(method.Comment)
					method.Examples = extractExamples(method.Comment)
				}
				if ft.Params != nil {
					method.Parameters = p.parseParameters(ft.Params, filename)
				}
				if ft.Results != nil {
					method.ReturnTypes = p.parseResults(ft.Results)
				}
				if p.shouldInclude(method.Exported) {
					methods = append(methods, method)
				}
			}
		}
	}

	return methods
}

func (p *Parser) parseFuncDecl(fd *ast.FuncDecl, filename string) Function {
	fn := Function{
		Name:     fd.Name.Name,
		Comment:  extractComment(fd.Doc),
		Exported: token.IsExported(fd.Name.Name),
		Location: Location{
			Filename: filename,
			Line:     p.fset.Position(fd.Pos()).Line,
		},
	}
	fn.Summary = extractSummary(fn.Comment)
	fn.Examples = extractExamples(fn.Comment)

	if fd.Type.Params != nil {
		fn.Parameters = p.parseParameters(fd.Type.Params, filename)
	}
	if fd.Type.Results != nil {
		fn.ReturnTypes = p.parseResults(fd.Type.Results)
	}

	return fn
}

func (p *Parser) parseMethodDecl(fd *ast.FuncDecl, filename string) Method {
	method := Method{
		Name:     fd.Name.Name,
		Comment:  extractComment(fd.Doc),
		Exported: token.IsExported(fd.Name.Name),
		Location: Location{
			Filename: filename,
			Line:     p.fset.Position(fd.Pos()).Line,
		},
	}
	method.Summary = extractSummary(method.Comment)
	method.Examples = extractExamples(method.Comment)

	if fd.Recv != nil && len(fd.Recv.List) > 0 {
		recv := fd.Recv.List[0]
		recvType := p.nodeToString(recv.Type)
		if star, ok := recv.Type.(*ast.StarExpr); ok {
			recvType = "*" + p.nodeToString(star.X)
		}
		recvName := ""
		if len(recv.Names) > 0 {
			recvName = recv.Names[0].Name
		}
		method.Receiver = &Parameter{
			Name: recvName,
			Type: recvType,
		}
	}

	if fd.Type.Params != nil {
		method.Parameters = p.parseParameters(fd.Type.Params, filename)
	}
	if fd.Type.Results != nil {
		method.ReturnTypes = p.parseResults(fd.Type.Results)
	}

	return method
}

func (p *Parser) parseParameters(params *ast.FieldList, filename string) []Parameter {
	var result []Parameter

	for _, field := range params.List {
		paramType := p.nodeToString(field.Type)
		_, isVariadic := field.Type.(*ast.Ellipsis)

		if len(field.Names) == 0 {
			result = append(result, Parameter{
				Type:       paramType,
				IsVariadic: isVariadic,
			})
		} else {
			for _, name := range field.Names {
				param := Parameter{
					Name:       name.Name,
					Type:       paramType,
					IsVariadic: isVariadic,
				}
				if field.Comment != nil {
					param.Comment = field.Comment.Text()
				}
				result = append(result, param)
			}
		}
	}

	return result
}

func (p *Parser) parseResults(results *ast.FieldList) []string {
	var result []string

	for _, field := range results.List {
		resultType := p.nodeToString(field.Type)
		if len(field.Names) == 0 {
			result = append(result, resultType)
		} else {
			for range field.Names {
				result = append(result, resultType)
			}
		}
	}

	return result
}

func (p *Parser) nodeToString(node ast.Node) string {
	if node == nil {
		return ""
	}

	switch n := node.(type) {
	case *ast.Ident:
		return n.Name
	case *ast.StarExpr:
		return "*" + p.nodeToString(n.X)
	case *ast.SelectorExpr:
		return p.nodeToString(n.X) + "." + p.nodeToString(n.Sel)
	case *ast.ArrayType:
		if n.Len == nil {
			return "[]" + p.nodeToString(n.Elt)
		}
		return "[" + p.nodeToString(n.Len) + "]" + p.nodeToString(n.Elt)
	case *ast.MapType:
		return "map[" + p.nodeToString(n.Key) + "]" + p.nodeToString(n.Value)
	case *ast.ChanType:
		if n.Dir == ast.SEND {
			return "chan<- " + p.nodeToString(n.Value)
		}
		if n.Dir == ast.RECV {
			return "<-chan " + p.nodeToString(n.Value)
		}
		return "chan " + p.nodeToString(n.Value)
	case *ast.Ellipsis:
		return "..." + p.nodeToString(n.Elt)
	case *ast.ParenExpr:
		return "(" + p.nodeToString(n.X) + ")"
	case *ast.FuncType:
		return p.funcTypeToString(n)
	case *ast.BasicLit:
		return n.Value
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.IndexExpr:
		return p.nodeToString(n.X) + "[" + p.nodeToString(n.Index) + "]"
	case *ast.UnaryExpr:
		return n.Op.String() + p.nodeToString(n.X)
	case *ast.BinaryExpr:
		return p.nodeToString(n.X) + " " + n.Op.String() + " " + p.nodeToString(n.Y)
	case *ast.IndexListExpr:
		var indices []string
		for _, idx := range n.Indices {
			indices = append(indices, p.nodeToString(idx))
		}
		return p.nodeToString(n.X) + "[" + strings.Join(indices, ", ") + "]"
	case *ast.FuncLit:
		return "func"
	default:
		var buf bytes.Buffer
		err := printer.Fprint(&buf, p.fset, node)
		if err != nil {
			return ""
		}
		return buf.String()
	}
}

func (p *Parser) funcTypeToString(ft *ast.FuncType) string {
	var params []string
	if ft.Params != nil {
		for _, field := range ft.Params.List {
			paramType := p.nodeToString(field.Type)
			if len(field.Names) == 0 {
				params = append(params, paramType)
			} else {
				for range field.Names {
					params = append(params, paramType)
				}
			}
		}
	}

	var results []string
	if ft.Results != nil {
		for _, field := range ft.Results.List {
			resultType := p.nodeToString(field.Type)
			if len(field.Names) == 0 {
				results = append(results, resultType)
			} else {
				for range field.Names {
					results = append(results, resultType)
				}
			}
		}
	}

	resultStr := ""
	if len(results) == 1 {
		resultStr = results[0]
	} else if len(results) > 1 {
		resultStr = "(" + strings.Join(results, ", ") + ")"
	}

	return "func(" + strings.Join(params, ", ") + ") " + resultStr
}

func mergePackageDocs(dest, src *PackageDoc) {
	if dest.Comment == "" {
		dest.Comment = src.Comment
		dest.Summary = src.Summary
		dest.Examples = src.Examples
	}

	dest.Imports = mergeImports(dest.Imports, src.Imports)
	dest.Constants = append(dest.Constants, src.Constants...)
	dest.Variables = append(dest.Variables, src.Variables...)
	dest.Types = append(dest.Types, src.Types...)
	dest.Functions = append(dest.Functions, src.Functions...)
	dest.Methods = append(dest.Methods, src.Methods...)
	dest.Files = append(dest.Files, src.Files...)
}

func mergeImports(dest, src []Import) []Import {
	seen := make(map[string]bool)
	result := make([]Import, 0, len(dest)+len(src))

	for _, imp := range dest {
		key := imp.Name + "|" + imp.Path
		if !seen[key] {
			seen[key] = true
			result = append(result, imp)
		}
	}

	for _, imp := range src {
		key := imp.Name + "|" + imp.Path
		if !seen[key] {
			seen[key] = true
			result = append(result, imp)
		}
	}

	return result
}

func isExported(name string) bool {
	if name == "" {
		return false
	}
	r := []rune(name)[0]
	return unicode.IsUpper(r)
}

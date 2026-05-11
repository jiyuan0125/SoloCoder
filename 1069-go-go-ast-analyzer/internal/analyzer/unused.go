package analyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"

	"github.com/example/go-ast-analyzer/internal/api"
)

type varDecl struct {
	name string
	line int
	used bool
}

type unusedVarsAnalyzer struct {
	filename string
	fset     *token.FileSet
	scopes   []map[string]*varDecl
}

func newUnusedVarsAnalyzer(filename string, fset *token.FileSet) *unusedVarsAnalyzer {
	return &unusedVarsAnalyzer{
		filename: filepath.Base(filename),
		fset:     fset,
		scopes:   []map[string]*varDecl{make(map[string]*varDecl)},
	}
}

func (a *unusedVarsAnalyzer) pushScope() {
	a.scopes = append(a.scopes, make(map[string]*varDecl))
}

func (a *unusedVarsAnalyzer) popScope() []*varDecl {
	if len(a.scopes) <= 1 {
		return nil
	}
	scope := a.scopes[len(a.scopes)-1]
	a.scopes = a.scopes[:len(a.scopes)-1]
	
	var unused []*varDecl
	for _, v := range scope {
		if !v.used {
			unused = append(unused, v)
		}
	}
	return unused
}

func (a *unusedVarsAnalyzer) declare(name string, line int) {
	if name == "_" {
		return
	}
	currentScope := a.scopes[len(a.scopes)-1]
	currentScope[name] = &varDecl{
		name: name,
		line: line,
		used: false,
	}
}

func (a *unusedVarsAnalyzer) use(name string) {
	if name == "_" {
		return
	}
	for i := len(a.scopes) - 1; i >= 0; i-- {
		if v, ok := a.scopes[i][name]; ok {
			v.used = true
			return
		}
	}
}

func AnalyzeUnused(filename string, fset *token.FileSet, file *ast.File) ([]api.UnusedVariable, []api.UnusedImport) {
	varAnalyzer := newUnusedVarsAnalyzer(filename, fset)
	unusedVars := []api.UnusedVariable{}
	
	imports := make(map[string]string)
	importLines := make(map[string]int)
	usedImports := make(map[string]bool)
	
	for _, imp := range file.Imports {
		name := ""
		if imp.Name != nil {
			name = imp.Name.Name
		} else {
			name = filepath.Base(imp.Path.Value[1 : len(imp.Path.Value)-1])
		}
		path := imp.Path.Value[1 : len(imp.Path.Value)-1]
		line := fset.Position(imp.Pos()).Line
		if name != "_" {
			imports[name] = path
			importLines[name] = line
		}
	}
	
	useImport := func(name string) {
		if _, ok := imports[name]; ok {
			usedImports[name] = true
		}
	}
	
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.VAR {
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range vs.Names {
							varAnalyzer.declare(id.Name, fset.Position(id.Pos()).Line)
						}
						for _, val := range vs.Values {
							analyzeExprForUse(val, varAnalyzer.use, useImport)
						}
					}
				}
			} else if d.Tok == token.CONST {
				for _, spec := range d.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range vs.Names {
							varAnalyzer.declare(id.Name, fset.Position(id.Pos()).Line)
						}
						for _, val := range vs.Values {
							analyzeExprForUse(val, varAnalyzer.use, useImport)
						}
					}
				}
			}
		case *ast.FuncDecl:
			if d.Body != nil {
				varAnalyzer.pushScope()
				
				if d.Type.Params != nil {
					for _, field := range d.Type.Params.List {
						for _, id := range field.Names {
							varAnalyzer.declare(id.Name, fset.Position(id.Pos()).Line)
						}
					}
				}
				
				if d.Type.Results != nil {
					for _, field := range d.Type.Results.List {
						for _, id := range field.Names {
							varAnalyzer.declare(id.Name, fset.Position(id.Pos()).Line)
						}
					}
				}
				
				analyzeBlockStmt(d.Body, varAnalyzer, useImport)
				
				unused := varAnalyzer.popScope()
				for _, v := range unused {
					unusedVars = append(unusedVars, api.UnusedVariable{
						Filename: filepath.Base(filename),
						Line:     v.line,
						Name:     v.name,
					})
				}
			}
		}
	}
	
	for _, v := range varAnalyzer.scopes[0] {
		if !v.used {
			unusedVars = append(unusedVars, api.UnusedVariable{
				Filename: filepath.Base(filename),
				Line:     v.line,
				Name:     v.name,
			})
		}
	}
	
	var unusedImports []api.UnusedImport
	for name, path := range imports {
		if !usedImports[name] {
			unusedImports = append(unusedImports, api.UnusedImport{
				Filename: filepath.Base(filename),
				Line:     importLines[name],
				Path:     path,
			})
		}
	}
	
	return unusedVars, unusedImports
}

func analyzeBlockStmt(block *ast.BlockStmt, va *unusedVarsAnalyzer, useImport func(string)) {
	for _, stmt := range block.List {
		analyzeStmt(stmt, va, useImport)
	}
}

func analyzeStmt(stmt ast.Stmt, va *unusedVarsAnalyzer, useImport func(string)) {
	if stmt == nil {
		return
	}
	
	switch s := stmt.(type) {
	case *ast.AssignStmt:
		if s.Tok == token.DEFINE {
			for _, lhs := range s.Lhs {
				if id, ok := lhs.(*ast.Ident); ok {
					va.declare(id.Name, va.fset.Position(id.Pos()).Line)
				}
			}
		}
		for _, rhs := range s.Rhs {
			analyzeExprForUse(rhs, va.use, useImport)
		}
		
	case *ast.IncDecStmt:
		analyzeExprForUse(s.X, va.use, useImport)
		
	case *ast.ExprStmt:
		analyzeExprForUse(s.X, va.use, useImport)
		
	case *ast.ReturnStmt:
		for _, result := range s.Results {
			analyzeExprForUse(result, va.use, useImport)
		}
		
	case *ast.IfStmt:
		va.pushScope()
		analyzeStmt(s.Init, va, useImport)
		analyzeExprForUse(s.Cond, va.use, useImport)
		analyzeBlockStmt(s.Body, va, useImport)
		va.popScope()
		if s.Else != nil {
			va.pushScope()
			analyzeStmt(s.Else, va, useImport)
			va.popScope()
		}
		
	case *ast.ForStmt:
		va.pushScope()
		analyzeStmt(s.Init, va, useImport)
		analyzeExprForUse(s.Cond, va.use, useImport)
		analyzeStmt(s.Post, va, useImport)
		analyzeBlockStmt(s.Body, va, useImport)
		va.popScope()
		
	case *ast.RangeStmt:
		va.pushScope()
		if s.Tok == token.DEFINE {
			if s.Key != nil {
				if id, ok := s.Key.(*ast.Ident); ok {
					va.declare(id.Name, va.fset.Position(id.Pos()).Line)
				}
			}
			if s.Value != nil {
				if id, ok := s.Value.(*ast.Ident); ok {
					va.declare(id.Name, va.fset.Position(id.Pos()).Line)
				}
			}
		}
		analyzeExprForUse(s.X, va.use, useImport)
		analyzeBlockStmt(s.Body, va, useImport)
		va.popScope()
		
	case *ast.SwitchStmt:
		va.pushScope()
		analyzeStmt(s.Init, va, useImport)
		analyzeExprForUse(s.Tag, va.use, useImport)
		for _, cs := range s.Body.List {
			if cc, ok := cs.(*ast.CaseClause); ok {
				va.pushScope()
				for _, expr := range cc.List {
					analyzeExprForUse(expr, va.use, useImport)
				}
				for _, bodyStmt := range cc.Body {
					analyzeStmt(bodyStmt, va, useImport)
				}
				va.popScope()
			}
		}
		va.popScope()
		
	case *ast.TypeSwitchStmt:
		va.pushScope()
		analyzeStmt(s.Init, va, useImport)
		if assign, ok := s.Assign.(*ast.AssignStmt); ok {
			if assign.Tok == token.DEFINE {
				for _, lhs := range assign.Lhs {
					if id, ok := lhs.(*ast.Ident); ok {
						va.declare(id.Name, va.fset.Position(id.Pos()).Line)
					}
				}
			}
		}
		for _, cs := range s.Body.List {
			if cc, ok := cs.(*ast.CaseClause); ok {
				va.pushScope()
				for _, bodyStmt := range cc.Body {
					analyzeStmt(bodyStmt, va, useImport)
				}
				va.popScope()
			}
		}
		va.popScope()
		
	case *ast.SelectStmt:
		va.pushScope()
		for _, cs := range s.Body.List {
			if cc, ok := cs.(*ast.CommClause); ok {
				va.pushScope()
				analyzeStmt(cc.Comm, va, useImport)
				for _, bodyStmt := range cc.Body {
					analyzeStmt(bodyStmt, va, useImport)
				}
				va.popScope()
			}
		}
		va.popScope()
		
	case *ast.BlockStmt:
		va.pushScope()
		analyzeBlockStmt(s, va, useImport)
		va.popScope()
		
	case *ast.DeclStmt:
		if decl, ok := s.Decl.(*ast.GenDecl); ok {
			if decl.Tok == token.VAR {
				for _, spec := range decl.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range vs.Names {
							va.declare(id.Name, va.fset.Position(id.Pos()).Line)
						}
						for _, val := range vs.Values {
							analyzeExprForUse(val, va.use, useImport)
						}
					}
				}
			} else if decl.Tok == token.CONST {
				for _, spec := range decl.Specs {
					if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, id := range vs.Names {
							va.declare(id.Name, va.fset.Position(id.Pos()).Line)
						}
						for _, val := range vs.Values {
							analyzeExprForUse(val, va.use, useImport)
						}
					}
				}
			}
		}
		
	case *ast.GoStmt:
		analyzeExprForUse(s.Call, va.use, useImport)
		
	case *ast.DeferStmt:
		analyzeExprForUse(s.Call, va.use, useImport)
		
	case *ast.SendStmt:
		analyzeExprForUse(s.Chan, va.use, useImport)
		analyzeExprForUse(s.Value, va.use, useImport)
		
	case *ast.LabeledStmt:
		analyzeStmt(s.Stmt, va, useImport)
		
	case *ast.BranchStmt:
		
	case *ast.EmptyStmt:
	}
}

func analyzeExprForUse(expr ast.Expr, useVar func(string), useImport func(string)) {
	if expr == nil {
		return
	}
	
	switch e := expr.(type) {
	case *ast.Ident:
		useVar(e.Name)
		
	case *ast.SelectorExpr:
		if id, ok := e.X.(*ast.Ident); ok {
			useImport(id.Name)
			useVar(id.Name)
		} else {
			analyzeExprForUse(e.X, useVar, useImport)
		}
		
	case *ast.CallExpr:
		analyzeExprForUse(e.Fun, useVar, useImport)
		for _, arg := range e.Args {
			analyzeExprForUse(arg, useVar, useImport)
		}
		
	case *ast.ParenExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		
	case *ast.UnaryExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		
	case *ast.BinaryExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		analyzeExprForUse(e.Y, useVar, useImport)
		
	case *ast.IndexExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		analyzeExprForUse(e.Index, useVar, useImport)
		
	case *ast.IndexListExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		for _, idx := range e.Indices {
			analyzeExprForUse(idx, useVar, useImport)
		}
		
	case *ast.SliceExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		analyzeExprForUse(e.Low, useVar, useImport)
		analyzeExprForUse(e.High, useVar, useImport)
		analyzeExprForUse(e.Max, useVar, useImport)
		
	case *ast.TypeAssertExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		
	case *ast.StarExpr:
		analyzeExprForUse(e.X, useVar, useImport)
		
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			analyzeExprForUse(elt, useVar, useImport)
		}
		
	case *ast.KeyValueExpr:
		analyzeExprForUse(e.Key, useVar, useImport)
		analyzeExprForUse(e.Value, useVar, useImport)
		
	case *ast.ArrayType:
	case *ast.StructType:
	case *ast.FuncType:
	case *ast.InterfaceType:
	case *ast.MapType:
	case *ast.ChanType:
	case *ast.BasicLit:
		
	case *ast.FuncLit:
		if e.Body != nil {
			analyzeFuncLitBodyForImports(e.Body, useImport)
		}
	}
}

func analyzeFuncLitBodyForImports(body *ast.BlockStmt, useImport func(string)) {
	ast.Inspect(body, func(node ast.Node) bool {
		if node == nil {
			return true
		}
		switch n := node.(type) {
		case *ast.SelectorExpr:
			if id, ok := n.X.(*ast.Ident); ok {
				useImport(id.Name)
			}
		case *ast.Ident:
			useImport(n.Name)
		}
		return true
	})
}

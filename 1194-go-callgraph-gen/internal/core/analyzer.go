package core

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"unicode"

	"github.com/example/callgraph/internal/model"
)

type Analyzer struct {
	fset    *token.FileSet
	pkgPath string
	files   []model.File

	graph  model.Graph
	nodes  map[string]*model.Node
	anonID int

	scopeStack []*scope
}

type scope struct {
	funcName   string
	funcDecl   *ast.FuncDecl
	anonFuncs  map[string]*ast.FuncLit
	vars       map[string]string
}

func NewAnalyzer(pkgPath string, files []model.File) *Analyzer {
	return &Analyzer{
		fset:    token.NewFileSet(),
		pkgPath: pkgPath,
		files:   files,
		nodes:   make(map[string]*model.Node),
	}
}

func (a *Analyzer) Analyze() (*model.Graph, error) {
	parsedFiles := make(map[string]*ast.File)
	for _, f := range a.files {
		if !strings.HasSuffix(f.Name, ".go") {
			continue
		}
		pf, err := parser.ParseFile(a.fset, f.Name, f.Content, 0)
		if err != nil {
			return nil, fmt.Errorf("parse file %s: %w", f.Name, err)
		}
		parsedFiles[f.Name] = pf
	}

	for name, pf := range parsedFiles {
		a.analyzeDecls(name, pf)
	}

	for name, pf := range parsedFiles {
		a.analyzeFileCalls(name, pf)
	}

	a.finalizeGraph()
	return &a.graph, nil
}

func (a *Analyzer) analyzeDecls(fileName string, f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			a.processFuncDecl(fileName, node)
		}
		return true
	})
}

func (a *Analyzer) processFuncDecl(fileName string, fd *ast.FuncDecl) {
	name := a.funcName(fd)
	if name == "" {
		return
	}

	n := &model.Node{
		Name:       name,
		Package:    a.pkgPath,
		File:       fileName,
		IsMain:     fd.Name.Name == "main" && fd.Recv == nil,
		IsInit:     fd.Name.Name == "init" && fd.Recv == nil,
		IsExported: fd.Recv == nil && isExported(fd.Name.Name),
		IsAnon:     false,
	}

	if _, exists := a.nodes[name]; !exists {
		a.nodes[name] = n
	}
}

func (a *Analyzer) funcName(fd *ast.FuncDecl) string {
	if fd.Name == nil {
		return ""
	}

	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return fd.Name.Name
	}

	recv := fd.Recv.List[0]
	if recv.Type == nil {
		return fd.Name.Name
	}

	recvType := a.exprName(recv.Type)
	if strings.HasPrefix(recvType, "*") {
		recvType = recvType[1:]
	}

	return fmt.Sprintf("(*%s).%s", recvType, fd.Name.Name)
}

func (a *Analyzer) exprName(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.Ident:
		return ex.Name
	case *ast.StarExpr:
		return "*" + a.exprName(ex.X)
	case *ast.SelectorExpr:
		return a.exprName(ex.X) + "." + ex.Sel.Name
	case *ast.IndexExpr:
		return a.exprName(ex.X)
	default:
		return ""
	}
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	r := []rune(name)[0]
	return unicode.IsUpper(r)
}

func (a *Analyzer) analyzeFileCalls(fileName string, f *ast.File) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.FuncDecl:
			a.analyzeFuncDeclCalls(fileName, node)
			return false
		}
		return true
	})
}

func (a *Analyzer) analyzeFuncDeclCalls(fileName string, fd *ast.FuncDecl) {
	enclosingName := a.funcName(fd)
	if enclosingName == "" {
		return
	}

	a.scopeStack = append(a.scopeStack, &scope{
		funcName:  enclosingName,
		funcDecl:  fd,
		anonFuncs: make(map[string]*ast.FuncLit),
		vars:      make(map[string]string),
	})
	defer func() {
		a.scopeStack = a.scopeStack[:len(a.scopeStack)-1]
	}()

	if fd.Body == nil {
		return
	}

	a.analyzeBlock(enclosingName, fd.Body)
}

func (a *Analyzer) analyzeBlock(enclosing string, body *ast.BlockStmt) {
	if body == nil {
		return
	}

	for _, stmt := range body.List {
		a.analyzeStmt(enclosing, stmt)
	}
}

func (a *Analyzer) analyzeStmt(enclosing string, stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		a.analyzeExpr(enclosing, s.X)
	case *ast.GoStmt:
		a.analyzeGoStmt(enclosing, s)
	case *ast.DeferStmt:
		a.analyzeDeferStmt(enclosing, s)
	case *ast.AssignStmt:
		for _, rhs := range s.Rhs {
			a.analyzeExpr(enclosing, rhs)
		}
		if len(s.Lhs) == len(s.Rhs) {
			for i, lhs := range s.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
					rhsType := a.inferReceiverType(s.Rhs[i])
					if rhsType != "" {
						a.registerVarType(id.Name, rhsType)
					}
				}
			}
		}
	case *ast.ReturnStmt:
		for _, ret := range s.Results {
			a.analyzeExpr(enclosing, ret)
		}
	case *ast.IfStmt:
		a.analyzeExpr(enclosing, s.Cond)
		a.analyzeBlock(enclosing, s.Body)
		if s.Else != nil {
			a.analyzeStmt(enclosing, s.Else)
		}
	case *ast.ForStmt:
		if s.Init != nil {
			a.analyzeStmt(enclosing, s.Init)
		}
		if s.Cond != nil {
			a.analyzeExpr(enclosing, s.Cond)
		}
		if s.Post != nil {
			a.analyzeStmt(enclosing, s.Post)
		}
		a.analyzeBlock(enclosing, s.Body)
	case *ast.RangeStmt:
		a.analyzeExpr(enclosing, s.X)
		a.analyzeBlock(enclosing, s.Body)
	case *ast.SwitchStmt:
		if s.Init != nil {
			a.analyzeStmt(enclosing, s.Init)
		}
		if s.Tag != nil {
			a.analyzeExpr(enclosing, s.Tag)
		}
		for _, cc := range s.Body.List {
			ccs := cc.(*ast.CaseClause)
			for _, c := range ccs.List {
				a.analyzeExpr(enclosing, c)
			}
			for _, bs := range ccs.Body {
				a.analyzeStmt(enclosing, bs)
			}
		}
	case *ast.SelectStmt:
		for _, cc := range s.Body.List {
			ccs := cc.(*ast.CommClause)
			if ccs.Comm != nil {
				a.analyzeStmt(enclosing, ccs.Comm)
			}
			for _, bs := range ccs.Body {
				a.analyzeStmt(enclosing, bs)
			}
		}
	case *ast.BlockStmt:
		a.analyzeBlock(enclosing, s)
	case *ast.DeclStmt:
		for _, decl := range s.Decl.(*ast.GenDecl).Specs {
			if vs, ok := decl.(*ast.ValueSpec); ok {
				if vs.Type != nil {
					explicitType := a.exprName(vs.Type)
					for _, name := range vs.Names {
						if name.Name != "_" {
							a.registerVarType(name.Name, explicitType)
						}
					}
				}
				if len(vs.Names) == len(vs.Values) {
					for i, v := range vs.Values {
						a.analyzeExpr(enclosing, v)
						name := vs.Names[i]
						if name.Name != "_" {
							rhsType := a.inferReceiverType(v)
							if rhsType != "" {
								a.registerVarType(name.Name, rhsType)
							}
						}
					}
				} else {
					for _, v := range vs.Values {
						a.analyzeExpr(enclosing, v)
					}
				}
			}
		}
	}
}

func (a *Analyzer) analyzeExpr(enclosing string, expr ast.Expr) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpr:
		a.analyzeCall(enclosing, e)
	case *ast.SelectorExpr:
		a.analyzeExpr(enclosing, e.X)
	case *ast.ParenExpr:
		a.analyzeExpr(enclosing, e.X)
	case *ast.UnaryExpr:
		a.analyzeExpr(enclosing, e.X)
	case *ast.BinaryExpr:
		a.analyzeExpr(enclosing, e.X)
		a.analyzeExpr(enclosing, e.Y)
	case *ast.TypeAssertExpr:
		a.analyzeExpr(enclosing, e.X)
	case *ast.IndexExpr:
		a.analyzeExpr(enclosing, e.X)
		a.analyzeExpr(enclosing, e.Index)
	case *ast.SliceExpr:
		a.analyzeExpr(enclosing, e.X)
		if e.Low != nil {
			a.analyzeExpr(enclosing, e.Low)
		}
		if e.High != nil {
			a.analyzeExpr(enclosing, e.High)
		}
		if e.Max != nil {
			a.analyzeExpr(enclosing, e.Max)
		}
	case *ast.CompositeLit:
		for _, elt := range e.Elts {
			a.analyzeExpr(enclosing, elt)
		}
	case *ast.FuncLit:
		anonName := a.registerAnonFunc(enclosing, e)
		a.analyzeFuncLit(enclosing, anonName, e)
	case *ast.MapType:
	case *ast.ArrayType:
	case *ast.StructType:
	case *ast.Ident:
	case *ast.BasicLit:
	}
}

func (a *Analyzer) analyzeCall(enclosing string, call *ast.CallExpr) {
	for _, arg := range call.Args {
		a.analyzeExpr(enclosing, arg)
	}

	fun := call.Fun

	switch f := fun.(type) {
	case *ast.Ident:
		if a.isLocalFunction(f.Name) {
			a.addEdge(enclosing, f.Name, model.EdgeTypeCall, false)
		}
	case *ast.SelectorExpr:
		a.analyzeSelectorCall(enclosing, f)
	case *ast.FuncLit:
		anonName := a.registerAnonFunc(enclosing, f)
		a.analyzeFuncLit(enclosing, anonName, f)
	}
}

func (a *Analyzer) ensureNodeExists(name, fileName string, isCrossPkg bool) {
	if _, ok := a.nodes[name]; !ok {
		a.nodes[name] = &model.Node{
			Name:       name,
			Package:    a.pkgPath,
			File:       fileName,
			IsMain:     false,
			IsInit:     false,
			IsExported: false,
			IsAnon:     false,
		}
	}
}

func (a *Analyzer) analyzeSelectorCall(enclosing string, sel *ast.SelectorExpr) {
	a.analyzeExpr(enclosing, sel.X)

	methodName := sel.Sel.Name
	receiverType := a.inferReceiverType(sel.X)

	if receiverType != "" {
		fullName := fmt.Sprintf("(*%s).%s", receiverType, methodName)
		if a.isLocalMethod(fullName) {
			a.addEdge(enclosing, fullName, model.EdgeTypeCall, false)
			return
		}
	}

	if pkgIdent, ok := sel.X.(*ast.Ident); ok {
		crossPkgName := fmt.Sprintf("%s.%s", pkgIdent.Name, methodName)
		var fileName string
		if enclosingNode, ok := a.nodes[enclosing]; ok {
			fileName = enclosingNode.File
		}
		a.ensureNodeExists(crossPkgName, fileName, true)
		a.addEdge(enclosing, crossPkgName, model.EdgeTypeCall, false)
	}
}

func (a *Analyzer) inferReceiverType(e ast.Expr) string {
	switch ex := e.(type) {
	case *ast.Ident:
		return a.lookupIdentType(ex.Name)
	case *ast.SelectorExpr:
		return a.lookupIdentType(ex.Sel.Name)
	case *ast.CallExpr:
		return ""
	case *ast.CompositeLit:
		return a.exprName(ex.Type)
	case *ast.IndexExpr:
		return a.inferReceiverType(ex.X)
	case *ast.ParenExpr:
		return a.inferReceiverType(ex.X)
	case *ast.UnaryExpr:
		if ex.Op == token.AND {
			return a.inferReceiverType(ex.X)
		}
	case *ast.TypeAssertExpr:
		if ex.Type != nil {
			return a.exprName(ex.Type)
		}
	}
	return ""
}

func (a *Analyzer) registerVarType(name, typ string) {
	if len(a.scopeStack) == 0 {
		return
	}
	scope := a.scopeStack[len(a.scopeStack)-1]
	scope.vars[name] = typ
}

func (a *Analyzer) lookupIdentType(name string) string {
	for i := len(a.scopeStack) - 1; i >= 0; i-- {
		scope := a.scopeStack[i]
		if t, ok := scope.vars[name]; ok {
			return t
		}
		if scope.funcDecl == nil || scope.funcDecl.Type == nil {
			continue
		}
		if scope.funcDecl.Type.Params != nil {
			for _, p := range scope.funcDecl.Type.Params.List {
				for _, n := range p.Names {
					if n.Name == name {
						return a.exprName(p.Type)
					}
				}
			}
		}
		if scope.funcDecl.Type.Results != nil {
			for _, r := range scope.funcDecl.Type.Results.List {
				for _, n := range r.Names {
					if n.Name == name {
						return a.exprName(r.Type)
					}
				}
			}
		}
	}
	return ""
}

func (a *Analyzer) isLocalFunction(name string) bool {
	_, ok := a.nodes[name]
	return ok
}

func (a *Analyzer) isLocalMethod(fullName string) bool {
	for n := range a.nodes {
		if n == fullName || strings.HasSuffix(n, "."+strings.Split(fullName, ".")[1]) {
			return true
		}
	}
	return false
}

func (a *Analyzer) analyzeGoStmt(enclosing string, gs *ast.GoStmt) {
	if gs.Call == nil {
		return
	}

	call := gs.Call
	for _, arg := range call.Args {
		a.analyzeExpr(enclosing, arg)
	}

	if fl, ok := call.Fun.(*ast.FuncLit); ok {
		anonName := a.registerAnonFunc(enclosing, fl)
		a.addEdge(enclosing, anonName, model.EdgeTypeStarts, false)
		a.analyzeFuncLit(enclosing, anonName, fl)
	} else if id, ok := call.Fun.(*ast.Ident); ok {
		if a.isLocalFunction(id.Name) {
			a.addEdge(enclosing, id.Name, model.EdgeTypeStarts, false)
		}
	} else {
		a.analyzeExpr(enclosing, call.Fun)
	}
}

func (a *Analyzer) analyzeDeferStmt(enclosing string, ds *ast.DeferStmt) {
	if ds.Call == nil {
		return
	}

	call := ds.Call
	for _, arg := range call.Args {
		a.analyzeExpr(enclosing, arg)
	}

	if fl, ok := call.Fun.(*ast.FuncLit); ok {
		anonName := a.registerAnonFunc(enclosing, fl)
		a.addEdge(enclosing, anonName, model.EdgeTypeDefer, false)
		a.analyzeFuncLit(enclosing, anonName, fl)
	} else if id, ok := call.Fun.(*ast.Ident); ok {
		if a.isLocalFunction(id.Name) {
			a.addEdge(enclosing, id.Name, model.EdgeTypeDefer, false)
		}
	} else {
		a.analyzeExpr(enclosing, call.Fun)
	}
}

func (a *Analyzer) registerAnonFunc(enclosing string, _ *ast.FuncLit) string {
	a.anonID++
	anonName := fmt.Sprintf("%s.func%d", enclosing, a.anonID)

	fileName := ""
	if len(a.scopeStack) > 0 {
		if n, ok := a.nodes[enclosing]; ok {
			fileName = n.File
		}
	}

	n := &model.Node{
		Name:       anonName,
		Package:    a.pkgPath,
		File:       fileName,
		IsMain:     false,
		IsInit:     false,
		IsExported: false,
		IsAnon:     true,
	}
	a.nodes[anonName] = n
	return anonName
}

func (a *Analyzer) analyzeFuncLit(enclosing string, anonName string, fl *ast.FuncLit) {
	a.scopeStack = append(a.scopeStack, &scope{
		funcName:  anonName,
		anonFuncs: make(map[string]*ast.FuncLit),
		vars:      make(map[string]string),
	})
	defer func() {
		a.scopeStack = a.scopeStack[:len(a.scopeStack)-1]
	}()

	if fl.Body != nil {
		a.analyzeBlock(anonName, fl.Body)
	}
}

func (a *Analyzer) addEdge(from, to string, et model.EdgeType, iface bool) {
	for _, e := range a.graph.Edges {
		if e.From == from && e.To == to && e.Type == et {
			return
		}
	}
	a.graph.Edges = append(a.graph.Edges, model.Edge{
		From:      from,
		To:        to,
		Type:      et,
		Interface: iface,
	})
}

func (a *Analyzer) finalizeGraph() {
	for _, n := range a.nodes {
		a.graph.Nodes = append(a.graph.Nodes, *n)
	}
	sort.Slice(a.graph.Nodes, func(i, j int) bool {
		return a.graph.Nodes[i].Name < a.graph.Nodes[j].Name
	})
	sort.Slice(a.graph.Edges, func(i, j int) bool {
		if a.graph.Edges[i].From == a.graph.Edges[j].From {
			return a.graph.Edges[i].To < a.graph.Edges[j].To
		}
		return a.graph.Edges[i].From < a.graph.Edges[j].From
	})
}

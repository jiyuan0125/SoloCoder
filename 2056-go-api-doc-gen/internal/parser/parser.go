package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go-api-doc-gen/internal/models"
)

type ParseResult struct {
	APIs    []models.API
	Module  string
}

type Parser struct {
	moduleName string
}

func New(moduleName string) *Parser {
	return &Parser{moduleName: moduleName}
}

func (p *Parser) Parse(dir string) ([]models.API, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", dir)
	}

	goFiles, err := findGoFiles(dir)
	if err != nil {
		return nil, err
	}

	if len(goFiles) == 0 {
		return []models.API{}, nil
	}

	var allAPIs []models.API
	handlerDocs := make(map[string]*DocContent)
	routes := make([]RouteInfo, 0)

	for _, file := range goFiles {
		fileHandlerDocs, fileRoutes, err := p.parseFile(file)
		if err != nil {
			continue
		}
		for k, v := range fileHandlerDocs {
			handlerDocs[k] = v
		}
		routes = append(routes, fileRoutes...)
	}

	for _, route := range routes {
		api := p.buildAPI(route, handlerDocs[route.HandlerName])
		if api != nil {
			allAPIs = append(allAPIs, *api)
		}
	}

	return allAPIs, nil
}

func findGoFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

type RouteInfo struct {
	Path        string
	Method      string
	HandlerName string
}

type DocContent struct {
	Description   string
	Params        []models.Param
	Returns       []models.ReturnValue
	ExampleReq    string
	ExampleResp   string
	IsComplete    bool
	MissingFields []string
}

func (p *Parser) parseFile(filePath string) (map[string]*DocContent, []RouteInfo, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}

	handlerDocs := make(map[string]*DocContent)
	routes := make([]RouteInfo, 0)

	for _, decl := range node.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			if doc := parseFuncDoc(funcDecl); doc != nil {
				handlerDocs[funcDecl.Name.Name] = doc
			}
		}
	}

	routes = append(routes, extractRoutes(node)...)

	return handlerDocs, routes, nil
}

func extractRoutes(node *ast.File) []RouteInfo {
	var routes []RouteInfo

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.CallExpr:
			r := extractRouteFromCall(x)
			if r != nil {
				routes = append(routes, *r)
			}
		}
		return true
	})

	return routes
}

var httpMethodRegex = regexp.MustCompile(`^[A-Z]+$`)

func extractRouteFromCall(call *ast.CallExpr) *RouteInfo {
	isHandleFunc := false
	pathArgIndex := 0
	handlerArgIndex := 1

	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		method := sel.Sel.Name
		if method == "HandleFunc" {
			isHandleFunc = true
			pathArgIndex = 0
			handlerArgIndex = 1
		} else if httpMethodRegex.MatchString(method) && len(call.Args) >= 2 {
			path := extractStringLiteral(call.Args[0])
			handler := extractHandlerName(call.Args[1])
			if path != "" && handler != "" {
				return &RouteInfo{
					Path:        path,
					Method:      method,
					HandlerName: handler,
				}
			}
		}
	}

	if ident, ok := call.Fun.(*ast.Ident); ok {
		if ident.Name == "HandleFunc" {
			isHandleFunc = true
			if len(call.Args) >= 3 {
				pathArgIndex = 1
				handlerArgIndex = 2
			} else if len(call.Args) >= 2 {
				pathArgIndex = 0
				handlerArgIndex = 1
			}
		}
	}

	if isHandleFunc && len(call.Args) > handlerArgIndex {
		pattern := extractStringLiteral(call.Args[pathArgIndex])
		handler := extractHandlerName(call.Args[handlerArgIndex])
		if pattern != "" && handler != "" {
			method, path := parsePattern(pattern)
			return &RouteInfo{
				Path:        path,
				Method:      method,
				HandlerName: handler,
			}
		}
	}

	return nil
}

func parsePattern(pattern string) (string, string) {
	pattern = strings.TrimSpace(pattern)
	parts := strings.SplitN(pattern, " ", 2)
	if len(parts) == 2 && httpMethodRegex.MatchString(parts[0]) {
		return parts[0], parts[1]
	}
	return "ALL", pattern
}

func extractStringLiteral(expr ast.Expr) string {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		value := lit.Value
		value = strings.TrimPrefix(value, `"`)
		value = strings.TrimSuffix(value, `"`)
		value = strings.TrimPrefix(value, "`")
		value = strings.TrimSuffix(value, "`")
		return value
	}
	return ""
}

func extractHandlerName(expr ast.Expr) string {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name
	}
	if sel, ok := expr.(*ast.SelectorExpr); ok {
		return sel.Sel.Name
	}
	return ""
}

func parseFuncDoc(fn *ast.FuncDecl) *DocContent {
	if fn.Doc == nil {
		return nil
	}

	doc := &DocContent{
		IsComplete:    false,
		MissingFields: []string{},
	}

	text := fn.Doc.Text()
	lines := strings.Split(text, "\n")

	haveDesc := false
	haveParams := false
	haveReturns := false
	haveExample := false

	var exampleBuilder strings.Builder
	inExampleReq := false
	inExampleResp := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "@desc") {
			doc.Description = strings.TrimSpace(strings.TrimPrefix(trimmed, "@desc"))
			haveDesc = true
			continue
		}

		if strings.HasPrefix(trimmed, "@param") {
			haveParams = true
			param := parseParamLine(trimmed)
			if param != nil {
				doc.Params = append(doc.Params, *param)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "@return") {
			haveReturns = true
			ret := parseReturnLine(trimmed)
			if ret != nil {
				doc.Returns = append(doc.Returns, *ret)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "@example-request") {
			inExampleReq = true
			inExampleResp = false
			exampleBuilder.Reset()
			haveExample = true
			continue
		}

		if strings.HasPrefix(trimmed, "@example-response") {
			if inExampleReq {
				doc.ExampleReq = strings.TrimSpace(exampleBuilder.String())
				exampleBuilder.Reset()
			}
			inExampleReq = false
			inExampleResp = true
			continue
		}

		if strings.HasPrefix(trimmed, "@") {
			if inExampleReq {
				doc.ExampleReq = strings.TrimSpace(exampleBuilder.String())
			} else if inExampleResp {
				doc.ExampleResp = strings.TrimSpace(exampleBuilder.String())
			}
			inExampleReq = false
			inExampleResp = false
			continue
		}

		if inExampleReq || inExampleResp {
			if exampleBuilder.Len() > 0 {
				exampleBuilder.WriteString("\n")
			}
			exampleBuilder.WriteString(line)
		}
	}

	if inExampleReq {
		doc.ExampleReq = strings.TrimSpace(exampleBuilder.String())
	} else if inExampleResp {
		doc.ExampleResp = strings.TrimSpace(exampleBuilder.String())
	}

	if !haveDesc {
		doc.MissingFields = append(doc.MissingFields, "description")
	}
	if !haveParams {
		doc.MissingFields = append(doc.MissingFields, "params")
	}
	if !haveReturns {
		doc.MissingFields = append(doc.MissingFields, "returns")
	}
	if !haveExample {
		doc.MissingFields = append(doc.MissingFields, "example")
	}

	doc.IsComplete = len(doc.MissingFields) == 0

	return doc
}

func parseParamLine(line string) *models.Param {
	parts := strings.Fields(strings.TrimPrefix(line, "@param"))
	if len(parts) < 2 {
		return nil
	}

	param := &models.Param{
		Name:     parts[0],
		Type:     "string",
		In:       "query",
		Required: false,
	}

	if len(parts) >= 2 {
		param.Type = parts[1]
	}

	if len(parts) >= 3 {
		rest := strings.Join(parts[2:], " ")
		rest = strings.TrimSpace(rest)
		rest = strings.TrimPrefix(rest, "-")
		rest = strings.TrimSpace(rest)

		if strings.Contains(rest, "[required]") {
			param.Required = true
			rest = strings.Replace(rest, "[required]", "", 1)
		}
		if strings.Contains(rest, "[query]") {
			param.In = "query"
			rest = strings.Replace(rest, "[query]", "", 1)
		}
		if strings.Contains(rest, "[path]") {
			param.In = "path"
			rest = strings.Replace(rest, "[path]", "", 1)
		}
		if strings.Contains(rest, "[header]") {
			param.In = "header"
			rest = strings.Replace(rest, "[header]", "", 1)
		}
		if strings.Contains(rest, "[body]") {
			param.In = "body"
			rest = strings.Replace(rest, "[body]", "", 1)
		}

		param.Description = strings.TrimSpace(rest)
	}

	return param
}

func parseReturnLine(line string) *models.ReturnValue {
	parts := strings.Fields(strings.TrimPrefix(line, "@return"))
	if len(parts) < 2 {
		return nil
	}

	ret := &models.ReturnValue{
		Code: parts[0],
	}

	rest := strings.Join(parts[1:], " ")
	rest = strings.TrimSpace(rest)
	rest = strings.TrimPrefix(rest, "-")
	ret.Description = strings.TrimSpace(rest)

	return ret
}

func (p *Parser) buildAPI(route RouteInfo, doc *DocContent) *models.API {
	api := &models.API{
		Module:      p.moduleName,
		Path:        route.Path,
		Method:      route.Method,
		HandlerName: route.HandlerName,
		IsComplete:  true,
	}

	if doc != nil {
		api.Description = doc.Description
		api.Params = doc.Params
		api.Returns = doc.Returns
		api.Example.Request = doc.ExampleReq
		api.Example.Response = doc.ExampleResp
		api.IsComplete = doc.IsComplete
		api.MissingFields = doc.MissingFields
	} else {
		api.IsComplete = false
		api.MissingFields = []string{"description", "params", "returns", "example"}
	}

	return api
}

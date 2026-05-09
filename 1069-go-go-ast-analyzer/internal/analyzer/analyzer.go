package analyzer

import (
	"go/parser"
	"go/token"

	"github.com/example/go-ast-analyzer/internal/api"
)

func Analyze(filename string, source string) api.AnalysisResponse {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, parser.AllErrors)
	
	if err != nil {
		return api.AnalysisResponse{
			SyntaxErrors: parseSyntaxErrors(filename, source, fset, err),
		}
	}
	
	complexities := AnalyzeComplexity(filename, fset, file)
	callGraph := AnalyzeCallGraph(fset, file)
	unusedVars, unusedImports := AnalyzeUnused(filename, fset, file)
	
	return api.AnalysisResponse{
		Complexities:    complexities,
		CallGraph:       callGraph,
		UnusedVariables: unusedVars,
		UnusedImports:   unusedImports,
	}
}

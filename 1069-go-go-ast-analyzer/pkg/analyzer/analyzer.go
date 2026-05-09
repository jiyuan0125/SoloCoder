package analyzer

import (
	"github.com/example/go-ast-analyzer/internal/analyzer"
	"github.com/example/go-ast-analyzer/internal/api"
)

func Analyze(filename string, source string) api.AnalysisResponse {
	return analyzer.Analyze(filename, source)
}

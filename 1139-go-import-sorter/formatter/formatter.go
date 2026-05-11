package formatter

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
)

func Format(source string, cfg Config) (string, Stats, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "", source, parser.ParseComments)
	if err != nil {
		return "", Stats{}, err
	}

	stats := Stats{}

	var buf bytes.Buffer
	printerCfg := printer.Config{
		Mode:     printer.UseSpaces | printer.TabIndent,
		Tabwidth: 8,
	}
	if err := printerCfg.Fprint(&buf, fset, file); err != nil {
		return "", Stats{}, err
	}

	result := buf.String()

	formatted, importCount := processImportBlock(result, cfg)
	stats.ImportsMoved = importCount
	result = formatted

	lineStats := processLineWidth(&result, cfg)
	stats.LinesSplit = lineStats

	emptyLineStats := processEmptyLines(&result, cfg)
	stats.LinesModified = emptyLineStats + lineStats

	return result, stats, nil
}

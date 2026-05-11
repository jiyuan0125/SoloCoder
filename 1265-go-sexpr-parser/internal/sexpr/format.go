package sexpr

import (
	"strings"
)

func Format(expr Expr) string {
	return formatExpr(expr, 0)
}

func formatExpr(expr Expr, indent int) string {
	switch v := expr.(type) {
	case *Atom:
		switch v.Type {
		case AtomInt:
			return strings.Repeat("  ", indent) + v.String()
		case AtomFloat:
			return strings.Repeat("  ", indent) + v.String()
		case AtomString:
			return strings.Repeat("  ", indent) + v.String()
		case AtomSymbol:
			return strings.Repeat("  ", indent) + v.SymbolValue
		}
	case *List:
		if v.IsNil() {
			return strings.Repeat("  ", indent) + "()"
		}

		allAtoms := true
		for _, e := range v.Elements {
			if _, ok := e.(*List); ok {
				allAtoms = false
				break
			}
		}

		if allAtoms {
			return strings.Repeat("  ", indent) + v.String()
		}

		var sb strings.Builder
		sb.WriteString(strings.Repeat("  ", indent))
		sb.WriteString("(")

		first := v.Elements[0]
		if atom, ok := first.(*Atom); ok {
			sb.WriteString(atom.String())
		} else {
			sb.WriteString("\n")
			sb.WriteString(formatExpr(first, indent+1))
		}

		for i := 1; i < len(v.Elements); i++ {
			sb.WriteString("\n")
			sb.WriteString(formatExpr(v.Elements[i], indent+1))
		}

		sb.WriteString(")")
		return sb.String()
	}
	return ""
}

func FormatAll(exprs []Expr) string {
	var sb strings.Builder
	for i, expr := range exprs {
		if i > 0 {
			sb.WriteString("\n\n")
		}
		sb.WriteString(Format(expr))
	}
	return sb.String()
}

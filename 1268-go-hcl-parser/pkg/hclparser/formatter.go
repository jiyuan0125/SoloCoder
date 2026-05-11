package hclparser

import (
	"fmt"
	"strconv"
	"strings"
)

type Formatter struct {
	indent int
}

func NewFormatter() *Formatter {
	return &Formatter{indent: 0}
}

func (f *Formatter) Format(body *Body) string {
	var sb strings.Builder
	for i, stmt := range body.Statements {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.formatStatement(stmt))
	}
	return sb.String()
}

func (f *Formatter) formatStatement(stmt Statement) string {
	switch s := stmt.(type) {
	case *Block:
		return f.formatBlock(s)
	case *Attribute:
		return f.formatAttribute(s)
	default:
		return ""
	}
}

func (f *Formatter) formatBlock(b *Block) string {
	var sb strings.Builder
	sb.WriteString(strings.Repeat("  ", f.indent))
	sb.WriteString(b.Type)
	for _, label := range b.Labels {
		sb.WriteString(" ")
		sb.WriteString(fmt.Sprintf("%q", label))
	}
	sb.WriteString(" {\n")
	f.indent++
	for i, stmt := range b.Body.Statements {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(f.formatStatement(stmt))
		sb.WriteString("\n")
	}
	f.indent--
	sb.WriteString(strings.Repeat("  ", f.indent))
	sb.WriteString("}")
	return sb.String()
}

func (f *Formatter) formatAttribute(a *Attribute) string {
	var sb strings.Builder
	sb.WriteString(strings.Repeat("  ", f.indent))
	sb.WriteString(a.Key)
	sb.WriteString(" = ")
	sb.WriteString(f.formatExpression(a.Value))
	return sb.String()
}

func (f *Formatter) formatExpression(e Expression) string {
	switch expr := e.(type) {
	case *LiteralValue:
		return f.formatLiteral(expr)
	case *VariableReference:
		return f.formatVariableReference(expr)
	case *FunctionCall:
		return f.formatFunctionCall(expr)
	case *ConditionalExpr:
		return f.formatConditional(expr)
	case *BinaryOp:
		return f.formatBinaryOp(expr)
	case *UnaryOp:
		return f.formatUnaryOp(expr)
	case *ListExpr:
		return f.formatList(expr)
	case *MapExpr:
		return f.formatMap(expr)
	case *HeredocValue:
		return f.formatHeredoc(expr)
	default:
		return ""
	}
}

func (f *Formatter) formatLiteral(l *LiteralValue) string {
	switch v := l.Value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func (f *Formatter) formatVariableReference(vr *VariableReference) string {
	result := strings.Join(vr.Path, ".")
	for _, idx := range vr.Indices {
		result += "[" + f.formatExpression(idx) + "]"
	}
	return result
}

func (f *Formatter) formatFunctionCall(fn *FunctionCall) string {
	var sb strings.Builder
	sb.WriteString(fn.Name)
	sb.WriteString("(")
	args := make([]string, 0, len(fn.PositionalArgs)+len(fn.NamedArgs))
	for _, arg := range fn.PositionalArgs {
		args = append(args, f.formatExpression(arg))
	}
	for k, v := range fn.NamedArgs {
		args = append(args, fmt.Sprintf("%s = %s", k, f.formatExpression(v)))
	}
	sb.WriteString(strings.Join(args, ", "))
	sb.WriteString(")")
	return sb.String()
}

func (f *Formatter) formatConditional(c *ConditionalExpr) string {
	return fmt.Sprintf("%s ? %s : %s",
		f.formatExpression(c.Condition),
		f.formatExpression(c.TrueValue),
		f.formatExpression(c.FalseValue))
}

func (f *Formatter) formatBinaryOp(b *BinaryOp) string {
	return fmt.Sprintf("(%s %s %s)",
		f.formatExpression(b.Left),
		b.Op,
		f.formatExpression(b.Right))
}

func (f *Formatter) formatUnaryOp(u *UnaryOp) string {
	return fmt.Sprintf("(%s%s)", u.Op, f.formatExpression(u.Operand))
}

func (f *Formatter) formatList(l *ListExpr) string {
	if len(l.Elements) == 0 {
		return "[]"
	}
	var sb strings.Builder
	sb.WriteString("[\n")
	f.indent++
	for i, elem := range l.Elements {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(strings.Repeat("  ", f.indent))
		sb.WriteString(f.formatExpression(elem))
	}
	sb.WriteString(",\n")
	f.indent--
	sb.WriteString(strings.Repeat("  ", f.indent))
	sb.WriteString("]")
	return sb.String()
}

func (f *Formatter) formatMap(m *MapExpr) string {
	if len(m.Elements) == 0 {
		return "{}"
	}
	var sb strings.Builder
	sb.WriteString("{\n")
	f.indent++
	i := 0
	for k, v := range m.Elements {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(strings.Repeat("  ", f.indent))
		sb.WriteString(fmt.Sprintf("%s = %s", k, f.formatExpression(v)))
		i++
	}
	sb.WriteString("\n")
	f.indent--
	sb.WriteString(strings.Repeat("  ", f.indent))
	sb.WriteString("}")
	return sb.String()
}

func (f *Formatter) formatHeredoc(h *HeredocValue) string {
	arrow := "<<"
	if h.StripIndent {
		arrow = "<<-"
	}
	delim := h.Delimiter
	content := h.Content
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	return fmt.Sprintf("%s%s\n%s%s", arrow, delim, content, delim)
}

func Format(body *Body) string {
	return NewFormatter().Format(body)
}

func GetValueByPath(body *Body, path string) (interface{}, error) {
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty path")
	}

	var current interface{} = body

	for _, part := range parts {
		switch v := current.(type) {
		case *Body:
			result := findInBody(v, part)
			if result == nil {
				return nil, fmt.Errorf("path not found: %s", path)
			}
			current = result
		case *Block:
			result := findInBody(v.Body, part)
			if result == nil {
				return nil, fmt.Errorf("path not found: %s", path)
			}
			current = result
		default:
			return nil, fmt.Errorf("cannot navigate from %T at path: %s", current, path)
		}
	}

	return current, nil
}

func findInBody(body *Body, key string) interface{} {
	for _, stmt := range body.Statements {
		switch s := stmt.(type) {
		case *Block:
			if s.Type == key {
				return s
			}
			for _, label := range s.Labels {
				if label == key {
					return s
				}
			}
		case *Attribute:
			if s.Key == key {
				return s.Value
			}
		}
	}
	return nil
}

func splitPath(path string) []string {
	var parts []string
	var current strings.Builder
	inBracket := false

	for i := 0; i < len(path); i++ {
		c := path[i]
		switch {
		case c == '.' && !inBracket:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		case c == '[' && !inBracket:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = true
		case c == ']' && inBracket:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = false
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts
}

package parser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func ToJSON(doc *PackageDoc) ([]byte, error) {
	return json.MarshalIndent(doc, "", "  ")
}

func ToMarkdown(doc *PackageDoc) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString("# Package " + doc.Name + "\n\n")

	if doc.Summary != "" {
		buf.WriteString(doc.Summary + "\n\n")
	}

	if doc.Comment != "" && doc.Comment != doc.Summary {
		buf.WriteString(doc.Comment + "\n\n")
	}

	writeExamples(&buf, doc.Examples, "Package")

	if len(doc.Imports) > 0 {
		buf.WriteString("## Imports\n\n")
		buf.WriteString("```go\n")
		for _, imp := range doc.Imports {
			if imp.Name != "" {
				buf.WriteString(fmt.Sprintf("%s %q\n", imp.Name, imp.Path))
			} else {
				buf.WriteString(fmt.Sprintf("%q\n", imp.Path))
			}
		}
		buf.WriteString("```\n\n")
	}

	if len(doc.Constants) > 0 {
		buf.WriteString("## Constants\n\n")
		for _, c := range doc.Constants {
			writeConst(&buf, c)
		}
	}

	if len(doc.Variables) > 0 {
		buf.WriteString("## Variables\n\n")
		for _, v := range doc.Variables {
			writeVar(&buf, v)
		}
	}

	if len(doc.Types) > 0 {
		buf.WriteString("## Types\n\n")
		for _, t := range doc.Types {
			writeType(&buf, t)
		}
	}

	if len(doc.Functions) > 0 {
		buf.WriteString("## Functions\n\n")
		for _, f := range doc.Functions {
			writeFunction(&buf, f)
		}
	}

	if len(doc.Methods) > 0 {
		buf.WriteString("## Methods\n\n")
		for _, m := range doc.Methods {
			writeMethod(&buf, m)
		}
	}

	if len(doc.Files) > 0 {
		buf.WriteString("## Files\n\n")
		for _, f := range doc.Files {
			buf.WriteString("- " + f + "\n")
		}
		buf.WriteString("\n")
	}

	return buf.Bytes(), nil
}

func writeConst(buf *bytes.Buffer, c Const) {
	buf.WriteString("### " + c.Name + "\n\n")
	if !c.Exported {
		buf.WriteString("*[Unexported]*\n\n")
	}
	if c.Summary != "" {
		buf.WriteString(c.Summary + "\n\n")
	}
	if c.Value != "" {
		buf.WriteString("```go\nconst " + c.Name + " = " + c.Value + "\n```\n\n")
	}
	if c.Comment != "" && c.Comment != c.Summary {
		buf.WriteString(c.Comment + "\n\n")
	}
}

func writeVar(buf *bytes.Buffer, v Variable) {
	buf.WriteString("### " + v.Name + "\n\n")
	if !v.Exported {
		buf.WriteString("*[Unexported]*\n\n")
	}
	if v.Summary != "" {
		buf.WriteString(v.Summary + "\n\n")
	}
	if v.Type != "" {
		buf.WriteString("```go\nvar " + v.Name + " " + v.Type)
		if v.Value != "" {
			buf.WriteString(" = " + v.Value)
		}
		buf.WriteString("\n```\n\n")
	}
	if v.Comment != "" && v.Comment != v.Summary {
		buf.WriteString(v.Comment + "\n\n")
	}
}

func writeType(buf *bytes.Buffer, t Type) {
	buf.WriteString("### " + t.Name + "\n\n")
	if !t.Exported {
		buf.WriteString("*[Unexported]*\n\n")
	}
	if t.Summary != "" {
		buf.WriteString(t.Summary + "\n\n")
	}

	buf.WriteString("```go\ntype " + t.Name + " " + string(t.Kind) + " {")
	if t.Kind == TypeKindStruct && t.StructInfo != nil && len(t.StructInfo.Fields) > 0 {
		buf.WriteString("\n")
		for _, f := range t.StructInfo.Fields {
			buf.WriteString("    " + f.Name + " " + f.Type)
			if len(f.Tags) > 0 {
				buf.WriteString(" `")
				var tags []string
				for _, tag := range f.Tags {
					tags = append(tags, tag.Key+":"+tag.Value)
				}
				buf.WriteString(strings.Join(tags, " "))
				buf.WriteString("`")
			}
			buf.WriteString("\n")
		}
	}
	buf.WriteString("}\n```\n\n")

	if t.Kind == TypeKindStruct && t.StructInfo != nil && len(t.StructInfo.Fields) > 0 {
		buf.WriteString("#### Fields\n\n")
		buf.WriteString("| Name | Type | Comment | Tags |\n")
		buf.WriteString("|------|------|---------|------|\n")
		for _, f := range t.StructInfo.Fields {
			tags := ""
			if len(f.Tags) > 0 {
				var tagStrs []string
				for _, tag := range f.Tags {
					tagStrs = append(tagStrs, "`"+tag.Key+":"+tag.Value+"`")
				}
				tags = strings.Join(tagStrs, ", ")
			}
			buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				escapeMarkdown(f.Name),
				escapeMarkdown(f.Type),
				escapeMarkdown(f.Comment),
				escapeMarkdown(tags)))
		}
		buf.WriteString("\n")
	}

	if t.Kind == TypeKindInterface && t.InterfaceInfo != nil && len(t.InterfaceInfo.Methods) > 0 {
		buf.WriteString("#### Methods\n\n")
		buf.WriteString("| Name | Signature | Comment |\n")
		buf.WriteString("|------|-----------|---------|\n")
		for _, m := range t.InterfaceInfo.Methods {
			sig := formatMethodSignature(m)
			buf.WriteString(fmt.Sprintf("| %s | `%s` | %s |\n",
				escapeMarkdown(m.Name),
				escapeMarkdown(sig),
				escapeMarkdown(m.Summary)))
		}
		buf.WriteString("\n")
	}

	if t.Kind == TypeKindAlias && t.AliasInfo != nil {
		buf.WriteString("Alias for: `" + t.AliasInfo.OriginalType + "`\n\n")
	}

	writeExamples(buf, t.Examples, t.Name)

	if t.Comment != "" && t.Comment != t.Summary {
		buf.WriteString(t.Comment + "\n\n")
	}
}

func writeFunction(buf *bytes.Buffer, f Function) {
	buf.WriteString("### " + f.Name + "\n\n")
	if !f.Exported {
		buf.WriteString("*[Unexported]*\n\n")
	}
	if f.Summary != "" {
		buf.WriteString(f.Summary + "\n\n")
	}

	sig := formatFunctionSignature(f)
	buf.WriteString("```go\nfunc " + f.Name + "(" + sig + ")\n")
	if len(f.ReturnTypes) > 0 {
		buf.WriteString("    -> ")
		if len(f.ReturnTypes) == 1 {
			buf.WriteString(f.ReturnTypes[0])
		} else {
			buf.WriteString("(" + strings.Join(f.ReturnTypes, ", ") + ")")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("```\n\n")

	if len(f.Parameters) > 0 {
		buf.WriteString("#### Parameters\n\n")
		for _, p := range f.Parameters {
			name := p.Name
			if name == "" {
				name = "(unnamed)"
			}
			variadic := ""
			if p.IsVariadic {
				variadic = " [variadic]"
			}
			buf.WriteString("- **" + name + "** (`" + p.Type + "`)" + variadic)
			if p.Comment != "" {
				buf.WriteString(": " + p.Comment)
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	writeExamples(buf, f.Examples, f.Name)

	if f.Comment != "" && f.Comment != f.Summary {
		buf.WriteString(f.Comment + "\n\n")
	}
}

func writeMethod(buf *bytes.Buffer, m Method) {
	receiver := ""
	if m.Receiver != nil {
		receiver = m.Receiver.Type + "."
	}

	buf.WriteString("### " + receiver + m.Name + "\n\n")
	if !m.Exported {
		buf.WriteString("*[Unexported]*\n\n")
	}
	if m.Summary != "" {
		buf.WriteString(m.Summary + "\n\n")
	}

	sig := formatMethodSignature(m)
	buf.WriteString("```go\nfunc " + m.Name + "(" + sig + ")\n")
	if len(m.ReturnTypes) > 0 {
		buf.WriteString("    -> ")
		if len(m.ReturnTypes) == 1 {
			buf.WriteString(m.ReturnTypes[0])
		} else {
			buf.WriteString("(" + strings.Join(m.ReturnTypes, ", ") + ")")
		}
		buf.WriteString("\n")
	}
	buf.WriteString("```\n\n")

	if len(m.Parameters) > 0 {
		buf.WriteString("#### Parameters\n\n")
		for _, p := range m.Parameters {
			name := p.Name
			if name == "" {
				name = "(unnamed)"
			}
			variadic := ""
			if p.IsVariadic {
				variadic = " [variadic]"
			}
			buf.WriteString("- **" + name + "** (`" + p.Type + "`)" + variadic)
			if p.Comment != "" {
				buf.WriteString(": " + p.Comment)
			}
			buf.WriteString("\n")
		}
		buf.WriteString("\n")
	}

	writeExamples(buf, m.Examples, m.Name)

	if m.Comment != "" && m.Comment != m.Summary {
		buf.WriteString(m.Comment + "\n\n")
	}
}

func writeExamples(buf *bytes.Buffer, examples []Example, owner string) {
	if len(examples) == 0 {
		return
	}

	buf.WriteString("#### Examples\n\n")
	for _, ex := range examples {
		if ex.Title != "" && ex.Title != "Code Example" {
			buf.WriteString("**" + ex.Title + "**\n\n")
		}
		if ex.Code != "" {
			buf.WriteString("```go\n" + ex.Code + "\n```\n\n")
		}
	}
}

func formatFunctionSignature(f Function) string {
	var params []string
	for _, p := range f.Parameters {
		paramType := p.Type
		if p.IsVariadic {
			paramType = strings.TrimPrefix(paramType, "...")
		}
		if p.Name != "" {
			params = append(params, p.Name+" "+paramType)
		} else {
			params = append(params, paramType)
		}
	}
	return strings.Join(params, ", ")
}

func formatMethodSignature(m Method) string {
	var params []string
	for _, p := range m.Parameters {
		paramType := p.Type
		if p.IsVariadic {
			paramType = strings.TrimPrefix(paramType, "...")
		}
		if p.Name != "" {
			params = append(params, p.Name+" "+paramType)
		} else {
			params = append(params, paramType)
		}
	}
	return strings.Join(params, ", ")
}

func escapeMarkdown(s string) string {
	s = strings.ReplaceAll(s, "|", "\\|")
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}

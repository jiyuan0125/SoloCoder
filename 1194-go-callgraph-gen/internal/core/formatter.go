package core

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/example/callgraph/internal/model"
)

var fileColors = []string{
	"#FFCDD2",
	"#C8E6C9",
	"#BBDEFB",
	"#FFF9C4",
	"#E1BEE7",
	"#B2EBF2",
	"#FFE0B2",
	"#F8BBD0",
	"#D7CCC8",
	"#F5F5F5",
}

func GenerateDOT(graph *model.Graph) string {
	if graph == nil {
		return "digraph G {}"
	}

	fileColorMap := make(map[string]string)
	idx := 0
	for _, n := range graph.Nodes {
		if n.File == "" {
			continue
		}
		if _, ok := fileColorMap[n.File]; !ok {
			fileColorMap[n.File] = fileColors[idx%len(fileColors)]
			idx++
		}
	}

	var sb strings.Builder
	sb.WriteString("digraph G {\n")
	sb.WriteString("\trankdir=TB;\n")
	sb.WriteString("\tcompound=true;\n")
	sb.WriteString("\tnode [fontname=\"Courier New\", fontsize=10];\n")
	sb.WriteString("\tedge [fontname=\"Courier New\", fontsize=8];\n\n")

	for file, color := range fileColorMap {
		sb.WriteString(fmt.Sprintf("\t\"%s\" [shape=plaintext, style=filled, fillcolor=%s];\n", file, color))
	}
	sb.WriteString("\n")

	for _, n := range graph.Nodes {
		sb.WriteString(fmt.Sprintf("\t\"%s\" ", n.Name))
		attrs := []string{}

		if n.IsMain || n.IsInit || n.IsExported {
			attrs = append(attrs, "shape=doublecircle")
		} else if n.IsAnon {
			attrs = append(attrs, "shape=diamond")
		} else {
			attrs = append(attrs, "shape=ellipse")
		}

		if n.File != "" {
			if c, ok := fileColorMap[n.File]; ok {
				attrs = append(attrs, fmt.Sprintf("style=filled, fillcolor=%s", c))
			}
		}

		var labels []string
		labels = append(labels, n.Name)
		if n.IsInit {
			labels = append(labels, "{init}")
		}
		if n.IsMain {
			labels = append(labels, "{main}")
		}
		if n.IsExported {
			labels = append(labels, "{exported}")
		}
		if n.IsAnon {
			labels = append(labels, "{anonymous}")
		}
		label := strings.Join(labels, "\\n")
		attrs = append(attrs, fmt.Sprintf("label=\"%s\"", label))

		sb.WriteString(fmt.Sprintf("[%s];\n", strings.Join(attrs, ", ")))
	}

	sb.WriteString("\n")

	for _, e := range graph.Edges {
		var attrs []string

		switch e.Type {
		case model.EdgeTypeStarts:
			attrs = append(attrs, "style=dashed", "color=blue", "label=\"starts\"")
		case model.EdgeTypeDefer:
			attrs = append(attrs, "style=dotted", "color=purple", "label=\"defer\"")
		default:
			attrs = append(attrs, "style=solid")
		}

		if e.Interface {
			attrs = append(attrs, "comment=\"interface dispatch\"")
			attrs = append(attrs, "xlabel=\"*\"")
		}

		sb.WriteString(fmt.Sprintf("\t\"%s\" -> \"%s\" [%s];\n", e.From, e.To, strings.Join(attrs, ", ")))
	}

	sb.WriteString("}\n")
	return sb.String()
}

type AdjacencyEntry struct {
	From  string   `json:"from"`
	Calls []string `json:"calls"`
	Starts []string `json:"starts,omitempty"`
	Defers []string `json:"defers,omitempty"`
}

func GenerateJSON(graph *model.Graph) (string, error) {
	if graph == nil {
		return "{}", nil
	}

	adjMap := make(map[string]*AdjacencyEntry)

	for _, n := range graph.Nodes {
		adjMap[n.Name] = &AdjacencyEntry{
			From:   n.Name,
			Calls:  []string{},
			Starts: []string{},
			Defers: []string{},
		}
	}

	for _, e := range graph.Edges {
		if entry, ok := adjMap[e.From]; ok {
			switch e.Type {
			case model.EdgeTypeCall:
				entry.Calls = append(entry.Calls, e.To)
			case model.EdgeTypeStarts:
				entry.Starts = append(entry.Starts, e.To)
			case model.EdgeTypeDefer:
				entry.Defers = append(entry.Defers, e.To)
			}
		}
	}

	adjList := []*AdjacencyEntry{}
	for _, entry := range adjMap {
		adjList = append(adjList, entry)
	}

	data, err := json.MarshalIndent(adjList, "", "  ")
	if err != nil {
		return "", fmt.Errorf("json marshal: %w", err)
	}

	return string(data), nil
}

func FormatGraph(graph *model.Graph, format string) (string, error) {
	format = strings.ToLower(strings.TrimSpace(format))
	switch format {
	case "json":
		return GenerateJSON(graph)
	case "dot", "":
		return GenerateDOT(graph), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

package toposort

import (
	"bytes"
	"fmt"
)

func (g *Graph) GenerateDot() string {
	var buf bytes.Buffer
	buf.WriteString("digraph tasks {\n")
	buf.WriteString("    node [shape=box, style=filled, fillcolor=lightblue];\n")

	for taskID, task := range g.tasks {
		label := fmt.Sprintf("%s\\n%s", taskID, task.Name)
		buf.WriteString(fmt.Sprintf("    %q [label=%q];\n", taskID, label))
	}

	for taskID, task := range g.tasks {
		for _, dep := range task.Dependencies {
			buf.WriteString(fmt.Sprintf("    %q -> %q;\n", dep, taskID))
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}

func (g *Graph) GenerateDotWithCriticalPath(criticalPath []string) string {
	var buf bytes.Buffer
	buf.WriteString("digraph tasks {\n")
	buf.WriteString("    node [shape=box];\n")

	criticalSet := make(map[string]bool)
	for _, id := range criticalPath {
		criticalSet[id] = true
	}

	for taskID, task := range g.tasks {
		label := fmt.Sprintf("%s\\n%s", taskID, task.Name)
		if criticalSet[taskID] {
			buf.WriteString(fmt.Sprintf("    %q [label=%q, style=filled, fillcolor=red, fontcolor=white];\n", taskID, label))
		} else {
			buf.WriteString(fmt.Sprintf("    %q [label=%q, style=filled, fillcolor=lightblue];\n", taskID, label))
		}
	}

	criticalEdges := make(map[string]map[string]bool)
	for i := 0; i < len(criticalPath)-1; i++ {
		from := criticalPath[i]
		to := criticalPath[i+1]
		if criticalEdges[from] == nil {
			criticalEdges[from] = make(map[string]bool)
		}
		criticalEdges[from][to] = true
	}

	for taskID, task := range g.tasks {
		for _, dep := range task.Dependencies {
			isCritical := criticalEdges[dep] != nil && criticalEdges[dep][taskID]
			if isCritical {
				buf.WriteString(fmt.Sprintf("    %q -> %q [color=red, penwidth=2];\n", dep, taskID))
			} else {
				buf.WriteString(fmt.Sprintf("    %q -> %q;\n", dep, taskID))
			}
		}
	}

	buf.WriteString("}\n")
	return buf.String()
}

package client

import (
	"encoding/json"
	"fmt"
	"strings"
	
	"dep-tree/protocol"
)

type Config struct {
	ProjectPath      string
	ServerAddr       string
	JSONOutput       bool
	FilterStandard   bool
	FilterThirdParty bool
	SearchPattern    string
}

func Run(cfg Config) error {
	conn, err := protocol.ConnectToServer(cfg.ServerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %w\nMake sure the server is running with: dep-tree-server", err)
	}
	defer conn.Close()

	req := &protocol.Request{
		ProjectPath: cfg.ProjectPath,
		FilterOpts: protocol.FilterOptions{
			FilterStandard:   cfg.FilterStandard,
			FilterThirdParty: cfg.FilterThirdParty,
			SearchPattern:    cfg.SearchPattern,
		},
	}

	requestData, err := protocol.EncodeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to encode request: %w", err)
	}

	err = protocol.WriteMessage(conn, requestData)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	responseData, err := protocol.ReadMessage(conn)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	resp, err := protocol.DecodeResponse(responseData)
	if err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("%s", resp.ErrorMsg)
	}

	if cfg.JSONOutput {
		return outputJSON(resp.Project)
	}

	return outputText(resp.Project)
}

func outputJSON(project *protocol.ProjectInfo) error {
	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputText(project *protocol.ProjectInfo) error {
	fmt.Println("Dependency Tree for:", project.ModuleName)
	fmt.Println()

	printTree(project.Root, 0)

	if len(project.CircularDeps) > 0 {
		fmt.Println()
		fmt.Println("Warning: Circular dependencies detected:")
		for i, dep := range project.CircularDeps {
			fmt.Printf("  %d. %s\n", i+1, strings.Join(dep.Path, " -> "))
		}
	}

	fmt.Println()
	fmt.Println("Statistics:")
	fmt.Printf("  Total dependencies: %d\n", project.TotalDeps)
	fmt.Printf("  Third-party dependencies: %d\n", project.ThirdPartyDeps)
	fmt.Printf("  Standard library dependencies: %d\n", project.TotalDeps-project.ThirdPartyDeps)

	return nil
}

func printTree(node *protocol.DependencyNode, level int) {
	indent := strings.Repeat("  ", level)
	
	label := node.Name
	if node.Version != "" && !node.IsRoot {
		label = fmt.Sprintf("%s %s", node.Name, node.Version)
	}
	if node.IsIndirect {
		label = fmt.Sprintf("%s [indirect]", label)
	}
	if node.IsReused {
		label = fmt.Sprintf("%s (已在上方展开)", label)
	}
	
	fmt.Printf("%s%s\n", indent, label)
	
	if !node.IsStandard && !node.IsReused {
		for _, child := range node.Children {
			printTree(child, level+1)
		}
	}
}

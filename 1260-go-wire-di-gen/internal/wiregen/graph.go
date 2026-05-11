package wiregen

import (
	"fmt"
	"strings"
)

type Graph struct {
	pkg      *PackageInfo
	nodes    map[string]*GraphNode
	edges    map[string][]string
	inDegree map[string]int
}

type GraphNode struct {
	ID       string
	Provider *Provider
	Type     string
	IsExternal bool
	ExternalParam *ExternalParam
}

type CircularDependencyError struct {
	Path []string
}

func (e *CircularDependencyError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s", strings.Join(e.Path, " -> "))
}

type MultipleProvidersError struct {
	Type      string
	Providers []string
}

func (e *MultipleProvidersError) Error() string {
	return fmt.Sprintf("multiple providers found for type %s: %s", e.Type, strings.Join(e.Providers, ", "))
}

type MissingProviderError struct {
	Type string
}

func (e *MissingProviderError) Error() string {
	return fmt.Sprintf("no provider found for type %s", e.Type)
}

func NewGraph(pkg *PackageInfo) *Graph {
	return &Graph{
		pkg:      pkg,
		nodes:    make(map[string]*GraphNode),
		edges:    make(map[string][]string),
		inDegree: make(map[string]int),
	}
}

func (g *Graph) Build() error {
	selectedProviders := make(map[string]bool)

	providerByReturnType := make(map[string][]*Provider)
	for _, p := range g.pkg.Providers {
		providerByReturnType[p.ReturnType] = append(providerByReturnType[p.ReturnType], p)
	}

	needsProvider := make(map[string]bool)
	for _, provider := range g.pkg.Providers {
		for _, param := range provider.Params {
			needsProvider[param.Type] = true
		}
	}

	rootCandidates := make([]*Provider, 0)
	for _, provider := range g.pkg.Providers {
		if !needsProvider[provider.ReturnType] {
			rootCandidates = append(rootCandidates, provider)
		}
	}

	if len(rootCandidates) == 0 {
		for _, p := range g.pkg.Providers {
			rootCandidates = append(rootCandidates, p)
		}
	}

	for _, root := range rootCandidates {
		if err := g.selectProvidersRecursive(root, selectedProviders, providerByReturnType); err != nil {
			return err
		}
	}

	for providerID := range selectedProviders {
		provider := g.pkg.Providers[providerID]
		nodeID := g.nodeIDForProvider(provider)
		g.nodes[nodeID] = &GraphNode{
			ID:       nodeID,
			Provider: provider,
			Type:     provider.ReturnType,
		}
		g.inDegree[nodeID] = 0
	}

	for providerID := range selectedProviders {
		provider := g.pkg.Providers[providerID]
		nodeID := g.nodeIDForProvider(provider)
		for _, param := range provider.Params {
			if err := g.processParam(nodeID, param, selectedProviders, providerByReturnType); err != nil {
				return err
			}
		}
	}

	if err := g.detectCycles(); err != nil {
		return err
	}

	return nil
}

func (g *Graph) selectProvidersRecursive(provider *Provider, selected map[string]bool, providerByReturnType map[string][]*Provider) error {
	if selected[provider.ID] {
		return nil
	}

	selected[provider.ID] = true

	for _, param := range provider.Params {
		if IsBasicType(param.Type) || IsContextType(param.Type) {
			continue
		}

		providersForType := providerByReturnType[param.Type]
		if len(providersForType) == 0 {
			return &MissingProviderError{Type: param.Type}
		}

		var selectedDep *Provider
		if len(providersForType) == 1 {
			selectedDep = providersForType[0]
		} else {
			explicitCount := 0
			var explicitProvider *Provider
			for _, p := range providersForType {
				if p.IsExplicitProvider {
					explicitCount++
					explicitProvider = p
				}
			}
			if explicitCount == 1 {
				selectedDep = explicitProvider
			} else if explicitCount > 1 {
				return &MultipleProvidersError{
					Type:      param.Type,
					Providers: func() []string {
						names := make([]string, 0)
						for _, p := range providersForType {
							if p.IsExplicitProvider {
								names = append(names, p.ID)
							}
						}
						return names
					}(),
				}
			} else {
				return &MultipleProvidersError{
					Type:      param.Type,
					Providers: func() []string {
						names := make([]string, 0, len(providersForType))
						for _, p := range providersForType {
							names = append(names, p.ID)
						}
						return names
					}(),
				}
			}
		}

		if err := g.selectProvidersRecursive(selectedDep, selected, providerByReturnType); err != nil {
			return err
		}
	}

	return nil
}

func (g *Graph) nodeIDForProvider(p *Provider) string {
	return "provider:" + p.ID
}

func (g *Graph) nodeIDForType(t string) string {
	return "type:" + t
}

func (g *Graph) processParam(fromNode string, param ParamInfo, selectedProviders map[string]bool, providerByReturnType map[string][]*Provider) error {
	providers := providerByReturnType[param.Type]

	if len(providers) == 0 {
		if IsBasicType(param.Type) || IsContextType(param.Type) {
			extNodeID := g.nodeIDForType(param.Type)
			if _, exists := g.nodes[extNodeID]; !exists {
				g.nodes[extNodeID] = &GraphNode{
					ID:         extNodeID,
					Type:       param.Type,
					IsExternal: true,
					ExternalParam: &ExternalParam{
						Name: param.Name,
						Type: param.Type,
					},
				}
				g.inDegree[extNodeID] = 0
			}
			g.addEdge(extNodeID, fromNode)
			return nil
		}
		return &MissingProviderError{Type: param.Type}
	}

	var selectedProvider *Provider
	for _, p := range providers {
		if selectedProviders[p.ID] {
			selectedProvider = p
			break
		}
	}

	if selectedProvider == nil {
		if IsBasicType(param.Type) || IsContextType(param.Type) {
			extNodeID := g.nodeIDForType(param.Type)
			if _, exists := g.nodes[extNodeID]; !exists {
				g.nodes[extNodeID] = &GraphNode{
					ID:         extNodeID,
					Type:       param.Type,
					IsExternal: true,
					ExternalParam: &ExternalParam{
						Name: param.Name,
						Type: param.Type,
					},
				}
				g.inDegree[extNodeID] = 0
			}
			g.addEdge(extNodeID, fromNode)
			return nil
		}
		return &MissingProviderError{Type: param.Type}
	}

	toNode := g.nodeIDForProvider(selectedProvider)
	g.addEdge(toNode, fromNode)
	return nil
}

func (g *Graph) addEdge(from, to string) {
	g.edges[from] = append(g.edges[from], to)
	g.inDegree[to]++
}

func (g *Graph) detectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(nodeID string) error
	dfs = func(nodeID string) error {
		visited[nodeID] = true
		recStack[nodeID] = true
		path = append(path, g.nodeLabel(nodeID))

		for _, neighbor := range g.edges[nodeID] {
			if !visited[neighbor] {
				if err := dfs(neighbor); err != nil {
					return err
				}
			} else if recStack[neighbor] {
				cyclePath := append(path, g.nodeLabel(neighbor))
				return &CircularDependencyError{Path: cyclePath}
			}
		}

		recStack[nodeID] = false
		path = path[:len(path)-1]
		return nil
	}

	for nodeID := range g.nodes {
		if !visited[nodeID] {
			if err := dfs(nodeID); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Graph) nodeLabel(nodeID string) string {
	node, ok := g.nodes[nodeID]
	if !ok {
		return nodeID
	}
	if node.IsExternal {
		return fmt.Sprintf("[External] %s", node.Type)
	}
	return node.Provider.Name
}

func (g *Graph) TopologicalSort() ([]*GraphNode, error) {
	inDegree := make(map[string]int)
	for k, v := range g.inDegree {
		inDegree[k] = v
	}

	queue := make([]string, 0)
	for nodeID, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, nodeID)
		}
	}

	result := make([]*GraphNode, 0)
	visited := 0

	for len(queue) > 0 {
		nodeID := queue[0]
		queue = queue[1:]

		node := g.nodes[nodeID]
		if !node.IsExternal {
			result = append(result, node)
		}
		visited++

		for _, neighbor := range g.edges[nodeID] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if visited != len(g.nodes) {
		return nil, fmt.Errorf("graph has a cycle")
	}

	return result, nil
}

func (g *Graph) GetExternalParams() []*ExternalParam {
	params := make([]*ExternalParam, 0)
	for _, node := range g.nodes {
		if node.IsExternal && node.ExternalParam != nil {
			params = append(params, node.ExternalParam)
		}
	}
	return params
}

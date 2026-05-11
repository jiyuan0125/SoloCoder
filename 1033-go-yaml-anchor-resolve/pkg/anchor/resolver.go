package anchor

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
)

type Resolver struct {
	anchors    map[string]*yaml.Node
	anchorList []AnchorInfo
	refs       []ReferenceInfo
	relations  []ReferenceRelation
	warnings   []string
}

func NewResolver() *Resolver {
	return &Resolver{
		anchors:    make(map[string]*yaml.Node),
		anchorList: []AnchorInfo{},
		refs:       []ReferenceInfo{},
		relations:  []ReferenceRelation{},
		warnings:   []string{},
	}
}

func (r *Resolver) Resolve(yamlText string) (*AnalysisResult, error) {
	preprocessed := PreprocessYAML(yamlText)
	
	var doc yaml.Node
	err := yaml.Unmarshal([]byte(preprocessed), &doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	if len(doc.Content) == 0 {
		return &AnalysisResult{
			Anchors:     []AnchorInfo{},
			References:  []ReferenceInfo{},
			Relations:   []ReferenceRelation{},
			Warnings:    []string{},
			ResolvedData: nil,
		}, nil
	}

	root := doc.Content[0]

	r.collectAnchors(root, "")

	resolvedNode, err := r.expandReferences(root, []string{})
	if err != nil {
		return nil, err
	}

	resolvedData, err := r.nodeToInterface(resolvedNode)
	if err != nil {
		return nil, fmt.Errorf("failed to convert resolved node: %w", err)
	}

	return &AnalysisResult{
		Anchors:      r.anchorList,
		References:   r.refs,
		Relations:    r.relations,
		Warnings:     r.warnings,
		ResolvedData: resolvedData,
	}, nil
}

func (r *Resolver) collectAnchors(node *yaml.Node, path string) {
	if node == nil {
		return
	}

	if node.Anchor != "" {
		location := path
		if location == "" {
			location = "<root>"
		}
		r.anchors[node.Anchor] = node
		r.anchorList = append(r.anchorList, AnchorInfo{
			Name:     node.Anchor,
			Location: location,
		})
	}

	switch node.Kind {
	case yaml.DocumentNode, yaml.SequenceNode:
		for i, child := range node.Content {
			childPath := fmt.Sprintf("%s[%d]", path, i)
			if path == "" {
				childPath = fmt.Sprintf("[%d]", i)
			}
			r.collectAnchors(child, childPath)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			keyStr := nodeToString(keyNode)
			keyPath := fmt.Sprintf("%s.%s", path, keyStr)
			if path == "" {
				keyPath = keyStr
			}

			r.collectAnchors(keyNode, keyPath+"(key)")
			r.collectAnchors(valueNode, keyPath)
		}
	case yaml.AliasNode:
		aliasPath := path
		if aliasPath == "" {
			aliasPath = "<root>"
		}
		r.refs = append(r.refs, ReferenceInfo{
			Name:       node.Value,
			Location:   aliasPath,
			IsMergeKey: false,
		})
		r.relations = append(r.relations, ReferenceRelation{
			Anchor:    node.Value,
			Reference: aliasPath,
			IsMerge:   false,
		})
	case yaml.ScalarNode:
		if isAlias, aliasName := IsAliasMarker(node.Value); isAlias {
			aliasPath := path
			if aliasPath == "" {
				aliasPath = "<root>"
			}
			r.refs = append(r.refs, ReferenceInfo{
				Name:       aliasName,
				Location:   aliasPath,
				IsMergeKey: false,
			})
			r.relations = append(r.relations, ReferenceRelation{
				Anchor:    aliasName,
				Reference: aliasPath,
				IsMerge:   false,
			})
		}
	}
}

func (r *Resolver) expandReferences(node *yaml.Node, visitedChain []string) (*yaml.Node, error) {
	if node == nil {
		return nil, nil
	}

	result := &yaml.Node{
		Kind:    node.Kind,
		Style:   node.Style,
		Tag:     node.Tag,
		Value:   node.Value,
		Anchor:  "",
		Alias:   nil,
		Content: []*yaml.Node{},
		HeadComment: node.HeadComment,
		LineComment: node.LineComment,
		FootComment: node.FootComment,
		Line:        node.Line,
		Column:      node.Column,
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			expanded, err := r.expandReferences(child, visitedChain)
			if err != nil {
				return nil, err
			}
			result.Content = append(result.Content, expanded)
		}

	case yaml.SequenceNode:
		for _, child := range node.Content {
			expanded, err := r.expandReferences(child, visitedChain)
			if err != nil {
				return nil, err
			}
			result.Content = append(result.Content, expanded)
		}

	case yaml.MappingNode:
		mergeKeys := []*yaml.Node{}
		regularPairs := [][2]*yaml.Node{}

		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			if isMergeKey(keyNode) {
				mergeKeys = append(mergeKeys, valueNode)
			} else {
				regularPairs = append(regularPairs, [2]*yaml.Node{keyNode, valueNode})
			}
		}

		mergedMap := make(map[string]*yaml.Node)
		mergedKeys := []string{}

		for _, mergeValue := range mergeKeys {
			var itemsToMerge []*yaml.Node
			
			if mergeValue.Kind == yaml.SequenceNode {
				itemsToMerge = mergeValue.Content
			} else {
				itemsToMerge = []*yaml.Node{mergeValue}
			}

			for _, item := range itemsToMerge {
				var targetAnchor string
				var needsExpansion bool

				if item.Kind == yaml.AliasNode {
					targetAnchor = item.Value
					needsExpansion = true
				} else if item.Kind == yaml.ScalarNode {
					if isAlias, aliasName := IsAliasMarker(item.Value); isAlias {
						targetAnchor = aliasName
						needsExpansion = true
					}
				}

				var expandedMerge *yaml.Node
				var err error

				if needsExpansion {
					if contains(visitedChain, targetAnchor) {
						return nil, &CircularReferenceError{Chain: append(visitedChain, targetAnchor)}
					}

					target, ok := r.anchors[targetAnchor]
					if !ok {
						r.warnings = append(r.warnings, fmt.Sprintf("undefined anchor reference: %s", targetAnchor))
						continue
					}

					newChain := append(visitedChain, targetAnchor)
					expandedMerge, err = r.expandReferences(target, newChain)
					if err != nil {
						return nil, err
					}
				} else {
					expandedMerge, err = r.expandReferences(item, visitedChain)
					if err != nil {
						return nil, err
					}
				}

				if expandedMerge.Kind != yaml.MappingNode {
					r.warnings = append(r.warnings, "merge key reference is not a mapping, ignoring")
					continue
				}

				for i := 0; i < len(expandedMerge.Content); i += 2 {
					k := expandedMerge.Content[i]
					v := expandedMerge.Content[i+1]
					keyStr := nodeToString(k)
					if !contains(mergedKeys, keyStr) {
						mergedKeys = append(mergedKeys, keyStr)
					}
					mergedMap[keyStr] = v
				}
			}
		}

		for _, pair := range regularPairs {
			expandedKey, err := r.expandReferences(pair[0], visitedChain)
			if err != nil {
				return nil, err
			}
			expandedValue, err := r.expandReferences(pair[1], visitedChain)
			if err != nil {
				return nil, err
			}

			keyStr := nodeToString(expandedKey)
			if !contains(mergedKeys, keyStr) {
				mergedKeys = append(mergedKeys, keyStr)
			}
			mergedMap[keyStr] = expandedValue
		}

		for _, keyStr := range mergedKeys {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: keyStr,
			}
			result.Content = append(result.Content, keyNode, mergedMap[keyStr])
		}

	case yaml.AliasNode:
		targetAnchor := node.Value
		if contains(visitedChain, targetAnchor) {
			return nil, &CircularReferenceError{Chain: append(visitedChain, targetAnchor)}
		}

		target, ok := r.anchors[targetAnchor]
		if !ok {
			return nil, fmt.Errorf("undefined anchor reference: %s", targetAnchor)
		}

		newChain := append(visitedChain, targetAnchor)
		return r.expandReferences(target, newChain)

	case yaml.ScalarNode:
		if isAlias, aliasName := IsAliasMarker(node.Value); isAlias {
			targetAnchor := aliasName
			if contains(visitedChain, targetAnchor) {
				return nil, &CircularReferenceError{Chain: append(visitedChain, targetAnchor)}
			}

			target, ok := r.anchors[targetAnchor]
			if !ok {
				return nil, fmt.Errorf("undefined anchor reference: %s", targetAnchor)
			}

			newChain := append(visitedChain, targetAnchor)
			return r.expandReferences(target, newChain)
		}
		result.Value = node.Value
		result.Tag = node.Tag
	}

	return result, nil
}

func (r *Resolver) nodeToInterface(node *yaml.Node) (interface{}, error) {
	if node == nil {
		return nil, nil
	}

	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) == 0 {
			return nil, nil
		}
		return r.nodeToInterface(node.Content[0])

	case yaml.SequenceNode:
		result := []interface{}{}
		for _, child := range node.Content {
			item, err := r.nodeToInterface(child)
			if err != nil {
				return nil, err
			}
			result = append(result, item)
		}
		return result, nil

	case yaml.MappingNode:
		result := make(map[string]interface{})
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			keyStr := nodeToString(keyNode)
			value, err := r.nodeToInterface(valueNode)
			if err != nil {
				return nil, err
			}
			result[keyStr] = value
		}
		return result, nil

	case yaml.ScalarNode:
		switch node.Tag {
		case "!!int":
			var i int
			err := yaml.Unmarshal([]byte(node.Value), &i)
			if err == nil {
				return i, nil
			}
		case "!!float":
			var f float64
			err := yaml.Unmarshal([]byte(node.Value), &f)
			if err == nil {
				return f, nil
			}
		case "!!bool":
			var b bool
			err := yaml.Unmarshal([]byte(node.Value), &b)
			if err == nil {
				return b, nil
			}
		case "!!null":
			return nil, nil
		}
		return node.Value, nil

	case yaml.AliasNode:
		return nil, fmt.Errorf("unexpected alias node after expansion")
	}

	return nil, fmt.Errorf("unexpected node kind: %v", node.Kind)
}

func (r *Resolver) ToJSON(data interface{}) (string, error) {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal to JSON: %w", err)
	}
	return string(bytes), nil
}

func (r *Resolver) ToYAML(data interface{}) (string, error) {
	bytes, err := yaml.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}
	return string(bytes), nil
}

func nodeToString(node *yaml.Node) string {
	if node == nil {
		return ""
	}
	if node.Kind == yaml.ScalarNode {
		return node.Value
	}
	bytes, err := yaml.Marshal(node)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return strings.TrimSpace(string(bytes))
}

func isMergeKey(node *yaml.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == yaml.ScalarNode && node.Value == "<<" {
		return true
	}
	if node.Kind == yaml.AliasNode {
		return false
	}
	return false
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

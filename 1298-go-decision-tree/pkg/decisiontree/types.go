package decisiontree

import (
	"encoding/csv"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

type FeatureType int

const (
	Discrete FeatureType = iota
	Continuous
)

type Node struct {
	IsLeaf       bool
	Label        string
	Feature      string
	FeatureIndex int
	Threshold    float64
	Children     map[string]*Node
	FeatureType  FeatureType
}

type DataSet struct {
	Features   []string
	Data       [][]string
	Labels     []string
	FeatureTypes []FeatureType
}

type Parameters struct {
	MaxDepth      int
	MinSamples    int
	HandleMissing string
}

func NewParameters() *Parameters {
	return &Parameters{
		MaxDepth:      10,
		MinSamples:    1,
		HandleMissing: "skip",
	}
}

func ParseCSV(csvData string) (*DataSet, error) {
	reader := csv.NewReader(strings.NewReader(csvData))
	
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV: %v", err)
	}
	
	if len(records) < 2 {
		return nil, fmt.Errorf("dataset is empty or has no data rows")
	}
	
	headers := records[0]
	if len(headers) < 2 {
		return nil, fmt.Errorf("CSV must have at least feature and label columns")
	}
	
	features := headers[:len(headers)-1]
	data := make([][]string, 0, len(records)-1)
	labels := make([]string, 0, len(records)-1)
	
	for i, row := range records[1:] {
		if len(row) != len(headers) {
			return nil, fmt.Errorf("row %d has %d columns, expected %d", i+2, len(row), len(headers))
		}
		data = append(data, row[:len(row)-1])
		labels = append(labels, row[len(row)-1])
	}
	
	featureTypes := detectFeatureTypes(data)
	
	return &DataSet{
		Features:     features,
		Data:         data,
		Labels:       labels,
		FeatureTypes: featureTypes,
	}, nil
}

func detectFeatureTypes(data [][]string) []FeatureType {
	if len(data) == 0 {
		return nil
	}
	
	featureTypes := make([]FeatureType, len(data[0]))
	
	for i := 0; i < len(data[0]); i++ {
		isContinuous := true
		for _, row := range data {
			if row[i] == "" || strings.EqualFold(row[i], "NaN") {
				continue
			}
			if _, err := strconv.ParseFloat(row[i], 64); err != nil {
				isContinuous = false
				break
			}
		}
		if isContinuous {
			featureTypes[i] = Continuous
		} else {
			featureTypes[i] = Discrete
		}
	}
	
	return featureTypes
}

func (ds *DataSet) HandleMissingValues(method string) (*DataSet, error) {
	if method != "skip" && method != "mean" {
		return nil, fmt.Errorf("unknown missing value handling method: %s", method)
	}
	
	if method == "skip" {
		newData := make([][]string, 0)
		newLabels := make([]string, 0)
		
		for i, row := range ds.Data {
			hasMissing := false
			for _, val := range row {
				if val == "" || strings.EqualFold(val, "NaN") {
					hasMissing = true
					break
				}
			}
			if !hasMissing {
				newData = append(newData, row)
				newLabels = append(newLabels, ds.Labels[i])
			}
		}
		
		if len(newData) == 0 {
			return nil, fmt.Errorf("all rows have missing values")
		}
		
		return &DataSet{
			Features:     ds.Features,
			Data:         newData,
			Labels:       newLabels,
			FeatureTypes: ds.FeatureTypes,
		}, nil
	}
	
	newData := make([][]string, len(ds.Data))
	for i := range newData {
		newData[i] = make([]string, len(ds.Data[i]))
		copy(newData[i], ds.Data[i])
	}
	
	for j := 0; j < len(ds.Features); j++ {
		if ds.FeatureTypes[j] == Continuous {
			var sum float64
			var count int
			for _, row := range ds.Data {
				if row[j] != "" && !strings.EqualFold(row[j], "NaN") {
					if val, err := strconv.ParseFloat(row[j], 64); err == nil {
						sum += val
						count++
					}
				}
			}
			if count > 0 {
				mean := sum / float64(count)
				meanStr := strconv.FormatFloat(mean, 'f', -1, 64)
				for i, row := range newData {
					if row[j] == "" || strings.EqualFold(row[j], "NaN") {
						newData[i][j] = meanStr
					}
				}
			}
		}
	}
	
	return &DataSet{
		Features:     ds.Features,
		Data:         newData,
		Labels:       ds.Labels,
		FeatureTypes: ds.FeatureTypes,
	}, nil
}

func entropy(labels []string) float64 {
	if len(labels) == 0 {
		return 0.0
	}
	
	labelCounts := make(map[string]int)
	for _, label := range labels {
		labelCounts[label]++
	}
	
	var e float64
	total := float64(len(labels))
	for _, count := range labelCounts {
		p := float64(count) / total
		if p > 0 {
			e -= p * math.Log2(p)
		}
	}
	
	return e
}

func informationGain(parentLabels []string, childLabels [][]string) float64 {
	parentEntropy := entropy(parentLabels)
	
	total := float64(len(parentLabels))
	var weightedChildEntropy float64
	
	for _, labels := range childLabels {
		if len(labels) == 0 {
			continue
		}
		weight := float64(len(labels)) / total
		weightedChildEntropy += weight * entropy(labels)
	}
	
	return parentEntropy - weightedChildEntropy
}

func (ds *DataSet) calculateContinuousSplit(featureIndex int, labels []string) (float64, float64) {
	if len(labels) == 0 {
		return 0, 0
	}
	
	type valueLabel struct {
		value float64
		label string
	}
	
	var values []valueLabel
	for i, row := range ds.Data {
		if row[featureIndex] == "" || strings.EqualFold(row[featureIndex], "NaN") {
			continue
		}
		if val, err := strconv.ParseFloat(row[featureIndex], 64); err == nil {
			values = append(values, valueLabel{value: val, label: labels[i]})
		}
	}
	
	if len(values) < 2 {
		return 0, 0
	}
	
	sort.Slice(values, func(i, j int) bool {
		return values[i].value < values[j].value
	})
	
	bestThreshold := 0.0
	bestGain := 0.0
	
	for i := 0; i < len(values)-1; i++ {
		if values[i].value == values[i+1].value {
			continue
		}
		
		threshold := (values[i].value + values[i+1].value) / 2.0
		
		var leftLabels []string
		var rightLabels []string
		
		for _, vl := range values {
			if vl.value <= threshold {
				leftLabels = append(leftLabels, vl.label)
			} else {
				rightLabels = append(rightLabels, vl.label)
			}
		}
		
		gain := informationGain(labels, [][]string{leftLabels, rightLabels})
		if gain > bestGain {
			bestGain = gain
			bestThreshold = threshold
		}
	}
	
	return bestThreshold, bestGain
}

func (ds *DataSet) calculateDiscreteGain(featureIndex int, labels []string) float64 {
	if len(labels) == 0 {
		return 0
	}
	
	valueLabels := make(map[string][]string)
	for i, row := range ds.Data {
		value := row[featureIndex]
		if value == "" || strings.EqualFold(value, "NaN") {
			continue
		}
		valueLabels[value] = append(valueLabels[value], labels[i])
	}
	
	if len(valueLabels) <= 1 {
		return 0
	}
	
	var childLabels [][]string
	for _, lbls := range valueLabels {
		childLabels = append(childLabels, lbls)
	}
	
	return informationGain(labels, childLabels)
}

func Train(ds *DataSet, params *Parameters) (*Node, error) {
	if len(ds.Data) == 0 {
		return nil, fmt.Errorf("dataset is empty")
	}
	
	if params == nil {
		params = NewParameters()
	}
	
	return buildTree(ds, ds.Labels, make(map[int]bool), 0, params)
}

func allSameLabel(labels []string) bool {
	if len(labels) == 0 {
		return true
	}
	first := labels[0]
	for _, lbl := range labels[1:] {
		if lbl != first {
			return false
		}
	}
	return true
}

func majorityVote(labels []string) string {
	if len(labels) == 0 {
		return ""
	}
	
	counts := make(map[string]int)
	maxCount := 0
	maxLabel := ""
	
	for _, lbl := range labels {
		counts[lbl]++
		if counts[lbl] > maxCount {
			maxCount = counts[lbl]
			maxLabel = lbl
		}
	}
	
	return maxLabel
}

func buildTree(ds *DataSet, labels []string, usedFeatures map[int]bool, depth int, params *Parameters) (*Node, error) {
	if len(labels) == 0 {
		return nil, fmt.Errorf("empty subset")
	}
	
	if depth >= params.MaxDepth {
		return &Node{
			IsLeaf: true,
			Label:  majorityVote(labels),
		}, nil
	}
	
	if len(labels) < params.MinSamples {
		return &Node{
			IsLeaf: true,
			Label:  majorityVote(labels),
		}, nil
	}
	
	if allSameLabel(labels) {
		return &Node{
			IsLeaf: true,
			Label:  labels[0],
		}, nil
	}
	
	if len(usedFeatures) >= len(ds.Features) {
		return &Node{
			IsLeaf: true,
			Label:  majorityVote(labels),
		}, nil
	}
	
	bestFeatureIndex := -1
	bestThreshold := 0.0
	bestGain := 0.0
	bestFeatureType := Discrete
	
	for i := 0; i < len(ds.Features); i++ {
		if usedFeatures[i] {
			continue
		}
		
		var gain float64
		var threshold float64
		
		if ds.FeatureTypes[i] == Continuous {
			threshold, gain = ds.calculateContinuousSplit(i, labels)
		} else {
			gain = ds.calculateDiscreteGain(i, labels)
			if gain <= 0 {
				continue
			}
		}
		
		if gain > bestGain {
			bestGain = gain
			bestFeatureIndex = i
			bestFeatureType = ds.FeatureTypes[i]
			if bestFeatureType == Continuous {
				bestThreshold = threshold
			}
		}
	}
	
	if bestGain <= 0 {
		return &Node{
			IsLeaf: true,
			Label:  majorityVote(labels),
		}, nil
	}
	
	node := &Node{
		IsLeaf:       false,
		Feature:      ds.Features[bestFeatureIndex],
		FeatureIndex: bestFeatureIndex,
		FeatureType:  bestFeatureType,
		Threshold:    bestThreshold,
		Children:     make(map[string]*Node),
	}
	
	if bestFeatureType == Discrete {
		usedFeatures[bestFeatureIndex] = true
	}
	
	if bestFeatureType == Discrete {
		valueSubsets := make(map[string]*DataSet)
		valueLabels := make(map[string][]string)
		
		for i, row := range ds.Data {
			value := row[bestFeatureIndex]
			if value == "" || strings.EqualFold(value, "NaN") {
				continue
			}
			
			if valueSubsets[value] == nil {
				valueSubsets[value] = &DataSet{
					Features:     ds.Features,
					FeatureTypes: ds.FeatureTypes,
				}
			}
			
			newRow := make([]string, len(row))
			copy(newRow, row)
			valueSubsets[value].Data = append(valueSubsets[value].Data, newRow)
			valueLabels[value] = append(valueLabels[value], labels[i])
		}
		
		for value, subset := range valueSubsets {
			child, err := buildTree(subset, valueLabels[value], usedFeatures, depth+1, params)
			if err != nil {
				return nil, err
			}
			node.Children[value] = child
		}
	} else {
		leftover := &DataSet{
			Features:     ds.Features,
			FeatureTypes: ds.FeatureTypes,
		}
		rightover := &DataSet{
			Features:     ds.Features,
			FeatureTypes: ds.FeatureTypes,
		}
		var leftLabels []string
		var rightLabels []string
		
		for i, row := range ds.Data {
			if row[bestFeatureIndex] == "" || strings.EqualFold(row[bestFeatureIndex], "NaN") {
				continue
			}
			
			if val, err := strconv.ParseFloat(row[bestFeatureIndex], 64); err == nil {
				newRow := make([]string, len(row))
				copy(newRow, row)
				
				if val <= bestThreshold {
					leftover.Data = append(leftover.Data, newRow)
					leftLabels = append(leftLabels, labels[i])
				} else {
					rightover.Data = append(rightover.Data, newRow)
					rightLabels = append(rightLabels, labels[i])
				}
			}
		}
		
		if len(leftLabels) > 0 {
			leftChild, err := buildTree(leftover, leftLabels, usedFeatures, depth+1, params)
			if err != nil {
				return nil, err
			}
			node.Children["<="] = leftChild
		} else {
			node.Children["<="] = &Node{IsLeaf: true, Label: majorityVote(labels)}
		}
		
		if len(rightLabels) > 0 {
			rightChild, err := buildTree(rightover, rightLabels, usedFeatures, depth+1, params)
			if err != nil {
				return nil, err
			}
			node.Children[">"] = rightChild
		} else {
			node.Children[">"] = &Node{IsLeaf: true, Label: majorityVote(labels)}
		}
	}
	
	return node, nil
}

func (n *Node) Predict(features []string, featureTypes []FeatureType) (string, error) {
	if n.IsLeaf {
		return n.Label, nil
	}
	
	if n.FeatureIndex < 0 || n.FeatureIndex >= len(features) {
		return "", fmt.Errorf("feature index %d out of range", n.FeatureIndex)
	}
	
	value := features[n.FeatureIndex]
	
	if n.FeatureType == Continuous {
		if value == "" || strings.EqualFold(value, "NaN") {
			return "", fmt.Errorf("missing value for continuous feature %s", n.Feature)
		}
		
		val, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("invalid numeric value %s for feature %s", value, n.Feature)
		}
		
		var key string
		if val <= n.Threshold {
			key = "<="
		} else {
			key = ">"
		}
		
		child, ok := n.Children[key]
		if !ok {
			return "", fmt.Errorf("no child node for %s %f", key, n.Threshold)
		}
		
		return child.Predict(features, featureTypes)
	}
	
	child, ok := n.Children[value]
	if !ok {
		return "", fmt.Errorf("unknown value %s for feature %s", value, n.Feature)
	}
	
	return child.Predict(features, featureTypes)
}

func (n *Node) ExportJSON() string {
	if n.IsLeaf {
		return fmt.Sprintf(`{"is_leaf": true, "label": "%s"}`, n.Label)
	}
	
	children := "{"
	first := true
	for k, v := range n.Children {
		if !first {
			children += ", "
		}
		first = false
		children += fmt.Sprintf(`"%s": %s`, k, v.ExportJSON())
	}
	children += "}"
	
	if n.FeatureType == Continuous {
		return fmt.Sprintf(`{"is_leaf": false, "feature": "%s", "threshold": %f, "feature_type": "continuous", "children": %s}`, 
			n.Feature, n.Threshold, children)
	}
	
	return fmt.Sprintf(`{"is_leaf": false, "feature": "%s", "feature_type": "discrete", "children": %s}`, 
		n.Feature, children)
}

func (n *Node) ExportText() string {
	var builder strings.Builder
	exportTextRecursive(n, &builder, 0, "")
	return builder.String()
}

func exportTextRecursive(node *Node, builder *strings.Builder, depth int, condition string) {
	indent := strings.Repeat("  ", depth)
	
	if node.IsLeaf {
		if condition != "" {
			builder.WriteString(fmt.Sprintf("%s%s -> Label: %s\n", indent, condition, node.Label))
		} else {
			builder.WriteString(fmt.Sprintf("%sLabel: %s\n", indent, node.Label))
		}
		return
	}
	
	if condition != "" {
		builder.WriteString(fmt.Sprintf("%s%s\n", indent, condition))
	}
	
	if node.FeatureType == Continuous {
		leqChild := node.Children["<="]
		gtChild := node.Children[">"]
		
		if leqChild != nil {
			leqCondition := fmt.Sprintf("%s <= %.4f", node.Feature, node.Threshold)
			exportTextRecursive(leqChild, builder, depth+1, leqCondition)
		}
		
		if gtChild != nil {
			gtCondition := fmt.Sprintf("%s > %.4f", node.Feature, node.Threshold)
			exportTextRecursive(gtChild, builder, depth+1, gtCondition)
		}
	} else {
		for value, child := range node.Children {
			childCondition := fmt.Sprintf("%s = %s", node.Feature, value)
			exportTextRecursive(child, builder, depth+1, childCondition)
		}
	}
}

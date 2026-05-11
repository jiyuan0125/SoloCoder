package jsonpath

import "fmt"

type Match struct {
	Value interface{}
	Path  string
}

type pathBuilder struct {
	segments []string
}

func (b *pathBuilder) clone() *pathBuilder {
	segs := make([]string, len(b.segments))
	copy(segs, b.segments)
	return &pathBuilder{segments: segs}
}

func (b *pathBuilder) addKey(key string) {
	b.segments = append(b.segments, fmt.Sprintf("[%q]", key))
}

func (b *pathBuilder) addDot(key string) {
	b.segments = append(b.segments, "."+key)
}

func (b *pathBuilder) addIndex(idx int) {
	b.segments = append(b.segments, fmt.Sprintf("[%d]", idx))
}

func (b *pathBuilder) build() string {
	result := "$"
	for _, s := range b.segments {
		result += s
	}
	return result
}

func Evaluate(root interface{}, path string) ([]Match, error) {
	parser := NewParser(path)
	parsed, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	matches := []Match{{Value: root, Path: "$"}}
	for _, seg := range parsed.Segments {
		matches = evalSegment(matches, seg, root)
		if len(matches) == 0 {
			return []Match{}, nil
		}
	}
	return matches, nil
}

func evalSegment(matches []Match, seg Segment, root interface{}) []Match {
	out := []Match{}
	for _, m := range matches {
		out = append(out, applySegment(m, seg, root)...)
	}
	return out
}

func applySegment(m Match, seg Segment, root interface{}) []Match {
	switch s := seg.(type) {
	case *DotSegment:
		return applyDot(m, s.Name)
	case *DotDotSegment:
		return applyDotDot(m, s.Name)
	case *BracketSegment:
		out := []Match{}
		for _, sel := range s.Selectors {
			out = append(out, applySelector(m, sel, root)...)
		}
		return out
	}
	return nil
}

func applyDot(m Match, name string) []Match {
	if obj, ok := m.Value.(map[string]interface{}); ok {
		if v, ok := obj[name]; ok {
			b := pathBuilder{segments: parsePathSegments(m.Path)}
			b.addDot(name)
			return []Match{{Value: v, Path: b.build()}}
		}
	}
	return nil
}

func applyDotDot(m Match, name string) []Match {
	result := []Match{}
	collectDotDot(m, name, &result)
	return result
}

func collectDotDot(m Match, name string, out *[]Match) {
	b := &pathBuilder{segments: parsePathSegments(m.Path)}
	visit(m.Value, b, func(v interface{}, pb *pathBuilder, key string) bool {
		if key == name || (name == "" && key != "") {
			newPB := pb.clone()
			newPB.addDot(key)
			*out = append(*out, Match{Value: v, Path: newPB.build()})
		}
		return true
	})
}

func collectDescendants(node interface{}, name string) []interface{} {
	out := []interface{}{}
	visitSimple(node, func(v interface{}, key string) bool {
		if key == name || (name == "" && key != "") {
			out = append(out, v)
		}
		return true
	})
	return out
}

func visit(node interface{}, pb *pathBuilder, fn func(v interface{}, pb *pathBuilder, key string) bool) {
	switch n := node.(type) {
	case map[string]interface{}:
		for k, v := range n {
			if !fn(v, pb, k) {
				return
			}
			childPB := pb.clone()
			childPB.addKey(k)
			visit(v, childPB, fn)
		}
	case []interface{}:
		for i, v := range n {
			if !fn(v, pb, fmt.Sprintf("%d", i)) {
				return
			}
			childPB := pb.clone()
			childPB.addIndex(i)
			visit(v, childPB, fn)
		}
	}
}

func visitSimple(node interface{}, fn func(v interface{}, key string) bool) {
	switch n := node.(type) {
	case map[string]interface{}:
		for k, v := range n {
			if !fn(v, k) {
				return
			}
			visitSimple(v, fn)
		}
	case []interface{}:
		for _, v := range n {
			if !fn(v, "") {
				return
			}
			visitSimple(v, fn)
		}
	}
}

func applySelector(m Match, sel Selector, root interface{}) []Match {
	switch s := sel.(type) {
	case *WildcardSelector:
		return applyWildcard(m)
	case *NameSelector:
		return applyName(m, s.Name)
	case *IndexSelector:
		return applyIndex(m, s.Index)
	case *SliceSelector:
		return applySlice(m, s)
	case *FilterSelector:
		return applyFilter(m, s.Expr, root)
	}
	return nil
}

func applyWildcard(m Match) []Match {
	out := []Match{}
	b := pathBuilder{segments: parsePathSegments(m.Path)}
	switch v := m.Value.(type) {
	case map[string]interface{}:
		for k, val := range v {
			childB := b.clone()
			childB.addKey(k)
			out = append(out, Match{Value: val, Path: childB.build()})
		}
	case []interface{}:
		for i, val := range v {
			childB := b.clone()
			childB.addIndex(i)
			out = append(out, Match{Value: val, Path: childB.build()})
		}
	}
	return out
}

func applyName(m Match, name string) []Match {
	if obj, ok := m.Value.(map[string]interface{}); ok {
		if v, ok := obj[name]; ok {
			b := pathBuilder{segments: parsePathSegments(m.Path)}
			b.addKey(name)
			return []Match{{Value: v, Path: b.build()}}
		}
	}
	return nil
}

func applyIndex(m Match, idx int) []Match {
	if arr, ok := m.Value.([]interface{}); ok {
		i := idx
		if i < 0 {
			i = len(arr) + i
		}
		if i >= 0 && i < len(arr) {
			b := pathBuilder{segments: parsePathSegments(m.Path)}
			b.addIndex(i)
			return []Match{{Value: arr[i], Path: b.build()}}
		}
	}
	return nil
}

func applySlice(m Match, sel *SliceSelector) []Match {
	arr, ok := m.Value.([]interface{})
	if !ok {
		return nil
	}
	n := len(arr)
	start := 0
	if sel.Start != nil {
		start = *sel.Start
		if start < 0 {
			start = n + start
		}
	}
	end := n
	if sel.End != nil {
		end = *sel.End
		if end < 0 {
			end = n + end
		}
	}
	step := 1
	if sel.Step != nil {
		step = *sel.Step
		if step == 0 {
			return nil
		}
	}

	out := []Match{}
	if step > 0 {
		for i := start; i < end; i += step {
			if i >= 0 && i < n {
				b := pathBuilder{segments: parsePathSegments(m.Path)}
				b.addIndex(i)
				out = append(out, Match{Value: arr[i], Path: b.build()})
			}
		}
	} else {
		for i := start; i > end; i += step {
			if i >= 0 && i < n {
				b := pathBuilder{segments: parsePathSegments(m.Path)}
				b.addIndex(i)
				out = append(out, Match{Value: arr[i], Path: b.build()})
			}
		}
	}
	return out
}

func applyFilter(m Match, expr Expression, root interface{}) []Match {
	arr, ok := m.Value.([]interface{})
	if !ok {
		obj, isObj := m.Value.(map[string]interface{})
		if !isObj {
			return nil
		}
		out := []Match{}
		b := pathBuilder{segments: parsePathSegments(m.Path)}
		for k, v := range obj {
			result, err := expr.Eval(v, root)
			if err == nil && truthy(result) {
				childB := b.clone()
				childB.addKey(k)
				out = append(out, Match{Value: v, Path: childB.build()})
			}
		}
		return out
	}
	out := []Match{}
	b := pathBuilder{segments: parsePathSegments(m.Path)}
	for i, v := range arr {
		result, err := expr.Eval(v, root)
		if err == nil && truthy(result) {
			childB := b.clone()
			childB.addIndex(i)
			out = append(out, Match{Value: arr[i], Path: childB.build()})
		}
	}
	return out
}

func parsePathSegments(p string) []string {
	if p == "$" {
		return []string{}
	}
	rest := p[1:]
	return splitPathParts(rest)
}

func splitPathParts(s string) []string {
	parts := []string{}
	i := 0
	for i < len(s) {
		if s[i] == '.' {
			j := i + 1
			for j < len(s) && s[j] != '.' && s[j] != '[' {
				j++
			}
			parts = append(parts, s[i:j])
			i = j
		} else if s[i] == '[' {
			depth := 1
			j := i + 1
			for j < len(s) && depth > 0 {
				if s[j] == '[' {
					depth++
				} else if s[j] == ']' {
					depth--
				}
				j++
			}
			parts = append(parts, s[i:j])
			i = j
		} else {
			i++
		}
	}
	return parts
}

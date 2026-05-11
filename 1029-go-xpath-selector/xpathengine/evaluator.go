package xpathengine

import (
	"strconv"
	"strings"
)

type Evaluator struct {
	nsPrefixes map[string]string
}

func NewEvaluator() *Evaluator {
	return &Evaluator{
		nsPrefixes: make(map[string]string),
	}
}

func (e *Evaluator) SetNamespacePrefix(prefix string, uri string) {
	e.nsPrefixes[prefix] = uri
}

func (e *Evaluator) Evaluate(expr Expr, context *Node) (*XPathResult, error) {
	contextList := NodeSet{context}
	result, err := e.evaluateWithContext(expr, contextList, 0)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *Evaluator) evaluateWithContext(expr Expr, contextList NodeSet, contextPos int) (*XPathResult, error) {
	var currentNodeList NodeSet
	if len(contextList) > 0 && contextPos < len(contextList) {
		currentNodeList = NodeSet{contextList[contextPos]}
	} else {
		currentNodeList = NodeSet{}
	}
	
	switch ex := expr.(type) {
	case *LocationPath:
		return e.evaluateLocationPath(ex, currentNodeList)
	
	case *PathExpr:
		filterResult, err := e.evaluateWithContext(ex.Filter, contextList, contextPos)
		if err != nil {
			return nil, err
		}
		filterNodes, ok := filterResult.Value.(NodeSet)
		if !ok {
			return nil, nil
		}
		if ex.Relative != nil {
			return e.evaluateLocationPath(ex.Relative, filterNodes)
		}
		return filterResult, nil
	
	case *LiteralExpr:
		return &XPathResult{Type: StringResult, Value: ex.Value}, nil
	
	case *NumberExpr:
		return &XPathResult{Type: NumberResult, Value: ex.Value}, nil
	
	case *BinaryOpExpr:
		return e.evaluateBinaryOp(ex, contextList, contextPos)
	
	case *FunctionCallExpr:
		return e.evaluateFunctionCall(ex, contextList, contextPos)
	
	case *UnaryExpr:
		valResult, err := e.evaluateWithContext(ex.Value, contextList, contextPos)
		if err != nil {
			return nil, err
		}
		num, ok := e.toNumber(valResult)
		if !ok {
			return nil, nil
		}
		return &XPathResult{Type: NumberResult, Value: -num}, nil
	
	default:
		return nil, nil
	}
}

func (e *Evaluator) evaluateLocationPath(path *LocationPath, contextList NodeSet) (*XPathResult, error) {
	var currentNodes NodeSet
	if path.Absolute {
		if len(contextList) == 0 {
			return &XPathResult{Type: NodeSetResult, Value: NodeSet{}}, nil
		}
		
		root := contextList[0]
		for root.Parent != nil {
			root = root.Parent
		}
		currentNodes = NodeSet{root}
	} else {
		currentNodes = contextList
	}
	
	for _, step := range path.Steps {
		currentNodes = e.evaluateStep(step, currentNodes)
	}
	
	return &XPathResult{Type: NodeSetResult, Value: currentNodes}, nil
}

func (e *Evaluator) evaluateStep(step *Step, nodes NodeSet) NodeSet {
	var result NodeSet
	
	for _, node := range nodes {
		axisNodes := e.getAxisNodes(step.Axis, node)
		matched := e.filterByNodeTest(axisNodes, step.NodeTest)
		filtered := e.applyPredicates(matched, step.Predicates)
		result = appendUnique(result, filtered)
	}
	
	return result
}

func (e *Evaluator) getAxisNodes(axis AxisType, node *Node) NodeSet {
	var result NodeSet
	
	switch axis {
	case AxisChild:
		for _, child := range node.Children {
			if child.Type == ElementNode || child.Type == TextNode {
				result = append(result, child)
			}
		}
	
	case AxisDescendant:
		result = e.getDescendants(node, false)
	
	case AxisDescendantOrSelf:
		result = e.getDescendants(node, true)
	
	case AxisParent:
		if node.Parent != nil {
			result = append(result, node.Parent)
		}
	
	case AxisSelf:
		result = append(result, node)
	
	case AxisAttribute:
		result = append(result, node.Attributes...)
	
	case AxisFollowingSibling:
		if node.Parent != nil {
			found := false
			for _, sibling := range node.Parent.Children {
				if sibling == node {
					found = true
					continue
				}
				if found && sibling.Type == ElementNode {
					result = append(result, sibling)
				}
			}
		}
	}
	
	return result
}

func (e *Evaluator) getDescendants(node *Node, includeSelf bool) NodeSet {
	var result NodeSet
	if includeSelf {
		result = append(result, node)
	}
	
	var visit func(n *Node)
	visit = func(n *Node) {
		for _, child := range n.Children {
			if child.Type == ElementNode {
				result = append(result, child)
				visit(child)
			}
		}
	}
	
	visit(node)
	return result
}

func (e *Evaluator) filterByNodeTest(nodes NodeSet, test *NodeTest) NodeSet {
	var result NodeSet
	
	for _, node := range nodes {
		if e.matchesNodeTest(node, test) {
			result = append(result, node)
		}
	}
	
	return result
}

func (e *Evaluator) matchesNodeTest(node *Node, test *NodeTest) bool {
	switch test.Type {
	case NodeTestAll:
		return node.Type == ElementNode
	
	case NodeTestText:
		return node.Type == TextNode
	
	case NodeTestNode:
		return true
	
	case NodeTestName:
		if test.IsWild {
			if node.Type == ElementNode {
				if test.QName.Prefix == "" {
					return true
				}
				uri, hasPrefix := e.nsPrefixes[test.QName.Prefix]
				return hasPrefix && node.Namespace == uri
			}
			if node.Type == AttributeNode {
				if test.QName.Prefix == "" {
					return true
				}
				uri, hasPrefix := e.nsPrefixes[test.QName.Prefix]
				return hasPrefix && node.Namespace == uri
			}
			return false
		}
		
		if node.Type == ElementNode || node.Type == AttributeNode {
			if node.Name.Local != test.QName.Local {
				return false
			}
			
			if test.QName.Prefix != "" {
				uri, hasPrefix := e.nsPrefixes[test.QName.Prefix]
				if !hasPrefix {
					return false
				}
				return node.Namespace == uri
			}
			
			if node.Namespace != "" {
				return false
			}
			
			return true
		}
		return false
	
	default:
		return false
	}
}

func (e *Evaluator) applyPredicates(nodes NodeSet, predicates []*Predicate) NodeSet {
	result := nodes
	
	for _, pred := range predicates {
		var filtered NodeSet
		
		for i, node := range result {
			predResult, err := e.evaluateWithContext(pred.Expr, result, i)
			if err != nil {
				continue
			}
			
			if e.isTrue(predResult, result, i) {
				filtered = append(filtered, node)
			}
		}
		
		result = filtered
	}
	
	return result
}

func (e *Evaluator) isTrue(result *XPathResult, contextList NodeSet, index int) bool {
	if result.Type == NumberResult {
		num := result.Value.(float64)
		return num == float64(index+1)
	}
	
	if result.Type == BooleanResult {
		return result.Value.(bool)
	}
	
	if result.Type == StringResult {
		return len(result.Value.(string)) > 0
	}
	
	if result.Type == NodeSetResult {
		ns := result.Value.(NodeSet)
		return len(ns) > 0
	}
	
	return false
}

func (e *Evaluator) evaluateBinaryOp(op *BinaryOpExpr, contextList NodeSet, contextPos int) (*XPathResult, error) {
	leftResult, err := e.evaluateWithContext(op.Left, contextList, contextPos)
	if err != nil {
		return nil, err
	}
	rightResult, err := e.evaluateWithContext(op.Right, contextList, contextPos)
	if err != nil {
		return nil, err
	}
	
	switch op.Op {
	case TokenOr:
		leftBool := e.toBoolean(leftResult)
		rightBool := e.toBoolean(rightResult)
		return &XPathResult{Type: BooleanResult, Value: leftBool || rightBool}, nil
	
	case TokenAnd:
		leftBool := e.toBoolean(leftResult)
		rightBool := e.toBoolean(rightResult)
		return &XPathResult{Type: BooleanResult, Value: leftBool && rightBool}, nil
	
	case TokenEqual:
		return e.compareEquals(leftResult, rightResult)
	
	case TokenNotEqual:
		eqResult, err := e.compareEquals(leftResult, rightResult)
		if err != nil {
			return nil, err
		}
		return &XPathResult{Type: BooleanResult, Value: !eqResult.Value.(bool)}, nil
	
	case TokenLess, TokenLessEqual, TokenGreater, TokenGreaterEqual:
		return e.compareRelational(op.Op, leftResult, rightResult)
	
	case TokenPlus, TokenMinus, TokenAsterisk:
		leftNum, lOk := e.toNumber(leftResult)
		rightNum, rOk := e.toNumber(rightResult)
		if !lOk || !rOk {
			return nil, nil
		}
		var result float64
		switch op.Op {
		case TokenPlus:
			result = leftNum + rightNum
		case TokenMinus:
			result = leftNum - rightNum
		case TokenAsterisk:
			result = leftNum * rightNum
		}
		return &XPathResult{Type: NumberResult, Value: result}, nil
	
	case TokenPipe:
		leftNodes, lOk := leftResult.NodeSet()
		rightNodes, rOk := rightResult.NodeSet()
		if !lOk || !rOk {
			return nil, nil
		}
		return &XPathResult{Type: NodeSetResult, Value: appendUnique(leftNodes, rightNodes)}, nil
	
	default:
		return nil, nil
	}
}

func (e *Evaluator) compareEquals(left, right *XPathResult) (*XPathResult, error) {
	if left.Type == NodeSetResult && right.Type == NodeSetResult {
		leftNodes := left.Value.(NodeSet)
		rightNodes := right.Value.(NodeSet)
		for _, ln := range leftNodes {
			for _, rn := range rightNodes {
				if e.nodeEquals(ln, rn) {
					return &XPathResult{Type: BooleanResult, Value: true}, nil
				}
			}
		}
		return &XPathResult{Type: BooleanResult, Value: false}, nil
	}
	
	if left.Type == NodeSetResult {
		nodes := left.Value.(NodeSet)
		for _, node := range nodes {
			if e.nodeValueEquals(node, right) {
				return &XPathResult{Type: BooleanResult, Value: true}, nil
			}
		}
		return &XPathResult{Type: BooleanResult, Value: false}, nil
	}
	
	if right.Type == NodeSetResult {
		nodes := right.Value.(NodeSet)
		for _, node := range nodes {
			if e.nodeValueEquals(node, left) {
				return &XPathResult{Type: BooleanResult, Value: true}, nil
			}
		}
		return &XPathResult{Type: BooleanResult, Value: false}, nil
	}
	
	leftStr := e.toString(left)
	rightStr := e.toString(right)
	return &XPathResult{Type: BooleanResult, Value: leftStr == rightStr}, nil
}

func (e *Evaluator) nodeEquals(a, b *Node) bool {
	return a == b
}

func (e *Evaluator) nodeValueEquals(node *Node, value *XPathResult) bool {
	nodeVal := e.toStringNode(node)
	
	if value.Type == StringResult {
		return nodeVal == value.Value.(string)
	}
	
	if value.Type == NumberResult {
		nodeNum, ok := e.parseNumber(nodeVal)
		return ok && nodeNum == value.Value.(float64)
	}
	
	if value.Type == BooleanResult {
		nodeBool := e.toBooleanNode(node)
		return nodeBool == value.Value.(bool)
	}
	
	return false
}

func (e *Evaluator) compareRelational(op TokenType, left, right *XPathResult) (*XPathResult, error) {
	leftNum, lOk := e.toNumber(left)
	rightNum, rOk := e.toNumber(right)
	
	if !lOk || !rOk {
		return &XPathResult{Type: BooleanResult, Value: false}, nil
	}
	
	var result bool
	switch op {
	case TokenLess:
		result = leftNum < rightNum
	case TokenLessEqual:
		result = leftNum <= rightNum
	case TokenGreater:
		result = leftNum > rightNum
	case TokenGreaterEqual:
		result = leftNum >= rightNum
	}
	
	return &XPathResult{Type: BooleanResult, Value: result}, nil
}

func (e *Evaluator) evaluateFunctionCall(call *FunctionCallExpr, contextList NodeSet, contextPos int) (*XPathResult, error) {
	funcName := strings.ToLower(call.Name.Local)
	
	switch funcName {
	case "text":
		if len(contextList) == 0 || contextPos >= len(contextList) {
			return &XPathResult{Type: NodeSetResult, Value: NodeSet{}}, nil
		}
		currentNode := contextList[contextPos]
		var texts NodeSet
		for _, child := range currentNode.Children {
			if child.Type == TextNode {
				texts = append(texts, child)
			}
		}
		return &XPathResult{Type: NodeSetResult, Value: texts}, nil
	
	case "string":
		var argResult *XPathResult
		var err error
		if len(call.Args) == 0 {
			if len(contextList) > 0 && contextPos < len(contextList) {
				argResult = &XPathResult{Type: NodeSetResult, Value: NodeSet{contextList[contextPos]}}
			} else {
				argResult = &XPathResult{Type: StringResult, Value: ""}
			}
		} else {
			argResult, err = e.evaluateWithContext(call.Args[0], contextList, contextPos)
			if err != nil {
				return nil, err
			}
		}
		strVal := e.toString(argResult)
		return &XPathResult{Type: StringResult, Value: strVal}, nil
	
	case "last":
		return &XPathResult{Type: NumberResult, Value: float64(len(contextList))}, nil
	
	case "position":
		return &XPathResult{Type: NumberResult, Value: float64(contextPos + 1)}, nil
	
	case "count":
		if len(call.Args) != 1 {
			return nil, nil
		}
		argResult, err := e.evaluateWithContext(call.Args[0], contextList, contextPos)
		if err != nil {
			return nil, err
		}
		nodes, ok := argResult.NodeSet()
		if !ok {
			return &XPathResult{Type: NumberResult, Value: 0}, nil
		}
		return &XPathResult{Type: NumberResult, Value: float64(len(nodes))}, nil
	
	case "name":
		if len(contextList) == 0 || contextPos >= len(contextList) {
			return &XPathResult{Type: StringResult, Value: ""}, nil
		}
		node := contextList[contextPos]
		return &XPathResult{Type: StringResult, Value: node.Name.Local}, nil
	
	default:
		return nil, nil
	}
}

func (e *Evaluator) toBoolean(result *XPathResult) bool {
	switch result.Type {
	case BooleanResult:
		return result.Value.(bool)
	case NumberResult:
		return result.Value.(float64) != 0
	case StringResult:
		return len(result.Value.(string)) > 0
	case NodeSetResult:
		ns := result.Value.(NodeSet)
		return len(ns) > 0
	default:
		return false
	}
}

func (e *Evaluator) toBooleanNode(node *Node) bool {
	return len(node.StringValue()) > 0
}

func (e *Evaluator) toString(result *XPathResult) string {
	switch result.Type {
	case StringResult:
		return result.Value.(string)
	case NumberResult:
		num := result.Value.(float64)
		if num == float64(int64(num)) {
			return strconv.FormatInt(int64(num), 10)
		}
		return strconv.FormatFloat(num, 'g', -1, 64)
	case BooleanResult:
		if result.Value.(bool) {
			return "true"
		}
		return "false"
	case NodeSetResult:
		ns := result.Value.(NodeSet)
		if len(ns) == 0 {
			return ""
		}
		return e.toStringNode(ns[0])
	default:
		return ""
	}
}

func (e *Evaluator) toStringNode(node *Node) string {
	return node.StringValue()
}

func (e *Evaluator) toNumber(result *XPathResult) (float64, bool) {
	switch result.Type {
	case NumberResult:
		return result.Value.(float64), true
	case StringResult:
		return e.parseNumber(result.Value.(string))
	case BooleanResult:
		if result.Value.(bool) {
			return 1.0, true
		}
		return 0.0, true
	case NodeSetResult:
		ns := result.Value.(NodeSet)
		if len(ns) == 0 {
			return 0, false
		}
		return e.parseNumber(e.toStringNode(ns[0]))
	default:
		return 0, false
	}
}

func (e *Evaluator) parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false
	}
	return num, true
}

func (e *Evaluator) ToString(result *XPathResult) string {
	return e.toString(result)
}

func (e *Evaluator) ToNumber(result *XPathResult) (float64, bool) {
	return e.toNumber(result)
}

func (e *Evaluator) ToBoolean(result *XPathResult) bool {
	return e.toBoolean(result)
}

func appendUnique(a, b NodeSet) NodeSet {
	result := make(NodeSet, len(a))
	copy(result, a)
	
	for _, node := range b {
		found := false
		for _, existing := range result {
			if node == existing {
				found = true
				break
			}
		}
		if !found {
			result = append(result, node)
		}
	}
	
	return result
}

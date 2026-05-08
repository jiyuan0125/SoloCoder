package xpathengine

import (
	"errors"
)

type XPathEngine struct {
	parser    *Parser
	evaluator *Evaluator
}

func New() *XPathEngine {
	return &XPathEngine{
		parser:    NewParser(),
		evaluator: NewEvaluator(),
	}
}

func (xe *XPathEngine) SetNamespacePrefix(prefix string, uri string) {
	xe.evaluator.SetNamespacePrefix(prefix, uri)
}

func (xe *XPathEngine) Query(xmlData string, xpath string) (*XPathResult, error) {
	doc, err := xe.parser.Parse(xmlData)
	if err != nil {
		return nil, err
	}
	
	xpathParser := NewXPathParser(xpath)
	expr, err := xpathParser.Parse()
	if err != nil {
		return nil, err
	}
	
	return xe.evaluator.Evaluate(expr, doc)
}

func (xe *XPathEngine) QueryNode(xmlData string, xpath string) (*Node, error) {
	result, err := xe.Query(xmlData, xpath)
	if err != nil {
		return nil, err
	}
	
	nodes, ok := result.NodeSet()
	if !ok {
		return nil, errors.New("expression does not return a node set")
	}
	
	if len(nodes) == 0 {
		return nil, errors.New("no nodes found")
	}
	
	return nodes[0], nil
}

func (xe *XPathEngine) QueryNodes(xmlData string, xpath string) (NodeSet, error) {
	result, err := xe.Query(xmlData, xpath)
	if err != nil {
		return nil, err
	}
	
	nodes, ok := result.NodeSet()
	if !ok {
		return nil, errors.New("expression does not return a node set")
	}
	
	return nodes, nil
}

func (xe *XPathEngine) QueryString(xmlData string, xpath string) (string, error) {
	result, err := xe.Query(xmlData, xpath)
	if err != nil {
		return "", err
	}
	
	return xe.evaluator.toString(result), nil
}

func (xe *XPathEngine) QueryNumber(xmlData string, xpath string) (float64, error) {
	result, err := xe.Query(xmlData, xpath)
	if err != nil {
		return 0, err
	}
	
	num, ok := xe.evaluator.toNumber(result)
	if !ok {
		return 0, errors.New("cannot convert result to number")
	}
	
	return num, nil
}

func (xe *XPathEngine) QueryBoolean(xmlData string, xpath string) (bool, error) {
	result, err := xe.Query(xmlData, xpath)
	if err != nil {
		return false, err
	}
	
	return xe.evaluator.toBoolean(result), nil
}

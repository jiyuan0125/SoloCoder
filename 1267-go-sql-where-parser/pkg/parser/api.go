package parser

import (
	"encoding/json"

	"sqlparser/pkg/parser/ast"
	"sqlparser/pkg/parser/matcher"
)

func ParseToJSON(input string) (json.RawMessage, error) {
	node, err := Parse(input)
	if err != nil {
		return nil, err
	}
	return json.Marshal(node)
}

func MatchData(where string, data map[string]interface{}) (bool, error) {
	node, err := Parse(where)
	if err != nil {
		return false, err
	}
	return matcher.Match(node, data)
}

func MatchNode(node ast.Node, data map[string]interface{}) (bool, error) {
	return matcher.Match(node, data)
}

func Validate(input string) error {
	_, err := Parse(input)
	return err
}

func Explain(input string) (json.RawMessage, []string, error) {
	node, steps, err := ParseWithExplain(input)
	if err != nil {
		return nil, steps, err
	}
	astJSON, err := json.Marshal(node)
	return astJSON, steps, err
}

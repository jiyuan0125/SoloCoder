package core

import "trie-router/common"

type ParamType string

const (
	ParamTypeString ParamType = "string"
	ParamTypeInt    ParamType = "int"
	ParamTypeBool   ParamType = "bool"
)

type node struct {
	children   map[string]*node
	param      *node
	wildcard   *node
	handlers   map[string]*common.RouteInfo
	static     string
	paramName  string
	paramType  ParamType
	isParam    bool
	isWildcard bool
}

func newRootNode() *node {
	return newNode()
}

func newNode() *node {
	return &node{
		children: make(map[string]*node),
	}
}

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/example/xpath-selector/xpathengine"
	"github.com/example/xpath-selector/xproto"
)

type Server struct {
	port string
}

func NewServer(port string) *Server {
	return &Server{port: port}
}

func (s *Server) queryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	
	var req xproto.QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON: "+err.Error())
		return
	}
	
	if strings.TrimSpace(req.XML) == "" {
		writeError(w, http.StatusBadRequest, "XML content is required")
		return
	}
	
	if strings.TrimSpace(req.XPath) == "" {
		writeError(w, http.StatusBadRequest, "XPath expression is required")
		return
	}
	
	engine := xpathengine.New()
	
	for _, ns := range req.Namespaces {
		engine.SetNamespacePrefix(ns.Prefix, ns.URI)
	}
	
	result, err := engine.Query(req.XML, req.XPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Query error: "+err.Error())
		return
	}
	
	response := buildResponse(result, req.ReturnType)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func buildResponse(result *xpathengine.XPathResult, returnType string) *xproto.QueryResult {
	response := &xproto.QueryResult{Success: true}
	
	switch returnType {
	case xproto.ReturnTypeString:
		response.SetString(resultToString(result))
	
	case xproto.ReturnTypeNumber:
		num, ok := resultToNumber(result)
		if ok {
			response.SetNumber(num)
		} else {
			response.SetString(resultToString(result))
		}
	
	case xproto.ReturnTypeBoolean:
		response.SetBoolean(resultToBoolean(result))
	
	default:
		switch result.Type {
		case xpathengine.NodeSetResult:
			nodes, _ := result.NodeSet()
			response.ResultType = xproto.ResultTypeNodeSet
			response.Nodes = make([]xproto.NodeInfo, 0, len(nodes))
			for _, node := range nodes {
				response.Nodes = append(response.Nodes, nodeToInfo(node))
			}
		case xpathengine.StringResult:
			s, _ := result.String()
			response.SetString(s)
		case xpathengine.NumberResult:
			n, _ := result.Number()
			response.SetNumber(n)
		case xpathengine.BooleanResult:
			b, _ := result.Boolean()
			response.SetBoolean(b)
		}
	}
	
	return response
}

func resultToString(result *xpathengine.XPathResult) string {
	evaluator := xpathengine.NewEvaluator()
	return evaluator.ToString(result)
}

func resultToNumber(result *xpathengine.XPathResult) (float64, bool) {
	evaluator := xpathengine.NewEvaluator()
	return evaluator.ToNumber(result)
}

func resultToBoolean(result *xpathengine.XPathResult) bool {
	evaluator := xpathengine.NewEvaluator()
	return evaluator.ToBoolean(result)
}

func nodeToInfo(node *xpathengine.Node) xproto.NodeInfo {
	info := xproto.NodeInfo{
		LocalName: node.GetLocalName(),
		Prefix:    node.GetPrefix(),
		Namespace: node.GetNamespace(),
		Text:      node.StringValue(),
	}
	
	switch node.Type {
	case xpathengine.ElementNode:
		info.Type = xproto.NodeTypeElement
	case xpathengine.TextNode:
		info.Type = xproto.NodeTypeText
	case xpathengine.AttributeNode:
		info.Type = xproto.NodeTypeAttribute
	case xpathengine.DocumentNode:
		info.Type = xproto.NodeTypeDocument
	default:
		info.Type = "unknown"
	}
	
	if node.Type == xpathengine.ElementNode {
		info.Attributes = make(map[string]string)
		for _, attr := range node.Attributes {
			if attr.GetPrefix() != "" {
				info.Attributes[attr.GetPrefix()+":"+attr.GetLocalName()] = attr.Text
			} else {
				info.Attributes[attr.GetLocalName()] = attr.Text
			}
		}
	}
	
	return info
}

func writeError(w http.ResponseWriter, status int, message string) {
	response := &xproto.QueryResult{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func main() {
	port := "8300"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}
	
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
	
	server := NewServer(port)
	
	http.HandleFunc("/query", server.queryHandler)
	
	fmt.Printf("XPath Query Server listening on port %s\n", port)
	fmt.Printf("Endpoint: POST http://localhost%s/query\n", port)
	
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}

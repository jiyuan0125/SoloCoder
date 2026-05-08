package search

import (
	"fmt"
	"sort"
	"strings"
)

type SearchResult struct {
	DocID string
	Score float64
}

type Engine struct {
	index *InvertedIndex
}

func NewEngine() *Engine {
	return &Engine{
		index: NewInvertedIndex(),
	}
}

func (e *Engine) AddDoc(docID, content string) {
	e.index.AddDoc(docID, content)
}

func (e *Engine) DeleteDoc(docID string) {
	e.index.DeleteDoc(docID)
}

func (e *Engine) Search(query string) ([]SearchResult, error) {
	ast, err := ParseQuery(query)
	if err != nil {
		return nil, err
	}

	if err := validateQuery(ast); err != nil {
		return nil, err
	}

	docIDs, err := e.evaluateQuery(ast)
	if err != nil {
		return nil, err
	}

	results := make([]SearchResult, 0, len(docIDs))
	for docID := range docIDs {
		score := e.computeScore(docID, ast)
		results = append(results, SearchResult{DocID: docID, Score: score})
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].DocID < results[j].DocID
		}
		return results[i].Score > results[j].Score
	})

	return results, nil
}

func validateQuery(ast AstNode) error {
	switch node := ast.(type) {
	case *NotNode:
		return fmt.Errorf("NOT cannot be used independently: query must be of form 'A AND NOT B'")
	case *AndNode:
		if err := validateQuery(node.Left); err != nil {
			return err
		}
		if not, ok := node.Right.(*NotNode); ok {
			return validateQuery(not.Operand)
		}
		return validateQuery(node.Right)
	case *OrNode:
		if err := validateQuery(node.Left); err != nil {
			return err
		}
		return validateQuery(node.Right)
	case *TermNode:
		return nil
	default:
		return fmt.Errorf("unknown node type: %T", ast)
	}
}

func (e *Engine) evaluateQuery(ast AstNode) (map[string]struct{}, error) {
	switch node := ast.(type) {
	case *TermNode:
		return e.termDocs(node.Value), nil
	case *AndNode:
		return e.andDocs(node.Left, node.Right)
	case *OrNode:
		return e.orDocs(node.Left, node.Right)
	case *NotNode:
		return e.notDocs(node.Operand)
	default:
		return nil, fmt.Errorf("unknown node type: %T", ast)
	}
}

func (e *Engine) termDocs(term string) map[string]struct{} {
	docs := make(map[string]struct{})
	tokens := Tokenize(term)
	if len(tokens) == 0 {
		return docs
	}
	token := strings.ToLower(tokens[0])
	postings := e.index.GetPostings(token)
	for _, p := range postings {
		docs[p.DocID] = struct{}{}
	}
	return docs
}

func (e *Engine) andDocs(left, right AstNode) (map[string]struct{}, error) {
	leftDocs, err := e.evaluateQuery(left)
	if err != nil {
		return nil, err
	}

	if not, ok := right.(*NotNode); ok {
		excludeDocs, err := e.evaluateQuery(not.Operand)
		if err != nil {
			return nil, err
		}
		docs := make(map[string]struct{})
		for docID := range leftDocs {
			if _, excluded := excludeDocs[docID]; !excluded {
				docs[docID] = struct{}{}
			}
		}
		return docs, nil
	}

	rightDocs, err := e.evaluateQuery(right)
	if err != nil {
		return nil, err
	}

	docs := make(map[string]struct{})
	for docID := range leftDocs {
		if _, exists := rightDocs[docID]; exists {
			docs[docID] = struct{}{}
		}
	}
	return docs, nil
}

func (e *Engine) orDocs(left, right AstNode) (map[string]struct{}, error) {
	leftDocs, err := e.evaluateQuery(left)
	if err != nil {
		return nil, err
	}
	rightDocs, err := e.evaluateQuery(right)
	if err != nil {
		return nil, err
	}

	docs := make(map[string]struct{})
	for docID := range leftDocs {
		docs[docID] = struct{}{}
	}
	for docID := range rightDocs {
		docs[docID] = struct{}{}
	}
	return docs, nil
}

func (e *Engine) notDocs(operand AstNode) (map[string]struct{}, error) {
	return nil, fmt.Errorf("NOT cannot be used in this context")
}

func (e *Engine) computeScore(docID string, ast AstNode) float64 {
	switch node := ast.(type) {
	case *TermNode:
		tokens := Tokenize(node.Value)
		if len(tokens) == 0 {
			return 0
		}
		return e.index.ComputeTFIDF(docID, strings.ToLower(tokens[0]))
	case *AndNode:
		if _, ok := node.Right.(*NotNode); ok {
			return e.computeScore(docID, node.Left)
		}
		return e.computeScore(docID, node.Left) + e.computeScore(docID, node.Right)
	case *OrNode:
		return max(e.computeScore(docID, node.Left), e.computeScore(docID, node.Right))
	default:
		return 0
	}
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

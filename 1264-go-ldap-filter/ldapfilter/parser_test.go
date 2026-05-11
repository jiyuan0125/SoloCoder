package ldapfilter

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParseValues(t *testing.T) {
	testCases := []struct {
		name   string
		filter string
		want   string
	}{
		{"cn=John", "(cn=John)", "John"},
		{"cn=John Doe", "(cn=John Doe)", "John Doe"},
		{"mail=john@example.com", "(mail=john@example.com)", "john@example.com"},
		{"telephoneNumber=1234", "(telephoneNumber=1234)", "1234"},
		{"sn=Smith", "(sn=Smith)", "Smith"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := Parse(tc.filter)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.filter, err)
			}
			simple, ok := node.(*SimpleNode)
			if !ok {
				t.Fatalf("expected SimpleNode, got %T", node)
			}
			if simple.Value != tc.want {
				b, _ := json.MarshalIndent(ToASTJSON(node), "", "  ")
				fmt.Printf("AST: %s\n", string(b))
				t.Errorf("got value=%q, want %q", simple.Value, tc.want)
			}
		})
	}
}

func TestParseOtherOperators(t *testing.T) {
	testCases := []struct {
		name   string
		filter string
		wantOp Operator
		wantVal string
	}{
		{"approx", "(cn~=John Doe)", OpApprox, "John Doe"},
		{"ge", "(cn>=ABC)", OpGreaterEqual, "ABC"},
		{"le", "(sn<=Z)", OpLessEqual, "Z"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := Parse(tc.filter)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.filter, err)
			}
			simple, ok := node.(*SimpleNode)
			if !ok {
				t.Fatalf("expected SimpleNode, got %T", node)
			}
			if simple.Operator != tc.wantOp {
				t.Errorf("got operator=%v, want %v", simple.Operator, tc.wantOp)
			}
			if simple.Value != tc.wantVal {
				b, _ := json.MarshalIndent(ToASTJSON(node), "", "  ")
				fmt.Printf("AST: %s\n", string(b))
				t.Errorf("got value=%q, want %q", simple.Value, tc.wantVal)
			}
		})
	}
}

func TestParseWildcards(t *testing.T) {
	testCases := []struct {
		name    string
		filter  string
		initial string
		any     []string
		final   string
	}{
		{"prefix", "(cn=Joh*)", "Joh", nil, ""},
		{"suffix", "(cn=*hn)", "", nil, "hn"},
		{"contains", "(cn=*oh*)", "", []string{"oh"}, ""},
		{"middle", "(cn=J*n)", "J", nil, "n"},
		{"multiple", "(cn=J*h*n)", "J", []string{"h"}, "n"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			node, err := Parse(tc.filter)
			if err != nil {
				t.Fatalf("Parse(%q) error: %v", tc.filter, err)
			}
			sub, ok := node.(*SubstringNode)
			if !ok {
				t.Fatalf("expected SubstringNode, got %T", node)
			}
			if sub.Initial != tc.initial {
				t.Errorf("initial=%q, want %q", sub.Initial, tc.initial)
			}
			if len(sub.Any) != len(tc.any) {
				t.Errorf("any len=%d, want %d", len(sub.Any), len(tc.any))
			}
			for i, v := range sub.Any {
				if v != tc.any[i] {
					t.Errorf("any[%d]=%q, want %q", i, v, tc.any[i])
				}
			}
			if sub.Final != tc.final {
				t.Errorf("final=%q, want %q", sub.Final, tc.final)
			}
		})
	}
}

func TestParsePresent(t *testing.T) {
	node, err := Parse("(cn=*)")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	_, ok := node.(*PresentNode)
	if !ok {
		t.Fatalf("expected PresentNode, got %T", node)
	}
}

func TestParseEscape(t *testing.T) {
	node, err := Parse("(cn=A\\2aB)")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	simple, ok := node.(*SimpleNode)
	if !ok {
		t.Fatalf("expected SimpleNode, got %T", node)
	}
	expected := "A*B"
	if simple.Value != expected {
		b, _ := json.MarshalIndent(ToASTJSON(node), "", "  ")
		fmt.Printf("AST: %s\n", string(b))
		t.Errorf("got value=%q, want %q", simple.Value, expected)
	}
}

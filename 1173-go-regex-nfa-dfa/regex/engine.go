package regex

import "unicode/utf8"

type CompiledRegex struct {
	Pattern           string
	AST               ASTNode
	NFAStart          *NFAState
	NFAStateCount     int
	DFAStart          *DFAState
	DFAStateCount     int
	MinimizedDFAStart *MinimizedDFAState
	MinimizedDFACount int
}

func Compile(pattern string) (*CompiledRegex, error) {
	lexer := NewLexer(pattern)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, err
	}

	parser := NewParser(tokens)
	ast, err := parser.Parse()
	if err != nil {
		return nil, err
	}

	nfaBuilder := NewNFABuilder()
	nfaStart, nfaCount := nfaBuilder.Build(ast)

	dfaBuilder := NewDFABuilder()
	dfaStart, dfaCount := dfaBuilder.Build(nfaStart)

	minDFAStart, minCount := HopcroftMinimize(dfaStart)

	return &CompiledRegex{
		Pattern:           pattern,
		AST:               ast,
		NFAStart:          nfaStart,
		NFAStateCount:     nfaCount,
		DFAStart:          dfaStart,
		DFAStateCount:     dfaCount,
		MinimizedDFAStart: minDFAStart,
		MinimizedDFACount: minCount,
	}, nil
}

func (r *CompiledRegex) Match(text string) bool {
	if r.MinimizedDFAStart == nil {
		return r.MatchWithNFA(text)
	}
	return r.MatchWithMinimizedDFA(text)
}

func (r *CompiledRegex) MatchWithNFA(text string) bool {
	current := EpsilonClosure(map[int]*NFAState{r.NFAStart.ID: r.NFAStart})

	for len(text) > 0 {
		sym, w := utf8.DecodeRuneInString(text)
		text = text[w:]

		moveResult := Move(current, sym)
		if len(moveResult) == 0 {
			return false
		}
		current = EpsilonClosure(moveResult)
	}

	return hasFinalState(current)
}

func (r *CompiledRegex) MatchWithDFA(text string) bool {
	current := r.DFAStart

	for len(text) > 0 {
		sym, w := utf8.DecodeRuneInString(text)
		text = text[w:]

		next, ok := current.Trans[sym]
		if !ok {
			return false
		}
		current = next
	}

	return current.IsFinal
}

func (r *CompiledRegex) MatchWithMinimizedDFA(text string) bool {
	current := r.MinimizedDFAStart
	if current == nil {
		return r.MatchWithDFA(text)
	}

	for len(text) > 0 {
		sym, w := utf8.DecodeRuneInString(text)
		text = text[w:]

		next, ok := current.Trans[sym]
		if !ok {
			return false
		}
		current = next
	}

	return current.IsFinal
}

type VisualizationData struct {
	Pattern       string
	AST           string
	NFA           *NFARepresentation
	DFA           *DFARepresentation
	MinimizedDFA  *MinimizedDFARepresentation
	NFAStates     int
	DFAStates     int
	MinDFAStates  int
}

func (r *CompiledRegex) Visualize() *VisualizationData {
	return &VisualizationData{
		Pattern:       r.Pattern,
		AST:           r.AST.String(),
		NFA:           BuildNFARepresentation(r.NFAStart),
		DFA:           BuildDFARepresentation(r.DFAStart),
		MinimizedDFA:  BuildMinimizedDFARepresentation(r.MinimizedDFAStart),
		NFAStates:     r.NFAStateCount,
		DFAStates:     r.DFAStateCount,
		MinDFAStates:  r.MinimizedDFACount,
	}
}

func Match(pattern, text string) (bool, error) {
	re, err := Compile(pattern)
	if err != nil {
		return false, err
	}
	return re.Match(text), nil
}

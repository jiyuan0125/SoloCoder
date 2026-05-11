package api

type CompileRequest struct {
	Pattern string `json:"pattern"`
}

type CompileResponse struct {
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Pattern       string `json:"pattern,omitempty"`
	NFAStates     int    `json:"nfa_states,omitempty"`
	DFAStates     int    `json:"dfa_states,omitempty"`
	MinDFAStates  int    `json:"min_dfa_states,omitempty"`
}

type MatchRequest struct {
	Pattern string `json:"pattern"`
	Text    string `json:"text"`
}

type MatchResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Match   bool   `json:"match,omitempty"`
}

type VizRequest struct {
	Pattern string `json:"pattern"`
}

type NFATransitionJSON struct {
	From   int    `json:"from"`
	Symbol string `json:"symbol"`
	To     []int  `json:"to"`
}

type NFAJSON struct {
	StartState  int                `json:"start_state"`
	FinalStates []int              `json:"final_states"`
	States      int                `json:"states"`
	Transitions []NFATransitionJSON `json:"transitions"`
}

type DFATransitionJSON struct {
	From   int    `json:"from"`
	Symbol string `json:"symbol"`
	To     int    `json:"to"`
}

type DFAJSON struct {
	StartState  int                `json:"start_state"`
	FinalStates []int              `json:"final_states"`
	States      int                `json:"states"`
	Transitions []DFATransitionJSON `json:"transitions"`
}

type VizResponse struct {
	Success       bool   `json:"success"`
	Error         string `json:"error,omitempty"`
	Pattern       string `json:"pattern,omitempty"`
	AST           string `json:"ast,omitempty"`
	NFA           NFAJSON `json:"nfa,omitempty"`
	DFA           DFAJSON `json:"dfa,omitempty"`
	MinimizedDFA  DFAJSON `json:"minimized_dfa,omitempty"`
	NFAStates     int     `json:"nfa_states,omitempty"`
	DFAStates     int     `json:"dfa_states,omitempty"`
	MinDFAStates  int     `json:"min_dfa_states,omitempty"`
}

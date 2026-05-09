package regex

import "strings"

type Regex struct {
	startState      *State
	stateCount      int
	transitionCount int
}

func NewRegex(pattern string) (*Regex, error) {
	parser := newParser(pattern)
	ast, err := parser.parse()
	if err != nil {
		return nil, err
	}

	start, stateCount, transitionCount := buildNFA(ast)

	if ast == nil && pattern != "" {
		return nil, parser.error("invalid pattern")
	}

	return &Regex{
		startState:      start,
		stateCount:      stateCount,
		transitionCount: transitionCount,
	}, nil
}

func MustCompile(pattern string) *Regex {
	re, err := NewRegex(pattern)
	if err != nil {
		panic(err)
	}
	return re
}

func (r *Regex) MatchString(input string) bool {
	currentStates := epsilonClosure([]*State{r.startState})

	for i := 0; i < len(input); i++ {
		ch := input[i]
		nextStates := []*State{}

		for _, state := range currentStates {
			if targets, ok := state.charTrans[ch]; ok {
				nextStates = append(nextStates, targets...)
			}
		}

		currentStates = epsilonClosure(nextStates)

		if len(currentStates) == 0 {
			return false
		}
	}

	for _, state := range currentStates {
		if state.isAccept {
			return true
		}
	}

	return false
}

func (r *Regex) FindAllString(input string) []string {
	var results []string
	
	for start := 0; start <= len(input); start++ {
		for end := start; end <= len(input); end++ {
			if r.MatchString(input[start:end]) {
				if start < end {
					results = append(results, input[start:end])
					for end < len(input) && r.MatchString(input[start:end+1]) {
						end++
					}
					start = end - 1
				}
				break
			}
		}
	}
	
	return results
}

func (r *Regex) FindAllStringIndex(input string) [][]int {
	var results [][]int
	
	for start := 0; start <= len(input); start++ {
		for end := start; end <= len(input); end++ {
			if r.MatchString(input[start:end]) {
				if start < end {
					results = append(results, []int{start, end})
					for end < len(input) && r.MatchString(input[start:end+1]) {
						end++
					}
					start = end - 1
				}
				break
			}
		}
	}
	
	return results
}

func epsilonClosure(states []*State) []*State {
	closure := make(map[*State]bool)
	var stack []*State

	for _, state := range states {
		if !closure[state] {
			closure[state] = true
			stack = append(stack, state)
		}
	}

	for len(stack) > 0 {
		state := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		for _, next := range state.epsilonTrans {
			if !closure[next] {
				closure[next] = true
				stack = append(stack, next)
			}
		}
	}

	result := make([]*State, 0, len(closure))
	for state := range closure {
		result = append(result, state)
	}

	return result
}

func (r *Regex) StateCount() int {
	return r.stateCount
}

func (r *Regex) TransitionCount() int {
	return r.transitionCount
}

func (r *Regex) Visualize() string {
	var builder strings.Builder

	states := collectAllStates(r.startState)
	stateMap := make(map[*State]int)
	
	for i, state := range states {
		stateMap[state] = i
	}

	builder.WriteString("NFA State Transition Graph:\n")
	builder.WriteString("===========================\n\n")
	
	for i, state := range states {
		acceptStr := ""
		if state.isAccept {
			acceptStr = " [ACCEPT]"
		}
		
		builder.WriteString(string("State "))
		builder.WriteString(string(rune('0' + i)))
		builder.WriteString(acceptStr)
		builder.WriteString(":\n")

		if len(state.epsilonTrans) > 0 {
			builder.WriteString("  ε-transitions to: ")
			for j, next := range state.epsilonTrans {
				if j > 0 {
					builder.WriteString(", ")
				}
				builder.WriteString(string("State "))
				builder.WriteString(string(rune('0' + stateMap[next])))
			}
			builder.WriteString("\n")
		}

		if len(state.charTrans) > 0 {
			builder.WriteString("  char-transitions:\n")
			for ch, targets := range state.charTrans {
				builder.WriteString("    '")
				if ch == '\n' {
					builder.WriteString("\\n")
				} else if ch == '\t' {
					builder.WriteString("\\t")
				} else if ch == '\r' {
					builder.WriteString("\\r")
				} else if ch < 32 || ch >= 127 {
					builder.WriteString(string(rune('0' + (ch/100)%10)))
					builder.WriteString(string(rune('0' + (ch/10)%10)))
					builder.WriteString(string(rune('0' + ch%10)))
				} else {
					builder.WriteString(string(ch))
				}
				builder.WriteString("' to: ")
				for j, target := range targets {
					if j > 0 {
						builder.WriteString(", ")
					}
					builder.WriteString(string("State "))
					builder.WriteString(string(rune('0' + stateMap[target])))
				}
				builder.WriteString("\n")
			}
		}

		builder.WriteString("\n")
	}

	return builder.String()
}

package regex

type State struct {
	id          int
	epsilonTrans []*State
	charTrans   map[byte][]*State
	isAccept    bool
}

type Fragment struct {
	start *State
	out   []*State
}

var nextStateID = 0

func newState() *State {
	s := &State{
		id:        nextStateID,
		charTrans: make(map[byte][]*State),
	}
	nextStateID++
	return s
}

func newFragment(start *State, out []*State) *Fragment {
	return &Fragment{start: start, out: out}
}

func connectOuts(frag *Fragment, state *State) {
	for _, out := range frag.out {
		out.epsilonTrans = append(out.epsilonTrans, state)
	}
}

func buildNFA(node *Node) (*State, int, int) {
	if node == nil {
		accept := newState()
		accept.isAccept = true
		return accept, 1, 0
	}
	return buildNFAFromNode(node)
}

func buildNFAFromNode(node *Node) (*State, int, int) {
	stateCount := 0
	transitionCount := 0

	switch node.typ {
	case NodeEmpty:
		start := newState()
		accept := newState()
		accept.isAccept = true
		start.epsilonTrans = append(start.epsilonTrans, accept)
		stateCount = 2
		transitionCount = 1
		return start, stateCount, transitionCount

	case NodeLiteral:
		start := newState()
		accept := newState()
		accept.isAccept = true
		start.charTrans[node.value] = append(start.charTrans[node.value], accept)
		stateCount = 2
		transitionCount = 1
		return start, stateCount, transitionCount

	case NodeDot:
		start := newState()
		accept := newState()
		accept.isAccept = true
		for ch := 0; ch < 256; ch++ {
			if ch != '\n' {
				start.charTrans[byte(ch)] = append(start.charTrans[byte(ch)], accept)
				transitionCount++
			}
		}
		stateCount = 2
		return start, stateCount, transitionCount

	case NodeConcat:
		left, leftStates, leftTrans := buildNFAFromNode(node.left)
		right, rightStates, rightTrans := buildNFAFromNode(node.right)

		acceptStates := collectAcceptStates(left)
		nonAcceptStates := []*State{}
		for _, s := range acceptStates {
			s.isAccept = false
			nonAcceptStates = append(nonAcceptStates, s)
		}
		
		for _, s := range nonAcceptStates {
			s.epsilonTrans = append(s.epsilonTrans, right)
			transitionCount++
		}

		stateCount = leftStates + rightStates
		transitionCount = leftTrans + rightTrans + len(nonAcceptStates)
		return left, stateCount, transitionCount

	case NodeAlternate:
		left, leftStates, leftTrans := buildNFAFromNode(node.left)
		right, rightStates, rightTrans := buildNFAFromNode(node.right)

		start := newState()
		accept := newState()
		accept.isAccept = true

		start.epsilonTrans = append(start.epsilonTrans, left, right)

		leftAccepts := collectAcceptStates(left)
		for _, s := range leftAccepts {
			s.isAccept = false
			s.epsilonTrans = append(s.epsilonTrans, accept)
		}

		rightAccepts := collectAcceptStates(right)
		for _, s := range rightAccepts {
			s.isAccept = false
			s.epsilonTrans = append(s.epsilonTrans, accept)
		}

		stateCount = leftStates + rightStates + 2
		transitionCount = leftTrans + rightTrans + 2 + len(leftAccepts) + len(rightAccepts)
		return start, stateCount, transitionCount

	case NodeStar:
		child, childStates, childTrans := buildNFAFromNode(node.child)

		start := newState()
		accept := newState()
		accept.isAccept = true

		childAccepts := collectAcceptStates(child)
		for _, s := range childAccepts {
			s.isAccept = false
			s.epsilonTrans = append(s.epsilonTrans, child, accept)
		}

		start.epsilonTrans = append(start.epsilonTrans, child, accept)

		stateCount = childStates + 2
		transitionCount = childTrans + 2 + 2*len(childAccepts)
		return start, stateCount, transitionCount
	}

	return nil, 0, 0
}

func collectAcceptStates(start *State) []*State {
	visited := make(map[*State]bool)
	var accepts []*State
	var queue []*State
	queue = append(queue, start)
	visited[start] = true

	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]

		if s.isAccept {
			accepts = append(accepts, s)
		}

		for _, t := range s.epsilonTrans {
			if !visited[t] {
				visited[t] = true
				queue = append(queue, t)
			}
		}

		for _, targets := range s.charTrans {
			for _, t := range targets {
				if !visited[t] {
					visited[t] = true
					queue = append(queue, t)
				}
			}
		}
	}

	return accepts
}

func collectAllStates(start *State) []*State {
	visited := make(map[*State]bool)
	var states []*State
	var queue []*State
	queue = append(queue, start)
	visited[start] = true

	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]
		states = append(states, s)

		for _, t := range s.epsilonTrans {
			if !visited[t] {
				visited[t] = true
				queue = append(queue, t)
			}
		}

		for _, targets := range s.charTrans {
			for _, t := range targets {
				if !visited[t] {
					visited[t] = true
					queue = append(queue, t)
				}
			}
		}
	}

	return states
}

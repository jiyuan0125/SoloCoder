package fst

func (f *FST) ExactSearch(key string) (uint64, bool) {
	if f.Root == nil {
		return 0, false
	}

	current := f.Root
	runes := []rune(key)

	for _, char := range runes {
		if trans, found := current.getTransition(char); found {
			current = trans.Next
		} else {
			return 0, false
		}
	}

	if current.IsFinal {
		return current.FinalValue, true
	}
	return 0, false
}

func (f *FST) PrefixSearch(prefix string) []SearchResult {
	results := make([]SearchResult, 0)
	if f.Root == nil {
		return results
	}

	current := f.Root
	runes := []rune(prefix)

	for _, char := range runes {
		if trans, found := current.getTransition(char); found {
			current = trans.Next
		} else {
			return results
		}
	}

	f.collectAll(current, prefix, &results)
	return results
}

func (f *FST) collectAll(state *State, currentKey string, results *[]SearchResult) {
	if state.IsFinal {
		*results = append(*results, SearchResult{
			Key:   currentKey,
			Value: state.FinalValue,
			Found: true,
		})
	}

	for _, trans := range state.Transitions {
		f.collectAll(trans.Next, currentKey+string(trans.Char), results)
	}
}

func (f *FST) Walk(visit func(key string, value uint64) error) error {
	if f.Root == nil {
		return nil
	}
	return f.walkDFS(f.Root, "", visit)
}

func (f *FST) walkDFS(state *State, currentKey string, visit func(key string, value uint64) error) error {
	if state.IsFinal {
		if err := visit(currentKey, state.FinalValue); err != nil {
			return err
		}
	}

	for _, trans := range state.Transitions {
		if err := f.walkDFS(trans.Next, currentKey+string(trans.Char), visit); err != nil {
			return err
		}
	}
	return nil
}

func (f *FST) Len() int {
	return f.Count
}

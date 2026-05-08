package filter

import (
	"sensitive-word-filter/pkg/common"
	"sort"
	"strings"
	"unicode"
)

func NewFilter() *Filter {
	return &Filter{
		trie:          newTrieNode(),
		wildcards:     []WildcardPattern{},
		whitelist:     make(map[string]bool),
		whitelistTrie: newTrieNode(),
		dailyStats:    make(map[string]*DailyCount),
		hitStats:      make(map[string]int64),
	}
}

func (f *Filter) Filter(text string) *common.FilterResponse {
	f.mu.RLock()
	defer f.mu.RUnlock()

	dateKey := getDateKey()
	if f.dailyStats[dateKey] == nil {
		f.dailyStats[dateKey] = &DailyCount{}
	}
	f.dailyStats[dateKey].Total++

	textRunes := []rune(text)

	whitelistRanges := f.findWhitelistRanges(textRunes)

	normalResults := f.matchNormal(textRunes, whitelistRanges)
	skipResults := f.matchWithSkip(textRunes, whitelistRanges)
	wildcardResults := f.matchWildcards(textRunes, whitelistRanges)

	allResults := append(normalResults, skipResults...)
	allResults = append(allResults, wildcardResults...)

	allResults = f.deduplicateAndPrioritize(allResults)

	response := &common.FilterResponse{
		Text: text,
		Hit:  len(allResults) > 0,
	}

	for _, r := range allResults {
		response.Hits = append(response.Hits, common.HitResult{
			SensitiveWord: r.word,
			MatchType:     r.matchType,
			Positions:     r.positions,
		})

		f.hitStats[r.word]++
	}

	if response.Hit {
		f.dailyStats[dateKey].HitCount++
	}

	return response
}

func (f *Filter) findWhitelistRanges(textRunes []rune) [][2]int {
	var ranges [][2]int
	whitelistWords := f.whitelistTrie.listWords()

	for _, word := range whitelistWords {
		wordRunes := []rune(word)
		if len(wordRunes) == 0 {
			continue
		}

		for i := 0; i <= len(textRunes)-len(wordRunes); i++ {
			match := true
			for j, ch := range wordRunes {
				if textRunes[i+j] != ch {
					match = false
					break
				}
			}
			if match {
				ranges = append(ranges, [2]int{i, i + len(wordRunes) - 1})
			}
		}
	}

	return ranges
}

func (f *Filter) isInWhitelistRange(idx int, ranges [][2]int) bool {
	for _, r := range ranges {
		if idx >= r[0] && idx <= r[1] {
			return true
		}
	}
	return false
}

func (f *Filter) matchNormal(textRunes []rune, whitelistRanges [][2]int) []matchResult {
	var results []matchResult

	for i := 0; i < len(textRunes); i++ {
		if f.isInWhitelistRange(i, whitelistRanges) {
			continue
		}

		current := f.trie
		for j := i; j < len(textRunes); j++ {
			ch := textRunes[j]
			child, exists := current.children[ch]
			if !exists {
				break
			}
			current = child

			if current.isEnd {
				word := current.word
				positions := []common.Position{{
					Start:  runeIndexToByteIndex(string(textRunes), i),
					Length: runeIndexToByteIndex(string(textRunes), j+1) - runeIndexToByteIndex(string(textRunes), i),
				}}
				results = append(results, matchResult{
					word:      word,
					matchType: common.DirectMatch,
					positions: positions,
					startIdx:  i,
					endIdx:    j,
				})
			}
		}
	}

	return results
}

func (f *Filter) matchWithSkip(textRunes []rune, whitelistRanges [][2]int) []matchResult {
	var results []matchResult

	for startIdx := 0; startIdx < len(textRunes); startIdx++ {
		if f.isInWhitelistRange(startIdx, whitelistRanges) {
			continue
		}

		startCh := textRunes[startIdx]
		current, exists := f.trie.children[startCh]
		if !exists {
			continue
		}

		pattern := current.word
		if pattern == "" {
			f.traverseSkip(textRunes, startIdx, startIdx, current, whitelistRanges, &results)
		} else {
			positions := []common.Position{{
				Start:  runeIndexToByteIndex(string(textRunes), startIdx),
				Length: runeLenToByteLen(startCh),
			}}
			results = append(results, matchResult{
				word:      pattern,
				matchType: common.DirectMatch,
				positions: positions,
				startIdx:  startIdx,
				endIdx:    startIdx,
			})
			f.traverseSkip(textRunes, startIdx, startIdx, current, whitelistRanges, &results)
		}
	}

	return results
}

func (f *Filter) traverseSkip(
	textRunes []rune,
	patternIdx int,
	textIdx int,
	node *TrieNode,
	whitelistRanges [][2]int,
	results *[]matchResult,
) {
	patternWord := node.word
	if patternWord != "" && patternIdx == len([]rune(patternWord))-1 {
		positions := collectSkipPositions(textRunes, patternWord, textIdx)
		if len(positions) > 0 {
			matchType := common.SkipMatch
			if len(positions) == 1 {
				matchType = common.DirectMatch
			}
			*results = append(*results, matchResult{
				word:      patternWord,
				matchType: matchType,
				positions: positions,
				startIdx:  findStartIdx(textRunes, patternWord, textIdx),
				endIdx:    textIdx,
			})
		}
	}

	for nextCh, childNode := range node.children {
		for j := textIdx + 1; j < len(textRunes); j++ {
			if f.isInWhitelistRange(j, whitelistRanges) {
				continue
			}

			ch := textRunes[j]
			if ch == nextCh {
				f.traverseSkip(textRunes, patternIdx+1, j, childNode, whitelistRanges, results)
				break
			} else if isInterference(ch) {
				continue
			} else {
				break
			}
		}
	}
}

func collectSkipPositions(textRunes []rune, pattern string, endTextIdx int) []common.Position {
	patternRunes := []rune(pattern)
	if len(patternRunes) == 0 {
		return nil
	}

	var positions []common.Position
	currentTextIdx := endTextIdx

	for i := len(patternRunes) - 1; i >= 0 && currentTextIdx >= 0; i-- {
		target := patternRunes[i]

		for currentTextIdx >= 0 {
			ch := textRunes[currentTextIdx]
			if ch == target {
				byteStart := runeIndexToByteIndex(string(textRunes), currentTextIdx)
				byteLen := runeLenToByteLen(ch)
				positions = append([]common.Position{{
					Start:  byteStart,
					Length: byteLen,
				}}, positions...)
				currentTextIdx--
				break
			} else if isInterference(ch) {
				currentTextIdx--
			} else {
				return nil
			}
		}
	}

	if len(positions) == len(patternRunes) {
		return positions
	}
	return nil
}

func findStartIdx(textRunes []rune, pattern string, endIdx int) int {
	patternRunes := []rune(pattern)
	if len(patternRunes) == 0 {
		return 0
	}

	current := endIdx
	for i := len(patternRunes) - 1; i >= 0 && current >= 0; i-- {
		target := patternRunes[i]
		for current >= 0 {
			if textRunes[current] == target {
				if i == 0 {
					return current
				}
				current--
				break
			} else if isInterference(textRunes[current]) {
				current--
			} else {
				return 0
			}
		}
	}
	return 0
}

func (f *Filter) matchWildcards(textRunes []rune, whitelistRanges [][2]int) []matchResult {
	var results []matchResult

	for _, pattern := range f.wildcards {
		if len(pattern.parts) < 2 {
			continue
		}

		firstPart := pattern.parts[0]
		firstPartRunes := []rune(firstPart)

		for i := 0; i <= len(textRunes)-len(firstPartRunes); i++ {
			if f.isInWhitelistRange(i, whitelistRanges) {
				continue
			}

			match := true
			for j, ch := range firstPartRunes {
				if textRunes[i+j] != ch {
					match = false
					break
				}
			}
			if !match {
				continue
			}

			currentIdx := i + len(firstPartRunes)
			var positions []common.Position
			positions = append(positions, common.Position{
				Start:  runeIndexToByteIndex(string(textRunes), i),
				Length: runeIndexToByteIndex(string(textRunes), i+len(firstPartRunes)) - runeIndexToByteIndex(string(textRunes), i),
			})

			allMatched := true
			for partIdx := 1; partIdx < len(pattern.parts); partIdx++ {
				part := pattern.parts[partIdx]
				partRunes := []rune(part)

				found := false
				for j := currentIdx; j <= len(textRunes)-len(partRunes); j++ {
					partMatch := true
					for k, ch := range partRunes {
						if textRunes[j+k] != ch {
							partMatch = false
							break
						}
					}
					if partMatch {
						positions = append(positions, common.Position{
							Start:  runeIndexToByteIndex(string(textRunes), j),
							Length: runeIndexToByteIndex(string(textRunes), j+len(partRunes)) - runeIndexToByteIndex(string(textRunes), j),
						})
						currentIdx = j + len(partRunes)
						found = true
						break
					}
				}

				if !found {
					allMatched = false
					break
				}
			}

			if allMatched {
				results = append(results, matchResult{
					word:      pattern.original,
					matchType: common.WildcardMatch,
					positions: positions,
					startIdx:  i,
					endIdx:    currentIdx - 1,
				})
			}
		}
	}

	return results
}

func (f *Filter) deduplicateAndPrioritize(results []matchResult) []matchResult {
	if len(results) == 0 {
		return results
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].startIdx != results[j].startIdx {
			return results[i].startIdx < results[j].startIdx
		}
		if results[i].endIdx != results[j].endIdx {
			return results[i].endIdx > results[j].endIdx
		}
		return priority(results[i].matchType) > priority(results[j].matchType)
	})

	var unique []matchResult
	var lastEndIdx = -1

	for _, r := range results {
		if r.startIdx > lastEndIdx {
			unique = append(unique, r)
			lastEndIdx = r.endIdx
		}
	}

	return unique
}

func priority(mt common.MatchType) int {
	switch mt {
	case common.DirectMatch:
		return 3
	case common.SkipMatch:
		return 2
	case common.WildcardMatch:
		return 1
	default:
		return 0
	}
}

func runeIndexToByteIndex(s string, runeIdx int) int {
	runes := []rune(s)
	if runeIdx >= len(runes) {
		return len(s)
	}
	byteIdx := 0
	for i := 0; i < runeIdx; i++ {
		byteIdx += len(string(runes[i]))
	}
	return byteIdx
}

func runeLenToByteLen(r rune) int {
	return len(string(r))
}

func (f *Filter) AddSensitiveWord(word string) error {
	if word == "" {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	if strings.Contains(word, "*") {
		return f.addWildcardPattern(word)
	}

	f.trie.insert(word)
	return nil
}

func (f *Filter) addWildcardPattern(pattern string) error {
	parts := strings.Split(pattern, "*")
	
	if parts[0] == "" || parts[len(parts)-1] == "" {
		return nil
	}

	for _, part := range parts {
		if part == "" {
			return nil
		}
	}

	for i, existing := range f.wildcards {
		if existing.original == pattern {
			f.wildcards[i].parts = parts
			return nil
		}
	}

	f.wildcards = append(f.wildcards, WildcardPattern{
		original: pattern,
		parts:    parts,
	})

	return nil
}

func (f *Filter) RemoveSensitiveWord(word string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if strings.Contains(word, "*") {
		for i, existing := range f.wildcards {
			if existing.original == word {
				f.wildcards = append(f.wildcards[:i], f.wildcards[i+1:]...)
				return nil
			}
		}
		return nil
	}

	f.trie.remove(word)
	return nil
}

func (f *Filter) UpdateSensitiveWord(oldWord, newWord string) error {
	if err := f.RemoveSensitiveWord(oldWord); err != nil {
		return err
	}
	return f.AddSensitiveWord(newWord)
}

func (f *Filter) ListSensitiveWords() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	words := f.trie.listWords()
	for _, pattern := range f.wildcards {
		words = append(words, pattern.original)
	}
	return words
}

func (f *Filter) AddWhitelistWord(word string) error {
	if word == "" {
		return nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	f.whitelist[word] = true
	f.whitelistTrie.insert(word)
	return nil
}

func (f *Filter) RemoveWhitelistWord(word string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	delete(f.whitelist, word)
	f.whitelistTrie.remove(word)
	return nil
}

func (f *Filter) ListWhitelistWords() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.whitelistTrie.listWords()
}

func (f *Filter) GetStats() *common.StatsResponse {
	f.mu.RLock()
	defer f.mu.RUnlock()

	response := &common.StatsResponse{
		DailyStats: make([]common.DailyStats, 0),
		TopHits:    make([]common.TopSensitiveWord, 0),
	}

	for date, count := range f.dailyStats {
		response.DailyStats = append(response.DailyStats, common.DailyStats{
			Date:     date,
			Total:    count.Total,
			HitCount: count.HitCount,
		})
		response.TotalTotal += count.Total
		response.TotalHits += count.HitCount
	}

	sort.Slice(response.DailyStats, func(i, j int) bool {
		return response.DailyStats[i].Date > response.DailyStats[j].Date
	})

	type hitCount struct {
		word  string
		count int64
	}
	var hcs []hitCount
	for word, count := range f.hitStats {
		hcs = append(hcs, hitCount{word: word, count: count})
	}

	sort.Slice(hcs, func(i, j int) bool {
		return hcs[i].count > hcs[j].count
	})

	for i, hc := range hcs {
		if i >= 10 {
			break
		}
		response.TopHits = append(response.TopHits, common.TopSensitiveWord{
			Word:  hc.word,
			Count: hc.count,
		})
	}

	return response
}

func isNumber(r rune) bool {
	return unicode.IsDigit(r)
}

func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}

func isChinese(r rune) bool {
	return r >= 0x4E00 && r <= 0x9FFF
}

package analyzer

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"

	"goroutinelab/api"
)

var (
	goroutineHeaderRegex = regexp.MustCompile(`^goroutine (\d+) \[([^\]]*)\]:$`)
	frameRegex           = regexp.MustCompile(`^([^\s(]+)(?:\([^)]*\))?$`)
	locationRegex        = regexp.MustCompile(`^\s+(.+):(\d+)(?:\s+.*)?$`)
	additionalFrames     = "...additional frames elided"
)

func Parse(text string) []*api.GoroutineInfo {
	var result []*api.GoroutineInfo
	scanner := bufio.NewScanner(strings.NewReader(text))
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var current *api.GoroutineInfo
	var rawStack []string

	for scanner.Scan() {
		line := scanner.Text()

		if match := goroutineHeaderRegex.FindStringSubmatch(line); match != nil {
			if current != nil {
				processStack(current, rawStack)
				result = append(result, current)
			}
			current = &api.GoroutineInfo{}
			rawStack = nil

			id, _ := strconv.Atoi(match[1])
			current.ID = id
			parseState(current, match[2])
		} else if current != nil {
			if strings.HasPrefix(line, "goroutine ") {
				continue
			}
			rawStack = append(rawStack, line)
		}
	}

	if current != nil {
		processStack(current, rawStack)
		result = append(result, current)
	}

	return result
}

func parseState(g *api.GoroutineInfo, stateStr string) {
	stateStr = strings.TrimSpace(stateStr)
	if stateStr == "" {
		g.State = "unknown"
		g.HasWaitTime = false
		return
	}

	parts := strings.Split(stateStr, ",")
	if len(parts) == 1 {
		maybeTime := strings.TrimSpace(parts[0])
		if strings.HasSuffix(maybeTime, " minutes") {
			g.State = "unknown"
			g.HasWaitTime = true
			g.WaitMinutes = parseMinutes(maybeTime)
			return
		}
		if _, err := strconv.Atoi(maybeTime); err == nil {
			g.State = "unknown"
			g.HasWaitTime = false
			return
		}
		g.State = normalizeState(maybeTime)
		g.HasWaitTime = false
		return
	}

	g.State = normalizeState(strings.TrimSpace(parts[0]))
	if len(parts) >= 2 {
		g.HasWaitTime = true
		g.WaitMinutes = parseMinutes(strings.TrimSpace(parts[1]))
	}
}

func normalizeState(s string) string {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "chan receive"):
		return api.BlockTypeChanReceive
	case strings.Contains(s, "chan send"):
		return api.BlockTypeChanSend
	case strings.Contains(s, "select"):
		return api.BlockTypeSelect
	case strings.Contains(s, "i/o wait"):
		return api.BlockTypeIOWait
	case strings.Contains(s, "syscall"):
		return api.BlockTypeSyscall
	case strings.Contains(s, "sleep"):
		return api.BlockTypeSleep
	case strings.Contains(s, "semacquire"):
		return api.BlockTypeSemacquire
	}
	return api.BlockTypeUnknown
}

func parseMinutes(s string) int {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, " minutes")
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func processStack(g *api.GoroutineInfo, rawStack []string) {
	g.RawStack = rawStack

	for _, line := range rawStack {
		if strings.Contains(line, additionalFrames) {
			g.StackTruncated = true
			break
		}
	}

	filtered := extractUserStack(rawStack)
	g.UserStack = filtered

	if len(filtered) > 0 {
		topFrame := filtered[0]
		g.BlockingOn = detectBlocking(topFrame, g.State)
	}
}

func extractUserStack(raw []string) []api.StackFrame {
	var frames []api.StackFrame
	var i = 0
	for i < len(raw) {
		line := raw[i]
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}
		if strings.HasPrefix(line, "...") {
			i++
			continue
		}

		match := frameRegex.FindStringSubmatch(strings.TrimSpace(line))
		if match == nil {
			i++
			continue
		}

		fn := match[1]
		if isRuntimeInternal(fn) {
			if i+1 < len(raw) {
				i += 2
			} else {
				i++
			}
			continue
		}

		frame := api.StackFrame{Function: fn}
		if i+1 < len(raw) {
			loc := raw[i+1]
			if locMatch := locationRegex.FindStringSubmatch(loc); locMatch != nil {
				frame.File = locMatch[1]
				if ln, err := strconv.Atoi(locMatch[2]); err == nil {
					frame.Line = ln
				}
			}
		}

		frames = append(frames, frame)
		i += 2
	}
	return frames
}

func isRuntimeInternal(fn string) bool {
	prefixes := []string{
		"runtime.",
		"runtime/debug.",
		"runtime/pprof.",
	}
	for _, p := range prefixes {
		if strings.HasPrefix(fn, p) {
			return true
		}
	}
	return false
}

func detectBlocking(frame api.StackFrame, state string) *api.BlockingInfo {
	bi := &api.BlockingInfo{Type: state, Function: frame.Function}
	switch state {
	case api.BlockTypeIOWait:
		bi.HasTimeout = false
	}
	return bi
}

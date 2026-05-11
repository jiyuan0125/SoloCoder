package benchmark

import (
	"bufio"
	"regexp"
	"strconv"
	"strings"
)

type RawBenchmark struct {
	Name       string
	Iterations int
	NsPerOp    float64
	BytesPerOp int
	AllocsPerOp int
	Uncertain  bool
}

var benchmarkLineRegex = regexp.MustCompile(`^(~?)Benchmark(.+)\s+(\d+)\s+([\d.]+)\s+([µmn]?s/op)(?:\s+(\d+)\s+B/op(?:\s+(\d+)\s+allocs/op)?)?$`)

func ParseOutput(output string) map[string][]RawBenchmark {
	results := make(map[string][]RawBenchmark)
	scanner := bufio.NewScanner(strings.NewReader(output))
	
	for scanner.Scan() {
		line := scanner.Text()
		matches := benchmarkLineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}
		
		uncertain := matches[1] == "~"
		name := "Benchmark" + matches[2]
		iterations, _ := strconv.Atoi(matches[3])
		timeValue, _ := strconv.ParseFloat(matches[4], 64)
		timeUnit := matches[5]
		
		nsPerOp := convertToNs(timeValue, timeUnit)
		
		var bytesPerOp int
		var allocsPerOp int
		
		if matches[6] != "" {
			bytesPerOp, _ = strconv.Atoi(matches[6])
		}
		if matches[7] != "" {
			allocsPerOp, _ = strconv.Atoi(matches[7])
		}
		
		raw := RawBenchmark{
			Name:        name,
			Iterations:  iterations,
			NsPerOp:     nsPerOp,
			BytesPerOp:  bytesPerOp,
			AllocsPerOp: allocsPerOp,
			Uncertain:   uncertain,
		}
		
		results[name] = append(results[name], raw)
	}
	
	return results
}

func convertToNs(value float64, unit string) float64 {
	switch unit {
	case "ns/op":
		return value
	case "µs/op":
		return value * 1000.0
	case "ms/op":
		return value * 1000000.0
	default:
		return value
	}
}

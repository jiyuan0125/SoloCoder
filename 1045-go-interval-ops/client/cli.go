package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"intervalops/common"
)

func parseIntegerLine(line string) (*common.IntegerInterval, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid integer interval format: %s", line)
	}
	min, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid min value: %s", parts[0])
	}
	max, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid max value: %s", parts[1])
	}
	if min > max {
		min, max = max, min
	}
	return &common.IntegerInterval{Min: min, Max: max}, nil
}

func parseTimeLine(line string, loc *time.Location) (*common.TimeInterval, error) {
	parts := strings.Split(line, ",")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid time interval format: %s", line)
	}
	startStr := strings.TrimSpace(parts[0])
	endStr := strings.TrimSpace(parts[1])

	start, err := parseTime(startStr, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid start time: %s", startStr)
	}
	end, err := parseTime(endStr, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid end time: %s", endStr)
	}
	return &common.TimeInterval{Start: start, End: end}, nil
}

func parseTime(s string, loc *time.Location) (time.Time, error) {
	t, err := time.ParseInLocation(time.RFC3339, s, loc)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02 15:04:05", s, loc)
	if err == nil {
		return t, nil
	}
	t, err = time.ParseInLocation("2006-01-02T15:04:05", s, loc)
	if err == nil {
		return t, nil
	}
	return time.Time{}, err
}

func readIntegerIntervalsFromFile(path string) ([]common.IntegerInterval, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var intervals []common.IntegerInterval
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		iv, err := parseIntegerLine(line)
		if err != nil {
			return nil, err
		}
		intervals = append(intervals, *iv)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return intervals, nil
}

func readTimeIntervalsFromFile(path string, loc *time.Location) ([]common.TimeInterval, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var intervals []common.TimeInterval
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		iv, err := parseTimeLine(line, loc)
		if err != nil {
			return nil, err
		}
		intervals = append(intervals, *iv)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return intervals, nil
}

func writeIntegerIntervalsToFile(path string, intervals []common.IntegerInterval) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, iv := range intervals {
		fmt.Fprintf(writer, "%d,%d\n", iv.Min, iv.Max)
	}
	return writer.Flush()
}

func writeTimeIntervalsToFile(path string, intervals []common.TimeInterval, loc *time.Location) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := bufio.NewWriter(file)
	for _, iv := range intervals {
		start := iv.Start.In(loc).Format(time.RFC3339)
		end := iv.End.In(loc).Format(time.RFC3339)
		fmt.Fprintf(writer, "%s,%s\n", start, end)
	}
	return writer.Flush()
}

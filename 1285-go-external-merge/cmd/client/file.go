package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

func ReadDataFromFile(filename string) ([]int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := make([]int64, 0)
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		if line == "" {
			continue
		}
		val, err := strconv.ParseInt(line, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid integer: %s", lineNum, line)
		}
		data = append(data, val)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return data, nil
}

func WriteDataToFile(filename string, data []int64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, val := range data {
		_, err := fmt.Fprintln(writer, val)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

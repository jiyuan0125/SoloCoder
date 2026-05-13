package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("需要PDF路径")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}

	if len(data) < 8 || !bytes.HasPrefix(data, []byte("%PDF-")) {
		fmt.Println("不是PDF")
		return
	}

	fmt.Println("===== 解析PDF =====")

	startxrefIdx := bytes.LastIndex(data, []byte("startxref"))
	if startxrefIdx == -1 {
		fmt.Println("找不到startxref")
		return
	}

	rest := data[startxrefIdx:]
	parts := strings.Fields(string(rest))
	if len(parts) < 2 {
		fmt.Println("无效startxref")
		return
	}
	start, _ := strconv.Atoi(parts[1])
	fmt.Printf("xref偏移: %d\n", start)

	if !bytes.HasPrefix(data[start:], []byte("xref")) {
		fmt.Println("不是传统xref格式")
		return
	}

	xrefs := make(map[int]int)
	lines := strings.Split(string(data[start:]), "\n")
	i := 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			i++
			continue
		}
		if line == "trailer" {
			break
		}

		fields := strings.Fields(line)
		if len(fields) == 2 {
			first, _ := strconv.Atoi(fields[0])
			count, _ := strconv.Atoi(fields[1])

			i++
			for j := 0; j < count && i < len(lines); j++ {
				line = strings.TrimSpace(lines[i])
				fields = strings.Fields(line)
				if len(fields) == 3 && fields[2] == "n" {
					off, _ := strconv.Atoi(fields[0])
					xrefs[first+j] = off
				}
				i++
			}
		} else {
			i++
		}
	}

	fmt.Printf("XRefs: %v\n", xrefs)

	trailerStart := bytes.Index(data[start:], []byte("<<"))
	if trailerStart == -1 {
		fmt.Println("找不到trailer <<")
		return
	}

	trailer, _, err := parseDict(data[start+trailerStart:])
	if err != nil {
		fmt.Printf("trailer解析失败: %v\n", err)
		return
	}

	trailerDict, ok := trailer.(map[string]interface{})
	if !ok {
		fmt.Printf("trailer不是字典: %T\n", trailer)
		return
	}

	fmt.Printf("Trailer keys: ")
	for k := range trailerDict {
		fmt.Printf("/%s ", k)
	}
	fmt.Println()

	rootRef, ok := trailerDict["Root"]
	if !ok {
		fmt.Println("没有Root")
		return
	}

	fmt.Printf("Root value: %v (type: %T)\n", rootRef, rootRef)

	refArr, ok := rootRef.([]interface{})
	if !ok || len(refArr) != 2 {
		fmt.Println("Root不是引用数组")
		return
	}

	rootNum := refArr[0].(int)
	fmt.Printf("Root对象号: %d\n", rootNum)

	rootOff, ok := xrefs[rootNum]
	if !ok {
		fmt.Printf("xrefs中没有对象 %d\n", rootNum)
		return
	}

	fmt.Printf("Root对象偏移: %d\n", rootOff)

	// 模拟 readObjectAt
	idx := rootOff
	for idx < len(data) && (data[idx] == '\n' || data[idx] == '\r' || data[idx] == ' ') {
		idx++
	}

	fmt.Printf("调整后idx: %d\n", idx)

	rest2 := data[idx:]
	fmt.Printf("起始处: %q\n", string(rest2[:min(50, len(rest2))]))

	parts2 := strings.Fields(string(rest2))
	fmt.Printf("前3个fields: %v\n", parts2[:min(3, len(parts2))])

	if len(parts2) < 3 || parts2[1] != "obj" {
		fmt.Println("无效对象头部")
		return
	}

	startIdx := idx
	for counter := 0; counter < 3 && startIdx < len(data); counter++ {
		for startIdx < len(data) && data[startIdx] != ' ' && data[startIdx] != '\n' && data[startIdx] != '\r' {
			startIdx++
		}
		for startIdx < len(data) && (data[startIdx] == ' ' || data[startIdx] == '\n' || data[startIdx] == '\r') {
			startIdx++
		}
	}

	fmt.Printf("内容起始偏移: %d\n", startIdx)
	fmt.Printf("内容起始: %q\n", string(data[startIdx:min(startIdx+50, len(data))]))

	obj, consumed, err := parseValue(data[startIdx:])
	if err != nil {
		fmt.Printf("解析值失败: %v\n", err)
		return
	}

	fmt.Printf("解析成功! 消耗 %d 字节\n", consumed)
	fmt.Printf("对象类型: %T\n", obj)

	if dict, ok := obj.(map[string]interface{}); ok {
		fmt.Printf("Catalog keys: ")
		for k := range dict {
			fmt.Printf("/%s ", k)
		}
		fmt.Println()

		if pagesRef, ok := dict["Pages"]; ok {
			fmt.Printf("/Pages = %v (type: %T)\n", pagesRef, pagesRef)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseValue(data []byte) (interface{}, int, error) {
	i := 0
	for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
		i++
	}

	if i >= len(data) {
		return nil, i, fmt.Errorf("空值")
	}

	switch data[i] {
	case '(':
		return parseString(data[i:])
	case '/':
		return parseName(data[i:])
	case '[':
		return parseArray(data[i:])
	case '<':
		if i+1 < len(data) && data[i+1] == '<' {
			return parseDict(data[i:])
		}
		return parseHex(data[i:])
	}

	if (data[i] >= '0' && data[i] <= '9') || data[i] == '-' || data[i] == '+' || data[i] == '.' {
		return parseNumOrRef(data[i:])
	}

	return nil, i, fmt.Errorf("未知类型: %c", data[i])
}

func parseString(data []byte) (string, int, error) {
	i := 1
	depth := 1
	var buf bytes.Buffer

	for i < len(data) && depth > 0 {
		switch data[i] {
		case '\\':
			if i+1 < len(data) {
				buf.WriteByte(data[i])
				buf.WriteByte(data[i+1])
				i += 2
			} else {
				i++
			}
		case '(':
			depth++
			buf.WriteByte(data[i])
			i++
		case ')':
			depth--
			if depth > 0 {
				buf.WriteByte(data[i])
			}
			i++
		default:
			buf.WriteByte(data[i])
			i++
		}
	}

	return buf.String(), i, nil
}

func parseHex(data []byte) (string, int, error) {
	i := 1
	var buf bytes.Buffer
	for i < len(data) && data[i] != '>' {
		buf.WriteByte(data[i])
		i++
	}
	return buf.String(), i + 1, nil
}

func parseName(data []byte) (string, int, error) {
	i := 1
	var buf bytes.Buffer
	for i < len(data) {
		c := data[i]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '/' || c == '[' || c == ']' || c == '(' || c == ')' || c == '<' || c == '>' {
			break
		}
		buf.WriteByte(c)
		i++
	}
	return buf.String(), i, nil
}

func parseArray(data []byte) (interface{}, int, error) {
	i := 1
	var result []interface{}

	for i < len(data) {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i >= len(data) || data[i] == ']' {
			return result, i + 1, nil
		}

		val, consumed, err := parseValue(data[i:])
		if err != nil {
			return nil, i, err
		}
		result = append(result, val)
		i += consumed
	}

	return result, i, nil
}

func parseDict(data []byte) (interface{}, int, error) {
	i := 2
	result := make(map[string]interface{})

	for i < len(data) {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i+1 < len(data) && data[i] == '>' && data[i+1] == '>' {
			return result, i + 2, nil
		}
		if i >= len(data) {
			return nil, i, fmt.Errorf("字典未终止")
		}

		if data[i] != '/' {
			fmt.Printf("警告: 在字典中遇到非'/'字符: %q at pos %d\n", data[i], i)
			i++
			continue
		}

		key, keyConsumed, err := parseName(data[i:])
		if err != nil {
			return nil, i, err
		}
		i += keyConsumed

		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}

		val, valConsumed, err := parseValue(data[i:])
		if err != nil {
			return nil, i, err
		}
		i += valConsumed

		result[key] = val
	}

	return result, i, nil
}

func parseNumOrRef(data []byte) (interface{}, int, error) {
	i := 0
	var numBuf bytes.Buffer

	for i < len(data) && ((data[i] >= '0' && data[i] <= '9') || data[i] == '-' || data[i] == '+' || data[i] == '.' || data[i] == 'e' || data[i] == 'E') {
		numBuf.WriteByte(data[i])
		i++
	}

	num1Str := numBuf.String()

	j := i
	for j < len(data) && (data[j] == ' ' || data[j] == '\n' || data[j] == '\r' || data[j] == '\t') {
		j++
	}

	k := j
	var num2Buf bytes.Buffer
	for k < len(data) && data[k] >= '0' && data[k] <= '9' {
		num2Buf.WriteByte(data[k])
		k++
	}

	if num2Buf.Len() > 0 {
		m := k
		for m < len(data) && (data[m] == ' ' || data[m] == '\n' || data[m] == '\r' || data[m] == '\t') {
			m++
		}
		if m < len(data) && data[m] == 'R' {
			num1, _ := strconv.Atoi(num1Str)
			num2, _ := strconv.Atoi(num2Buf.String())
			return []interface{}{num1, num2}, m + 1, nil
		}
	}

	if strings.Contains(num1Str, ".") {
		f, _ := strconv.ParseFloat(num1Str, 64)
		return f, i, nil
	}
	n, _ := strconv.Atoi(num1Str)
	return n, i, nil
}

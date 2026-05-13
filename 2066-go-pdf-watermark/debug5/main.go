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

	startxrefIdx := bytes.LastIndex(data, []byte("startxref"))
	rest := data[startxrefIdx:]
	parts := strings.Fields(string(rest))
	start, _ := strconv.Atoi(parts[1])

	xrefs := make(map[int]int)
	lines := strings.Split(string(data[start:]), "\n")
	i := 1
	for i < len(lines) {
		line := strings.TrimSpace(lines[i])
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

	trailerStart := bytes.Index(data[start:], []byte("<<"))
	trailer, _, _ := parseDict(data[start+trailerStart:])
	trailerDict := trailer.(map[string]interface{})

	rootRef := trailerDict["Root"].([]interface{})
	rootNum := rootRef[0].(int)

	rootObj, _ := readObject(data, xrefs, rootNum)
	rootDict := rootObj.(map[string]interface{})

	fmt.Printf("=== 对象 %d (Catalog) ===\n", rootNum)
	for k, v := range rootDict {
		fmt.Printf("  /%s = %v (type=%T)\n", k, v, v)
	}

	pagesRef := rootDict["Pages"].([]interface{})
	pagesNum := pagesRef[0].(int)

	pagesObj, _ := readObject(data, xrefs, pagesNum)
	pagesDict := pagesObj.(map[string]interface{})

	fmt.Printf("\n=== 对象 %d (Pages) ===\n", pagesNum)
	for k, v := range pagesDict {
		fmt.Printf("  /%s = %v (type=%T)\n", k, v, v)
	}

	kids, ok := pagesDict["Kids"].([]interface{})
	fmt.Printf("\nKids存在: %v\n", ok)
	if ok {
		fmt.Printf("Kids: %v\n", kids)
		for _, kid := range kids {
			fmt.Printf("  Kid类型: %T, 值: %v\n", kid, kid)
			if refArr, ok := kid.([]interface{}); ok && len(refArr) == 2 {
				kidNum := refArr[0].(int)
				kidObj, err := readObject(data, xrefs, kidNum)
				fmt.Printf("  读取对象 %d: err=%v\n", kidNum, err)
				if kidDict, ok := kidObj.(map[string]interface{}); ok {
					typ, _ := kidDict["Type"].(string)
					fmt.Printf("  /Type = %q\n", typ)
				}
			}
		}
	}

	typ, _ := pagesDict["Type"].(string)
	fmt.Printf("\nPages /Type = %q\n", typ)
	fmt.Printf("len(kids)=%d\n", len(kids))
	fmt.Printf("条件 typ==Pages: %v, 条件 typ==''&&len(kids)>0: %v\n", typ == "Pages", typ == "" && len(kids) > 0)
}

func readObject(data []byte, xrefs map[int]int, num int) (interface{}, error) {
	off, ok := xrefs[num]
	if !ok {
		return nil, fmt.Errorf("没有对象 %d", num)
	}

	idx := off
	for idx < len(data) && (data[idx] == '\n' || data[idx] == '\r' || data[idx] == ' ') {
		idx++
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

	obj, _, err := parseValue(data[startIdx:])
	return obj, err
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
	return nil, i, fmt.Errorf("未知")
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
		if data[i] != '/' {
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

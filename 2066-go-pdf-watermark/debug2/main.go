package main

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("需要PDF文件路径")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("读取失败: %v\n", err)
		os.Exit(1)
	}

	idx := bytes.LastIndex(data, []byte("startxref"))
	if idx == -1 {
		fmt.Println("找不到startxref")
		return
	}
	rest := data[idx:]
	parts := strings.Fields(string(rest))
	if len(parts) < 2 {
		fmt.Println("无效startxref")
		return
	}
	start, _ := strconv.Atoi(parts[1])
	fmt.Printf("xref偏移: %d\n", start)

	if !bytes.HasPrefix(data[start:], []byte("xref")) {
		fmt.Println("不是传统xref")
		return
	}

	lines := strings.Split(string(data[start:]), "\n")
	i := 1
	xrefs := make(map[int]int)
	totalSize := 0

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
			totalSize = first + count
			fmt.Printf("xref段: first=%d, count=%d\n", first, count)

			i++
			for j := 0; j < count && i < len(lines); j++ {
				line = strings.TrimSpace(lines[i])
				fields = strings.Fields(line)
				if len(fields) == 3 {
					offset, _ := strconv.Atoi(fields[0])
					if fields[2] == "n" {
						xrefs[first+j] = offset
						fmt.Printf("  obj %d -> offset %d\n", first+j, offset)
					}
				}
				i++
			}
		} else {
			i++
		}
	}

	fmt.Printf("\n找到 %d 个xref条目, totalSize=%d\n", len(xrefs), totalSize)

	for ; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "<<" {
			break
		}
	}

	trailerStart := bytes.Index(data[start:], []byte("<<"))
	if trailerStart == -1 {
		fmt.Println("找不到trailer字典")
		return
	}

	trailerData := data[start+trailerStart:]
	fmt.Printf("\ntrailer起始位置 (在start偏移后): %d\n", trailerStart)
	fmt.Printf("trailer前50字节: %q\n", string(trailerData[:min(50, len(trailerData))]))

	trailerDict, _, err := parseDictSimple(trailerData)
	if err != nil {
		fmt.Printf("解析trailer失败: %v\n", err)
	} else {
		fmt.Printf("\ntrailer字典内容:\n")
		for k, v := range trailerDict {
			fmt.Printf("  /%s = %v (类型: %T)\n", k, v, v)
		}
	}

	rootRef, ok := trailerDict["Root"]
	if !ok {
		fmt.Println("\n没有Root!")
		return
	}
	fmt.Printf("\nRoot值: %v (类型: %T)\n", rootRef, rootRef)

	if refArr, ok := rootRef.([]interface{}); ok && len(refArr) == 2 {
		rootNum := refArr[0].(int)
		fmt.Printf("Root对象号: %d\n", rootNum)
		offset, ok := xrefs[rootNum]
		if !ok {
			fmt.Printf("错误: xrefs中找不到对象 %d\n", rootNum)
			return
		}
		fmt.Printf("Root偏移: %d\n", offset)

		objData := data[offset:]
		fmt.Printf("\n对象 %d 起始: %q\n", rootNum, string(objData[:min(80, len(objData))]))

		headerEnd := bytes.Index(objData, []byte(" obj\n"))
		if headerEnd == -1 {
			headerEnd = bytes.Index(objData, []byte(" obj\r\n"))
		}
		if headerEnd != -1 {
			contentStart := headerEnd + 5
			for contentStart < len(objData) && (objData[contentStart] == '\n' || objData[contentStart] == '\r' || objData[contentStart] == ' ') {
				contentStart++
			}
			fmt.Printf("对象内容起始偏移: %d\n", offset+contentStart)
			fmt.Printf("对象内容前50字节: %q\n", string(objData[contentStart:min(contentStart+50, len(objData))]))

			val, consumed, err := parseValueSimple(objData[contentStart:])
			if err != nil {
				fmt.Printf("解析对象失败: %v\n", err)
			} else {
				fmt.Printf("解析成功! 消耗了 %d 字节\n", consumed)
				fmt.Printf("对象类型: %T\n", val)
				if dict, ok := val.(map[string]interface{}); ok {
					fmt.Printf("字典键: ")
					for k := range dict {
						fmt.Printf("/%s ", k)
					}
					fmt.Println()
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseDictSimple(data []byte) (map[string]interface{}, int, error) {
	i := 2
	result := make(map[string]interface{})

	for i < len(data) {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i+1 < len(data) && data[i] == '>' && data[i+1] == '>' {
			break
		}
		if i >= len(data) {
			break
		}

		if data[i] != '/' {
			i++
			continue
		}

		key, keyConsumed, err := parseNameSimple(data[i:])
		if err != nil {
			return nil, 0, err
		}
		i += keyConsumed

		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}

		val, valConsumed, err := parseValueSimple(data[i:])
		if err != nil {
			return nil, 0, err
		}
		i += valConsumed

		result[key] = val
	}

	return result, i + 2, nil
}

func parseNameSimple(data []byte) (string, int, error) {
	i := 1
	var result bytes.Buffer

	for i < len(data) {
		c := data[i]
		if c == ' ' || c == '\n' || c == '\r' || c == '\t' || c == '/' || c == '[' || c == ']' || c == '(' || c == ')' || c == '<' || c == '>' {
			break
		}
		result.WriteByte(c)
		i++
	}

	return result.String(), i, nil
}

func parseValueSimple(data []byte) (interface{}, int, error) {
	i := 0
	for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
		i++
	}

	if i >= len(data) {
		return nil, i, fmt.Errorf("空值")
	}

	switch data[i] {
	case '(':
		return parseStringSimple(data[i:])
	case '/':
		return parseNameSimple(data[i:])
	case '[':
		return parseArraySimple(data[i:])
	case '<':
		if i+1 < len(data) && data[i+1] == '<' {
			d, n, err := parseDictSimple(data[i:])
			return d, n, err
		}
		return parseHexSimple(data[i:])
	}

	if (data[i] >= '0' && data[i] <= '9') || data[i] == '-' || data[i] == '+' || data[i] == '.' {
		return parseNumOrRefSimple(data[i:])
	}

	return nil, i, fmt.Errorf("未知")
}

func parseStringSimple(data []byte) (string, int, error) {
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

func parseHexSimple(data []byte) (string, int, error) {
	i := 1
	var buf bytes.Buffer
	for i < len(data) && data[i] != '>' {
		buf.WriteByte(data[i])
		i++
	}
	return buf.String(), i + 1, nil
}

func parseArraySimple(data []byte) ([]interface{}, int, error) {
	i := 1
	var result []interface{}

	for i < len(data) {
		for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
			i++
		}
		if i >= len(data) || data[i] == ']' {
			return result, i + 1, nil
		}

		val, consumed, err := parseValueSimple(data[i:])
		if err != nil {
			return nil, i, err
		}
		result = append(result, val)
		i += consumed
	}

	return result, i, nil
}

var numRe = regexp.MustCompile(`^[-+]?[0-9]*\.?[0-9]+([eE][-+]?[0-9]+)?`)

func parseNumOrRefSimple(data []byte) (interface{}, int, error) {
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

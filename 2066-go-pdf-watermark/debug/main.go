package main

import (
	"fmt"
	"os"

	"pdf-watermark/internal/pdf"
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

	fmt.Printf("文件大小: %d 字节\n", len(data))

	idx := -1
	for i := len(data) - 20; i >= 0; i-- {
		if data[i] == 's' && i+8 < len(data) {
			if string(data[i:i+9]) == "startxref" {
				idx = i
				break
			}
		}
	}

	if idx == -1 {
		fmt.Println("找不到startxref")
		return
	}

	fmt.Printf("startxref索引: %d\n", idx)
	rest := data[idx:]
	fmt.Printf("end部分: %s\n", string(rest))

	// 测试直接解析trailer
	// 先找到xref
	xrefIdx := -1
	for i := len(data) - 300; i >= 0; i-- {
		if data[i] == 'x' && i+4 < len(data) && string(data[i:i+4]) == "xref" {
			xrefIdx = i
			break
		}
	}
	if xrefIdx != -1 {
		fmt.Printf("\nxref索引: %d\n", xrefIdx)
		xrefPart := data[xrefIdx:]
		lines := 0
		for _, b := range xrefPart {
			fmt.Printf("%c", b)
			if b == '\n' {
				lines++
				if lines > 15 {
					break
				}
			}
		}
	}

	// 测试PDF解析
	_, err = pdf.NewPDF(os.Args[1])
	if err != nil {
		fmt.Printf("\n\nPDF解析错误: %v\n", err)
	}
}

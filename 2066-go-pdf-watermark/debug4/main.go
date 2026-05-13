package main

import (
	"fmt"
	"os"

	"pdf-watermark/internal/pdf"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("需要PDF路径")
		os.Exit(1)
	}

	p, err := pdf.NewPDF(os.Args[1])
	if err != nil {
		fmt.Printf("打开PDF失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("页数: %d\n", p.PageCount())
	fmt.Printf("是否加密: %v\n", p.IsEncrypted())
}

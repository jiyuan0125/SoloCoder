package main

import (
	"fmt"
	"os"

	"pdf-watermark/internal/pdf"
	"pdf-watermark/internal/types"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run main.go <pdf文件>")
		os.Exit(1)
	}

	p, err := pdf.NewPDF(os.Args[1])
	if err != nil {
		fmt.Printf("打开PDF失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("页数: %d\n", p.PageCount())
	fmt.Printf("是否加密: %v\n", p.IsEncrypted())

	config := &types.WatermarkConfig{
		Mode:     "text",
		Text:     "WATERMARK",
		FontSize: 36,
		Color:    "255,0,0",
		Opacity:  0.5,
		Position: "center",
	}

	if err := p.AddWatermark("/tmp/out.pdf", config); err != nil {
		fmt.Printf("添加水印失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("成功! 输出到 /tmp/out.pdf")
}

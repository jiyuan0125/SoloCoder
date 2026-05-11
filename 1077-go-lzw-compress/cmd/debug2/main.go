package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/lzwtool/lzwcompress/pkg/lzw"
)

func main() {
	fmt.Println("=== 详细调试：min_code_size=2, 输入=[0,1,2,3,4] ===")
	debugRoundTrip([]byte{0, 1, 2, 3, 4}, 2)
}

func debugRoundTrip(input []byte, minCodeSize int) {
	fmt.Printf("输入: %v (len=%d)\n", input, len(input))
	fmt.Printf("minCodeSize=%d\n", minCodeSize)
	fmt.Printf("clearCode=%d, eoiCode=%d\n", 1<<minCodeSize, (1<<minCodeSize)+1)

	var compressed bytes.Buffer
	enc, err := lzw.NewEncoder(minCodeSize, &compressed)
	if err != nil {
		fmt.Printf("创建Encoder失败: %v\n", err)
		return
	}
	fmt.Println("--- 开始压缩 ---")
	err = enc.Compress(bytes.NewReader(input))
	if err != nil {
		fmt.Printf("压缩失败: %v\n", err)
		return
	}
	fmt.Printf("压缩结果: %v (len=%d)\n", compressed.Bytes(), compressed.Len())

	fmt.Println("--- 开始解压 ---")
	var decompressed bytes.Buffer
	dec, err := lzw.NewDecoder(minCodeSize, bytes.NewReader(compressed.Bytes()))
	if err != nil {
		fmt.Printf("创建Decoder失败: %v\n", err)
		return
	}
	err = dec.Decompress(&decompressed)
	if err != nil {
		fmt.Printf("解压失败: %v\n", err)
	}

	fmt.Printf("解压结果: %v (len=%d)\n", decompressed.Bytes(), len(decompressed.Bytes()))

	if bytes.Equal(input, decompressed.Bytes()) {
		fmt.Println("✓ PASS")
	} else {
		fmt.Println("✗ FAIL")
		os.Exit(1)
	}
}

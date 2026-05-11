package main

import (
	"bytes"
	"fmt"

	"github.com/lzwtool/lzwcompress/pkg/lzw"
)

func main() {
	fmt.Println("=== 测试：min_code_size=2，输入只有0-3 ===")
	test([]byte{0, 1, 2, 3, 0, 1, 2, 3}, 2)

	fmt.Println("\n=== 测试：min_code_size=3，输入只有0-7 ===")
	test([]byte{0, 1, 2, 3, 4, 5, 6, 7}, 3)

	fmt.Println("\n=== 测试：min_code_size=8，长数据（触发码宽变化）===")
	var input bytes.Buffer
	for i := 0; i < 1000; i++ {
		input.WriteString("TOBEORNOTTOBEORTOBEORNOT")
	}
	test(input.Bytes(), 8)
}

func test(input []byte, minCodeSize int) {
	var compressed bytes.Buffer
	enc, _ := lzw.NewEncoder(minCodeSize, &compressed)
	enc.Compress(bytes.NewReader(input))

	var decompressed bytes.Buffer
	dec, _ := lzw.NewDecoder(minCodeSize, bytes.NewReader(compressed.Bytes()))
	err := dec.Decompress(&decompressed)

	if err != nil {
		fmt.Printf("  解压错误: %v\n", err)
		return
	}

	if bytes.Equal(input, decompressed.Bytes()) {
		fmt.Printf("  ✓ PASS (ratio: %.2f%%)\n", float64(compressed.Len())/float64(len(input))*100)
	} else {
		fmt.Printf("  ✗ FAIL\n")
		fmt.Printf("  原始: %v\n", input)
		fmt.Printf("  解压: %v\n", decompressed.Bytes())
	}
}

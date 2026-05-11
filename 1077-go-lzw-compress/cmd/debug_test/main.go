package main

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/lzwtool/lzwcompress/pkg/lzw"
)

func main() {
	fmt.Println("=== 测试1: 长数据（触发码宽变化）===")
	testLongData()

	fmt.Println("\n=== 测试2: min_code_size=2，包含字节值>=4 ===")
	testMinCodeSize2()

	fmt.Println("\n=== 测试3: 随机数据测试 ===")
	testRandomData()
}

func testLongData() {
	var input bytes.Buffer
	for i := 0; i < 1000; i++ {
		input.WriteString("TOBEORNOTTOBEORTOBEORNOT")
	}

	testRoundTrip(input.Bytes(), 8)
}

func testMinCodeSize2() {
	input := []byte{0, 1, 2, 3, 4, 5, 10, 20, 100, 200, 255}
	testRoundTrip(input, 2)
}

func testRandomData() {
	for minCode := 2; minCode <= 8; minCode++ {
		for length := 10; length <= 1000; length *= 10 {
			input := make([]byte, length)
			for i := range input {
				input[i] = byte(rand.Intn(256))
			}

			var compressed bytes.Buffer
			enc, _ := lzw.NewEncoder(minCode, &compressed)
			enc.Compress(bytes.NewReader(input))

			var decompressed bytes.Buffer
			dec, _ := lzw.NewDecoder(minCode, bytes.NewReader(compressed.Bytes()))
			err := dec.Decompress(&decompressed)

			if err != nil {
				fmt.Printf("  minCode=%d, len=%d: 解压错误 - %v\n", minCode, length, err)
				continue
			}

			if bytes.Equal(input, decompressed.Bytes()) {
				fmt.Printf("  ✓ minCode=%d, len=%d: PASS\n", minCode, length)
			} else {
				fmt.Printf("  ✗ minCode=%d, len=%d: FAIL (数据不匹配)\n", minCode, length)
			}
		}
	}
}

func testRoundTrip(input []byte, minCodeSize int) {
	var compressed bytes.Buffer
	enc, err := lzw.NewEncoder(minCodeSize, &compressed)
	if err != nil {
		fmt.Printf("  创建Encoder失败: %v\n", err)
		return
	}
	err = enc.Compress(bytes.NewReader(input))
	if err != nil {
		fmt.Printf("  压缩失败: %v\n", err)
		return
	}

	var decompressed bytes.Buffer
	dec, err := lzw.NewDecoder(minCodeSize, bytes.NewReader(compressed.Bytes()))
	if err != nil {
		fmt.Printf("  创建Decoder失败: %v\n", err)
		return
	}
	err = dec.Decompress(&decompressed)
	if err != nil {
		fmt.Printf("  解压失败: %v\n", err)
		return
	}

	if bytes.Equal(input, decompressed.Bytes()) {
		fmt.Println("  ✓ PASS")
	} else {
		fmt.Println("  ✗ FAIL")
		fmt.Printf("  原始长度: %d, 解压后长度: %d\n", len(input), len(decompressed.Bytes()))
		if len(input) < 50 && len(decompressed.Bytes()) < 50 {
			fmt.Printf("  原始: %v\n", input)
			fmt.Printf("  解压: %v\n", decompressed.Bytes())
		}
	}
}

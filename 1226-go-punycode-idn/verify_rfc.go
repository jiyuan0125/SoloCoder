package main

import (
	"fmt"

	"punycode-idn/pkg/punycode"
)

func main() {
	fmt.Println("=== 验证标准Punycode编解码 ===")
	fmt.Println()

	encodeTests := []struct {
		input    string
		expected string
	}{
		{"例", "xn--fsq"},
		{"例子", "xn--fsqu00a"},
		{"日本", "xn--wgv71a"},
		{"日本語", "xn--wgv71a119e"},
	}

	allEncodeMatch := true
	fmt.Println("--- 编码测试 ---")
	for _, tc := range encodeTests {
		result, err := punycode.Encode(tc.input)
		if err != nil {
			fmt.Printf("[错误] Encode(%q): %v\n", tc.input, err)
			allEncodeMatch = false
			continue
		}
		if result == tc.expected {
			fmt.Printf("[匹配] Encode(%q) = %q\n", tc.input, result)
		} else {
			fmt.Printf("[不匹配] Encode(%q)\n", tc.input)
			fmt.Printf("  期望: %q\n", tc.expected)
			fmt.Printf("  实际: %q\n", result)
			allEncodeMatch = false
		}
	}

	decodeTests := []struct {
		input    string
		expected string
	}{
		{"xn--fsq", "例"},
		{"xn--fsqu00a", "例子"},
		{"xn--wgv71a", "日本"},
		{"xn--wgv71a119e", "日本語"},
	}

	allDecodeMatch := true
	fmt.Println("\n--- 解码测试 ---")
	for _, tc := range decodeTests {
		result, err := punycode.Decode(tc.input)
		if err != nil {
			fmt.Printf("[错误] Decode(%q): %v\n", tc.input, err)
			allDecodeMatch = false
			continue
		}
		if result == tc.expected {
			fmt.Printf("[匹配] Decode(%q) = %q\n", tc.input, result)
		} else {
			fmt.Printf("[不匹配] Decode(%q)\n", tc.input)
			fmt.Printf("  期望: %q\n", tc.expected)
			fmt.Printf("  实际: %q\n", result)
			allDecodeMatch = false
		}
	}

	fmt.Println()
	if allEncodeMatch && allDecodeMatch {
		fmt.Println("=== 所有测试通过！与RFC 3492标准一致 ===")
	} else {
		fmt.Println("=== 存在测试失败 ===")
	}
}

//go:build ignore

package main

import (
	"encoding/hex"
	"fmt"

	"encoding-detector/pkg/detector"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"bytes"
	"io/ioutil"
)

func toBig5(text string) []byte {
	encoder := traditionalchinese.Big5.NewEncoder()
	reader := transform.NewReader(bytes.NewReader([]byte(text)), encoder)
	result, _ := ioutil.ReadAll(reader)
	return result
}

func toGBK(text string) []byte {
	encoder := simplifiedchinese.GBK.NewEncoder()
	reader := transform.NewReader(bytes.NewReader([]byte(text)), encoder)
	result, _ := ioutil.ReadAll(reader)
	return result
}

func toShiftJIS(text string) []byte {
	encoder := japanese.ShiftJIS.NewEncoder()
	reader := transform.NewReader(bytes.NewReader([]byte(text)), encoder)
	result, _ := ioutil.ReadAll(reader)
	return result
}

func main() {
	testCases := []struct {
		name     string
		encoding string
		data     []byte
	}{
		{"短文本Big5: Hello World 你好", "big5", toBig5("Hello World 你好")},
		{"短文本Big5: 你好世界", "big5", toBig5("你好世界")},
		{"长文本Big5", "big5", toBig5("這是一個測試文本，用來測試Big5編碼的檢測功能。歡迎來到台灣！")},
		{"短文本GBK: Hello World 你好", "gbk", toGBK("Hello World 你好")},
		{"短文本Shift_JIS: Hello World こんにちは", "shift_jis", toShiftJIS("Hello World こんにちは")},
	}

	fmt.Println("=== 编码检测测试 ===\n")

	for _, tc := range testCases {
		result := detector.Detect(tc.data)
		status := "✓ PASS"
		if result.Encoding != tc.encoding {
			status = "✗ FAIL"
		}
		fmt.Printf("%s\n", status)
		fmt.Printf("  测试: %s\n", tc.name)
		fmt.Printf("  期望: %s\n", tc.encoding)
		fmt.Printf("  实际: %s (置信度: %.2f)\n", result.Encoding, result.Confidence)
		fmt.Printf("  字节: %s\n", hex.EncodeToString(tc.data))
		fmt.Println()
	}
}

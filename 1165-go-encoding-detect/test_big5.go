//go:build ignore

package main

import (
	"encoding/hex"
	"fmt"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"bytes"
	"io/ioutil"
)

func main() {
	chars := []string{"你", "好", "世", "界", "這", "是", "一", "個", "測", "試"}
	
	for _, c := range chars {
		utf8Bytes := []byte(c)
		fmt.Printf("字符: %s (UTF-8: %s)\n", c, hex.EncodeToString(utf8Bytes))
		
		encoder := traditionalchinese.Big5.NewEncoder()
		reader := transform.NewReader(bytes.NewReader(utf8Bytes), encoder)
		big5Bytes, err := ioutil.ReadAll(reader)
		if err != nil {
			fmt.Printf("  编码错误: %v\n", err)
			continue
		}
		
		fmt.Printf("  Big5: %s\n", hex.EncodeToString(big5Bytes))
		if len(big5Bytes) == 2 {
			fmt.Printf("  第一字节: 0x%02X (%d), 第二字节: 0x%02X (%d)\n", 
				big5Bytes[0], big5Bytes[0], big5Bytes[1], big5Bytes[1])
			fmt.Printf("  第二字节范围检查:\n")
			fmt.Printf("    Big5第二字节范围1 (0x40-0x7E): %v\n", 
				big5Bytes[1] >= 0x40 && big5Bytes[1] <= 0x7E)
			fmt.Printf("    Big5第二字节范围2 (0xA1-0xFE): %v\n", 
				big5Bytes[1] >= 0xA1 && big5Bytes[1] <= 0xFE)
			fmt.Printf("    Shift_JIS第二字节范围1 (0x40-0x7E): %v\n", 
				big5Bytes[1] >= 0x40 && big5Bytes[1] <= 0x7E)
			fmt.Printf("    Shift_JIS第二字节范围2 (0x80-0xFC): %v\n", 
				big5Bytes[1] >= 0x80 && big5Bytes[1] <= 0xFC)
			fmt.Printf("    第二字节在0x80-0x9F (Shift_JIS专属): %v\n", 
				big5Bytes[1] >= 0x80 && big5Bytes[1] <= 0x9F)
		}
		fmt.Println()
	}
}

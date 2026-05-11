//go:build ignore

package main

import (
	"fmt"

	"go-lunar-gcal/pkg/jieqi"
)

func main() {
	fmt.Println("=== 测试2024年二十四节气 ===")

	jieqis, err := jieqi.GetJieqi(2024)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}

	for _, jq := range jieqis {
		fmt.Printf("%s: %s\n", jq.Name, jq.Time.Format("2006-01-02 15:04"))
	}

	fmt.Println()
	fmt.Println("=== 重点验证几个节气 ===")
	for _, jq := range jieqis {
		switch jq.Name {
		case "立春":
			fmt.Printf("立春: %s (预期: 2024-02-04 16:26左右)\n", jq.Time.Format("2006-01-02 15:04"))
		case "春分":
			fmt.Printf("春分: %s (预期: 2024-03-20)\n", jq.Time.Format("2006-01-02 15:04"))
		case "夏至":
			fmt.Printf("夏至: %s (预期: 2024-06-21 04:50左右)\n", jq.Time.Format("2006-01-02 15:04"))
		case "秋分":
			fmt.Printf("秋分: %s (预期: 2024-09-22 02:13左右)\n", jq.Time.Format("2006-01-02 15:04"))
		case "冬至":
			fmt.Printf("冬至: %s (预期: 2024-12-21 17:13左右)\n", jq.Time.Format("2006-01-02 15:04"))
		}
	}
}

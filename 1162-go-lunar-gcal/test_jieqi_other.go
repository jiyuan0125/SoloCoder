//go:build ignore

package main

import (
	"fmt"

	"go-lunar-gcal/pkg/jieqi"
)

func main() {
	fmt.Println("=== 2023年二十四节气 ===")
	jq, _ := jieqi.GetJieqi(2023)
	for _, j := range jq {
		fmt.Printf("%s: %s\n", j.Name, j.Time.Format("2006-01-02 15:04"))
	}

	fmt.Println()
	fmt.Println("=== 2025年二十四节气 ===")
	jq2, _ := jieqi.GetJieqi(2025)
	for _, j := range jq2 {
		fmt.Printf("%s: %s\n", j.Name, j.Time.Format("2006-01-02 15:04"))
	}
}

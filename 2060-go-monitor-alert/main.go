package main

import (
	"fmt"
	"os"

	"monitor-alert/cmd"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("使用 'monitor-alert --help' 查看可用命令")
		fmt.Println("示例:")
		fmt.Println("  monitor-alert rules --config config.yaml")
		fmt.Println("  monitor-alert start --config config.yaml")
		return
	}
	cmd.Execute()
}

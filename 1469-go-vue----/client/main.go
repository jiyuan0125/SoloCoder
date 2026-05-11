package main

import (
	"fmt"
	"os"
)

type Command interface {
	Name() string
	Description() string
	Execute(args []string) error
}

var commands = []Command{
	&SupervisoryCommand{},
	&AcceptanceCommand{},
	&IssueCommand{},
	&DailyLogCommand{},
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	commandName := os.Args[1]
	args := os.Args[2:]

	for _, cmd := range commands {
		if cmd.Name() == commandName {
			if err := cmd.Execute(args); err != nil {
				fmt.Printf("错误: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	fmt.Printf("未知命令: %s\n\n", commandName)
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Println("监理日志系统客户端")
	fmt.Println("\n用法:")
	fmt.Println("  client <command> [arguments]")
	fmt.Println("\n命令:")
	for _, cmd := range commands {
		fmt.Printf("  %-20s %s\n", cmd.Name(), cmd.Description())
	}
	fmt.Println("\n示例:")
	fmt.Println("  client supervisory --help")
	fmt.Println("  client acceptance --help")
}

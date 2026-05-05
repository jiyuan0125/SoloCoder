package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	parseCmd := flag.NewFlagSet("parse", flag.ExitOnError)
	validateCmd := flag.NewFlagSet("validate", flag.ExitOnError)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "parse":
		parseCmd.Parse(os.Args[2:])
		if parseCmd.NArg() < 1 {
			fmt.Println("错误: 请提供身份证号")
			os.Exit(1)
		}
		idNumber := parseCmd.Arg(0)
		handleParse(idNumber)

	case "validate":
		validateCmd.Parse(os.Args[2:])
		if validateCmd.NArg() < 1 {
			fmt.Println("错误: 请提供身份证号")
			os.Exit(1)
		}
		idNumber := validateCmd.Arg(0)
		handleValidate(idNumber)

	case "help", "-h", "--help":
		printUsage()
		os.Exit(0)

	default:
		fmt.Printf("未知命令: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("身份证号解析和校验工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  client parse <身份证号>   解析身份证号信息")
	fmt.Println("  client validate <身份证号>  校验身份证号合法性")
	fmt.Println("  client help                 显示帮助信息")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client parse 110101199001011234")
	fmt.Println("  client validate 110101199001011234")
}

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

type Command string

const (
	CommandGenerate  Command = "generate"
	CommandHistory   Command = "history"
	CommandGet       Command = "get"
	CommandHelp      Command = "help"
)

type GenerateOptions struct {
	RepoPath string
	Since    string
	Author   string
	Format   string
	Output   string
}

type HistoryOptions struct {
	Limit int
}

type GetOptions struct {
	ReportID string
	Format   string
	Output   string
}

type ClientOptions struct {
	ServerAddr string
	Command    Command
	Generate   *GenerateOptions
	History    *HistoryOptions
	Get        *GetOptions
}

func ParseArgs() (*ClientOptions, error) {
	opts := &ClientOptions{
		ServerAddr: "localhost:8765",
	}

	if len(os.Args) < 2 {
		return nil, fmt.Errorf("请指定命令: generate, history, get, help")
	}

	cmd := os.Args[1]
	switch cmd {
	case "generate":
		opts.Command = CommandGenerate
		if err := parseGenerateArgs(opts); err != nil {
			return nil, err
		}
	case "history":
		opts.Command = CommandHistory
		if err := parseHistoryArgs(opts); err != nil {
			return nil, err
		}
	case "get":
		opts.Command = CommandGet
		if err := parseGetArgs(opts); err != nil {
			return nil, err
		}
	case "help", "-h", "--help":
		opts.Command = CommandHelp
		PrintHelp()
		os.Exit(0)
	default:
		return nil, fmt.Errorf("未知命令: %s", cmd)
	}

	return opts, nil
}

func parseGenerateArgs(opts *ClientOptions) error {
	genOpts := &GenerateOptions{}
	flagSet := flag.NewFlagSet("generate", flag.ExitOnError)

	flagSet.StringVar(&genOpts.RepoPath, "repo", "", "Git 仓库路径 (必需)")
	flagSet.StringVar(&genOpts.Since, "since", "1 week ago", "时间范围，如 \"2 days ago\", \"1 week ago\"")
	flagSet.StringVar(&genOpts.Author, "author", "", "作者名称，不指定则统计所有人")
	flagSet.StringVar(&genOpts.Format, "format", "", "输出格式，支持 markdown")
	flagSet.StringVar(&genOpts.Output, "output", "", "输出文件路径")
	flagSet.StringVar(&opts.ServerAddr, "server", "localhost:8765", "服务端地址")

	if len(os.Args) > 2 {
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			return err
		}
	}

	if genOpts.RepoPath == "" {
		return fmt.Errorf("必须指定 --repo 参数")
	}

	absPath, err := filepath.Abs(genOpts.RepoPath)
	if err != nil {
		return fmt.Errorf("解析仓库路径失败: %w", err)
	}
	genOpts.RepoPath = absPath

	opts.Generate = genOpts
	return nil
}

func parseHistoryArgs(opts *ClientOptions) error {
	histOpts := &HistoryOptions{}
	flagSet := flag.NewFlagSet("history", flag.ExitOnError)

	flagSet.IntVar(&histOpts.Limit, "limit", 0, "显示最近的 N 条记录，0 表示显示所有")
	flagSet.StringVar(&opts.ServerAddr, "server", "localhost:8765", "服务端地址")

	if len(os.Args) > 2 {
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			return err
		}
	}

	opts.History = histOpts
	return nil
}

func parseGetArgs(opts *ClientOptions) error {
	getOpts := &GetOptions{}
	flagSet := flag.NewFlagSet("get", flag.ExitOnError)

	flagSet.StringVar(&getOpts.ReportID, "id", "", "报告 ID (必需)")
	flagSet.StringVar(&getOpts.Format, "format", "", "输出格式，支持 markdown")
	flagSet.StringVar(&getOpts.Output, "output", "", "输出文件路径")
	flagSet.StringVar(&opts.ServerAddr, "server", "localhost:8765", "服务端地址")

	if len(os.Args) > 2 {
		if err := flagSet.Parse(os.Args[2:]); err != nil {
			return err
		}
	}

	if getOpts.ReportID == "" {
		return fmt.Errorf("必须指定 --id 参数")
	}

	opts.Get = getOpts
	return nil
}

func PrintHelp() {
	fmt.Println(`Go 日报/周报自动生成工具

用法:
  report-gen [command] [options]

命令:
  generate    生成新的报告
  history     查看历史报告列表
  get         获取指定的历史报告
  help        显示帮助信息

generate 命令选项:
  --repo      Git 仓库路径 (必需)
  --since     时间范围，如 "2 days ago", "1 week ago" (默认: 1 week ago)
  --author    作者名称，不指定则统计所有人
  --format    输出格式，支持 markdown
  --output    输出文件路径
  --server    服务端地址 (默认: localhost:8765)

history 命令选项:
  --limit     显示最近的 N 条记录，0 表示显示所有 (默认: 0)
  --server    服务端地址 (默认: localhost:8765)

get 命令选项:
  --id        报告 ID (必需)
  --format    输出格式，支持 markdown
  --output    输出文件路径
  --server    服务端地址 (默认: localhost:8765)

示例:
  report-gen generate --repo /path/to/repo --since "2 days ago" --author zhangsan
  report-gen generate --repo /path/to/repo --format markdown --output report.md
  report-gen history --limit 10
  report-gen get --id report_20240101_120000 --format markdown
`)
}

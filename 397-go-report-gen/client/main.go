package main

import (
	"fmt"
	"os"
)

func main() {
	opts, err := ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n\n", err)
		PrintHelp()
		os.Exit(1)
	}

	client := NewClient(opts.ServerAddr)
	if err := client.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	switch opts.Command {
	case CommandGenerate:
		handleGenerate(client, opts.Generate)
	case CommandHistory:
		handleHistory(client, opts.History)
	case CommandGet:
		handleGet(client, opts.Get)
	}
}

func handleGenerate(client *Client, opts *GenerateOptions) {
	report, err := client.GenerateReport(opts.RepoPath, opts.Since, opts.Author, opts.Format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	PrintReport(report, opts.Format)

	filePath, err := SaveReportToFile(report, opts.Format, opts.Output)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n警告: 保存报告到文件失败: %v\n", err)
	} else {
		fmt.Printf("\n报告已保存到: %s\n", filePath)
	}
}

func handleHistory(client *Client, opts *HistoryOptions) {
	history, err := client.ListHistory(opts.Limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	PrintHistory(history)
}

func handleGet(client *Client, opts *GetOptions) {
	report, err := client.GetReport(opts.ReportID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	PrintReport(report, opts.Format)

	if opts.Output != "" {
		filePath, err := SaveReportToFile(report, opts.Format, opts.Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\n警告: 保存报告到文件失败: %v\n", err)
		} else {
			fmt.Printf("\n报告已保存到: %s\n", filePath)
		}
	}
}

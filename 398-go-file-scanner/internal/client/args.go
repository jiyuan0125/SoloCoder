package client

import (
	"flag"
	"fmt"
	"os"
)

type ScanArgs struct {
	Directory    string
	ServerAddr   string
	ConfigFile   string
	ExcludePaths []string
	Severity     string
	OutputJSON   bool
}

func ParseArgs() (*ScanArgs, error) {
	var args ScanArgs

	flag.StringVar(&args.ServerAddr, "server", "localhost:8080", "Server address to connect to")
	flag.StringVar(&args.ConfigFile, "config", "", "Path to configuration file")
	flag.StringVar(&args.Severity, "severity", "", "Filter by severity level (low, medium, high, critical)")
	flag.BoolVar(&args.OutputJSON, "json", false, "Output results in JSON format")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] <directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s /path/to/scan\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -severity high /path/to/scan\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -json /path/to/scan\n", os.Args[0])
	}

	flag.Parse()

	// 第一个非flag参数是扫描目录
	if flag.NArg() < 1 {
		return nil, fmt.Errorf("missing directory argument")
	}
	args.Directory = flag.Arg(0)

	// 验证严重程度参数
	if args.Severity != "" {
		validSeverities := map[string]bool{
			"low":      true,
			"medium":   true,
			"high":     true,
			"critical": true,
		}
		if !validSeverities[args.Severity] {
			return nil, fmt.Errorf("invalid severity level: %s. Valid levels: low, medium, high, critical", args.Severity)
		}
	}

	return &args, nil
}

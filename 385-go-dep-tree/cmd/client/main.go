package main

import (
	"flag"
	"fmt"
	"os"
	
	"dep-tree/internal/client"
)

func main() {
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	filterStandard := flag.Bool("standard", false, "Show only standard library dependencies")
	filterThirdParty := flag.Bool("third-party", false, "Show only third-party dependencies")
	search := flag.String("search", "", "Search for specific package pattern")
	serverAddr := flag.String("server", "localhost:9876", "Server address (host:port)")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] /path/to/go/project\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Go Dependency Tree Analyzer\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s /path/to/project\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --json /path/to/project\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --standard /path/to/project\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --third-party /path/to/project\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s --search=\"net/http\" /path/to/project\n", os.Args[0])
	}
	
	flag.Parse()
	
	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(1)
	}
	
	projectPath := args[0]
	
	cfg := client.Config{
		ProjectPath:      projectPath,
		ServerAddr:       *serverAddr,
		JSONOutput:       *jsonOutput,
		FilterStandard:   *filterStandard,
		FilterThirdParty: *filterThirdParty,
		SearchPattern:    *search,
	}
	
	if err := client.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

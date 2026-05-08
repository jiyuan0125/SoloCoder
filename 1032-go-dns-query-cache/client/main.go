package main

import (
	"dns-cache-service/common"
	"fmt"
	"os"
	"strings"
)

func printUsage() {
	fmt.Println(`DNS Cache Client - A command-line tool for DNS Cache Service

Usage:
  dns-client <command> [options]

Commands:
  query <domain> [--types <type1,type2>]  Query a domain
  refresh <domain> [--types <type1,type2>]  Refresh a domain cache
  stats                                   Show cache statistics
  preload <domain1,domain2,...>           Preload domains into cache
  entries                                 List all cache entries
  help                                    Show this help message

Options:
  --server <url>                          Server URL (default: http://localhost:8080)
  --types <type1,type2,...>               Record types: A, AAAA, CNAME, MX, TXT

Examples:
  dns-client query google.com
  dns-client query google.com --types A,AAAA
  dns-client refresh google.com
  dns-client stats
  dns-client preload google.com,github.com
  dns-client entries
  dns-client query example.com --server http://localhost:8080`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := "http://localhost:8080"
	var types []common.RecordType
	var remainingArgs []string

	i := 1
	for i < len(os.Args) {
		arg := os.Args[i]
		switch arg {
		case "--server":
			if i+1 < len(os.Args) {
				serverURL = os.Args[i+1]
				i += 2
			} else {
				fmt.Println("Error: --server requires a value")
				os.Exit(1)
			}
		case "--types":
			if i+1 < len(os.Args) {
				types = parseTypes(os.Args[i+1])
				i += 2
			} else {
				fmt.Println("Error: --types requires a value")
				os.Exit(1)
			}
		default:
			remainingArgs = append(remainingArgs, arg)
			i++
		}
	}

	if len(remainingArgs) == 0 {
		printUsage()
		os.Exit(1)
	}

	command := remainingArgs[0]
	params := remainingArgs[1:]

	client := NewClient(serverURL)

	switch command {
	case "query":
		if len(params) < 1 {
			fmt.Println("Error: query requires a domain")
			os.Exit(1)
		}
		handleQuery(client, params[0], types)
	case "refresh":
		if len(params) < 1 {
			fmt.Println("Error: refresh requires a domain")
			os.Exit(1)
		}
		handleRefresh(client, params[0], types)
	case "stats":
		handleStats(client)
	case "preload":
		if len(params) < 1 {
			fmt.Println("Error: preload requires at least one domain")
			os.Exit(1)
		}
		domains := strings.Split(params[0], ",")
		handlePreload(client, domains)
	case "entries":
		handleEntries(client)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Error: unknown command '%s'\n", command)
		printUsage()
		os.Exit(1)
	}
}

func parseTypes(typesStr string) []common.RecordType {
	parts := strings.Split(typesStr, ",")
	types := make([]common.RecordType, 0, len(parts))
	for _, p := range parts {
		t := strings.ToUpper(strings.TrimSpace(p))
		switch t {
		case "A":
			types = append(types, common.TypeA)
		case "AAAA":
			types = append(types, common.TypeAAAA)
		case "CNAME":
			types = append(types, common.TypeCNAME)
		case "MX":
			types = append(types, common.TypeMX)
		case "TXT":
			types = append(types, common.TypeTXT)
		}
	}
	return types
}

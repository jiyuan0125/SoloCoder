package main

import (
    "bufio"
    "flag"
    "fmt"
    "net"
    "os"

    "batch-rename/protocol"
)

const (
    defaultServerAddr = "localhost:8888"
)

type renameFlags struct {
    prefix       string
    suffix       string
    sequence     bool
    date         bool
    dateFormat   string
    replaceOld   string
    replaceNew   string
    regexPattern string
    regexReplace string
    recursive    bool
    dryRun       bool
    serverAddr   string
    help         bool
}

func main() {
    flags := parseFlags()

    if flags.help {
        printUsage()
        os.Exit(0)
    }

    rules := collectRules(flags)
    if len(rules) == 0 {
        fmt.Println("Error: At least one rule is required")
        printUsage()
        os.Exit(1)
    }

    files := collectFilesFromArgs()

    req := protocol.Request{
        Type:      protocol.MsgTypePreview,
        Files:     files,
        Rules:     rules,
        Recursive: flags.recursive,
        DryRun:    flags.dryRun,
    }

    conn, err := net.Dial("tcp", flags.serverAddr)
    if err != nil {
        fmt.Printf("Error connecting to server: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    previewResp, err := sendPreviewRequest(writer, reader, &req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    printPreview(previewResp)

    if len(previewResp.Items) == 0 {
        fmt.Println("No files to process")
        os.Exit(0)
    }

    if flags.dryRun {
        fmt.Println("\n--dry-run: No changes will be made")
        os.Exit(0)
    }

    if !confirmExecution() {
        fmt.Println("Operation cancelled")
        os.Exit(0)
    }

    req.Type = protocol.MsgTypeExecute
    executeResp, err := sendExecuteRequest(writer, reader, &req)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }

    printExecuteResult(executeResp)
}

func parseFlags() *renameFlags {
    flags := &renameFlags{}

    flag.StringVar(&flags.prefix, "prefix", "", "Add prefix to file name")
    flag.StringVar(&flags.suffix, "suffix", "", "Add suffix to file name (before extension)")
    flag.BoolVar(&flags.sequence, "seq", false, "Replace {seq} with sequential number")
    flag.BoolVar(&flags.date, "date", false, "Replace {date} with file modification date")
    flag.StringVar(&flags.dateFormat, "date-format", "20060102", "Date format for {date} (Go format)")
    flag.StringVar(&flags.replaceOld, "replace-old", "", "Old string to replace")
    flag.StringVar(&flags.replaceNew, "replace-new", "", "New string to replace with")
    flag.StringVar(&flags.regexPattern, "regex-pattern", "", "Regex pattern for replacement")
    flag.StringVar(&flags.regexReplace, "regex-replace", "", "Replacement string for regex (use $1, $2 for groups)")
    flag.BoolVar(&flags.recursive, "recursive", false, "Process files recursively in subdirectories")
    flag.BoolVar(&flags.recursive, "r", false, "Short form of --recursive")
    flag.BoolVar(&flags.dryRun, "dry-run", false, "Only preview without executing")
    flag.StringVar(&flags.serverAddr, "server", defaultServerAddr, "Server address (host:port)")
    flag.BoolVar(&flags.help, "help", false, "Show this help message")
    flag.BoolVar(&flags.help, "h", false, "Short form of --help")

    flag.Parse()

    return flags
}

func collectRules(flags *renameFlags) []protocol.Rule {
    var rules []protocol.Rule

    if flags.prefix != "" {
        rules = append(rules, protocol.Rule{
            Type:   protocol.RuleTypePrefix,
            Prefix: flags.prefix,
        })
    }

    if flags.suffix != "" {
        rules = append(rules, protocol.Rule{
            Type:   protocol.RuleTypeSuffix,
            Suffix: flags.suffix,
        })
    }

    if flags.sequence {
        rules = append(rules, protocol.Rule{
            Type: protocol.RuleTypeSequence,
        })
    }

    if flags.date {
        rules = append(rules, protocol.Rule{
            Type:       protocol.RuleTypeDate,
            DateFormat: flags.dateFormat,
        })
    }

    if flags.replaceOld != "" {
        rules = append(rules, protocol.Rule{
            Type:      protocol.RuleTypeReplace,
            OldString: flags.replaceOld,
            NewString: flags.replaceNew,
        })
    }

    if flags.regexPattern != "" {
        rules = append(rules, protocol.Rule{
            Type:        protocol.RuleTypeRegexReplace,
            Pattern:     flags.regexPattern,
            Replacement: flags.regexReplace,
        })
    }

    return rules
}

func collectFilesFromArgs() []string {
    return flag.Args()
}

func printUsage() {
    fmt.Println(`Batch Rename Tool - Client

Usage:
  batch-rename [options] [files/directories...]

Options:
  --prefix <string>        Add prefix to file name
  --suffix <string>        Add suffix to file name (before extension)
  --seq                    Replace {seq} with sequential number (auto-padded)
  --date                   Replace {date} with file modification date
  --date-format <format>   Date format for {date} (default: 20060102)
  --replace-old <string>   Old string to replace
  --replace-new <string>   New string to replace with
  --regex-pattern <regex>  Regex pattern for replacement
  --regex-replace <string> Replacement string for regex (use $1, $2 for groups)
  -r, --recursive          Process files recursively in subdirectories
  --dry-run                Only preview without executing
  --server <addr>          Server address (default: localhost:8888)
  -h, --help               Show this help message

Rules are applied in the order they are specified.

Examples:
  # Add prefix "photo_" to all jpg files
  batch-rename --prefix "photo_" *.jpg

  # Add sequence number
  batch-rename --seq "image_{seq}.jpg" *.jpg

  # Add date from file modification time
  batch-rename --date "vacation_{date}_{seq}.jpg" *.jpg

  # Replace "IMG" with "Photo"
  batch-rename --replace-old "IMG" --replace-new "Photo" *.jpg

  # Regex replace: extract number and reformat
  batch-rename --regex-pattern "img(\d+)" --regex-replace "photo_$1" *.jpg

  # Multiple rules
  batch-rename --prefix "2024_" --seq "vacation_{seq}" *.jpg

  # Recursive
  batch-rename -r --prefix "backup_" ./photos

  # Dry run
  batch-rename --dry-run --prefix "test_" *.jpg`)
}

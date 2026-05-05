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
    prefix         string
    suffix         string
    sequence       bool
    date           bool
    dateFormat     string
    replaceOld     string
    replaceNew     string
    regexPattern   string
    regexReplace   string
    recursive      bool
    dryRun         bool
    serverAddr     string
    help           bool
    template       string
    sequenceFormat string
    dateFormatArg  string
    history        bool
}

func main() {
    flags := parseFlags()

    if flags.help {
        printUsage()
        os.Exit(0)
    }

    conn, err := net.Dial("tcp", flags.serverAddr)
    if err != nil {
        fmt.Printf("Error connecting to server: %v\n", err)
        os.Exit(1)
    }
    defer conn.Close()

    reader := bufio.NewReader(conn)
    writer := bufio.NewWriter(conn)

    if flags.history {
        handleHistoryRequest(writer, reader)
        return
    }

    rules := collectRules(flags)
    if len(rules) == 0 {
        fmt.Println("Error: At least one rule is required")
        printUsage()
        os.Exit(1)
    }

    files := collectFiles()

    req := protocol.Request{
        Type:      protocol.MsgTypePreview,
        Files:     files,
        Rules:     rules,
        Recursive: flags.recursive,
        DryRun:    flags.dryRun,
    }

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
    flag.BoolVar(&flags.sequence, "seq", false, "Use sequence number as new name (equivalent to --template \"{seq}\")")
    flag.StringVar(&flags.sequenceFormat, "seq-format", "", "Sequence format template, e.g., \"photo_{seq}\"")
    flag.BoolVar(&flags.date, "date", false, "Use date in file name (equivalent to --template \"{date}_{name}\")")
    flag.StringVar(&flags.dateFormat, "date-format", "20060102", "Date format for {date} (Go format, e.g., 2006-01-02)")
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
    flag.StringVar(&flags.template, "template", "", "Template for new file name (supports {seq}, {date}, {name} placeholders)")
    flag.BoolVar(&flags.history, "history", false, "Show rename history")

    flag.Parse()

    return flags
}

func collectRules(flags *renameFlags) []protocol.Rule {
    var rules []protocol.Rule

    if flags.template != "" {
        rules = append(rules, protocol.Rule{
            Type:     protocol.RuleTypeTemplate,
            Template: flags.template,
        })
    } else if flags.sequenceFormat != "" {
        rules = append(rules, protocol.Rule{
            Type:     protocol.RuleTypeTemplate,
            Template: flags.sequenceFormat,
        })
    } else if flags.sequence {
        rules = append(rules, protocol.Rule{
            Type:     protocol.RuleTypeTemplate,
            Template: "{seq}",
        })
    } else if flags.date {
        rules = append(rules, protocol.Rule{
            Type:     protocol.RuleTypeTemplate,
            Template: "{date}_{name}",
        })
    }

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

func collectFiles() []string {
    args := flag.Args()
    if len(args) > 0 {
        return args
    }

    stdinFiles := collectFilesFromStdin()
    if len(stdinFiles) > 0 {
        return stdinFiles
    }

    return []string{}
}

func collectFilesFromStdin() []string {
    var files []string
    
    fi, err := os.Stdin.Stat()
    if err != nil {
        return files
    }
    
    if (fi.Mode() & os.ModeNamedPipe) == 0 {
        return files
    }

    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        line := scanner.Text()
        if line != "" {
            files = append(files, line)
        }
    }

    if scanner.Err() != nil {
        return files
    }

    return files
}

func printUsage() {
    fmt.Println(`Batch Rename Tool - Client

Usage:
  batch-rename [options] [files/directories...]
  ls *.jpg | batch-rename [options]

Options:
  --template <string>       Template for new file name (supports placeholders)
                            Placeholders:
                              {seq}  - Sequential number (auto-padded)
                              {date} - File modification date
                              {name} - Original file name (without extension)
                            
  --seq                     Use sequence number as new name (equivalent to --template "{seq}")
  --seq-format <string>     Sequence format template, e.g., "photo_{seq}"
  --date                    Use date in file name (equivalent to --template "{date}_{name}")
  --date-format <format>    Date format for {date} (Go format, default: 20060102)
                            
  --prefix <string>         Add prefix to file name
  --suffix <string>         Add suffix to file name (before extension)
  --replace-old <string>    Old string to replace
  --replace-new <string>    New string to replace with
  --regex-pattern <regex>   Regex pattern for replacement
  --regex-replace <string>  Replacement string for regex (use $1, $2 for groups)
                            
  -r, --recursive           Process files recursively in subdirectories
  --dry-run                 Only preview without executing
  --history                 Show rename history
  --server <addr>           Server address (default: localhost:8888)
  -h, --help                Show this help message

Rules are applied in the order they are specified.

Examples:
  # Rename to sequential numbers: 001.jpg, 002.jpg, etc.
  batch-rename --seq *.jpg

  # Rename using template
  batch-rename --template "photo_{seq}" *.jpg

  # Rename using date and sequence
  batch-rename --template "vacation_{date}_{seq}" *.jpg

  # Add prefix
  batch-rename --prefix "2024_" *.jpg

  # Replace string
  batch-rename --replace-old "IMG" --replace-new "Photo" *.jpg

  # Regex replace: extract number and reformat
  batch-rename --regex-pattern "img(\d+)" --regex-replace "photo_$1" *.jpg

  # Multiple rules: template + prefix
  batch-rename --template "{seq}" --prefix "photo_" *.jpg

  # Read from pipe
  ls *.jpg | batch-rename --template "image_{seq}"

  # Recursive
  batch-rename -r --template "{name}_{seq}" ./photos

  # Dry run
  batch-rename --dry-run --template "test_{seq}" *.jpg

  # Show history
  batch-rename --history`)
}

package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Coverage analysis server URL")
	reportType := flag.String("report", "all", "Report type: all, summary, function, package, line")
	diffMode := flag.Bool("diff", false, "Diff mode: compare two coverprofile files")
	oldFile := flag.String("old", "", "Old coverprofile file for diff mode")
	newFile := flag.String("new", "", "New coverprofile file for diff mode")
	id := flag.String("id", "", "Analysis ID to retrieve (skip upload)")
	
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <coverprofile-file>\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s coverage.out\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -report=summary coverage.out\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -diff -old=old.out -new=new.out\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -id=abc123 -report=function\n", os.Args[0])
	}

	flag.Parse()

	client := NewAPIClient(*serverURL)
	printer := NewReportPrinter(nil)

	if *diffMode {
		if *oldFile == "" || *newFile == "" {
			fmt.Fprintln(os.Stderr, "Error: -old and -new are required for diff mode")
			flag.Usage()
			os.Exit(1)
		}

		runDiffMode(client, printer, *oldFile, *newFile)
		return
	}

	if *id != "" {
		runRetrieveMode(client, printer, *id, *reportType)
		return
	}

	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "Error: coverprofile file is required")
		flag.Usage()
		os.Exit(1)
	}

	filePath := args[0]
	runUploadMode(client, printer, filePath, *reportType)
}

func runUploadMode(client *APIClient, printer *ReportPrinter, filePath string, reportType string) {
	fmt.Printf("Uploading coverprofile: %s\n", filePath)

	uploadResp, err := client.UploadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Upload successful! Analysis ID: %s\n", uploadResp.ID)
	fmt.Printf("Coverage mode: %s\n", uploadResp.Mode)
	fmt.Printf("Use '-id=%s' to retrieve this analysis later\n", uploadResp.ID)

	printReports(client, printer, uploadResp.ID, reportType)
}

func runRetrieveMode(client *APIClient, printer *ReportPrinter, id string, reportType string) {
	fmt.Printf("Retrieving analysis: %s\n", id)
	printReports(client, printer, id, reportType)
}

func runDiffMode(client *APIClient, printer *ReportPrinter, oldFile, newFile string) {
	fmt.Printf("Diff mode: comparing %s vs %s\n", oldFile, newFile)

	fmt.Println("\nUploading old coverprofile...")
	oldResp, err := client.UploadFile(oldFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading old file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Old analysis ID: %s\n", oldResp.ID)

	fmt.Println("\nUploading new coverprofile...")
	newResp, err := client.UploadFile(newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error uploading new file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("New analysis ID: %s\n", newResp.ID)

	fmt.Println("\nGetting diff report...")
	diff, err := client.GetDiff(oldResp.ID, newResp.ID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
		os.Exit(1)
	}

	printer.PrintDiff(diff)

	if len(diff.Decreased) > 0 {
		os.Exit(2)
	}
}

func printReports(client *APIClient, printer *ReportPrinter, id string, reportType string) {
	switch reportType {
	case "summary":
		summary, err := client.GetSummary(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting summary: %v\n", err)
			os.Exit(1)
		}
		printer.PrintSummary(summary)

	case "function":
		report, err := client.GetFunctionCoverage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting function coverage: %v\n", err)
			os.Exit(1)
		}
		printer.PrintFunctionCoverage(report)

	case "package":
		report, err := client.GetPackageCoverage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting package coverage: %v\n", err)
			os.Exit(1)
		}
		printer.PrintPackageCoverage(report)

	case "line":
		report, err := client.GetLineCoverage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting line coverage: %v\n", err)
			os.Exit(1)
		}
		printer.PrintLineCoverage(report)

	case "all":
		summary, err := client.GetSummary(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting summary: %v\n", err)
			os.Exit(1)
		}
		printer.PrintSummary(summary)

		pkgReport, err := client.GetPackageCoverage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting package coverage: %v\n", err)
			os.Exit(1)
		}
		printer.PrintPackageCoverage(pkgReport)

		funcReport, err := client.GetFunctionCoverage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting function coverage: %v\n", err)
			os.Exit(1)
		}
		printer.PrintFunctionCoverage(funcReport)

		lineReport, err := client.GetLineCoverage(id)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting line coverage: %v\n", err)
			os.Exit(1)
		}
		printer.PrintLineCoverage(lineReport)

	default:
		fmt.Fprintf(os.Stderr, "Unknown report type: %s\n", reportType)
		os.Exit(1)
	}
}

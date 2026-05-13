package cmd

import (
	"fmt"
	"os"

	"github.com/solocoder/concurrent-crawler/pkg/crawler"
	"github.com/spf13/cobra"
)

var (
	startURL    string
	maxDepth    int
	concurrency int
	rateLimit   int
	timeout     int
	outputDir   string
)

var rootCmd = &cobra.Command{
	Use:   "concurrent-crawler",
	Short: "A concurrent web crawler built with Go and Cobra",
	Long: `A concurrent web crawler that fetches pages starting from a given URL,
following links up to a specified depth. Features include:
- Concurrent crawling with configurable workers
- Domain-based rate limiting
- Robots.txt compliance
- Progress persistence and resume capability
- Real-time progress display`,
	Run: func(cmd *cobra.Command, args []string) {
		if startURL == "" {
			fmt.Fprintln(os.Stderr, "Error: start URL is required")
			os.Exit(1)
		}

		config := &crawler.Config{
			StartURL:    startURL,
			MaxDepth:    maxDepth,
			Concurrency: concurrency,
			RateLimit:   rateLimit,
			Timeout:     timeout,
			OutputDir:   outputDir,
		}

		c := crawler.NewCrawler(config)
		if err := c.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&startURL, "url", "u", "", "Starting URL to crawl (required)")
	rootCmd.Flags().IntVarP(&maxDepth, "depth", "d", 2, "Maximum depth to crawl")
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 10, "Number of concurrent workers (max 100)")
	rootCmd.Flags().IntVarP(&rateLimit, "rate", "r", 2, "Requests per second per domain")
	rootCmd.Flags().IntVarP(&timeout, "timeout", "t", 0, "Timeout in seconds (0 for no timeout)")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "./crawled", "Output directory for crawled pages")

	rootCmd.MarkFlagRequired("url")
}

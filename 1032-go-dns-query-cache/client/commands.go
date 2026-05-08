package main

import (
	"dns-cache-service/common"
	"fmt"
)

func handleQuery(client *Client, domain string, types []common.RecordType) {
	result, err := client.Query(domain, types)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Query result for %s:\n\n", result.Domain)
	for _, r := range result.Results {
		fmt.Printf("Type: %s\n", r.Type)
		if r.IsNXDOMAIN {
			fmt.Printf("  Status: NXDOMAIN (Domain not found)\n")
		} else {
			fmt.Printf("  Records:\n")
			for _, rec := range r.Records {
				fmt.Printf("    %s (TTL: %ds)\n", rec.Value, rec.TTL)
			}
		}
		fmt.Printf("  Hit Cache: %v\n", r.HitCache)
		fmt.Printf("  Remaining TTL: %ds\n", r.RemainingTTL)
		fmt.Println()
	}
}

func handleRefresh(client *Client, domain string, types []common.RecordType) {
	result, err := client.Refresh(domain, types)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Refreshed %s:\n\n", result.Domain)
	for _, r := range result.Results {
		fmt.Printf("Type: %s\n", r.Type)
		if r.IsNXDOMAIN {
			fmt.Printf("  Status: NXDOMAIN\n")
		} else {
			fmt.Printf("  Records: %d\n", len(r.Records))
			for _, rec := range r.Records {
				fmt.Printf("    %s (TTL: %ds)\n", rec.Value, rec.TTL)
			}
		}
		fmt.Println()
	}
}

func handleStats(client *Client) {
	stats, err := client.Stats()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Cache Statistics:")
	fmt.Printf("  Total Entries: %d\n", stats.TotalEntries)
	fmt.Printf("  Total Queries: %d\n", stats.TotalQueries)
	fmt.Printf("  Cache Hits: %d\n", stats.CacheHits)
	fmt.Printf("  Cache Misses: %d\n", stats.CacheMisses)
	fmt.Printf("  Hit Rate: %.2f%%\n", stats.HitRate*100)
}

func handlePreload(client *Client, domains []string) {
	result, err := client.Preload(domains)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Preload Result:")
	fmt.Printf("  Total: %d\n", result.Total)
	fmt.Printf("  Successful: %d\n", result.Successful)
	fmt.Printf("  Failed: %d\n", result.Failed)
}

func handleEntries(client *Client) {
	entries, err := client.Entries()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Cache Entries (%d total):\n\n", entries.Total)
	for _, e := range entries.Entries {
		fmt.Printf("Domain: %s\n", e.Domain)
		fmt.Printf("Type: %s\n", e.Type)
		if e.IsNXDOMAIN {
			fmt.Printf("Status: NXDOMAIN\n")
		} else {
			fmt.Printf("Records:\n")
			for _, rec := range e.Records {
				fmt.Printf("  %s (TTL: %ds)\n", rec.Value, rec.TTL)
			}
		}
		fmt.Printf("Remaining TTL: %ds\n", e.RemainingTTL)
		fmt.Printf("Expires At: %s\n", e.ExpiresAt)
		fmt.Println()
	}
}

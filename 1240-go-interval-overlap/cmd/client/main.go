package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"interval-overlap/common"
)

const (
	defaultServerURL = "http://localhost:8080"
	envServerURLKey  = "SERVER_URL"
)

func main() {
	serverURL := getServerURL()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "add":
		handleAdd(serverURL, args)
	case "list":
		handleList(serverURL, args)
	case "conflict":
		handleConflict(serverURL, args)
	case "delete":
		handleDelete(serverURL, args)
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func getServerURL() string {
	serverURL := defaultServerURL
	if envURL := os.Getenv(envServerURLKey); envURL != "" {
		serverURL = envURL
	}
	return serverURL
}

func printUsage() {
	fmt.Println("Usage: client <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  add       Add a new booking")
	fmt.Println("  list      List bookings for a resource")
	fmt.Println("  conflict  Check for conflicts without booking")
	fmt.Println("  delete    Delete a booking by ID")
	fmt.Println("\nOptions for add:")
	fmt.Println("  --resource <name>    Resource name (required)")
	fmt.Println("  --start <time>       Start time (RFC3339, required)")
	fmt.Println("  --end <time>         End time (RFC3339, required)")
	fmt.Println("  --booker <name>      Booker name (required)")
	fmt.Println("\nOptions for list:")
	fmt.Println("  --resource <name>    Resource name (required)")
	fmt.Println("  --start <time>       Start time (optional)")
	fmt.Println("  --end <time>         End time (optional)")
	fmt.Println("\nOptions for conflict:")
	fmt.Println("  --resource <name>    Resource name (required)")
	fmt.Println("  --start <time>       Start time (RFC3339, required)")
	fmt.Println("  --end <time>         End time (RFC3339, required)")
	fmt.Println("\nOptions for delete:")
	fmt.Println("  --id <booking_id>    Booking ID (required)")
	fmt.Println("  --operator <name>    Operator name (required)")
	fmt.Println("  --admin              Operator is admin (optional)")
}

func handleAdd(serverURL string, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	resource := fs.String("resource", "", "Resource name")
	start := fs.String("start", "", "Start time (RFC3339)")
	end := fs.String("end", "", "End time (RFC3339)")
	booker := fs.String("booker", "", "Booker name")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *resource == "" || *start == "" || *end == "" || *booker == "" {
		fmt.Println("Error: resource, start, end, and booker are required")
		fs.Usage()
		os.Exit(1)
	}

	req := common.AddBookingRequest{
		Resource: *resource,
		Start:    *start,
		End:      *end,
		Booker:   *booker,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/bookings/add", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.AddBookingResponse
	json.Unmarshal(respBody, &result)

	if result.Success {
		fmt.Printf("Booking created successfully\nID: %s\n", result.BookingID)
	} else {
		fmt.Printf("Error: %s\n", result.Message)
		if len(result.Conflicts) > 0 {
			fmt.Println("\nConflicts:")
			printBookings(result.Conflicts)
		}
		os.Exit(1)
	}
}

func handleList(serverURL string, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	resource := fs.String("resource", "", "Resource name")
	start := fs.String("start", "", "Start time (RFC3339)")
	end := fs.String("end", "", "End time (RFC3339)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *resource == "" {
		fmt.Println("Error: resource is required")
		fs.Usage()
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/bookings/list?resource=%s", serverURL, *resource)
	if *start != "" && *end != "" {
		url += fmt.Sprintf("&start=%s&end=%s", *start, *end)
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.ListBookingsResponse
	json.Unmarshal(respBody, &result)

	if len(result.Bookings) == 0 {
		fmt.Println("No bookings found")
		return
	}

	printBookings(result.Bookings)
}

func handleConflict(serverURL string, args []string) {
	fs := flag.NewFlagSet("conflict", flag.ExitOnError)
	resource := fs.String("resource", "", "Resource name")
	start := fs.String("start", "", "Start time (RFC3339)")
	end := fs.String("end", "", "End time (RFC3339)")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *resource == "" || *start == "" || *end == "" {
		fmt.Println("Error: resource, start, and end are required")
		fs.Usage()
		os.Exit(1)
	}

	req := common.CheckConflictRequest{
		Resource: *resource,
		Start:    *start,
		End:      *end,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/bookings/conflict", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.CheckConflictResponse
	json.Unmarshal(respBody, &result)

	if result.HasConflict {
		fmt.Println("Conflicts found!")
		printBookings(result.Conflicts)
	} else {
		fmt.Println("No conflicts")
	}
}

func handleDelete(serverURL string, args []string) {
	fs := flag.NewFlagSet("delete", flag.ExitOnError)
	id := fs.String("id", "", "Booking ID")
	operator := fs.String("operator", "", "Operator name")
	isAdmin := fs.Bool("admin", false, "Operator is admin")
	if err := fs.Parse(args); err != nil {
		os.Exit(1)
	}

	if *id == "" || *operator == "" {
		fmt.Println("Error: id and operator are required")
		fs.Usage()
		os.Exit(1)
	}

	req := common.DeleteBookingRequest{
		BookingID: *id,
		Operator:  *operator,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	httpReq, err := http.NewRequest("DELETE", serverURL+"/bookings/delete", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if *isAdmin {
		httpReq.Header.Set("X-Admin", "true")
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result common.DeleteBookingResponse
	json.Unmarshal(respBody, &result)

	if result.Success {
		fmt.Println("Booking deleted successfully")
	} else {
		fmt.Printf("Error: %s\n", result.Message)
		os.Exit(1)
	}
}

func printBookings(bookings []common.BookingInfo) {
	if len(bookings) == 0 {
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "ID\tResource\tBooker\tStart\tEnd")

	for _, b := range bookings {
		start := formatTime(b.Start)
		end := formatTime(b.End)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", b.ID[:8]+"...", b.Resource, b.Booker, start, end)
	}

	w.Flush()
}

func formatTime(timeStr string) string {
	t, err := time.Parse("2006-01-02T15:04:05Z07:00", timeStr)
	if err != nil {
		return timeStr
	}
	local := t.Local()
	return local.Format("2006-01-02 15:04:05 MST")
}

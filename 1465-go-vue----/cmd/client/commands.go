package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"envmonitor/internal/common"
)

func printResponse(resp *common.Response, err error) {
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("Response: %+v\n", resp)
		return
	}
	fmt.Println(string(data))
}

func parseMeasurements(flagValue string) map[string]float64 {
	result := make(map[string]float64)
	if flagValue == "" {
		return result
	}
	
	pairs := strings.Split(flagValue, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			continue
		}
		result[key] = value
	}
	return result
}

func cmdRegisterGas(c *Client) {
	fs := flag.NewFlagSet("register-gas", flag.ExitOnError)
	id := fs.String("id", "", "Outlet ID")
	location := fs.String("location", "", "Outlet location")
	fs.Parse(os.Args[2:])

	if *id == "" {
		fmt.Println("Error: -id is required")
		os.Exit(1)
	}

	resp, err := c.RegisterGasOutlet(*id, *location)
	printResponse(resp, err)
}

func cmdSubmitGas(c *Client) {
	fs := flag.NewFlagSet("submit-gas", flag.ExitOnError)
	outletID := fs.String("outlet-id", "", "Outlet ID")
	measurements := fs.String("measurements", "", "Measurements (format: SO2:50,NOx:100,PM:30,VOCs:20)")
	fs.Parse(os.Args[2:])

	if *outletID == "" {
		fmt.Println("Error: -outlet-id is required")
		os.Exit(1)
	}

	ms := parseMeasurements(*measurements)
	if len(ms) == 0 {
		fmt.Println("Error: -measurements is required and should be in format SO2:50,NOx:100")
		os.Exit(1)
	}

	resp, err := c.SubmitGasReport(*outletID, ms, nil)
	printResponse(resp, err)
}

func cmdQueryGas(c *Client) {
	fs := flag.NewFlagSet("query-gas", flag.ExitOnError)
	outletID := fs.String("outlet-id", "", "Outlet ID (optional)")
	fromDays := fs.Int("from-days", 7, "Query from N days ago")
	toDays := fs.Int("to-days", 0, "Query to N days ago (0 = now)")
	fs.Parse(os.Args[2:])

	from := time.Now().AddDate(0, 0, -*fromDays)
	to := time.Now().AddDate(0, 0, -*toDays)

	resp, err := c.QueryGasReports(*outletID, &from, &to)
	printResponse(resp, err)
}

func cmdRegisterWastewater(c *Client) {
	fs := flag.NewFlagSet("register-wastewater", flag.ExitOnError)
	id := fs.String("id", "", "Outlet ID")
	fs.Parse(os.Args[2:])

	if *id == "" {
		fmt.Println("Error: -id is required")
		os.Exit(1)
	}

	resp, err := c.RegisterWastewaterOutlet(*id)
	printResponse(resp, err)
}

func cmdSubmitWastewater(c *Client) {
	fs := flag.NewFlagSet("submit-wastewater", flag.ExitOnError)
	outletID := fs.String("outlet-id", "", "Outlet ID")
	measurements := fs.String("measurements", "", "Measurements (format: COD:50,NH3-N:5,TP:0.3,pH:7.5)")
	fs.Parse(os.Args[2:])

	if *outletID == "" {
		fmt.Println("Error: -outlet-id is required")
		os.Exit(1)
	}

	ms := parseMeasurements(*measurements)
	if len(ms) == 0 {
		fmt.Println("Error: -measurements is required and should be in format COD:50,NH3-N:5")
		os.Exit(1)
	}

	resp, err := c.SubmitWastewaterReport(*outletID, ms, nil)
	printResponse(resp, err)
}

func cmdQueryWastewater(c *Client) {
	fs := flag.NewFlagSet("query-wastewater", flag.ExitOnError)
	outletID := fs.String("outlet-id", "", "Outlet ID (optional)")
	fromDays := fs.Int("from-days", 7, "Query from N days ago")
	toDays := fs.Int("to-days", 0, "Query to N days ago (0 = now)")
	fs.Parse(os.Args[2:])

	from := time.Now().AddDate(0, 0, -*fromDays)
	to := time.Now().AddDate(0, 0, -*toDays)

	resp, err := c.QueryWastewaterReports(*outletID, &from, &to)
	printResponse(resp, err)
}

func cmdDailyReport(c *Client) {
	fs := flag.NewFlagSet("daily-report", flag.ExitOnError)
	outletID := fs.String("outlet-id", "", "Outlet ID")
	dateStr := fs.String("date", "", "Date (YYYY-MM-DD, default: today)")
	fs.Parse(os.Args[2:])

	if *outletID == "" {
		fmt.Println("Error: -outlet-id is required")
		os.Exit(1)
	}

	var date *time.Time
	if *dateStr != "" {
		d, err := time.Parse("2006-01-02", *dateStr)
		if err != nil {
			fmt.Printf("Error: invalid date format: %v\n", err)
			os.Exit(1)
		}
		date = &d
	}

	resp, err := c.GetDailyReport(*outletID, date)
	printResponse(resp, err)
}

func cmdAddSolidWaste(c *Client) {
	fs := flag.NewFlagSet("add-solid-waste", flag.ExitOnError)
	name := fs.String("name", "", "Waste name")
	category := fs.Int("category", 0, "Category: 0=Hazardous, 1=General")
	amount := fs.Float64("amount", 0, "Amount in tons")
	storage := fs.String("storage", "", "Storage location")
	fs.Parse(os.Args[2:])

	if *name == "" || *amount <= 0 || *storage == "" {
		fmt.Println("Error: -name, -amount, and -storage are required")
		os.Exit(1)
	}

	req := &common.AddSolidWasteRequest{
		Name:            *name,
		Category:        *category,
		Amount:          *amount,
		StorageLocation: *storage,
	}

	resp, err := c.AddSolidWasteRecord(req)
	printResponse(resp, err)
}

func cmdQuerySolidWaste(c *Client) {
	fs := flag.NewFlagSet("query-solid-waste", flag.ExitOnError)
	category := fs.Int("category", -1, "Category filter: 0=Hazardous, 1=General (-1=all)")
	status := fs.Int("status", -1, "Status filter: 0=Stored, 1=Disposed, 2=Expired (-1=all)")
	fs.Parse(os.Args[2:])

	var catPtr, statusPtr *int
	if *category >= 0 {
		catPtr = category
	}
	if *status >= 0 {
		statusPtr = status
	}

	resp, err := c.QuerySolidWasteRecords(catPtr, statusPtr)
	printResponse(resp, err)
}

func cmdQueryAlarms(c *Client) {
	fs := flag.NewFlagSet("query-alarms", flag.ExitOnError)
	entityType := fs.Int("entity-type", -1, "Entity type filter: 0=Gas, 1=Wastewater, 2=SolidWaste, 3=DailyReport (-1=all)")
	entityID := fs.String("entity-id", "", "Entity ID filter (optional)")
	level := fs.Int("level", -1, "Level filter: 0=Warning, 1=Emergency (-1=all)")
	resolved := fs.Bool("resolved", false, "Show resolved alarms")
	all := fs.Bool("all", false, "Show all alarms (active and resolved)")
	fs.Parse(os.Args[2:])

	var entityTypePtr, levelPtr *int
	var resolvedPtr *bool

	if *entityType >= 0 {
		entityTypePtr = entityType
	}
	if *level >= 0 {
		levelPtr = level
	}
	if !*all {
		resolvedPtr = resolved
	}

	resp, err := c.QueryAlarms(entityTypePtr, levelPtr, *entityID, resolvedPtr)
	printResponse(resp, err)
}

func cmdResolveAlarm(c *Client) {
	fs := flag.NewFlagSet("resolve-alarm", flag.ExitOnError)
	id := fs.String("id", "", "Alarm ID")
	fs.Parse(os.Args[2:])

	if *id == "" {
		fmt.Println("Error: -id is required")
		os.Exit(1)
	}

	resp, err := c.ResolveAlarm(*id)
	printResponse(resp, err)
}

package main

import (
	"business-hours/protocol"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const serverURL = "http://localhost:8080"

var timePattern = regexp.MustCompile(`^(\d{1,2}):(\d{1,2})$`)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "check":
		handleCheck()
	case "calculate":
		handleCalculate()
	case "config":
		handleConfig()
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCheck() {
	flagSet := flag.NewFlagSet("check", flag.ExitOnError)
	timeStr := flagSet.String("time", "", "Time to check (format: 2006-01-02 15:04:05)")
	flagSet.Parse(os.Args[2:])

	var t time.Time
	if *timeStr == "" {
		t = time.Now()
	} else {
		var err error
		t, err = time.ParseInLocation("2006-01-02 15:04:05", *timeStr, time.Local)
		if err != nil {
			fmt.Printf("Invalid time format: %v\n", err)
			os.Exit(1)
		}
	}

	client := NewClient(serverURL)
	result, err := client.Check(t)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Time: %s\n", t.Format("2006-01-02 15:04:05"))
	fmt.Printf("Is Open: %v\n", result.IsOpen)
	if result.IsOpen {
		fmt.Println("Status: Currently open")
	} else {
		if !result.NextOpenTime.IsZero() {
			fmt.Printf("Next Open Time: %s\n", result.NextOpenTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("Time to Open: %v\n", result.TimeToOpen)
		} else {
			fmt.Println("No upcoming business hours found")
		}
	}
}

func handleCalculate() {
	flagSet := flag.NewFlagSet("calculate", flag.ExitOnError)
	startStr := flagSet.String("start", "", "Start time (format: 2006-01-02 15:04:05)")
	endStr := flagSet.String("end", "", "End time (format: 2006-01-02 15:04:05)")
	flagSet.Parse(os.Args[2:])

	if *startStr == "" || *endStr == "" {
		fmt.Println("Error: -start and -end are required")
		os.Exit(1)
	}

	start, err := time.ParseInLocation("2006-01-02 15:04:05", *startStr, time.Local)
	if err != nil {
		fmt.Printf("Invalid start time: %v\n", err)
		os.Exit(1)
	}

	end, err := time.ParseInLocation("2006-01-02 15:04:05", *endStr, time.Local)
	if err != nil {
		fmt.Printf("Invalid end time: %v\n", err)
		os.Exit(1)
	}

	client := NewClient(serverURL)
	result, err := client.CalculateHours(start, end)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Start: %s\n", start.Format("2006-01-02 15:04:05"))
	fmt.Printf("End: %s\n", end.Format("2006-01-02 15:04:05"))
	fmt.Printf("Total Business Hours: %.2f\n", result.TotalHours)
	if len(result.Slots) > 0 {
		fmt.Println("\nBusiness Slots:")
		for i, slot := range result.Slots {
			fmt.Printf("  %d. %s - %s\n", i+1, slot.Start.Format("15:04"), slot.End.Format("15:04"))
		}
	}
}

type dayConfig struct {
	openSlots  []protocol.SlotConfig
	breakSlots []protocol.SlotConfig
}

func handleConfig() {
	flagSet := flag.NewFlagSet("config", flag.ExitOnError)

	weekdays := []string{
		"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday",
		"sun", "mon", "tue", "wed", "thu", "fri", "sat",
	}

	var holidaysStr string
	flagSet.StringVar(&holidaysStr, "holidays", "", "Comma-separated holidays (format: 2024-01-01,2024-12-25)")

	dayConfigs := make(map[string]*dayConfig)
	for _, wd := range weekdays {
		dayConfigs[wd] = &dayConfig{}
	}

	openFlags := make(map[string]*string)
	breakFlags := make(map[string]*string)

	for _, wd := range weekdays {
		openFlags[wd] = flagSet.String(wd+"-open", "", fmt.Sprintf("%s open slots (e.g., 9:00-18:00 or 7:00-9:00,11:00-14:00)", wd))
		breakFlags[wd] = flagSet.String(wd+"-break", "", fmt.Sprintf("%s break slots (e.g., 12:00-13:00)", wd))
	}

	flagSet.Parse(os.Args[2:])

	config := &protocol.ConfigRequest{
		Weekly:   make(protocol.WeeklyConfig),
		Holidays: []time.Time{},
	}

	weekdayMap := map[string]string{
		"sun":  "sunday", "mon": "monday", "tue": "tuesday", "wed": "wednesday",
		"thu": "thursday", "fri": "friday", "sat": "saturday",
		"sunday": "sunday", "monday": "monday", "tuesday": "tuesday", "wednesday": "wednesday",
		"thursday": "thursday", "friday": "friday", "saturday": "saturday",
	}

	hasConfig := false

	for _, wd := range weekdays {
		normalized := weekdayMap[wd]
		if *openFlags[wd] != "" {
			slots, err := parseTimeSlots(*openFlags[wd])
			if err != nil {
				fmt.Printf("Error parsing %s-open: %v\n", wd, err)
				os.Exit(1)
			}
			daySchedule := config.Weekly[normalized]
			daySchedule.OpenSlots = slots
			config.Weekly[normalized] = daySchedule
			hasConfig = true
		}
		if *breakFlags[wd] != "" {
			slots, err := parseTimeSlots(*breakFlags[wd])
			if err != nil {
				fmt.Printf("Error parsing %s-break: %v\n", wd, err)
				os.Exit(1)
			}
			daySchedule := config.Weekly[normalized]
			daySchedule.BreakSlots = slots
			config.Weekly[normalized] = daySchedule
			hasConfig = true
		}
	}

	if holidaysStr != "" {
		holidayParts := strings.Split(holidaysStr, ",")
		for _, hp := range holidayParts {
			hp = strings.TrimSpace(hp)
			if hp == "" {
				continue
			}
			holiday, err := time.ParseInLocation("2006-01-02", hp, time.Local)
			if err != nil {
				fmt.Printf("Error parsing holiday '%s': %v\n", hp, err)
				os.Exit(1)
			}
			config.Holidays = append(config.Holidays, holiday)
		}
		hasConfig = true
	}

	if !hasConfig {
		printConfigUsage()
		return
	}

	client := NewClient(serverURL)
	result, err := client.Configure(config)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if result.Success {
		fmt.Println("Configuration applied successfully!")
	} else {
		fmt.Printf("Configuration failed: %s\n", result.Error)
		os.Exit(1)
	}
}

func parseTimeSlots(s string) ([]protocol.SlotConfig, error) {
	parts := strings.Split(s, ",")
	slots := []protocol.SlotConfig{}

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		slot, err := parseTimeSlot(p)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}

	return slots, nil
}

func parseTimeSlot(s string) (protocol.SlotConfig, error) {
	parts := strings.Split(s, "-")
	if len(parts) != 2 {
		return protocol.SlotConfig{}, fmt.Errorf("invalid time slot format: %s (expected HH:MM-HH:MM)", s)
	}

	startHour, startMin, err := parseTime(parts[0])
	if err != nil {
		return protocol.SlotConfig{}, fmt.Errorf("invalid start time '%s': %v", parts[0], err)
	}

	endHour, endMin, err := parseTime(parts[1])
	if err != nil {
		return protocol.SlotConfig{}, fmt.Errorf("invalid end time '%s': %v", parts[1], err)
	}

	return protocol.SlotConfig{
		StartHour:   startHour,
		StartMinute: startMin,
		EndHour:     endHour,
		EndMinute:   endMin,
	}, nil
}

func parseTime(s string) (hour, minute int, err error) {
	s = strings.TrimSpace(s)
	matches := timePattern.FindStringSubmatch(s)
	if matches == nil {
		return 0, 0, fmt.Errorf("invalid time format: %s (expected HH:MM)", s)
	}

	hour, _ = strconv.Atoi(matches[1])
	minute, _ = strconv.Atoi(matches[2])

	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return 0, 0, fmt.Errorf("invalid time: %s (hour 0-23, minute 0-59)", s)
	}

	return hour, minute, nil
}

func printUsage() {
	fmt.Println("Business Hours Client")
	fmt.Println("\nUsage:")
	fmt.Println("  client check [--time \"2006-01-02 15:04:05\"]")
	fmt.Println("  client calculate --start \"2006-01-02 15:04:05\" --end \"2006-01-02 15:04:05\"")
	fmt.Println("  client config [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  check      - Check if a time is within business hours")
	fmt.Println("  calculate  - Calculate business hours between two times")
	fmt.Println("  config     - Configure business hours")
}

func printConfigUsage() {
	fmt.Println("Config command - configures business hours")
	fmt.Println("\nUsage: client config [options]")
	fmt.Println("\nOptions:")
	fmt.Println("  --[weekday]-open      Open slots for a weekday (e.g., 9:00-18:00 or 7:00-9:00,11:00-14:00)")
	fmt.Println("  --[weekday]-break     Break slots for a weekday (e.g., 12:00-13:00)")
	fmt.Println("  --holidays            Comma-separated holidays (e.g., 2024-01-01,2024-12-25)")
	fmt.Println("\nWeekday names:")
	fmt.Println("  Full:   sunday, monday, tuesday, wednesday, thursday, friday, saturday")
	fmt.Println("  Short:  sun, mon, tue, wed, thu, fri, sat")
	fmt.Println("\nExamples:")
	fmt.Println("  # Set Monday-Friday 9:00-18:00 with 12:00-13:00 break")
	fmt.Println("  client config --mon-open 9:00-18:00 --mon-break 12:00-13:00 --tue-open 9:00-18:00 --tue-break 12:00-13:00 ...")
	fmt.Println("")
	fmt.Println("  # Set weekend 10:00-22:00")
	fmt.Println("  client config --sat-open 10:00-22:00 --sun-open 10:00-22:00")
	fmt.Println("")
	fmt.Println("  # Set bar hours (22:00 to 02:00 next day)")
	fmt.Println("  client config --fri-open 22:00-02:00 --sat-open 22:00-02:00")
	fmt.Println("")
	fmt.Println("  # Set holidays")
	fmt.Println("  client config --holidays 2024-10-01,2024-12-25")
}

package main

import (
	"flag"
	"fmt"
	"os"

	"agecalculator/agecalc"
	"agecalculator/common"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Age calculator server URL")

	ageCmd := flag.NewFlagSet("age", flag.ExitOnError)
	ageBirth := ageCmd.String("birth", "", "Birth date (YYYY-MM-DD)")
	ageCurrent := ageCmd.String("current", "", "Current date (YYYY-MM-DD, optional)")
	ageGender := ageCmd.String("gender", "", "Gender (male/female, optional)")
	ageTarget := ageCmd.Int("target", 0, "Target age to check (optional)")

	daysCmd := flag.NewFlagSet("days", flag.ExitOnError)
	daysStart := daysCmd.String("start", "", "Start date (YYYY-MM-DD)")
	daysEnd := daysCmd.String("end", "", "End date (YYYY-MM-DD)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [global-options] <command> [command-options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Global Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  age   - Calculate age and related information\n")
		fmt.Fprintf(os.Stderr, "  days  - Calculate days between two dates\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s age -birth 1990-05-15\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s age -birth 1990-05-15 -gender male -target 30\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s days -start 2020-01-01 -end 2020-12-31\n", os.Args[0])
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	args := os.Args[1:]
	cmdIndex := 0
	for i, arg := range args {
		if arg != "" && arg[0] != '-' {
			cmdIndex = i
			break
		}
	}

	cmd := args[cmdIndex]
	globalArgs := args[:cmdIndex]
	cmdArgs := args[cmdIndex+1:]

	flag.CommandLine.Parse(globalArgs)

	c := NewClient(*serverURL)

	switch cmd {
	case "age":
		ageCmd.Parse(cmdArgs)
		if *ageBirth == "" {
			fmt.Fprintln(os.Stderr, "Error: -birth flag is required")
			ageCmd.Usage()
			os.Exit(1)
		}

		req := common.AgeRequest{
			BirthDate:   *ageBirth,
			CurrentDate: *ageCurrent,
			TargetAge:   *ageTarget,
		}

		if *ageGender == "male" {
			req.Gender = agecalc.Male
		} else if *ageGender == "female" {
			req.Gender = agecalc.Female
		}

		resp, err := c.CalculateAge(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if resp.Error != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}

		fmt.Printf("精确年龄: %d岁%d月%d天\n", resp.Age.Years, resp.Age.Months, resp.Age.Days)
		fmt.Printf("周岁: %d岁\n", resp.AgeYears)
		fmt.Printf("是否成年(满18岁): %s\n", boolToYesNo(resp.IsAdult))
		fmt.Printf("距离下一个生日: %d天\n", resp.DaysUntilBirthday)

		if *ageGender != "" {
			fmt.Printf("是否已到退休年龄: %s\n", boolToYesNo(resp.IsRetired))
		}
		if *ageTarget > 0 {
			fmt.Printf("是否已满%d岁: %s\n", *ageTarget, boolToYesNo(resp.HasReachedTargetAge))
		}

	case "days":
		daysCmd.Parse(cmdArgs)
		if *daysStart == "" || *daysEnd == "" {
			fmt.Fprintln(os.Stderr, "Error: -start and -end flags are required")
			daysCmd.Usage()
			os.Exit(1)
		}

		req := common.DaysBetweenRequest{
			StartDate: *daysStart,
			EndDate:   *daysEnd,
		}

		resp, err := c.DaysBetween(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if resp.Error != "" {
			fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
			os.Exit(1)
		}

		fmt.Printf("两个日期之间的天数: %d天\n", resp.Days)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		flag.Usage()
		os.Exit(1)
	}
}

func boolToYesNo(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

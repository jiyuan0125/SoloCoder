package main

import (
	"fmt"
	"time"

	"smart-park/common"
)

func handleVisitorCommands(c *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Visitor commands: reserve, review, checkin, list")
		return
	}

	cmd := args[0]
	switch cmd {
	case "reserve":
		if len(args) < 6 {
			fmt.Println("Usage: reserve <employee_id> <name> <phone> <purpose> <arrival_time>")
			fmt.Println("  arrival_time format: 2006-01-02T15:04:05")
			return
		}
		arrivalTime, err := time.Parse(time.RFC3339, args[5])
		if err != nil {
			arrivalTime, err = time.Parse("2006-01-02 15:04:05", args[5])
			if err != nil {
				fmt.Printf("Invalid time format: %v\n", err)
				return
			}
		}
		req := common.CreateVisitorRequest{
			EmployeeID:      args[1],
			VisitorName:     args[2],
			VisitorPhone:    args[3],
			Purpose:         args[4],
			ExpectedArrival: arrivalTime,
		}
		var resp common.Response
		c.post("/visitor/reserve", req, &resp)
		printResponse(resp)

	case "review":
		if len(args) < 3 {
			fmt.Println("Usage: review <reservation_id> <approve/reject>")
			return
		}
		approve := args[2] == "approve"
		req := common.ReviewVisitorRequest{
			ReservationID: args[1],
			Approve:       approve,
		}
		var resp common.Response
		c.post("/visitor/review", req, &resp)
		printResponse(resp)

	case "checkin":
		if len(args) < 2 {
			fmt.Println("Usage: checkin <phone>")
			return
		}
		req := common.CheckInVisitorRequest{
			VisitorPhone: args[1],
		}
		var resp common.Response
		c.post("/visitor/checkin", req, &resp)
		printResponse(resp)

	case "list":
		var resp common.Response
		c.get("/visitor/list", &resp)
		printResponse(resp)

	default:
		fmt.Printf("Unknown visitor command: %s\n", cmd)
	}
}

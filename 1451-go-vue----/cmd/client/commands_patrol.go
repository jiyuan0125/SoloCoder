package main

import (
	"fmt"
	"strings"
	"time"

	"smart-park/common"
)

func handlePatrolCommands(c *Client, args []string) {
	if len(args) == 0 {
		fmt.Println("Patrol commands: add-point, add-route, list-routes, add-task, list-tasks, start, checkin, list-checkins")
		return
	}

	cmd := args[0]
	switch cmd {
	case "add-point":
		if len(args) < 3 {
			fmt.Println("Usage: add-point <id> <name>")
			return
		}
		req := common.PatrolPoint{
			ID:   args[1],
			Name: args[2],
		}
		var resp common.Response
		c.post("/patrol/point", req, &resp)
		printResponse(resp)

	case "add-route":
		if len(args) < 3 {
			fmt.Println("Usage: add-route <name> <point1,point2,...>")
			return
		}
		pointIDs := strings.Split(args[2], ",")
		req := common.CreatePatrolRouteRequest{
			Name:     args[1],
			PointIDs: pointIDs,
		}
		var resp common.Response
		c.post("/patrol/route", req, &resp)
		printResponse(resp)

	case "list-routes":
		var resp common.Response
		c.get("/patrol/routes", &resp)
		printResponse(resp)

	case "add-task":
		if len(args) < 5 {
			fmt.Println("Usage: add-task <route_id> <assignee> <frequency> <start_time>")
			fmt.Println("  start_time format: 2006-01-02T15:04:05")
			return
		}
		startTime, err := time.Parse(time.RFC3339, args[4])
		if err != nil {
			startTime, err = time.Parse("2006-01-02 15:04:05", args[4])
			if err != nil {
				fmt.Printf("Invalid time format: %v\n", err)
				return
			}
		}
		req := common.CreatePatrolTaskRequest{
			RouteID:    args[1],
			AssigneeID: args[2],
			Frequency:  args[3],
			StartTime:  startTime,
		}
		var resp common.Response
		c.post("/patrol/task", req, &resp)
		printResponse(resp)

	case "list-tasks":
		var resp common.Response
		c.get("/patrol/tasks", &resp)
		printResponse(resp)

	case "start":
		if len(args) < 2 {
			fmt.Println("Usage: start <task_id>")
			return
		}
		req := common.StartPatrolRequest{
			TaskID: args[1],
		}
		var resp common.Response
		c.post("/patrol/start", req, &resp)
		printResponse(resp)

	case "checkin":
		if len(args) < 3 {
			fmt.Println("Usage: checkin <execution_id> <point_id> [has_anomaly] [desc] [related_ap]")
			return
		}
		hasAnomaly := false
		anomalyDesc := ""
		relatedAP := ""
		if len(args) >= 4 {
			hasAnomaly = args[3] == "true" || args[3] == "1"
		}
		if len(args) >= 5 {
			anomalyDesc = args[4]
		}
		if len(args) >= 6 {
			relatedAP = args[5]
		}
		req := common.PatrolCheckInRequest{
			ExecutionID:          args[1],
			PointID:              args[2],
			HasAnomaly:           hasAnomaly,
			AnomalyDesc:          anomalyDesc,
			RelatedAccessPointID: relatedAP,
		}
		var resp common.Response
		c.post("/patrol/checkin", req, &resp)
		printResponse(resp)

	case "list-checkins":
		if len(args) < 2 {
			fmt.Println("Usage: list-checkins <execution_id>")
			return
		}
		var resp common.Response
		c.get("/patrol/checkins?execution_id="+args[1], &resp)
		printResponse(resp)

	default:
		fmt.Printf("Unknown patrol command: %s\n", cmd)
	}
}

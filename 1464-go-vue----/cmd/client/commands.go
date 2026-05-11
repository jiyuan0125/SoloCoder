package main

import (
	"encoding/json"
	"firemanagement/pkg/api"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func printTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, strings.Join(headers, "\t"))
	sep := make([]string, len(headers))
	for i := range headers {
		sep[i] = strings.Repeat("-", len(headers[i]))
	}
	fmt.Fprintln(w, strings.Join(sep, "\t"))
	for _, row := range rows {
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
	w.Flush()
}

func cmdDeviceCreate(client *Client, args []string) {
	if len(args) < 7 {
		fmt.Println("Usage: device create <code> <type> <location> <install_date> <expiry_date> <last_check_date>")
		fmt.Println("  type: fire_extinguisher, fire_hydrant, smoke_detector, sprinkler_head, emergency_light")
		fmt.Println("  dates format: 2006-01-02")
		os.Exit(1)
	}

	installDate, _ := time.Parse("2006-01-02", args[3])
	expiryDate, _ := time.Parse("2006-01-02", args[4])
	lastCheckDate, _ := time.Parse("2006-01-02", args[5])

	req := api.CreateDeviceRequest{
		Code:          args[0],
		Type:          api.DeviceType(args[1]),
		Location:      args[2],
		InstallDate:   installDate,
		ExpiryDate:    expiryDate,
		LastCheckDate: lastCheckDate,
	}

	resp, err := client.CreateDevice(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Device created with ID: %s\n", resp.ID)
}

func cmdDeviceList(client *Client, args []string) {
	var req api.ListDevicesRequest
	if len(args) > 0 {
		t := api.DeviceType(args[0])
		req.Type = &t
	}

	resp, err := client.ListDevices(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(resp.Devices) == 0 {
		fmt.Println("No devices found")
		return
	}

	headers := []string{"ID", "Code", "Type", "Location", "Status", "Expiry"}
	var rows [][]string
	for _, d := range resp.Devices {
		rows = append(rows, []string{
			d.ID[:min(len(d.ID), 12)],
			d.Code,
			string(d.Type),
			d.Location,
			string(d.Status),
			d.ExpiryDate.Format("2006-01-02"),
		})
	}
	printTable(headers, rows)
}

func cmdDeviceGet(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: device get <id>")
		os.Exit(1)
	}
	resp, err := client.GetDevice(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp.Device)
}

func cmdDeviceUpdateStatus(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: device update-status <id> <status>")
		fmt.Println("  status: normal, pending_repair, scrapped")
		os.Exit(1)
	}
	status := api.DeviceStatus(args[1])
	req := api.UpdateDeviceRequest{Status: &status}
	err := client.UpdateDevice(args[0], req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Device updated")
}

func cmdPointCreate(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: point create <name> <location>")
		os.Exit(1)
	}
	req := api.CreateInspectionPointRequest{
		Name:     args[0],
		Location: args[1],
	}
	resp, err := client.CreateInspectionPoint(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Point created with ID: %s\n", resp.ID)
}

func cmdPointList(client *Client, args []string) {
	resp, err := client.ListInspectionPoints()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Points) == 0 {
		fmt.Println("No inspection points found")
		return
	}
	headers := []string{"ID", "Name", "Location"}
	var rows [][]string
	for _, p := range resp.Points {
		rows = append(rows, []string{p.ID[:min(len(p.ID), 12)], p.Name, p.Location})
	}
	printTable(headers, rows)
}

func cmdRouteCreate(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: route create <name> <point_id1> [point_id2 ...]")
		os.Exit(1)
	}
	req := api.CreateInspectionRouteRequest{
		Name:     args[0],
		PointIDs: args[1:],
	}
	resp, err := client.CreateInspectionRoute(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Route created with ID: %s\n", resp.ID)
}

func cmdRouteList(client *Client, args []string) {
	resp, err := client.ListInspectionRoutes()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Routes) == 0 {
		fmt.Println("No inspection routes found")
		return
	}
	headers := []string{"ID", "Name", "Points"}
	var rows [][]string
	for _, r := range resp.Routes {
		pointNames := make([]string, len(r.Points))
		for i, p := range r.Points {
			pointNames[i] = p.Name
		}
		rows = append(rows, []string{
			r.ID[:min(len(r.ID), 12)],
			r.Name,
			strconv.Itoa(len(r.Points)) + " points",
		})
	}
	printTable(headers, rows)
}

func cmdPlanCreate(client *Client, args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: plan create <route_id> <inspector_id> <inspector_name> <frequency> <start_time>")
		fmt.Println("  frequency: daily, weekly, monthly")
		fmt.Println("  start_time format: 2006-01-02")
		os.Exit(1)
	}
	startTime, _ := time.Parse("2006-01-02", args[4])
	req := api.CreateInspectionPlanRequest{
		RouteID:       args[0],
		InspectorID:   args[1],
		InspectorName: args[2],
		Frequency:     api.InspectionFrequency(args[3]),
		StartTime:     startTime,
	}
	resp, err := client.CreateInspectionPlan(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Plan created with ID: %s\n", resp.ID)
}

func cmdPlanList(client *Client, args []string) {
	resp, err := client.ListInspectionPlans()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Plans) == 0 {
		fmt.Println("No inspection plans found")
		return
	}
	headers := []string{"ID", "Inspector", "Frequency", "Route"}
	var rows [][]string
	for _, p := range resp.Plans {
		routeName := ""
		if p.Route != nil {
			routeName = p.Route.Name
		}
		rows = append(rows, []string{
			p.ID[:min(len(p.ID), 12)],
			p.InspectorName,
			string(p.Frequency),
			routeName,
		})
	}
	printTable(headers, rows)
}

func cmdTaskList(client *Client, args []string) {
	resp, err := client.ListInspectionTasks()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Tasks) == 0 {
		fmt.Println("No inspection tasks found")
		return
	}
	headers := []string{"ID", "Inspector", "Status", "Created"}
	var rows [][]string
	for _, t := range resp.Tasks {
		rows = append(rows, []string{
			t.ID[:min(len(t.ID), 12)],
			t.InspectorName,
			string(t.Status),
			t.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	printTable(headers, rows)
}

func cmdTaskGet(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: task get <id>")
		os.Exit(1)
	}
	resp, err := client.GetInspectionTask(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	printJSON(resp.Task)
}

func cmdTaskCheck(client *Client, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: task check <task_id> <point_id> <normal:true|false>")
		os.Exit(1)
	}
	normal, err := strconv.ParseBool(args[2])
	if err != nil {
		fmt.Printf("Invalid normal value: %v\n", err)
		os.Exit(1)
	}
	req := api.CheckTaskPointRequest{
		TaskID:  args[0],
		PointID: args[1],
		Normal:  normal,
	}
	err = client.CheckTaskPoint(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Point checked")
}

func cmdTaskSummary(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: task summary <task_id>")
		os.Exit(1)
	}
	summary, err := client.GetTaskSummary(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	printJSON(summary)
}

func cmdDrillCreatePlan(client *Client, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: drill create-plan <type> <name> <scheduled_time> <planned_attendees>")
		fmt.Println("  type: fire_extinguisher_use, evacuation_escape, comprehensive")
		fmt.Println("  scheduled_time format: 2006-01-02")
		os.Exit(1)
	}
	scheduledTime, _ := time.Parse("2006-01-02", args[2])
	plannedAttendees, _ := strconv.Atoi(args[3])
	req := api.CreateDrillPlanRequest{
		Type:             api.DrillType(args[0]),
		Name:             args[1],
		ScheduledTime:    scheduledTime,
		PlannedAttendees: plannedAttendees,
	}
	resp, err := client.CreateDrillPlan(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Drill plan created with ID: %s\n", resp.ID)
}

func cmdDrillListPlans(client *Client, args []string) {
	resp, err := client.ListDrillPlans()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Plans) == 0 {
		fmt.Println("No drill plans found")
		return
	}
	headers := []string{"ID", "Type", "Name", "Scheduled", "Planned Attendees"}
	var rows [][]string
	for _, p := range resp.Plans {
		rows = append(rows, []string{
			p.ID[:min(len(p.ID), 12)],
			string(p.Type),
			p.Name,
			p.ScheduledTime.Format("2006-01-02"),
			strconv.Itoa(p.PlannedAttendees),
		})
	}
	printTable(headers, rows)
}

func cmdDrillComplete(client *Client, args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: drill complete <plan_id> <actual_attendees> <duration_minutes> <improvements>")
		os.Exit(1)
	}
	actualAttendees, _ := strconv.Atoi(args[1])
	durationMinutes, _ := strconv.Atoi(args[2])
	req := api.CompleteDrillRequest{
		ActualAttendees: actualAttendees,
		DurationMinutes: durationMinutes,
		Improvements:    strings.Join(args[3:], " "),
	}
	resp, err := client.CompleteDrill(args[0], req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Drill completed, record ID: %s\n", resp.ID)
}

func cmdDrillListRecords(client *Client, args []string) {
	resp, err := client.ListDrillRecords()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Records) == 0 {
		fmt.Println("No drill records found")
		return
	}
	headers := []string{"ID", "Plan Name", "Type", "Actual Attendees", "Duration", "Completed"}
	var rows [][]string
	for _, r := range resp.Records {
		name := ""
		dType := ""
		if r.Plan != nil {
			name = r.Plan.Name
			dType = string(r.Plan.Type)
		}
		rows = append(rows, []string{
			r.ID[:min(len(r.ID), 12)],
			name,
			dType,
			strconv.Itoa(r.ActualAttendees),
			strconv.Itoa(r.DurationMinutes) + " min",
			r.CompletedAt.Format("2006-01-02"),
		})
	}
	printTable(headers, rows)
}

func cmdReminderList(client *Client, args []string) {
	var req api.ListRemindersRequest
	if len(args) > 0 {
		unread := (args[0] == "unread")
		req.Read = &unread
	}
	resp, err := client.ListReminders(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Reminders) == 0 {
		fmt.Println("No reminders found")
		return
	}
	headers := []string{"ID", "Type", "Title", "Read", "Created"}
	var rows [][]string
	for _, r := range resp.Reminders {
		rows = append(rows, []string{
			r.ID[:min(len(r.ID), 12)],
			string(r.Type),
			r.Title,
			strconv.FormatBool(r.Read),
			r.CreatedAt.Format("2006-01-02 15:04"),
		})
	}
	printTable(headers, rows)
}

func cmdReminderRead(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: reminder read <id> [true|false]")
		os.Exit(1)
	}
	read := true
	if len(args) > 1 {
		read, _ = strconv.ParseBool(args[1])
	}
	err := client.MarkReminderRead(args[0], read)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Reminder updated")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

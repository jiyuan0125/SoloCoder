package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"coldchain/common"
)

type CommandRunner struct {
	client *APIClient
}

func NewCommandRunner(client *APIClient) *CommandRunner {
	return &CommandRunner{client: client}
}

func (r *CommandRunner) RunCreateTask(args []string) error {
	if len(args) < 4 {
		return fmt.Errorf("usage: create-task <device_id> <cargo_type> <start_warehouse> <end_warehouse>")
	}

	task, err := r.client.CreateTask(args[0], args[1], args[2], args[3])
	if err != nil {
		return err
	}

	printJSON(task)
	return nil
}

func (r *CommandRunner) RunListTasks(args []string) error {
	resp, err := r.client.ListTasks()
	if err != nil {
		return err
	}

	if len(resp.Tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TASK ID\tDEVICE ID\tCARGO TYPE\tSTATUS\tSTART WAREHOUSE\tEND WAREHOUSE")
	for _, t := range resp.Tasks {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			t.TaskID, t.DeviceID, t.CargoType, t.Status, t.StartWarehouse, t.EndWarehouse)
	}
	w.Flush()

	return nil
}

func (r *CommandRunner) RunGetTask(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-task <task_id>")
	}

	task, err := r.client.GetTask(args[0])
	if err != nil {
		return err
	}

	printJSON(task)
	return nil
}

func (r *CommandRunner) RunEndTask(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: end-task <task_id>")
	}

	if err := r.client.EndTask(args[0]); err != nil {
		return err
	}

	fmt.Printf("Task %s ended successfully\n", args[0])
	return nil
}

func (r *CommandRunner) RunSubmitSensorData(args []string) error {
	if len(args) < 7 {
		return fmt.Errorf("usage: submit-data <device_id> <temperature> <humidity> <latitude> <longitude> <speed> [timestamp]")
	}

	temp, _ := strconv.ParseFloat(args[1], 64)
	humidity, _ := strconv.ParseFloat(args[2], 64)
	lat, _ := strconv.ParseFloat(args[3], 64)
	lon, _ := strconv.ParseFloat(args[4], 64)
	speed, _ := strconv.ParseFloat(args[5], 64)

	timestamp := time.Now()
	if len(args) >= 7 && args[6] != "" {
		if t, err := time.Parse(time.RFC3339, args[6]); err == nil {
			timestamp = t
		}
	}

	req := &common.SubmitSensorDataRequest{
		DeviceID:    args[0],
		Timestamp:   timestamp,
		Temperature: temp,
		Humidity:    humidity,
		Latitude:    lat,
		Longitude:   lon,
		Speed:       speed,
	}

	resp, err := r.client.SubmitSensorData(req)
	if err != nil {
		return err
	}

	if resp.Alert != nil {
		fmt.Printf("Data submitted. ALERT: [%s] Temp: %.1f°C at %s\n",
			resp.Alert.Level, resp.Alert.Temperature, resp.Alert.Timestamp.Format(time.RFC3339))
	} else {
		fmt.Println("Data submitted successfully")
	}

	return nil
}

func (r *CommandRunner) RunGetSensorData(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-data <task_id>")
	}

	resp, err := r.client.GetSensorData(args[0])
	if err != nil {
		return err
	}

	if len(resp.Data) == 0 {
		fmt.Println("No sensor data found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tTEMP (°C)\tHUMIDITY (%)\tLAT\tLON\tSPEED")
	for _, d := range resp.Data {
		fmt.Fprintf(w, "%s\t%.1f\t%.1f\t%.6f\t%.6f\t%.1f\n",
			d.Timestamp.Format(time.RFC3339), d.Temperature, d.Humidity,
			d.Latitude, d.Longitude, d.Speed)
	}
	w.Flush()

	return nil
}

func (r *CommandRunner) RunGetAlerts(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-alerts <task_id>")
	}

	resp, err := r.client.GetAlerts(args[0])
	if err != nil {
		return err
	}

	if len(resp.Alerts) == 0 {
		fmt.Println("No alerts found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tLEVEL\tTEMP (°C)\tALERT ID")
	for _, a := range resp.Alerts {
		fmt.Fprintf(w, "%s\t%s\t%.1f\t%s\n",
			a.Timestamp.Format(time.RFC3339), a.Level, a.Temperature, a.ID)
	}
	w.Flush()

	return nil
}

func (r *CommandRunner) RunGetStats(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-stats <task_id>")
	}

	stats, err := r.client.GetStats(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("Task ID: %s\n", stats.TaskID)
	fmt.Printf("Total Samples: %d\n", stats.TotalSamples)
	fmt.Printf("Pass Rate: %.2f%%\n", stats.PassRate*100)
	fmt.Printf("Need Recheck: %v\n", stats.NeedRecheck)
	fmt.Printf("Temperature Stats:\n")
	fmt.Printf("  Max: %.1f°C\n", stats.TempMax)
	fmt.Printf("  Min: %.1f°C\n", stats.TempMin)
	fmt.Printf("  Avg: %.1f°C\n", stats.TempAvg)
	fmt.Printf("  StdDev: %.2f°C\n", stats.TempStdDev)

	return nil
}

func (r *CommandRunner) RunGetReport(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: get-report <task_id>")
	}

	report, err := r.client.GetReport(args[0])
	if err != nil {
		return err
	}

	fmt.Printf("=== Transport Report ===\n")
	fmt.Printf("Task ID: %s\n", report.TaskID)
	fmt.Printf("Exported At: %s\n\n", report.ExportedAt.Format(time.RFC3339))

	fmt.Printf("--- Statistics ---\n")
	fmt.Printf("Total Samples: %d\n", report.Stats.TotalSamples)
	fmt.Printf("Pass Rate: %.2f%%\n", report.Stats.PassRate*100)
	fmt.Printf("Need Recheck: %v\n", report.Stats.NeedRecheck)
	fmt.Printf("Temp: Max=%.1f°C, Min=%.1f°C, Avg=%.1f°C, StdDev=%.2f°C\n\n",
		report.Stats.TempMax, report.Stats.TempMin, report.Stats.TempAvg, report.Stats.TempStdDev)

	fmt.Printf("--- Alerts (%d total) ---\n", len(report.Alerts))
	if len(report.Alerts) > 0 {
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "TIMESTAMP\tLEVEL\tTEMP (°C)")
		for _, a := range report.Alerts {
			fmt.Fprintf(w, "%s\t%s\t%.1f\n",
				a.Timestamp.Format(time.RFC3339), a.Level, a.Temperature)
		}
		w.Flush()
	} else {
		fmt.Println("No alerts")
	}

	return nil
}

func (r *CommandRunner) RunExportCSV(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: export <task_id> <type:sensor|stats|alerts> <output_file>")
	}

	err := r.client.ExportCSV(args[0], args[1], args[2])
	if err != nil {
		return err
	}

	fmt.Printf("Exported to %s\n", args[2])
	return nil
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

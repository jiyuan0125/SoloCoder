package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"obd-platform/pkg/api"
)

const defaultServerURL = "http://localhost:9004"

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	if !strings.HasPrefix(serverURL, "http://") && !strings.HasPrefix(serverURL, "https://") {
		serverURL = "http://" + serverURL
	}
	return &Client{serverURL: serverURL}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.serverURL+path, reader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("server returned error: %s", string(respBody))
	}

	return respBody, nil
}

func (c *Client) UploadData(deviceID string, packets []*api.OBDPacket) error {
	req := &api.UploadDataRequest{
		DeviceID: deviceID,
		Packets:  packets,
	}

	data, err := c.doRequest(http.MethodPost, "/upload", req)
	if err != nil {
		return err
	}

	var resp api.UploadDataResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	fmt.Printf("Upload result: success=%v, uploaded=%d, deduplicated=%d\n",
		resp.Success, resp.Uploaded, resp.Deduplicated)
	return nil
}

func (c *Client) GetVehicleStats(deviceID, start, end string) error {
	path := "/vehicle/stats?device_id=" + deviceID
	if start != "" {
		path += "&start=" + start
	}
	if end != "" {
		path += "&end=" + end
	}

	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var resp api.VehicleDailyStatsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	fmt.Printf("Vehicle stats for %s:\n", deviceID)
	fmt.Printf("%-12s %-12s %-12s %-12s %-10s\n",
		"Date", "Mileage(km)", "Fuel(L)", "AvgFuel", "Safety")
	for _, s := range resp.Data {
		fmt.Printf("%-12s %-12.2f %-12.2f %-12.2f %-10d\n",
			s.Date.Format("2006-01-02"),
			s.Mileage, s.TotalFuel, s.AvgFuel, s.SafetyScore)
	}

	return nil
}

func (c *Client) GetFleetStats(start, end string) error {
	path := "/fleet/stats"
	if start != "" || end != "" {
		path += "?"
		if start != "" {
			path += "start=" + start
		}
		if end != "" {
			if start != "" {
				path += "&"
			}
			path += "end=" + end
		}
	}

	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var resp api.FleetDailyStatsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return err
	}

	fmt.Printf("Fleet stats:\n")
	fmt.Printf("%-12s %-16s %-16s %-16s\n",
		"Date", "TotalMileage", "AvgFuel", "AbnormalEvents")
	for _, s := range resp.Data {
		fmt.Printf("%-12s %-16.2f %-16.2f %-16d\n",
			s.Date.Format("2006-01-02"),
			s.TotalMileage, s.AvgFuelConsumption, s.AbnormalEvents)
	}

	return nil
}

func (c *Client) ExportCSV(yearMonth, outputFile string) error {
	path := "/export/csv?year_month=" + yearMonth

	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	if outputFile == "" {
		outputFile = fmt.Sprintf("fuel_report_%s.csv", yearMonth)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return err
	}

	fmt.Printf("CSV saved to: %s\n", outputFile)
	return nil
}

func (c *Client) GenerateAlerts() error {
	data, err := c.doRequest(http.MethodPost, "/alerts/generate", nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	fmt.Printf("Alerts generated: %v\n", result)
	return nil
}

func (c *Client) GetAlerts(deviceID, status string) error {
	path := "/alerts"
	if deviceID != "" || status != "" {
		path += "?"
		if deviceID != "" {
			path += "device_id=" + deviceID
		}
		if status != "" {
			if deviceID != "" {
				path += "&"
			}
			path += "status=" + status
		}
	}

	data, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}

	alertsJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(alertsJSON))

	return nil
}

func (c *Client) ResolveAlerts(alertIDs string) error {
	ids := strings.Split(alertIDs, ",")
	req := &api.AlertResolveRequest{AlertIDs: ids}

	data, err := c.doRequest(http.MethodPost, "/alerts", req)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func (c *Client) HealthCheck() error {
	data, err := c.doRequest(http.MethodGet, "/health", nil)
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func printUsage() {
	fmt.Println(`OBD Platform Client

Usage:
  client [flags] <command> [args]

Commands:
  upload <device_id>                     Upload simulated test data
  vehicle <device_id> [start] [end]     Get vehicle daily stats
  fleet [start] [end]                    Get fleet daily stats
  export <year_month> [output_file]      Export CSV (YYYY-MM)
  alerts-generate                        Generate alerts from existing data
  alerts [device_id] [status]           List alerts
  alerts-resolve <alert_ids>            Resolve alerts (comma-separated)
  health                                 Health check

Flags:
  -server URL    Server URL (default: http://localhost:9004)
  -help          Show this help`)
}

func main() {
	serverURL := flag.String("server", defaultServerURL, "Server URL")
	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help || flag.NArg() == 0 {
		printUsage()
		os.Exit(0)
	}

	client := NewClient(*serverURL)
	cmd := flag.Arg(0)
	args := flag.Args()[1:]

	var err error

	switch cmd {
	case "upload":
		if len(args) < 1 {
			err = fmt.Errorf("device_id required")
			break
		}
		testPackets := generateTestPackets(args[0])
		err = client.UploadData(args[0], testPackets)

	case "vehicle":
		if len(args) < 1 {
			err = fmt.Errorf("device_id required")
			break
		}
		start := ""
		end := ""
		if len(args) >= 2 {
			start = args[1]
		}
		if len(args) >= 3 {
			end = args[2]
		}
		err = client.GetVehicleStats(args[0], start, end)

	case "fleet":
		start := ""
		end := ""
		if len(args) >= 1 {
			start = args[0]
		}
		if len(args) >= 2 {
			end = args[1]
		}
		err = client.GetFleetStats(start, end)

	case "export":
		if len(args) < 1 {
			err = fmt.Errorf("year_month required (YYYY-MM)")
			break
		}
		output := ""
		if len(args) >= 2 {
			output = args[1]
		}
		err = client.ExportCSV(args[0], output)

	case "alerts-generate":
		err = client.GenerateAlerts()

	case "alerts":
		deviceID := ""
		status := ""
		if len(args) >= 1 {
			deviceID = args[0]
		}
		if len(args) >= 2 {
			status = args[1]
		}
		err = client.GetAlerts(deviceID, status)

	case "alerts-resolve":
		if len(args) < 1 {
			err = fmt.Errorf("alert_ids required (comma-separated)")
			break
		}
		err = client.ResolveAlerts(args[0])

	case "health":
		err = client.HealthCheck()

	default:
		err = fmt.Errorf("unknown command: %s", cmd)
		printUsage()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func generateTestPackets(deviceID string) []*api.OBDPacket {
	now := time.Now().Add(-12 * time.Hour)
	var packets []*api.OBDPacket

	lat := 39.9042
	lon := 116.4074
	speed := 50.0

	for i := 0; i < 10; i++ {
		timestamp := now.Add(time.Duration(i) * 30 * time.Second)
		lat += 0.001
		lon += 0.001
		speed += float64(i%5 - 2)
		if speed < 0 {
			speed = 0
		}
		if speed > 100 {
			speed = 100
		}

		packets = append(packets, &api.OBDPacket{
			DeviceID:          deviceID,
			Timestamp:         timestamp,
			Latitude:          lat,
			Longitude:         lon,
			Speed:             speed,
			EngineRPM:         2000 + i*100,
			FuelConsumption:   8.5 + float64(i%3)*0.5,
			HardAccelerations: i % 3,
			HardBrakes:        i % 2,
			HardTurns:         i % 4,
		})
	}

	return packets
}

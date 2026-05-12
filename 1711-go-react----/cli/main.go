package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

const baseURL = "http://localhost:8300/api"

type Sample struct {
	ID            string `json:"id"`
	SampleCode    string `json:"sample_code"`
	SampleType    string `json:"sample_type"`
	ReceivedDate  string `json:"received_date"`
	Status        string `json:"status"`
	UnitName      string `json:"unit_name"`
	Submitter     string `json:"submitter"`
	FinalPrice    float64 `json:"final_price"`
}

type ApiListResponse struct {
	Items []Sample `json:"items"`
	Total int      `json:"total"`
}

func main() {
	app := &cli.App{
		Name:  "genedeck-cli",
		Usage: "GeneDeck Sample Management CLI",
		Commands: []*cli.Command{
			{
				Name:  "receive-sample",
				Usage: "Receive a new sample",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "type", Required: true, Usage: "Sample type: saliva, blood, swab, tissue"},
					&cli.StringFlag{Name: "unit-id", Required: true, Usage: "Submitter unit ID"},
					&cli.StringFlag{Name: "submitter", Required: true, Usage: "Submitter name"},
					&cli.StringFlag{Name: "patient-name", Required: true, Usage: "Patient name"},
					&cli.StringFlag{Name: "patient-id", Required: true, Usage: "Patient ID card"},
					&cli.StringFlag{Name: "patient-phone", Required: true, Usage: "Patient phone"},
					&cli.StringFlag{Name: "test-items", Required: true, Usage: "Test item codes, comma separated (e.g. GEN001,GEN002)"},
					&cli.StringFlag{Name: "collection-date", Usage: "Collection date (YYYY-MM-DD)"},
				},
				Action: receiveSample,
			},
			{
				Name:  "start-testing",
				Usage: "Start testing for a sample",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "sample-id", Required: true, Usage: "Sample ID"},
				},
				Action: startTesting,
			},
			{
				Name:  "submit-report",
				Usage: "Submit a report for review",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "report-id", Required: true, Usage: "Report ID"},
				},
				Action: submitReport,
			},
			{
				Name:  "export-samples",
				Usage: "Export samples list to CSV",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "output", Aliases: []string{"o"}, Required: true, Usage: "Output file path"},
				},
				Action: exportSamples,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func receiveSample(c *cli.Context) error {
	testItems := strings.Split(c.String("test-items"), ",")
	for i := range testItems {
		testItems[i] = strings.TrimSpace(testItems[i])
	}

	collectionDate := c.String("collection-date")
	if collectionDate == "" {
		collectionDate = time.Now().Format("2006-01-02")
	}

	payload := map[string]interface{}{
		"sample_type":     c.String("type"),
		"collection_date": collectionDate,
		"unit_id":         c.String("unit-id"),
		"submitter":       c.String("submitter"),
		"patient": map[string]string{
			"name":    c.String("patient-name"),
			"id_card": c.String("patient-id"),
			"phone":   c.String("patient-phone"),
		},
		"test_item_codes": testItems,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(baseURL+"/samples", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed: %s - %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("Sample received successfully!\n")
	fmt.Printf("Sample ID: %v\n", result["id"])
	fmt.Printf("Sample Code: %v\n", result["sample_code"])
	return nil
}

func startTesting(c *cli.Context) error {
	sampleID := c.String("sample-id")

	payload := map[string]string{"new_status": "testing"}
	data, _ := json.Marshal(payload)

	req, _ := http.NewRequest("PATCH", baseURL+"/samples/"+sampleID+"/status", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed: %s - %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("Testing started for sample %s\n", sampleID)
	fmt.Printf("New status: %v\n", result["status"])
	return nil
}

func submitReport(c *cli.Context) error {
	reportID := c.String("report-id")

	resp, err := http.Post(baseURL+"/reports/"+reportID+"/submit", "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed: %s - %s", resp.Status, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	fmt.Printf("Report %s submitted for review\n", reportID)
	fmt.Printf("Status: %v\n", result["status"])
	fmt.Printf("Approval: %v\n", result["approval"])
	return nil
}

func exportSamples(c *cli.Context) error {
	output := c.String("output")

	resp, err := http.Get(baseURL + "/samples?size=1000")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed: %s - %s", resp.Status, string(body))
	}

	var apiResp ApiListResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return err
	}

	file, err := os.Create(output)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Sample Code", "Sample Type", "Received Date", "Status", "Unit", "Submitter", "Final Price"}
	writer.Write(headers)

	for _, s := range apiResp.Items {
		writer.Write([]string{
			s.SampleCode,
			s.SampleType,
			s.ReceivedDate,
			s.Status,
			s.UnitName,
			s.Submitter,
			strconv.FormatFloat(s.FinalPrice, 'f', 2, 64),
		})
	}

	fmt.Printf("Exported %d samples to %s\n", len(apiResp.Items), output)
	return nil
}

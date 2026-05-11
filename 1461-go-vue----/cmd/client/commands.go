package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"quality-trace/pkg/common"
)

func printJSON(data interface{}) {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(string(jsonData))
}

func cmdHealth(client *Client, args []string) {
	result, err := client.Health()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(result)
}

func cmdCreateBatch(client *Client, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: create-batch <product_name> <factory_code> <plan_quantity> [start_time]")
		os.Exit(1)
	}

	productName := args[0]
	factoryCode := args[1]
	planQuantity, _ := strconv.Atoi(args[2])

	startTime := time.Now()
	if len(args) > 3 {
		t, err := time.Parse(time.RFC3339, args[3])
		if err == nil {
			startTime = t
		}
	}

	batch, err := client.CreateBatch(common.CreateBatchRequest{
		ProductName:  productName,
		FactoryCode:  factoryCode,
		PlanQuantity: planQuantity,
		StartTime:    startTime,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(batch)
}

func cmdGetBatch(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: get-batch <batch_id>")
		os.Exit(1)
	}

	batch, err := client.GetBatch(args[0])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(batch)
}

func cmdCompleteBatch(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: complete-batch <batch_id> <actual_quantity> [end_time]")
		os.Exit(1)
	}

	batchID := args[0]
	actualQuantity, _ := strconv.Atoi(args[1])

	endTime := time.Now()
	if len(args) > 2 {
		t, err := time.Parse(time.RFC3339, args[2])
		if err == nil {
			endTime = t
		}
	}

	batch, err := client.CompleteBatch(common.CompleteBatchRequest{
		BatchID:        batchID,
		ActualQuantity: actualQuantity,
		EndTime:        endTime,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(batch)
}

func cmdListBatches(client *Client, args []string) {
	var status, productName string

	for i := 0; i < len(args); i++ {
		if args[i] == "--status" && i+1 < len(args) {
			status = args[i+1]
			i++
		} else if args[i] == "--product" && i+1 < len(args) {
			productName = args[i+1]
			i++
		}
	}

	batches, err := client.ListBatches(status, productName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(batches)
}

func cmdAddProcessFlow(client *Client, args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: add-process-flow <product_name> <seq1:name1> <seq2:name2> ...")
		os.Exit(1)
	}

	productName := args[0]
	processes := []common.ProcessDef{}

	for i := 1; i < len(args); i++ {
		parts := strings.SplitN(args[i], ":", 2)
		if len(parts) != 2 {
			fmt.Printf("Invalid process format: %s (expected seq:name)\n", args[i])
			os.Exit(1)
		}
		seq, _ := strconv.Atoi(parts[0])
		processes = append(processes, common.ProcessDef{
			Sequence:    seq,
			ProcessName: parts[1],
		})
	}

	flow, err := client.AddProcessFlow(common.AddProcessFlowRequest{
		ProductName: productName,
		Processes:   processes,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(flow)
}

func cmdStartProcess(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: start-process <batch_id> <operator> [start_time]")
		os.Exit(1)
	}

	batchID := args[0]
	operator := args[1]

	startTime := time.Now()
	if len(args) > 2 {
		t, err := time.Parse(time.RFC3339, args[2])
		if err == nil {
			startTime = t
		}
	}

	record, err := client.StartProcess(common.StartProcessRequest{
		BatchID:   batchID,
		Operator:  operator,
		StartTime: startTime,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(record)
}

func cmdCompleteProcess(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: complete-process <batch_id> <pass|fail> [end_time]")
		os.Exit(1)
	}

	batchID := args[0]
	resultStr := args[1]

	var result common.ProcessResult
	if resultStr == "pass" {
		result = common.ProcessResultPass
	} else if resultStr == "fail" {
		result = common.ProcessResultFail
	} else {
		fmt.Println("Result must be 'pass' or 'fail'")
		os.Exit(1)
	}

	endTime := time.Now()
	if len(args) > 2 {
		t, err := time.Parse(time.RFC3339, args[2])
		if err == nil {
			endTime = t
		}
	}

	record, err := client.CompleteProcess(common.CompleteProcessRequest{
		BatchID: batchID,
		Result:  result,
		EndTime: endTime,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(record)
}

func cmdReworkDecision(client *Client, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: rework-decision <batch_id> <rework|scrap>")
		os.Exit(1)
	}

	batchID := args[0]
	decision := args[1]

	if decision != "rework" && decision != "scrap" {
		fmt.Println("Decision must be 'rework' or 'scrap'")
		os.Exit(1)
	}

	batch, err := client.MakeReworkDecision(common.ReworkDecisionRequest{
		BatchID:  batchID,
		Decision: decision,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(batch)
}

func cmdListProcessRecords(client *Client, args []string) {
	var batchID string
	if len(args) > 0 {
		batchID = args[0]
	}

	records, err := client.ListProcessRecords(batchID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(records)
}

func cmdAddInspectionSpec(client *Client, args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: add-inspection-spec <product_name> <incoming|process|final> <metric_name> <lower_limit> <upper_limit>")
		os.Exit(1)
	}

	productName := args[0]
	typeStr := args[1]
	metricName := args[2]
	lowerLimit, _ := strconv.ParseFloat(args[3], 64)
	upperLimit, _ := strconv.ParseFloat(args[4], 64)

	var inspectionType common.InspectionType
	switch typeStr {
	case "incoming":
		inspectionType = common.InspectionTypeIncoming
	case "process":
		inspectionType = common.InspectionTypeProcess
	case "final":
		inspectionType = common.InspectionTypeFinal
	default:
		fmt.Println("Inspection type must be 'incoming', 'process', or 'final'")
		os.Exit(1)
	}

	spec, err := client.AddInspectionSpec(common.AddInspectionSpecRequest{
		ProductName:    productName,
		InspectionType: inspectionType,
		MetricName:     metricName,
		LowerLimit:     lowerLimit,
		UpperLimit:     upperLimit,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(spec)
}

func cmdAddInspectionRecord(client *Client, args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: add-inspection-record <batch_id> <incoming|process|final> <metric_name> <actual_value> <inspected_by> [inspected_at]")
		os.Exit(1)
	}

	batchID := args[0]
	typeStr := args[1]
	metricName := args[2]
	actualValue, _ := strconv.ParseFloat(args[3], 64)
	inspectedBy := args[4]

	var inspectionType common.InspectionType
	switch typeStr {
	case "incoming":
		inspectionType = common.InspectionTypeIncoming
	case "process":
		inspectionType = common.InspectionTypeProcess
	case "final":
		inspectionType = common.InspectionTypeFinal
	default:
		fmt.Println("Inspection type must be 'incoming', 'process', or 'final'")
		os.Exit(1)
	}

	inspectedAt := time.Now()
	if len(args) > 5 {
		t, err := time.Parse(time.RFC3339, args[5])
		if err == nil {
			inspectedAt = t
		}
	}

	record, err := client.AddInspectionRecord(common.AddInspectionRequest{
		BatchID:        batchID,
		InspectionType: inspectionType,
		MetricName:     metricName,
		ActualValue:    actualValue,
		InspectedBy:    inspectedBy,
		InspectedAt:    inspectedAt,
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(record)
}

func cmdListInspectionRecords(client *Client, args []string) {
	var batchID, inspectionType string

	for i := 0; i < len(args); i++ {
		if args[i] == "--batch" && i+1 < len(args) {
			batchID = args[i+1]
			i++
		} else if args[i] == "--type" && i+1 < len(args) {
			inspectionType = args[i+1]
			i++
		}
	}

	records, err := client.ListInspectionRecords(batchID, inspectionType)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(records)
}

func cmdGetMetrics(client *Client, args []string) {
	days := 30
	if len(args) > 0 {
		d, err := strconv.Atoi(args[0])
		if err == nil && d > 0 {
			days = d
		}
	}

	metrics, err := client.GetMetrics(days)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	printJSON(metrics)
}

func cmdExportBatches(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: export-batches <output_file> [--status status] [--product product_name]")
		os.Exit(1)
	}

	outputFile := args[0]
	var status, productName string

	for i := 1; i < len(args); i++ {
		if args[i] == "--status" && i+1 < len(args) {
			status = args[i+1]
			i++
		} else if args[i] == "--product" && i+1 < len(args) {
			productName = args[i+1]
			i++
		}
	}

	err := client.ExportBatches(status, productName, outputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported to %s\n", outputFile)
}

func cmdExportProcessRecords(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: export-processes <output_file> [--batch batch_id]")
		os.Exit(1)
	}

	outputFile := args[0]
	var batchID string

	for i := 1; i < len(args); i++ {
		if args[i] == "--batch" && i+1 < len(args) {
			batchID = args[i+1]
			i++
		}
	}

	err := client.ExportProcessRecords(batchID, outputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported to %s\n", outputFile)
}

func cmdExportInspectionRecords(client *Client, args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: export-inspections <output_file> [--batch batch_id] [--type inspection_type]")
		os.Exit(1)
	}

	outputFile := args[0]
	var batchID, inspectionType string

	for i := 1; i < len(args); i++ {
		if args[i] == "--batch" && i+1 < len(args) {
			batchID = args[i+1]
			i++
		} else if args[i] == "--type" && i+1 < len(args) {
			inspectionType = args[i+1]
			i++
		}
	}

	err := client.ExportInspectionRecords(batchID, inspectionType, outputFile)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Exported to %s\n", outputFile)
}

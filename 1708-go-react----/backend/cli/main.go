package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const baseURL = "http://localhost:8080/api"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "register-donor":
		registerDonor()
	case "record-test":
		recordTest()
	case "apply-blood":
		applyBlood()
	case "show-stats":
		showStats()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Blood Management System CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  bms-cli <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  register-donor   Register a new donor")
	fmt.Println("  record-test      Record a blood test")
	fmt.Println("  apply-blood      Submit a blood application")
	fmt.Println("  show-stats       Show statistics")
	fmt.Println("  help             Show this help")
	fmt.Println("\nExamples:")
	fmt.Println("  bms-cli register-donor -name=\"张三\" -idcard=\"110101199001011234\" -gender=male -abo=A -rh=Positive")
	fmt.Println("  bms-cli record-test -collection=\"BC000000000001\" -round=1 -vendor=\"厂家A\"")
	fmt.Println("  bms-cli apply-blood -hospital=\"北京医院\" -blood=A_Positive -product=Whole_Blood -qty=2")
	fmt.Println("  bms-cli show-stats -type=dashboard")
}

func registerDonor() {
	fs := flag.NewFlagSet("register-donor", flag.ExitOnError)
	name := fs.String("name", "", "Donor name")
	idcard := fs.String("idcard", "", "ID card number")
	gender := fs.String("gender", "", "Gender (male/female)")
	birthDate := fs.String("birthdate", time.Now().Format("2006-01-02"), "Birth date (YYYY-MM-DD)")
	abo := fs.String("abo", "", "ABO blood type (A/B/AB/O)")
	rh := fs.String("rh", "Positive", "Rh factor (Positive/Negative)")
	phone := fs.String("phone", "", "Phone number")
	address := fs.String("address", "", "Address")
	donationType := fs.String("donation-type", "Whole_400ml", "Donation type")

	fs.Parse(os.Args[2:])

	if *name == "" || *idcard == "" || *gender == "" || *abo == "" {
		fmt.Println("Error: name, idcard, gender, and abo are required")
		os.Exit(1)
	}

	birth, _ := time.Parse("2006-01-02", *birthDate)

	req := map[string]interface{}{
		"name":              *name,
		"id_card":           *idcard,
		"gender":            *gender,
		"birth_date":        birth,
		"abo_blood_type":    *abo,
		"rh_factor":         *rh,
		"phone":             *phone,
		"address":           *address,
		"donation_type":     *donationType,
		"health_check": map[string]interface{}{
			"is_fasting": true,
		},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/donors", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	fmt.Printf("Donor registered successfully!\n")
	fmt.Printf("Donor Number: %s\n", result["donor_number"])
	fmt.Printf("Donor ID: %s\n", result["id"])
}

func recordTest() {
	fs := flag.NewFlagSet("record-test", flag.ExitOnError)
	collection := fs.String("collection", "", "Collection ID or barcode")
	round := fs.Int("round", 1, "Test round (1/2/3)")
	vendor := fs.String("vendor", "", "Reagent vendor")
	operator := fs.String("operator", "", "Operator ID")

	fs.Parse(os.Args[2:])

	if *collection == "" {
		fmt.Println("Error: collection is required")
		os.Exit(1)
	}

	resultValues := []string{"Negative", "Negative", "Negative", "Negative", "Negative", "Negative", "Negative", "Negative", "Negative"}

	req := map[string]interface{}{
		"collection_id":  *collection,
		"test_round":     *round,
		"reagent_vendor": *vendor,
		"operator_id":    *operator,
		"results": map[string]string{
			"abo_front":     resultValues[0],
			"abo_back":      resultValues[1],
			"rhd":           resultValues[2],
			"alt":           resultValues[3],
			"hbsag":         resultValues[4],
			"anti_hcv":      resultValues[5],
			"anti_hiv":      resultValues[6],
			"anti_syphilis": resultValues[7],
			"nat":           resultValues[8],
		},
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/tests", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	fmt.Printf("Test recorded successfully!\n")
	fmt.Printf("Test Round: %d\n", int(result["test_round"].(float64)))
}

func applyBlood() {
	fs := flag.NewFlagSet("apply-blood", flag.ExitOnError)
	hospital := fs.String("hospital", "", "Hospital name")
	hospitalID := fs.String("hospital-id", "", "Hospital ID")
	bloodType := fs.String("blood", "", "Blood type (e.g., A_Positive)")
	product := fs.String("product", "Whole_Blood", "Product type")
	qty := fs.Int("qty", 1, "Quantity")
	urgency := fs.String("urgency", "Regular", "Urgency (Regular/Urgent/Special_Urgent)")

	fs.Parse(os.Args[2:])

	if *hospital == "" || *bloodType == "" {
		fmt.Println("Error: hospital and blood are required")
		os.Exit(1)
	}

	req := map[string]interface{}{
		"hospital_name": *hospital,
		"hospital_id":   *hospitalID,
		"blood_type":    *bloodType,
		"product_type":  *product,
		"quantity":      *qty,
		"urgency":       *urgency,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/requests", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error: HTTP %d - %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)
	fmt.Printf("Blood application submitted!\n")
	fmt.Printf("Request ID: %s\n", result["id"])
	fmt.Printf("Status: %s\n", result["status"])
}

func showStats() {
	fs := flag.NewFlagSet("show-stats", flag.ExitOnError)
	statType := fs.String("type", "dashboard", "Stats type (dashboard/monthly-collection/monthly-scrap/hospital-ranking/inventory-turnover)")

	fs.Parse(os.Args[2:])

	var url string
	switch *statType {
	case "dashboard":
		url = baseURL + "/stats/dashboard"
	case "monthly-collection":
		url = baseURL + "/stats/monthly-collection"
	case "monthly-scrap":
		url = baseURL + "/stats/monthly-scrap"
	case "hospital-ranking":
		url = baseURL + "/stats/hospital-ranking"
	case "inventory-turnover":
		url = baseURL + "/stats/inventory-turnover"
	default:
		url = baseURL + "/stats/dashboard"
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: HTTP %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var prettyJSON bytes.Buffer
	json.Indent(&prettyJSON, respBody, "", "  ")
	fmt.Println(prettyJSON.String())
}

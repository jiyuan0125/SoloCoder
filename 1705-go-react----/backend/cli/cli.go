package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

const baseURL = "http://localhost:8080/api"

var token string

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "login":
		if len(os.Args) < 4 {
			fmt.Println("Usage: cli login <username> <password>")
			return
		}
		login(os.Args[2], os.Args[3])

	case "create-consultation":
		if len(os.Args) < 10 {
			fmt.Println("Usage: cli create-consultation <patient_name> <gender> <age> <id_card> <phone> <chief_complaint>")
			return
		}
		createConsultation(os.Args[2:])

	case "list-consultations":
		listConsultations()

	case "export":
		if len(os.Args) < 4 {
			fmt.Println("Usage: cli export <start_date> <end_date>")
			return
		}
		exportByDate(os.Args[2], os.Args[3])

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Telemedicine CLI")
	fmt.Println("Commands:")
	fmt.Println("  login <username> <password>              - Login and get token")
	fmt.Println("  create-consultation <params>            - Create a consultation")
	fmt.Println("  list-consultations                    - List consultations")
	fmt.Println("  export <start_date> <end_date>          - Export consultations by date range")
}

func login(username, password string) {
	data := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(data)

	resp, err := http.Post(baseURL+"/login", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if resp.StatusCode != 200 {
		fmt.Println("Login failed:", result["error"])
		return
	}

	token = result["token"].(string)
	fmt.Println("Login successful")
	fmt.Println("Token:", token)
}

func createConsultation(args []string) {
	if token == "" {
		fmt.Println("Please login first")
		return
	}

	patientName, gender, ageStr, idCard, phone := args[0], args[1], args[2], args[3], args[4]
	chiefComplaint := args[5]
	pastHistory := ""
	if len(args) > 6 {
		pastHistory = args[6]
	}

	age, _ := strconv.Atoi(ageStr)

	data := map[string]interface{}{
		"patient_name":    patientName,
		"patient_gender":  gender,
		"patient_age":     age,
		"patient_id_card":   idCard,
		"patient_phone":        phone,
		"chief_complaint":      chiefComplaint,
		"past_history":        pastHistory,
	}

	body, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", baseURL+"/consultations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	fmt.Println(string(result))
}

func listConsultations() {
	if token == "" {
		fmt.Println("Please login first")
		return
	}

	req, _ := http.NewRequest("GET", baseURL+"/consultations", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	result, _ := io.ReadAll(resp.Body)
	fmt.Println(string(result))
}

func exportByDate(startDate, endDate string) {
	if token == "" {
		fmt.Println("Please login first")
		return
	}

	url := fmt.Sprintf("%s/admin/export/date-range?start_date=%s&end_date=%s", baseURL, startDate, endDate)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("Export failed:", string(body))
		return
	}

	filename := fmt.Sprintf("export_%s_%s.csv", startDate, endDate)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	io.Copy(file, resp.Body)
	fmt.Println("Exported to:", filename)
}

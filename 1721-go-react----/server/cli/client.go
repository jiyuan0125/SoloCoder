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

const baseURL = "http://localhost:8300/api"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "create-course":
		createCourse()
	case "recharge":
		recharge()
	case "enroll-course":
		enrollCourse()
	case "show-stats":
		showStats()
	default:
		fmt.Println("Unknown command:", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Education Platform CLI")
	fmt.Println("\nUsage:")
	fmt.Println("  client create-course <name> <instructor> <type:live|record> <price> [subject] [duration]")
	fmt.Println("  client recharge <student_id> <card_number> <password>")
	fmt.Println("  client enroll-course <student_id> <course_id>")
	fmt.Println("  client show-stats")
	fmt.Println("\nExamples:")
	fmt.Println("  client create-course \"数学入门\" \"张老师\" live 9900 math 90")
	fmt.Println("  client create-course \"语文基础\" \"李老师\" record 0 chinese")
	fmt.Println("  client recharge 1 CARD001 PASS123")
	fmt.Println("  client enroll-course 1 3")
	fmt.Println("  client show-stats")
}

func createCourse() {
	args := os.Args[2:]
	if len(args) < 4 {
		fmt.Println("Usage: client create-course <name> <instructor> <type:live|record> <price> [subject] [duration]")
		return
	}

	name := args[0]
	instructor := args[1]
	courseType := args[2]
	price, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		fmt.Println("Invalid price:", args[3])
		return
	}

	course := map[string]interface{}{
		"name":        name,
		"instructor":  instructor,
		"course_type": courseType,
		"price":       price,
	}

	if len(args) > 4 {
		course["subject"] = args[4]
	}

	if courseType == "live" {
		duration := 60
		if len(args) > 5 {
			if d, err := strconv.Atoi(args[5]); err == nil {
				duration = d
			}
		}
		course["duration"] = duration
		maxOnline := 500
		course["max_online"] = maxOnline
	}

	if courseType == "record" {
		videoDuration := int64(3600)
		course["video_duration"] = videoDuration
	}

	body, _ := json.Marshal(course)
	resp, err := http.Post(baseURL+"/courses", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 201 {
		fmt.Println("Course created successfully!")
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)
		fmt.Printf("Course ID: %.0f\n", result["id"])
	} else {
		fmt.Printf("Error [%d]: %s\n", resp.StatusCode, string(respBody))
	}
}

func recharge() {
	args := os.Args[2:]
	if len(args) < 3 {
		fmt.Println("Usage: client recharge <student_id> <card_number> <password>")
		return
	}

	studentID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid student ID:", args[0])
		return
	}

	cardNumber := args[1]
	password := args[2]

	req := map[string]interface{}{
		"student_id":  studentID,
		"card_number": cardNumber,
		"password":    password,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/recharge", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Println("Recharge successful!")
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)
		fmt.Printf("Amount: %.0f cents (%.2f yuan)\n", result["amount"], float64(result["amount"].(float64))/100)
	} else {
		fmt.Printf("Error [%d]: %s\n", resp.StatusCode, string(respBody))
	}
}

func enrollCourse() {
	args := os.Args[2:]
	if len(args) < 2 {
		fmt.Println("Usage: client enroll-course <student_id> <course_id>")
		return
	}

	studentID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		fmt.Println("Invalid student ID:", args[0])
		return
	}

	courseID, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		fmt.Println("Invalid course ID:", args[1])
		return
	}

	req := map[string]interface{}{
		"student_id": studentID,
		"course_id":  courseID,
	}

	body, _ := json.Marshal(req)
	resp, err := http.Post(baseURL+"/enroll", "application/json", bytes.NewReader(body))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Println("Enrolled successfully!")
	} else {
		fmt.Printf("Error [%d]: %s\n", resp.StatusCode, string(respBody))
	}
}

func showStats() {
	fmt.Println("\n=== Subject Statistics ===")
	resp, err := http.Get(baseURL + "/stats/subjects")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		var subjects []map[string]interface{}
		json.Unmarshal(body, &subjects)
		for i, s := range subjects {
			fmt.Printf("%d. %s: %d courses, %.0f enrollments\n",
				i+1, s["subject_name"], int(s["course_count"].(float64)), s["enrollment_count"])
		}
	}

	fmt.Println("\n=== Instructor Statistics ===")
	resp2, err := http.Get(baseURL + "/stats/instructors")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		defer resp2.Body.Close()
		body, _ := io.ReadAll(resp2.Body)
		var instructors []map[string]interface{}
		json.Unmarshal(body, &instructors)
		for i, inst := range instructors {
			avg := 0.0
			if inst["avg_rating"] != nil {
				avg = inst["avg_rating"].(float64)
			}
			fmt.Printf("%d. %s: %d courses, avg rating %.1f\n",
				i+1, inst["instructor_name"], int(inst["course_count"].(float64)), avg)
		}
	}

	fmt.Println("\n=== Live Course Statistics ===")
	resp3, err := http.Get(baseURL + "/stats/live")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		defer resp3.Body.Close()
		body, _ := io.ReadAll(resp3.Body)
		var lives []map[string]interface{}
		json.Unmarshal(body, &lives)
		if len(lives) == 0 {
			fmt.Println("No live stats available yet.")
		}
		for i, l := range lives {
			fmt.Printf("%d. %s: peak=%d, avg_online=%.1fs, danmaku=%.0f, participation=%.1f%%\n",
				i+1, l["course_name"], int(l["peak_online"].(float64)),
				l["average_online_time"], l["danmaku_count"], l["vote_participation"])
		}
	}
	fmt.Println()
}

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
)

const baseURL = "http://localhost:8300/api"

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "add-course":
		handleAddCourse(os.Args[2:])
	case "add-prerequisite":
		handleAddPrerequisite(os.Args[2:])
	case "start-learning":
		handleStartLearning(os.Args[2:])
	case "show-progress":
		handleShowProgress(os.Args[2:])
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Learning Platform CLI")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add-course           Add a new course")
	fmt.Println("  add-prerequisite     Add a prerequisite relationship")
	fmt.Println("  start-learning       Start learning a course")
	fmt.Println("  show-progress        Show progress for a student and course")
	fmt.Println("  help                 Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli add-course --id C1 --name \"JavaScript Basics\" --hours 20 --difficulty beginner --domain frontend")
	fmt.Println("  cli add-prerequisite --course C2 --prereq C1")
	fmt.Println("  cli start-learning --student S1 --course C1")
	fmt.Println("  cli show-progress --student S1 --course C1")
}

func handleAddCourse(args []string) {
	fs := flag.NewFlagSet("add-course", flag.ExitOnError)
	id := fs.String("id", "", "Course ID")
	name := fs.String("name", "", "Course name")
	description := fs.String("description", "", "Course description")
	hours := fs.Int("hours", 0, "Expected learning hours")
	difficulty := fs.String("difficulty", "beginner", "Difficulty (beginner/elementary/intermediate/advanced)")
	domain := fs.String("domain", "frontend", "Domain (frontend/backend/database/algorithm/devops)")
	fs.Parse(args)

	if *id == "" || *name == "" {
		fmt.Println("Error: --id and --name are required")
		os.Exit(1)
	}

	course := map[string]interface{}{
		"id":            *id,
		"name":          *name,
		"description":   *description,
		"expected_hours": *hours,
		"difficulty":    *difficulty,
		"domain":        *domain,
		"units":         []interface{}{},
		"quiz":          map[string]interface{}{"questions": []interface{}{}},
	}

	body, err := json.Marshal(course)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(baseURL+"/courses", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result APIResponse
	json.Unmarshal(respBody, &result)

	if resp.StatusCode == http.StatusCreated {
		fmt.Printf("Course added successfully: %s\n", *id)
	} else if resp.StatusCode == http.StatusConflict {
		fmt.Printf("Error: Course with id %s already exists (409)\n", *id)
	} else {
		fmt.Printf("Error: %s\n", result.Error)
	}
}

func handleAddPrerequisite(args []string) {
	fs := flag.NewFlagSet("add-prerequisite", flag.ExitOnError)
	courseID := fs.String("course", "", "Course ID")
	prereqID := fs.String("prereq", "", "Prerequisite course ID")
	fs.Parse(args)

	if *courseID == "" || *prereqID == "" {
		fmt.Println("Error: --course and --prereq are required")
		os.Exit(1)
	}

	req := map[string]string{
		"course_id":      *courseID,
		"prerequisite_id": *prereqID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(baseURL+"/courses/prerequisites", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result APIResponse
	json.Unmarshal(respBody, &result)

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Prerequisite added: %s -> %s\n", *prereqID, *courseID)
	} else if resp.StatusCode == http.StatusBadRequest {
		fmt.Printf("Error: Cycle dependency detected (400)\n")
	} else if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("Error: Course not found (404)\n")
	} else {
		fmt.Printf("Error: %s\n", result.Error)
	}
}

func handleStartLearning(args []string) {
	fs := flag.NewFlagSet("start-learning", flag.ExitOnError)
	studentID := fs.String("student", "", "Student ID")
	courseID := fs.String("course", "", "Course ID")
	fs.Parse(args)

	if *studentID == "" || *courseID == "" {
		fmt.Println("Error: --student and --course are required")
		os.Exit(1)
	}

	req := map[string]string{
		"student_id": *studentID,
		"course_id":  *courseID,
	}

	body, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(baseURL+"/students/start-learning", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result APIResponse
	json.Unmarshal(respBody, &result)

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Started learning: Student %s -> Course %s\n", *studentID, *courseID)
	} else if resp.StatusCode == http.StatusBadRequest {
		fmt.Printf("Error: Maximum concurrent courses exceeded (400)\n")
	} else if resp.StatusCode == http.StatusForbidden {
		fmt.Printf("Error: Prerequisites not completed (403)\n")
	} else {
		fmt.Printf("Error: %s\n", result.Error)
	}
}

func handleShowProgress(args []string) {
	fs := flag.NewFlagSet("show-progress", flag.ExitOnError)
	studentID := fs.String("student", "", "Student ID")
	courseID := fs.String("course", "", "Course ID")
	fs.Parse(args)

	if *studentID == "" || *courseID == "" {
		fmt.Println("Error: --student and --course are required")
		os.Exit(1)
	}

	url := fmt.Sprintf("%s/students/progress?student_id=%s&course_id=%s", baseURL, *studentID, *courseID)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result APIResponse
	json.Unmarshal(respBody, &result)

	if resp.StatusCode == http.StatusOK {
		dataMap, ok := result.Data.(map[string]interface{})
		if ok {
			progress, _ := dataMap["progress_percentage"].(float64)
			completedUnits, _ := dataMap["completed_unit_ids"].([]interface{})
			fmt.Printf("Progress for Student %s, Course %s:\n", *studentID, *courseID)
			fmt.Printf("  Percentage: %.1f%%\n", progress)
			fmt.Printf("  Completed units: %d\n", len(completedUnits))
		} else {
			fmt.Println(string(respBody))
		}
	} else if resp.StatusCode == http.StatusNotFound {
		fmt.Printf("Error: Progress not found (404)\n")
	} else {
		fmt.Printf("Error: %s\n", result.Error)
	}
}

func containsFlag(args []string, flag string) bool {
	for _, arg := range args {
		if strings.HasPrefix(arg, flag) {
			return true
		}
	}
	return false
}

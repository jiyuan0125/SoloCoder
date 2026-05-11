package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"housekeeping-training/common"
)

var baseURL = "http://localhost:9018/api"

func setServerURL() {
	if url := os.Getenv("SERVER_URL"); url != "" {
		baseURL = strings.TrimSuffix(url, "/api")
		if !strings.HasPrefix(baseURL, "http") {
			baseURL = "http://" + baseURL
		}
		baseURL += "/api"
	}
}

func callAPI(method, endpoint string, body interface{}) ([]byte, error) {
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func printResponse(data []byte) {
	var resp common.Response
	json.Unmarshal(data, &resp)
	if resp.Success {
		fmt.Println("✓ Success:", resp.Message)
		if resp.Data != nil {
			pretty, _ := json.MarshalIndent(resp.Data, "", "  ")
			fmt.Println(string(pretty))
		}
	} else {
		fmt.Println("✗ Error:", resp.Message)
	}
}

func cmdListCourses() {
	data, err := callAPI("GET", "/courses", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdCreateCourse(args []string) {
	if len(args) < 6 {
		fmt.Println("Usage: course create <name> <beginner|advanced> <hours> <instructor> <fee> <capacity>")
		return
	}
	name := args[0]
	level := common.CourseLevel(args[1])
	hours, _ := strconv.Atoi(args[2])
	instructor := args[3]
	fee, _ := strconv.Atoi(args[4])
	capacity, _ := strconv.Atoi(args[5])

	req := common.CreateCourseRequest{
		Name:        name,
		Level:       level,
		Hours:       hours,
		Instructor:  instructor,
		Fee:         fee,
		MaxCapacity: capacity,
	}
	data, err := callAPI("POST", "/courses", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListSchedules() {
	data, err := callAPI("GET", "/schedules", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdCreateSchedule(args []string) {
	if len(args) < 5 {
		fmt.Println("Usage: schedule create <course_id> <start_date(YYYY-MM-DD)> <time_slot> <classroom> <total_classes>")
		return
	}
	courseID := args[0]
	startDate := args[1]
	timeSlot := args[2]
	classroom := args[3]
	totalClasses, _ := strconv.Atoi(args[4])

	req := common.CreateScheduleRequest{
		CourseID:     courseID,
		StartDate:    startDate,
		TimeSlot:     timeSlot,
		Classroom:    classroom,
		TotalClasses: totalClasses,
	}
	data, err := callAPI("POST", "/schedules", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListStudents() {
	data, err := callAPI("GET", "/students", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdCreateStudent(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: student create <name> <id_card> <phone> <education>")
		return
	}
	req := common.CreateStudentRequest{
		Name:      args[0],
		IDCard:    args[1],
		Phone:     args[2],
		Education: args[3],
	}
	data, err := callAPI("POST", "/students", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListEnrollments() {
	data, err := callAPI("GET", "/enrollments", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdEnroll(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: enroll <student_id> <schedule_id>")
		return
	}
	req := common.EnrollRequest{
		StudentID:  args[0],
		ScheduleID: args[1],
	}
	data, err := callAPI("POST", "/enrollments", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdPayFirst(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pay-first <enrollment_id>")
		return
	}
	data, err := callAPI("POST", "/pay-first/"+args[0], nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdPaySecond(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: pay-second <enrollment_id>")
		return
	}
	data, err := callAPI("POST", "/pay-second/"+args[0], nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdRecordAttendance(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: attendance <enrollment_id> <date(YYYY-MM-DD)> <present|leave|absent>")
		return
	}
	req := common.RecordAttendanceRequest{
		EnrollmentID: args[0],
		ClassDate:    args[1],
		Status:       common.AttendanceStatus(args[2]),
	}
	data, err := callAPI("POST", "/attendance", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdAttendanceRate(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: attendance-rate <enrollment_id>")
		return
	}
	data, err := callAPI("GET", "/attendance-rate/"+args[0], nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListExams() {
	data, err := callAPI("GET", "/exams", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdRecordExam(args []string) {
	if len(args) < 4 {
		fmt.Println("Usage: exam <enrollment_id> <date(YYYY-MM-DD)> <written_score> <practical_score> [retake]")
		return
	}
	written, _ := strconv.Atoi(args[2])
	practical, _ := strconv.Atoi(args[3])
	isRetake := len(args) >= 5 && args[4] == "retake"

	req := common.RecordExamRequest{
		EnrollmentID:   args[0],
		ExamDate:       args[1],
		WrittenScore:   written,
		PracticalScore: practical,
		IsRetake:       isRetake,
	}
	data, err := callAPI("POST", "/exams", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListCertificates(args []string) {
	endpoint := "/certificates"
	if len(args) >= 1 {
		endpoint += "?student_id=" + args[0]
	}
	data, err := callAPI("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListEmployers() {
	data, err := callAPI("GET", "/employers", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdCreateEmployer(args []string) {
	if len(args) < 3 {
		fmt.Println("Usage: employer create <name> <contact> <phone>")
		return
	}
	req := common.CreateEmployerRequest{
		Name:    args[0],
		Contact: args[1],
		Phone:   args[2],
	}
	data, err := callAPI("POST", "/employers", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListJobs() {
	data, err := callAPI("GET", "/jobs", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdCreateJob(args []string) {
	if len(args) < 6 {
		fmt.Println("Usage: job create <employer_id> <position> <required_cert> <min_salary> <max_salary> <location>")
		return
	}
	minSalary, _ := strconv.Atoi(args[3])
	maxSalary, _ := strconv.Atoi(args[4])
	req := common.CreateJobPostingRequest{
		EmployerID:   args[0],
		Position:     args[1],
		RequiredCert: args[2],
		MinSalary:    minSalary,
		MaxSalary:    maxSalary,
		WorkLocation: args[5],
	}
	data, err := callAPI("POST", "/jobs", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdGenerateRecommendations(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: recommend generate <job_id>")
		return
	}
	data, err := callAPI("GET", "/recommendations?generate_for_job="+args[0], nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListRecommendations() {
	data, err := callAPI("GET", "/recommendations", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdProcessRecommendation(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: recommend process <rec_id> <accepted|rejected> [notes]")
		return
	}
	req := common.ProcessRecommendationRequest{
		Result: common.InterviewResult(args[1]),
	}
	if len(args) >= 3 {
		req.Notes = args[2]
	}
	data, err := callAPI("POST", "/recommendations/"+args[0]+"/process", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdRequestRefund(args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: refund <enrollment_id> <reason>")
		return
	}
	req := common.RefundRequestReq{
		EnrollmentID: args[0],
		Reason:       args[1],
	}
	data, err := callAPI("POST", "/refunds", req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func cmdListRefunds() {
	data, err := callAPI("GET", "/refunds", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printResponse(data)
}

func showHelp() {
	fmt.Println("家政人员培训管理平台 CLI")
	fmt.Println()
	fmt.Println("课程管理:")
	fmt.Println("  course list                      列出所有课程")
	fmt.Println("  course create ...                创建课程")
	fmt.Println("  schedule list                    列出开班计划")
	fmt.Println("  schedule create ...              创建开班计划")
	fmt.Println()
	fmt.Println("学员管理:")
	fmt.Println("  student list                     列出所有学员")
	fmt.Println("  student create ...               创建学员")
	fmt.Println("  enrollment list                  列出报名记录")
	fmt.Println("  enroll <student_id> <schedule_id> 学员报名课程")
	fmt.Println("  pay-first <enrollment_id>        支付首期学费")
	fmt.Println("  pay-second <enrollment_id>       支付二期学费")
	fmt.Println()
	fmt.Println("出勤管理:")
	fmt.Println("  attendance <enrollment_id> <date> <status>")
	fmt.Println("  attendance-rate <enrollment_id>  查看出勤率")
	fmt.Println()
	fmt.Println("考证管理:")
	fmt.Println("  exam list                        列出考试记录")
	fmt.Println("  exam <enrollment_id> <date> <written> <practical> [retake]")
	fmt.Println("  certificate list [student_id]    查看证书")
	fmt.Println()
	fmt.Println("就业推荐:")
	fmt.Println("  employer list                    列出雇主")
	fmt.Println("  employer create ...              创建雇主")
	fmt.Println("  job list                         列出岗位需求")
	fmt.Println("  job create ...                   创建岗位需求")
	fmt.Println("  recommend generate <job_id>      生成推荐")
	fmt.Println("  recommend list                   列出推荐")
	fmt.Println("  recommend process <id> <result>  处理推荐")
	fmt.Println()
	fmt.Println("退费:")
	fmt.Println("  refund <enrollment_id> <reason>  申请退费")
	fmt.Println("  refund list                      列出退费申请")
}

func main() {
	setServerURL()

	if len(os.Args) < 2 {
		showHelp()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "course":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "list":
			cmdListCourses()
		case "create":
			cmdCreateCourse(args[1:])
		}
	case "schedule":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "list":
			cmdListSchedules()
		case "create":
			cmdCreateSchedule(args[1:])
		}
	case "student":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "list":
			cmdListStudents()
		case "create":
			cmdCreateStudent(args[1:])
		}
	case "enrollment":
		if len(args) > 0 && args[0] == "list" {
			cmdListEnrollments()
		}
	case "enroll":
		cmdEnroll(args)
	case "pay-first":
		cmdPayFirst(args)
	case "pay-second":
		cmdPaySecond(args)
	case "attendance":
		cmdRecordAttendance(args)
	case "attendance-rate":
		cmdAttendanceRate(args)
	case "exam":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "list":
			cmdListExams()
		default:
			cmdRecordExam(args)
		}
	case "certificate":
		if len(args) > 0 && args[0] == "list" {
			cmdListCertificates(args[1:])
		}
	case "employer":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "list":
			cmdListEmployers()
		case "create":
			cmdCreateEmployer(args[1:])
		}
	case "job":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "list":
			cmdListJobs()
		case "create":
			cmdCreateJob(args[1:])
		}
	case "recommend":
		if len(args) == 0 {
			showHelp()
			return
		}
		switch args[0] {
		case "generate":
			cmdGenerateRecommendations(args[1:])
		case "list":
			cmdListRecommendations()
		case "process":
			cmdProcessRecommendation(args[1:])
		}
	case "refund":
		if len(args) > 0 && args[0] == "list" {
			cmdListRefunds()
		} else {
			cmdRequestRefund(args)
		}
	case "help", "-h", "--help":
		showHelp()
	default:
		showHelp()
	}
}

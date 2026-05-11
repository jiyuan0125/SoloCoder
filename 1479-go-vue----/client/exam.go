package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/drivingschool/common"
)

func handleExamCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("缺少考试管理子命令")
	}

	subCmd := args[0]
	switch subCmd {
	case "plan":
		return handleExamPlan(args[1:])
	case "plans":
		return handleExamPlans()
	case "book":
		return handleExamBook(args[1:])
	case "confirm":
		return handleExamConfirm(args[1:])
	case "cancel":
		return handleExamCancel(args[1:])
	case "approve":
		return handleExamApprove(args[1:])
	case "bookings":
		return handleExamBookings(args[1:])
	default:
		return fmt.Errorf("未知的考试管理子命令: %s", subCmd)
	}
}

func handleExamPlan(args []string) error {
	if len(args) < 4 {
		return errors.New("用法: exam plan <日期> <科目> <考场> <名额>")
	}

	date, err := time.Parse("2006-01-02", args[0])
	if err != nil {
		date, err = time.Parse(time.RFC3339, args[0])
		if err != nil {
			return errors.New("日期格式错误，使用 2006-01-02 或 RFC3339")
		}
	}

	subject, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("科目必须是数字")
	}

	quota, err := strconv.Atoi(args[3])
	if err != nil {
		return errors.New("名额必须是数字")
	}

	req := common.CreateExamPlanRequest{
		Date:       date,
		Subject:    common.Subject(subject),
		Venue:      args[2],
		TotalQuota: quota,
	}

	resp, err := makePostRequest("/api/exams/plans/create", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("考试计划创建成功:")
	prettyPrint(resp.Data)
	return nil
}

func handleExamPlans() error {
	resp, err := makeGetRequest("/api/exams/plans/list")
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("考试计划列表:")
	prettyPrint(resp.Data)
	return nil
}

func handleExamBook(args []string) error {
	if len(args) < 2 {
		return errors.New("用法: exam book <学员ID> <考试计划ID>")
	}

	req := common.BookExamRequest{
		StudentID:  args[0],
		ExamPlanID: args[1],
	}

	resp, err := makePostRequest("/api/exams/book", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("考试预约成功:")
	prettyPrint(resp.Data)
	return nil
}

func handleExamConfirm(args []string) error {
	if len(args) < 1 {
		return errors.New("用法: exam confirm <预约ID>")
	}

	req := common.ConfirmWaitingExamRequest{
		ExamBookingID: args[0],
	}

	resp, err := makePostRequest("/api/exams/confirm", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("候补确认成功")
	return nil
}

func handleExamCancel(args []string) error {
	if len(args) < 2 {
		return errors.New("用法: exam cancel <学员ID> <预约ID>")
	}

	req := common.CancelExamBookingRequest{
		StudentID:     args[0],
		ExamBookingID: args[1],
	}

	resp, err := makePostRequest("/api/exams/cancel", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	prettyPrint(resp.Data)
	return nil
}

func handleExamApprove(args []string) error {
	if len(args) < 2 {
		return errors.New("用法: exam approve <预约ID> <true/false>")
	}

	approved, err := strconv.ParseBool(args[1])
	if err != nil {
		return errors.New("第二个参数必须是 true 或 false")
	}

	req := common.ApproveCancelExamRequest{
		ExamBookingID: args[0],
		Approved:      approved,
	}

	resp, err := makePostRequest("/api/exams/approve", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	prettyPrint(resp.Data)
	return nil
}

func handleExamBookings(args []string) error {
	path := "/api/exams/bookings"
	if len(args) > 0 {
		path += "?student_id=" + args[0]
	}

	resp, err := makeGetRequest(path)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("考试预约列表:")
	prettyPrint(resp.Data)
	return nil
}

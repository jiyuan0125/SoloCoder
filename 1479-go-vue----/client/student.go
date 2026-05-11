package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/drivingschool/common"
)

func handleStudentCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("缺少学员管理子命令")
	}

	subCmd := args[0]
	switch subCmd {
	case "register":
		return handleStudentRegister(args[1:])
	case "list":
		return handleStudentList()
	case "get":
		return handleStudentGet(args[1:])
	case "status":
		return handleStudentStatus(args[1:])
	case "hours":
		return handleStudentHours(args[1:])
	default:
		return fmt.Errorf("未知的学员管理子命令: %s", subCmd)
	}
}

func handleStudentRegister(args []string) error {
	if len(args) < 4 {
		return errors.New("用法: student register <姓名> <身份证号> <手机号> <车型>")
	}

	req := common.RegisterStudentRequest{
		Name:        args[0],
		IDCard:      args[1],
		Phone:       args[2],
		VehicleType: common.VehicleType(args[3]),
	}

	resp, err := makePostRequest("/api/students/register", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("学员注册成功:")
	prettyPrint(resp.Data)
	return nil
}

func handleStudentList() error {
	resp, err := makeGetRequest("/api/students/list")
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("学员列表:")
	prettyPrint(resp.Data)
	return nil
}

func handleStudentGet(args []string) error {
	if len(args) < 1 {
		return errors.New("用法: student get <学员ID>")
	}

	resp, err := makeGetRequest("/api/students/" + args[0])
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("学员信息:")
	prettyPrint(resp.Data)
	return nil
}

func handleStudentStatus(args []string) error {
	if len(args) < 3 {
		return errors.New("用法: student status <学员ID> <科目> <状态>")
	}

	subject, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("科目必须是数字")
	}

	req := common.UpdateSubjectStatusRequest{
		StudentID: args[0],
		Subject:   common.Subject(subject),
		Status:    common.SubjectStatus(args[2]),
	}

	resp, err := makePostRequest("/api/students/status", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("状态更新成功")
	return nil
}

func handleStudentHours(args []string) error {
	if len(args) < 3 {
		return errors.New("用法: student hours <学员ID> <科目> <学时>")
	}

	subject, err := strconv.Atoi(args[1])
	if err != nil {
		return errors.New("科目必须是数字")
	}

	hours, err := strconv.Atoi(args[2])
	if err != nil {
		return errors.New("学时必须是数字")
	}

	req := common.AddStudyHoursRequest{
		StudentID: args[0],
		Subject:   common.Subject(subject),
		Hours:     hours,
	}

	resp, err := makePostRequest("/api/students/hours", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("学时添加成功")
	return nil
}

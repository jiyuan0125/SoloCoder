package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/drivingschool/common"
)

func handleCoachCommand(args []string) error {
	if len(args) == 0 {
		return errors.New("缺少教练管理子命令")
	}

	subCmd := args[0]
	switch subCmd {
	case "add":
		return handleCoachAdd(args[1:])
	case "list":
		return handleCoachList()
	case "get":
		return handleCoachGet(args[1:])
	case "practice":
		return handleCoachPractice(args[1:])
	default:
		return fmt.Errorf("未知的教练管理子命令: %s", subCmd)
	}
}

func handleCoachAdd(args []string) error {
	if len(args) < 3 {
		return errors.New("用法: coach add <姓名> <准教车型> <手机号>")
	}

	req := common.AddCoachRequest{
		Name:         args[0],
		TeachingType: common.VehicleType(args[1]),
		Phone:        args[2],
	}

	resp, err := makePostRequest("/api/coaches/add", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("教练添加成功:")
	prettyPrint(resp.Data)
	return nil
}

func handleCoachList() error {
	resp, err := makeGetRequest("/api/coaches/list")
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("教练列表:")
	prettyPrint(resp.Data)
	return nil
}

func handleCoachGet(args []string) error {
	if len(args) < 1 {
		return errors.New("用法: coach get <教练ID>")
	}

	resp, err := makeGetRequest("/api/coaches/" + args[0])
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("教练信息:")
	prettyPrint(resp.Data)
	return nil
}

func handleCoachPractice(args []string) error {
	if len(args) < 5 {
		return errors.New("用法: coach practice <学员ID> <教练ID> <开始时间> <结束时间> <科目>")
	}

	start, err := time.Parse(time.RFC3339, args[2])
	if err != nil {
		start, err = time.Parse("2006-01-02 15:04", args[2])
		if err != nil {
			return errors.New("开始时间格式错误，使用 RFC3339 或 2006-01-02 15:04")
		}
	}

	end, err := time.Parse(time.RFC3339, args[3])
	if err != nil {
		end, err = time.Parse("2006-01-02 15:04", args[3])
		if err != nil {
			return errors.New("结束时间格式错误，使用 RFC3339 或 2006-01-02 15:04")
		}
	}

	subject, err := strconv.Atoi(args[4])
	if err != nil {
		return errors.New("科目必须是数字")
	}

	req := common.BookPracticeRequest{
		StudentID: args[0],
		CoachID:   args[1],
		TimeSlot: common.TimeSlot{
			Start: start,
			End:   end,
		},
		Subject: common.Subject(subject),
	}

	resp, err := makePostRequest("/api/coaches/practice", req)
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New(resp.Message)
	}

	fmt.Println("练车预约成功:")
	prettyPrint(resp.Data)
	return nil
}

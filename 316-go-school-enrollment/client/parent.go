package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"school-enrollment/common"
	"strings"
)

func runParentMode(client *APIClient) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== 家长入口 ===")
		fmt.Println("1. 查看招生计划")
		fmt.Println("2. 提交报名")
		fmt.Println("3. 查看报名结果")
		fmt.Println("0. 返回上一级")
		fmt.Print("\n请选择: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			listPlans(client)
		case "2":
			submitRegistration(client, reader)
		case "3":
			queryResult(client, reader)
		case "0":
			return
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}

func listPlans(client *APIClient) {
	resp, err := client.ListPlans()
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取失败: %s\n", resp.Message)
		return
	}

	var plans []common.EnrollmentPlan
	plansBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(plansBytes, &plans)

	fmt.Println("\n=== 招生计划列表 ===")
	for i, p := range plans {
		status := "开放中"
		if !p.IsOpen {
			status = "已关闭"
		}
		fmt.Printf("\n[%d] 计划ID: %s\n", i+1, p.ID)
		fmt.Printf("    名称: %s\n", p.Name)
		fmt.Printf("    年级: %s\n", p.Grade)
		fmt.Printf("    班级数: %d, 每班人数: %d, 总容量: %d\n",
			p.Classes, p.MaxPerClass, p.TotalCapacity)
		fmt.Printf("    已报名: %d\n", p.EnrolledCount)
		fmt.Printf("    状态: %s\n", status)
	}
}

func submitRegistration(client *APIClient, reader *bufio.Reader) {
	listPlans(client)

	fmt.Print("\n请输入要报名的计划ID: ")
	planID, _ := reader.ReadString('\n')
	planID = strings.TrimSpace(planID)

	if planID == "" {
		fmt.Println("计划ID不能为空")
		return
	}

	fmt.Print("学生姓名: ")
	studentName, _ := reader.ReadString('\n')
	studentName = strings.TrimSpace(studentName)

	fmt.Print("身份证号 (18位): ")
	idCard, _ := reader.ReadString('\n')
	idCard = strings.TrimSpace(idCard)

	fmt.Print("家长联系电话 (11位手机号): ")
	parentPhone, _ := reader.ReadString('\n')
	parentPhone = strings.TrimSpace(parentPhone)

	fmt.Print("户籍地址: ")
	address, _ := reader.ReadString('\n')
	address = strings.TrimSpace(address)

	fmt.Printf("\n确认信息:\n")
	fmt.Printf("  计划ID: %s\n", planID)
	fmt.Printf("  学生姓名: %s\n", studentName)
	fmt.Printf("  身份证号: %s\n", idCard)
	fmt.Printf("  联系电话: %s\n", parentPhone)
	fmt.Printf("  户籍地址: %s\n", address)
	fmt.Print("\n确认提交? (y/n): ")

	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("已取消")
		return
	}

	resp, err := client.SubmitRegistration(planID, studentName, idCard, parentPhone, address)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("\n报名成功!\n")
		var reg common.Registration
		regBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(regBytes, &reg)
		fmt.Printf("报名ID: %s\n", reg.ID)
		fmt.Printf("当前状态: %s\n", reg.Status)
	} else {
		fmt.Printf("\n报名失败: %s\n", resp.Message)
	}
}

func queryResult(client *APIClient, reader *bufio.Reader) {
	listPlans(client)

	fmt.Print("\n请输入计划ID: ")
	planID, _ := reader.ReadString('\n')
	planID = strings.TrimSpace(planID)

	fmt.Print("请输入身份证号: ")
	idCard, _ := reader.ReadString('\n')
	idCard = strings.TrimSpace(idCard)

	resp, err := client.QueryRegistration(planID, idCard)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("查询失败: %s\n", resp.Message)
		return
	}

	var reg common.Registration
	regBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(regBytes, &reg)

	fmt.Println("\n=== 报名结果 ===")
	fmt.Printf("报名ID: %s\n", reg.ID)
	fmt.Printf("计划ID: %s\n", reg.PlanID)
	fmt.Printf("学生姓名: %s\n", reg.StudentName)
	fmt.Printf("身份证号: %s\n", reg.IDCard)
	fmt.Printf("联系电话: %s\n", reg.ParentPhone)
	fmt.Printf("户籍地址: %s\n", reg.Address)
	fmt.Printf("当前状态: %s\n", reg.Status)
	if reg.RejectReason != "" {
		fmt.Printf("拒绝原因: %s\n", reg.RejectReason)
	}
	fmt.Printf("提交时间: %s\n", reg.SubmittedAt.Format("2006-01-02 15:04:05"))
	if !reg.ReviewedAt.IsZero() {
		fmt.Printf("审核时间: %s\n", reg.ReviewedAt.Format("2006-01-02 15:04:05"))
	}
}

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"school-enrollment/common"
	"strconv"
	"strings"
)

func runAdminMode(client *APIClient) {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== 管理员入口 ===")
		fmt.Println("1. 发布招生计划")
		fmt.Println("2. 查看招生计划列表")
		fmt.Println("3. 修改招生计划容量")
		fmt.Println("4. 关闭招生计划")
		fmt.Println("5. 查看报名列表")
		fmt.Println("6. 审核报名（录取/拒绝）")
		fmt.Println("0. 返回上一级")
		fmt.Print("\n请选择: ")

		input, _ := reader.ReadString('\n')
		choice := strings.TrimSpace(input)

		switch choice {
		case "1":
			createPlan(client, reader)
		case "2":
			listPlans(client)
		case "3":
			updatePlan(client, reader)
		case "4":
			closePlan(client, reader)
		case "5":
			listRegistrations(client, reader)
		case "6":
			reviewRegistration(client, reader)
		case "0":
			return
		default:
			fmt.Println("无效选择，请重试")
		}
	}
}

func createPlan(client *APIClient, reader *bufio.Reader) {
	fmt.Print("招生计划名称 (最多50字): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("年级: ")
	grade, _ := reader.ReadString('\n')
	grade = strings.TrimSpace(grade)

	fmt.Print("班级数量: ")
	classesStr, _ := reader.ReadString('\n')
	classesStr = strings.TrimSpace(classesStr)
	classes, _ := strconv.Atoi(classesStr)

	fmt.Print("每班最多人数: ")
	maxPerClassStr, _ := reader.ReadString('\n')
	maxPerClassStr = strings.TrimSpace(maxPerClassStr)
	maxPerClass, _ := strconv.Atoi(maxPerClassStr)

	fmt.Printf("\n确认信息:\n")
	fmt.Printf("  名称: %s\n", name)
	fmt.Printf("  年级: %s\n", grade)
	fmt.Printf("  班级数: %d\n", classes)
	fmt.Printf("  每班人数: %d\n", maxPerClass)
	fmt.Printf("  总容量: %d\n", classes*maxPerClass)
	fmt.Print("\n确认创建? (y/n): ")

	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("已取消")
		return
	}

	resp, err := client.CreatePlan(name, grade, classes, maxPerClass)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("\n创建成功!\n")
		var plan common.EnrollmentPlan
		planBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(planBytes, &plan)
		fmt.Printf("计划ID: %s\n", plan.ID)
	} else {
		fmt.Printf("\n创建失败: %s\n", resp.Message)
	}
}

func updatePlan(client *APIClient, reader *bufio.Reader) {
	listPlans(client)

	fmt.Print("\n请输入要修改的计划ID: ")
	planID, _ := reader.ReadString('\n')
	planID = strings.TrimSpace(planID)

	if planID == "" {
		fmt.Println("计划ID不能为空")
		return
	}

	fmt.Print("新的班级数量 (留空保持不变): ")
	classesStr, _ := reader.ReadString('\n')
	classesStr = strings.TrimSpace(classesStr)
	classes := 0
	if classesStr != "" {
		classes, _ = strconv.Atoi(classesStr)
	}

	fmt.Print("新的每班最多人数 (留空保持不变): ")
	maxPerClassStr, _ := reader.ReadString('\n')
	maxPerClassStr = strings.TrimSpace(maxPerClassStr)
	maxPerClass := 0
	if maxPerClassStr != "" {
		maxPerClass, _ = strconv.Atoi(maxPerClassStr)
	}

	resp, err := client.UpdatePlan(planID, classes, maxPerClass)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("\n更新成功!\n")
		var plan common.EnrollmentPlan
		planBytes, _ := json.Marshal(resp.Data)
		json.Unmarshal(planBytes, &plan)
		fmt.Printf("班级数: %d, 每班人数: %d, 总容量: %d\n",
			plan.Classes, plan.MaxPerClass, plan.TotalCapacity)
	} else {
		fmt.Printf("\n更新失败: %s\n", resp.Message)
	}
}

func closePlan(client *APIClient, reader *bufio.Reader) {
	listPlans(client)

	fmt.Print("\n请输入要关闭的计划ID: ")
	planID, _ := reader.ReadString('\n')
	planID = strings.TrimSpace(planID)

	if planID == "" {
		fmt.Println("计划ID不能为空")
		return
	}

	fmt.Print("确认关闭该计划? 关闭后将无法接受新报名 (y/n): ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(strings.ToLower(confirm))

	if confirm != "y" && confirm != "yes" {
		fmt.Println("已取消")
		return
	}

	resp, err := client.ClosePlan(planID)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("\n计划已关闭!\n")
	} else {
		fmt.Printf("\n操作失败: %s\n", resp.Message)
	}
}

func listRegistrations(client *APIClient, reader *bufio.Reader) {
	listPlans(client)

	fmt.Print("\n请输入计划ID (查看所有报名): ")
	planID, _ := reader.ReadString('\n')
	planID = strings.TrimSpace(planID)

	if planID == "" {
		fmt.Println("计划ID不能为空")
		return
	}

	fmt.Print("状态筛选 (待审核/已录取/已拒绝, 留空显示全部): ")
	status, _ := reader.ReadString('\n')
	status = strings.TrimSpace(status)

	resp, err := client.ListRegistrations(planID, status)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取失败: %s\n", resp.Message)
		return
	}

	var regs []common.Registration
	regsBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(regsBytes, &regs)

	fmt.Printf("\n=== 报名列表 (共%d条) ===\n", len(regs))
	for i, reg := range regs {
		fmt.Printf("\n[%d] 报名ID: %s\n", i+1, reg.ID)
		fmt.Printf("    学生姓名: %s\n", reg.StudentName)
		fmt.Printf("    身份证号: %s\n", reg.IDCard)
		fmt.Printf("    联系电话: %s\n", reg.ParentPhone)
		fmt.Printf("    户籍地址: %s\n", reg.Address)
		fmt.Printf("    当前状态: %s\n", reg.Status)
		if reg.RejectReason != "" {
			fmt.Printf("    拒绝原因: %s\n", reg.RejectReason)
		}
		fmt.Printf("    提交时间: %s\n", reg.SubmittedAt.Format("2006-01-02 15:04:05"))
	}
}

func reviewRegistration(client *APIClient, reader *bufio.Reader) {
	listPlans(client)

	fmt.Print("\n请输入计划ID: ")
	planID, _ := reader.ReadString('\n')
	planID = strings.TrimSpace(planID)

	if planID == "" {
		fmt.Println("计划ID不能为空")
		return
	}

	resp, err := client.ListRegistrations(planID, common.StatusPending)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取失败: %s\n", resp.Message)
		return
	}

	var regs []common.Registration
	regsBytes, _ := json.Marshal(resp.Data)
	json.Unmarshal(regsBytes, &regs)

	if len(regs) == 0 {
		fmt.Println("\n没有待审核的报名")
		return
	}

	fmt.Printf("\n=== 待审核报名列表 (共%d条) ===\n", len(regs))
	for i, reg := range regs {
		fmt.Printf("\n[%d] 报名ID: %s\n", i+1, reg.ID)
		fmt.Printf("    学生姓名: %s\n", reg.StudentName)
		fmt.Printf("    身份证号: %s\n", reg.IDCard)
		fmt.Printf("    联系电话: %s\n", reg.ParentPhone)
		fmt.Printf("    提交时间: %s\n", reg.SubmittedAt.Format("2006-01-02 15:04:05"))
	}

	fmt.Print("\n请输入要审核的报名ID: ")
	regID, _ := reader.ReadString('\n')
	regID = strings.TrimSpace(regID)

	if regID == "" {
		fmt.Println("报名ID不能为空")
		return
	}

	fmt.Println("\n请选择操作:")
	fmt.Println("1. 录取")
	fmt.Println("2. 拒绝")
	fmt.Print("选择: ")
	actionChoice, _ := reader.ReadString('\n')
	actionChoice = strings.TrimSpace(actionChoice)

	var action string
	var reason string

	switch actionChoice {
	case "1":
		action = "accept"
	case "2":
		action = "reject"
		fmt.Print("请输入拒绝原因: ")
		reason, _ = reader.ReadString('\n')
		reason = strings.TrimSpace(reason)
		if reason == "" {
			fmt.Println("拒绝原因不能为空")
			return
		}
	default:
		fmt.Println("无效选择")
		return
	}

	resp2, err := client.ReviewRegistration(regID, action, reason)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if resp2.Success {
		if action == "accept" {
			fmt.Printf("\n已录取!\n")
		} else {
			fmt.Printf("\n已拒绝!\n")
		}
	} else {
		fmt.Printf("\n操作失败: %s\n", resp2.Message)
	}
}

package main

import (
	"announcement-board/common"
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

var scanner = bufio.NewScanner(os.Stdin)

func readLine(prompt string) string {
	fmt.Print(prompt)
	scanner.Scan()
	return strings.TrimSpace(scanner.Text())
}

func readLineWithDefault(prompt, defaultValue string) string {
	result := readLine(fmt.Sprintf("%s [%s]: ", prompt, defaultValue))
	if result == "" {
		return defaultValue
	}
	return result
}

func printMenu(title string, options []string) {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println(title)
	fmt.Println(strings.Repeat("=", 50))
	for i, opt := range options {
		fmt.Printf("%d. %s\n", i+1, opt)
	}
	fmt.Println("0. 退出")
	fmt.Print("请选择: ")
}

func loginUser() {
	fmt.Println("\n请选择用户身份:")
	fmt.Println("1. 管理员 (user_admin)")
	fmt.Println("2. 部门经理 (user_manager_1 - 技术部)")
	fmt.Println("3. 员工 (user_emp_1 - 技术部)")
	fmt.Println("4. 员工 (user_emp_2 - 产品部)")
	fmt.Println("5. 自定义输入")

	choice := readLine("请选择: ")
	var userID string

	switch choice {
	case "1":
		userID = "user_admin"
	case "2":
		userID = "user_manager_1"
	case "3":
		userID = "user_emp_1"
	case "4":
		userID = "user_emp_2"
	case "5":
		userID = readLine("请输入用户ID: ")
	default:
		userID = "user_emp_1"
	}

	SetCurrentUser(userID)

	user, err := GetUserInfo()
	if err != nil {
		fmt.Printf("获取用户信息失败: %v\n", err)
		return
	}

	fmt.Printf("\n已登录用户: %s\n", user.Name)
	fmt.Printf("角色: %s\n", user.Role)
	fmt.Printf("部门: %s\n", user.DeptName)
}

func printAnnouncementList(anns []common.Announcement, isAdmin bool) {
	if len(anns) == 0 {
		fmt.Println("暂无公告")
		return
	}

	for i, ann := range anns {
		pinMark := ""
		if ann.IsPinned {
			pinMark = "[置顶]"
		}
		fmt.Printf("\n[%d] %s %s\n", i+1, pinMark, ann.Title)
		fmt.Printf("    ID: %s\n", ann.ID)
		fmt.Printf("    状态: %s\n", ann.Status)
		fmt.Printf("    优先级: %s\n", ann.Priority)
		fmt.Printf("    发布范围: %s\n", ann.Scope)
		if ann.Scope == common.ScopeDept {
			fmt.Printf("    目标部门: %v\n", ann.TargetDeptIDs)
		}
		fmt.Printf("    生效时间: %s 至 %s\n",
			ann.EffectiveStart.Format("2006-01-02 15:04"),
			ann.EffectiveEnd.Format("2006-01-02 15:04"))
		fmt.Printf("    浏览量: %d\n", ann.ViewCount)
		if len(ann.Attachments) > 0 {
			fmt.Printf("    附件数: %d\n", len(ann.Attachments))
		}
	}
}

func printAnnouncementDetail(detail *common.AnnouncementDetailResponse, isAdmin bool) {
	ann := detail.Announcement
	stats := detail.ViewStats

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("标题: %s\n", ann.Title)
	if ann.IsPinned {
		fmt.Println("[置顶公告]")
	}
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("状态: %s\n", ann.Status)
	fmt.Printf("优先级: %s\n", ann.Priority)
	fmt.Printf("发布范围: %s\n", ann.Scope)
	if ann.Scope == common.ScopeDept {
		fmt.Printf("目标部门: %v\n", ann.TargetDeptIDs)
	}
	fmt.Printf("生效时间: %s 至 %s\n",
		ann.EffectiveStart.Format("2006-01-02 15:04"),
		ann.EffectiveEnd.Format("2006-01-02 15:04"))
	fmt.Printf("创建时间: %s\n", ann.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("正文内容:")
	fmt.Println(ann.Content)
	fmt.Println(strings.Repeat("-", 60))

	if len(ann.Attachments) > 0 {
		fmt.Println("附件列表:")
		for i, att := range ann.Attachments {
			fmt.Printf("  [%d] %s (%.2f KB)\n", i+1, att.FileName, float64(att.FileSize)/1024)
		}
	}

	if isAdmin {
		fmt.Println("\n--- 阅读统计 ---")
		fmt.Printf("浏览量: %d\n", stats.ViewCount)
		fmt.Printf("目标受众人数: %d\n", stats.TotalAudience)
		fmt.Printf("阅读率: %.2f%%\n", stats.ReadRate)
		if len(stats.ReadUserIDs) > 0 {
			fmt.Printf("已阅读用户ID: %v\n", stats.ReadUserIDs)
		}
	}

	if isAdmin && len(ann.ChangeHistory) > 0 {
		fmt.Println("\n--- 修改历史 ---")
		for i, ch := range ann.ChangeHistory {
			fmt.Printf("[%d] 修改时间: %s, 修改人ID: %s\n", i+1,
				ch.ModifiedAt.Format("2006-01-02 15:04:05"), ch.ModifiedBy)
		}
	}
}

func createAnnouncementCLI() {
	fmt.Println("\n=== 创建新公告 ===")

	title := readLine("公告标题: ")
	if title == "" {
		fmt.Println("标题不能为空")
		return
	}

	fmt.Println("请输入公告正文 (输入 . 结束):")
	var contentLines []string
	for {
		scanner.Scan()
		line := scanner.Text()
		if line == "." {
			break
		}
		contentLines = append(contentLines, line)
	}
	content := strings.Join(contentLines, "\n")
	if content == "" {
		fmt.Println("内容不能为空")
		return
	}

	fmt.Println("\n发布范围:")
	fmt.Println("1. 全体员工")
	fmt.Println("2. 指定部门")
	scopeChoice := readLineWithDefault("请选择", "1")
	scope := common.ScopeAll
	var targetDepts []string

	if scopeChoice == "2" {
		scope = common.ScopeDept
		depts, err := ListDepartments()
		if err != nil {
			fmt.Printf("获取部门列表失败: %v\n", err)
			return
		}
		fmt.Println("\n可用部门:")
		for i, d := range depts {
			fmt.Printf("%d. %s (%s)\n", i+1, d.Name, d.ID)
		}
		deptIDs := readLine("请选择目标部门ID (多个用逗号分隔): ")
		for _, id := range strings.Split(deptIDs, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				targetDepts = append(targetDepts, id)
			}
		}
		if len(targetDepts) == 0 {
			fmt.Println("未选择有效部门")
			return
		}
	}

	isPinned := false
	pinChoice := readLineWithDefault("是否置顶? (y/n)", "n")
	if strings.ToLower(pinChoice) == "y" || strings.ToLower(pinChoice) == "yes" {
		isPinned = true
	}

	now := time.Now()
	defaultStart := now.Format("2006-01-02 15:04")
	defaultEnd := now.Add(7 * 24 * time.Hour).Format("2006-01-02 15:04")

	startStr := readLineWithDefault("生效开始时间 (格式: 2006-01-02 15:04)", defaultStart)
	endStr := readLineWithDefault("生效结束时间 (格式: 2006-01-02 15:04)", defaultEnd)

	startTime, err := time.ParseInLocation("2006-01-02 15:04", startStr, time.Local)
	if err != nil {
		fmt.Printf("时间格式错误: %v\n", err)
		return
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04", endStr, time.Local)
	if err != nil {
		fmt.Printf("时间格式错误: %v\n", err)
		return
	}

	fmt.Println("\n优先级:")
	fmt.Println("1. 普通")
	fmt.Println("2. 紧急 (立即发布)")
	priorityChoice := readLineWithDefault("请选择", "1")
	priority := common.PriorityNormal
	if priorityChoice == "2" {
		priority = common.PriorityUrgent
	}

	req := &common.CreateAnnouncementRequest{
		Title:          title,
		Content:        content,
		Scope:          scope,
		TargetDeptIDs:  targetDepts,
		IsPinned:       isPinned,
		EffectiveStart: startTime,
		EffectiveEnd:   endTime,
		Priority:       priority,
	}

	ann, err := CreateAnnouncement(req)
	if err != nil {
		fmt.Printf("创建公告失败: %v\n", err)
		return
	}

	fmt.Printf("\n公告创建成功! ID: %s, 状态: %s\n", ann.ID, ann.Status)
}

func viewAnnouncementDetail() {
	annID := readLine("请输入公告ID: ")
	if annID == "" {
		return
	}

	detail, err := GetAnnouncementDetail(annID)
	if err != nil {
		fmt.Printf("获取公告详情失败: %v\n", err)
		return
	}

	user, _ := GetUserInfo()
	isAdmin := user != nil && user.Role == common.RoleAdmin
	printAnnouncementDetail(detail, isAdmin)
}

func updateAnnouncementCLI() {
	annID := readLine("请输入要修改的公告ID: ")
	if annID == "" {
		return
	}

	fmt.Println("请输入新的公告正文 (输入 . 结束):")
	var contentLines []string
	for {
		scanner.Scan()
		line := scanner.Text()
		if line == "." {
			break
		}
		contentLines = append(contentLines, line)
	}
	content := strings.Join(contentLines, "\n")

	if content == "" {
		fmt.Println("内容不能为空")
		return
	}

	ann, err := UpdateAnnouncement(annID, content)
	if err != nil {
		fmt.Printf("修改公告失败: %v\n", err)
		return
	}

	fmt.Printf("公告修改成功! ID: %s\n", ann.ID)
}

func deleteDraftCLI() {
	annID := readLine("请输入要删除的草稿公告ID: ")
	if annID == "" {
		return
	}

	confirm := readLineWithDefault("确认删除? (y/n)", "n")
	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("已取消删除")
		return
	}

	if err := DeleteDraft(annID); err != nil {
		fmt.Printf("删除失败: %v\n", err)
		return
	}
	fmt.Println("删除成功")
}

func togglePinCLI() {
	annID := readLine("请输入要切换置顶状态的公告ID: ")
	if annID == "" {
		return
	}

	if err := TogglePin(annID); err != nil {
		fmt.Printf("操作失败: %v\n", err)
		return
	}
	fmt.Println("置顶状态已更新")
}

func listPendingApprovalsCLI() {
	resp, err := ListPendingApprovals()
	if err != nil {
		fmt.Printf("获取待审批列表失败: %v\n", err)
		return
	}

	if resp.Total == 0 {
		fmt.Println("暂无待审批的公告")
		return
	}

	fmt.Printf("\n共 %d 条待审批公告:\n", resp.Total)
	for i, ar := range resp.Approvals {
		fmt.Printf("\n[%d] 审批ID: %s\n", i+1, ar.ID)
		fmt.Printf("    公告ID: %s\n", ar.AnnouncementID)
		fmt.Printf("    提交人ID: %s\n", ar.RequesterID)
		fmt.Printf("    提交时间: %s\n", ar.RequestAt.Format("2006-01-02 15:04:05"))
	}

	annID := readLine("\n请输入要审批的公告ID (直接回车返回): ")
	if annID == "" {
		return
	}

	action := readLine("操作 (approve=通过, reject=驳回): ")
	if action != "approve" && action != "reject" {
		fmt.Println("无效操作")
		return
	}

	var comment string
	if action == "reject" {
		comment = readLine("驳回原因: ")
	} else {
		comment = readLineWithDefault("审批意见 (可选)", "")
	}

	var err2 error
	if action == "approve" {
		err2 = ApproveAnnouncement(annID, comment)
	} else {
		err2 = RejectAnnouncement(annID, comment)
	}

	if err2 != nil {
		fmt.Printf("审批操作失败: %v\n", err2)
		return
	}
	fmt.Println("审批操作成功")
}

func submitApprovalCLI() {
	annID := readLine("请输入要提交审批的草稿公告ID: ")
	if annID == "" {
		return
	}

	if err := SubmitApproval(annID); err != nil {
		fmt.Printf("提交审批失败: %v\n", err)
		return
	}
	fmt.Println("已提交审批")
}

func searchAnnouncementsCLI(keyword string) {
	user, _ := GetUserInfo()
	isAdmin := user != nil && user.Role == common.RoleAdmin

	var resp *common.AnnouncementListResponse
	var err error

	if isAdmin {
		resp, err = ListAdminAnnouncements(keyword)
	} else {
		resp, err = ListAnnouncements(keyword)
	}

	if err != nil {
		fmt.Printf("搜索失败: %v\n", err)
		return
	}

	if keyword != "" {
		fmt.Printf("\n搜索关键词: \"%s\"\n", keyword)
	}
	fmt.Printf("共找到 %d 条公告\n", resp.Total)
	printAnnouncementList(resp.Announcements, isAdmin)
}

func runEmployeeMenu() {
	for {
		fmt.Println("")
		printMenu("员工功能菜单", []string{
			"查看我的公告列表",
			"搜索公告",
			"查看公告详情",
			"切换用户登录",
		})

		choice := readLine("")
		switch choice {
		case "1":
			searchAnnouncementsCLI("")
		case "2":
			keyword := readLine("请输入搜索关键词: ")
			searchAnnouncementsCLI(keyword)
		case "3":
			viewAnnouncementDetail()
		case "4":
			loginUser()
		case "0":
			fmt.Println("再见!")
			os.Exit(0)
		default:
			fmt.Println("无效选择")
		}
	}
}

func runManagerMenu() {
	for {
		fmt.Println("")
		printMenu("部门经理功能菜单", []string{
			"查看我的公告列表",
			"搜索公告",
			"查看公告详情",
			"查看待审批列表",
			"切换用户登录",
		})

		choice := readLine("")
		switch choice {
		case "1":
			searchAnnouncementsCLI("")
		case "2":
			keyword := readLine("请输入搜索关键词: ")
			searchAnnouncementsCLI(keyword)
		case "3":
			viewAnnouncementDetail()
		case "4":
			listPendingApprovalsCLI()
		case "5":
			loginUser()
		case "0":
			fmt.Println("再见!")
			os.Exit(0)
		default:
			fmt.Println("无效选择")
		}
	}
}

func runAdminMenu() {
	for {
		fmt.Println("")
		printMenu("管理员功能菜单", []string{
			"创建新公告",
			"查看所有公告列表",
			"搜索公告",
			"查看公告详情",
			"修改公告正文",
			"删除草稿公告",
			"切换置顶状态",
			"提交公告审批",
			"切换用户登录",
		})

		choice := readLine("")
		switch choice {
		case "1":
			createAnnouncementCLI()
		case "2":
			searchAnnouncementsCLI("")
		case "3":
			keyword := readLine("请输入搜索关键词: ")
			searchAnnouncementsCLI(keyword)
		case "4":
			viewAnnouncementDetail()
		case "5":
			updateAnnouncementCLI()
		case "6":
			deleteDraftCLI()
		case "7":
			togglePinCLI()
		case "8":
			submitApprovalCLI()
		case "9":
			loginUser()
		case "0":
			fmt.Println("再见!")
			os.Exit(0)
		default:
			fmt.Println("无效选择")
		}
	}
}

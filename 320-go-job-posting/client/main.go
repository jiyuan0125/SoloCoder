package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"jobposting/client/api"
	"jobposting/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := api.NewClient()
	command := os.Args[1]

	switch command {
	case "company":
		if len(os.Args) < 3 {
			printCompanyUsage()
			os.Exit(1)
		}
		handleCompanyCommand(client, os.Args[2:])
	case "jobseeker":
		if len(os.Args) < 3 {
			printJobseekerUsage()
			os.Exit(1)
		}
		handleJobseekerCommand(client, os.Args[2:])
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("用法:")
	fmt.Println("  client company <子命令>  - 企业端操作")
	fmt.Println("  client jobseeker <子命令> - 求职者端操作")
	fmt.Println("")
	fmt.Println("企业端子命令:")
	fmt.Println("  post_job   - 发布新职位")
	fmt.Println("  list_jobs  - 查看所有职位（包括已下架）")
	fmt.Println("  offline    - 下架职位")
	fmt.Println("  list_apps  - 查看投递简历列表")
	fmt.Println("  update_status - 更新简历状态")
	fmt.Println("")
	fmt.Println("求职者端子命令:")
	fmt.Println("  list_jobs  - 浏览职位列表")
	fmt.Println("  apply      - 投递简历")
	fmt.Println("  my_apps    - 查看我的投递记录")
}

func printCompanyUsage() {
	fmt.Println("企业端用法:")
	fmt.Println("  client company post_job")
	fmt.Println("  client company list_jobs")
	fmt.Println("  client company offline <职位ID>")
	fmt.Println("  client company list_apps [职位ID]")
	fmt.Println("  client company update_status <投递ID> <状态> [面试时间] [面试地点]")
	fmt.Println("")
	fmt.Println("状态选项: 查看中, 面试邀约, 不合适")
}

func printJobseekerUsage() {
	fmt.Println("求职者端用法:")
	fmt.Println("  client jobseeker list_jobs [最低薪资] [城市] [学历]")
	fmt.Println("  client jobseeker apply <职位ID>")
	fmt.Println("  client jobseeker my_apps <手机号>")
	fmt.Println("")
	fmt.Println("学历选项: 不限, 大专, 本科, 硕士, 博士")
}

func handleCompanyCommand(client *api.Client, args []string) {
	subCmd := args[0]

	switch subCmd {
	case "post_job":
		handlePostJob(client)
	case "list_jobs":
		handleListCompanyJobs(client)
	case "offline":
		if len(args) < 2 {
			fmt.Println("错误: 缺少职位ID")
			printCompanyUsage()
			os.Exit(1)
		}
		handleOfflineJob(client, args[1])
	case "list_apps":
		var jobID *string
		if len(args) >= 2 && args[1] != "" {
			jobID = &args[1]
		}
		handleListApplications(client, jobID)
	case "update_status":
		if len(args) < 3 {
			fmt.Println("错误: 缺少参数")
			printCompanyUsage()
			os.Exit(1)
		}
		appID := args[1]
		status := common.ApplicationStatus(args[2])
		var interviewTime, interviewLocation *string
		if len(args) >= 4 {
			interviewTime = &args[3]
		}
		if len(args) >= 5 {
			interviewLocation = &args[4]
		}
		handleUpdateStatus(client, appID, status, interviewTime, interviewLocation)
	default:
		fmt.Printf("未知企业端子命令: %s\n", subCmd)
		printCompanyUsage()
		os.Exit(1)
	}
}

func handleJobseekerCommand(client *api.Client, args []string) {
	subCmd := args[0]

	switch subCmd {
	case "list_jobs":
		var minSalary *int
		var city *string
		var education *common.EducationLevel

		if len(args) >= 2 && args[1] != "" {
			if val, err := strconv.Atoi(args[1]); err == nil {
				minSalary = &val
			}
		}
		if len(args) >= 3 && args[2] != "" {
			city = &args[2]
		}
		if len(args) >= 4 && args[3] != "" {
			edu := common.EducationLevel(args[3])
			education = &edu
		}
		handleListJobseekerJobs(client, minSalary, city, education)
	case "apply":
		if len(args) < 2 {
			fmt.Println("错误: 缺少职位ID")
			printJobseekerUsage()
			os.Exit(1)
		}
		handleApplyJob(client, args[1])
	case "my_apps":
		if len(args) < 2 {
			fmt.Println("错误: 缺少手机号")
			printJobseekerUsage()
			os.Exit(1)
		}
		handleMyApplications(client, args[1])
	default:
		fmt.Printf("未知求职者端子命令: %s\n", subCmd)
		printJobseekerUsage()
		os.Exit(1)
	}
}

func readInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func handlePostJob(client *api.Client) {
	title := readInput("职位名称: ")
	minSalaryStr := readInput("最低薪资: ")
	maxSalaryStr := readInput("最高薪资: ")
	city := readInput("工作城市: ")
	educationStr := readInput("学历要求 (不限/大专/本科/硕士/博士): ")
	experience := readInput("经验要求: ")
	description := readInput("职位描述: ")

	minSalary, err := strconv.Atoi(minSalaryStr)
	if err != nil {
		fmt.Printf("错误: 最低薪资格式错误: %v\n", err)
		os.Exit(1)
	}

	maxSalary, err := strconv.Atoi(maxSalaryStr)
	if err != nil {
		fmt.Printf("错误: 最高薪资格式错误: %v\n", err)
		os.Exit(1)
	}

	education := common.EducationLevel(educationStr)
	if education == "" {
		education = common.EducationAny
	}

	req := common.CreateJobRequest{
		Title:       title,
		MinSalary:   minSalary,
		MaxSalary:   maxSalary,
		City:        city,
		Education:   education,
		Experience:  experience,
		Description: description,
	}

	jobID, err := client.CreateJob(req)
	if err != nil {
		fmt.Printf("发布职位失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("职位发布成功，职位ID: %s\n", jobID)
}

func handleListCompanyJobs(client *api.Client) {
	jobs, err := client.ListJobsForCompany()
	if err != nil {
		fmt.Printf("获取职位列表失败: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Println("暂无职位")
		return
	}

	fmt.Println("职位列表:")
	fmt.Println("------------------------------------------------------------")
	for _, job := range jobs {
		status := "上架"
		if job.IsOffline {
			status = "已下架"
		}
		fmt.Printf("ID: %s\n", job.ID)
		fmt.Printf("  职位: %s\n", job.Title)
		fmt.Printf("  薪资: %d-%d\n", job.MinSalary, job.MaxSalary)
		fmt.Printf("  城市: %s\n", job.City)
		fmt.Printf("  学历: %s\n", job.Education)
		fmt.Printf("  状态: %s\n", status)
		fmt.Println("------------------------------------------------------------")
	}
}

func handleOfflineJob(client *api.Client, jobID string) {
	err := client.OfflineJob(jobID)
	if err != nil {
		fmt.Printf("下架职位失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("职位已下架")
}

func handleListApplications(client *api.Client, jobID *string) {
	apps, err := client.ListApplications(jobID)
	if err != nil {
		fmt.Printf("获取投递列表失败: %v\n", err)
		os.Exit(1)
	}

	if len(apps) == 0 {
		fmt.Println("暂无投递记录")
		return
	}

	fmt.Println("投递记录列表:")
	fmt.Println("------------------------------------------------------------")
	for _, app := range apps {
		fmt.Printf("投递ID: %s\n", app.ID)
		fmt.Printf("  职位: %s\n", app.JobTitle)
		fmt.Printf("  姓名: %s\n", app.ResumeName)
		fmt.Printf("  手机: %s\n", app.ResumePhone)
		fmt.Printf("  状态: %s\n", app.Status)
		if app.Status == common.StatusInterview {
			if app.InterviewTime != nil {
				fmt.Printf("  面试时间: %s\n", *app.InterviewTime)
			}
			if app.InterviewLocation != nil {
				fmt.Printf("  面试地点: %s\n", *app.InterviewLocation)
			}
		}
		fmt.Printf("  简历摘要: %s\n", app.ResumeSummary)
		fmt.Println("------------------------------------------------------------")
	}
}

func handleUpdateStatus(client *api.Client, appID string, status common.ApplicationStatus, interviewTime, interviewLocation *string) {
	req := common.UpdateApplicationStatusRequest{
		Status:            status,
		InterviewTime:     interviewTime,
		InterviewLocation: interviewLocation,
	}

	err := client.UpdateApplicationStatus(appID, req)
	if err != nil {
		fmt.Printf("更新状态失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("状态更新成功")
}

func handleListJobseekerJobs(client *api.Client, minSalary *int, city *string, education *common.EducationLevel) {
	jobs, err := client.ListJobsForJobseeker(minSalary, city, education)
	if err != nil {
		fmt.Printf("获取职位列表失败: %v\n", err)
		os.Exit(1)
	}

	if len(jobs) == 0 {
		fmt.Println("暂无符合条件的职位")
		return
	}

	fmt.Println("职位列表:")
	fmt.Println("------------------------------------------------------------")
	for _, job := range jobs {
		fmt.Printf("ID: %s\n", job.ID)
		fmt.Printf("  职位: %s\n", job.Title)
		fmt.Printf("  薪资: %d-%d\n", job.MinSalary, job.MaxSalary)
		fmt.Printf("  城市: %s\n", job.City)
		fmt.Printf("  学历: %s\n", job.Education)
		fmt.Printf("  经验要求: %s\n", job.Experience)
		fmt.Printf("  职位描述: %s\n", job.Description)
		fmt.Println("------------------------------------------------------------")
	}
}

func handleApplyJob(client *api.Client, jobID string) {
	name := readInput("姓名: ")
	phone := readInput("手机号: ")
	summary := readInput("简历摘要: ")

	req := common.ApplyJobRequest{
		Name:    name,
		Phone:   phone,
		Summary: summary,
	}

	appID, err := client.ApplyJob(jobID, req)
	if err != nil {
		fmt.Printf("投递失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("投递成功，投递ID: %s\n", appID)
}

func handleMyApplications(client *api.Client, phone string) {
	apps, err := client.ListMyApplications(phone)
	if err != nil {
		fmt.Printf("获取投递记录失败: %v\n", err)
		os.Exit(1)
	}

	if len(apps) == 0 {
		fmt.Println("暂无投递记录")
		return
	}

	fmt.Println("我的投递记录:")
	fmt.Println("------------------------------------------------------------")
	for _, app := range apps {
		fmt.Printf("投递ID: %s\n", app.ID)
		fmt.Printf("  职位: %s\n", app.JobTitle)
		fmt.Printf("  状态: %s\n", app.Status)
		if app.Status == common.StatusInterview {
			if app.InterviewTime != nil {
				fmt.Printf("  面试时间: %s\n", *app.InterviewTime)
			}
			if app.InterviewLocation != nil {
				fmt.Printf("  面试地点: %s\n", *app.InterviewLocation)
			}
		}
		fmt.Println("------------------------------------------------------------")
	}
}

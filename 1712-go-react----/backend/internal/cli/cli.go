package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const APIBase = "http://localhost:8300/api"

type CLI struct {
	apiBase string
}

func New(apiBase string) *CLI {
	if apiBase == "" {
		apiBase = APIBase
	}
	return &CLI{apiBase: apiBase}
}

func (c *CLI) Run(args []string) error {
	if len(args) < 1 {
		c.printUsage()
		return nil
	}

	command := strings.ToLower(args[0])
	switch command {
	case "create-record":
		return c.createRecord(args[1:])
	case "add-checkup":
		return c.addCheckup(args[1:])
	case "add-followup":
		return c.addFollowup(args[1:])
	case "export":
		return c.export(args[1:])
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		c.printUsage()
		return fmt.Errorf("unknown command")
	}
}

func (c *CLI) printUsage() {
	fmt.Println("居民健康档案管理系统 CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  client <command> [options]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  create-record    创建居民健康档案")
	fmt.Println("  add-checkup      录入体检记录")
	fmt.Println("  add-followup     录入慢性病随访记录")
	fmt.Println("  export           导出数据")
	fmt.Println("")
	fmt.Println("Use 'client <command> --help' for more information about a command.")
}

func (c *CLI) createRecord(args []string) error {
	fs := flag.NewFlagSet("create-record", flag.ExitOnError)
	name := fs.String("name", "", "居民姓名")
	idCard := fs.String("idcard", "", "身份证号")
	phone := fs.String("phone", "", "联系电话")
	address := fs.String("address", "", "家庭住址")
	emergency := fs.String("emergency", "", "紧急联系人")
	doctor := fs.String("doctor", "", "责任医生")
	bloodType := fs.String("bloodtype", "", "血型 (A/B/AB/O)")
	community := fs.String("community", "", "所属社区")
	guardianName := fs.String("guardian-name", "", "监护人姓名（未成年人必填）")
	guardianPhone := fs.String("guardian-phone", "", "监护人电话（未成年人必填）")
	help := fs.Bool("help", false, "显示帮助信息")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help || *name == "" || *idCard == "" {
		fmt.Println("创建居民健康档案")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  client create-record [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		return nil
	}

	allergies := []map[string]string{}

	guardian := map[string]interface{}{}
	if *guardianName != "" || *guardianPhone != "" {
		guardian = map[string]interface{}{
			"name":  *guardianName,
			"phone": *guardianPhone,
		}
	}

	data := map[string]interface{}{
		"name":             *name,
		"idCard":           *idCard,
		"phone":            *phone,
		"address":          *address,
		"emergencyContact": *emergency,
		"doctorInCharge":   *doctor,
		"bloodType":        *bloodType,
		"community":        *community,
		"allergies":        allergies,
	}

	if len(guardian) > 0 {
		data["guardian"] = guardian
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(c.apiBase+"/residents", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("创建失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	fmt.Println("居民档案创建成功！")
	fmt.Println(string(body))
	return nil
}

func (c *CLI) addCheckup(args []string) error {
	fs := flag.NewFlagSet("add-checkup", flag.ExitOnError)
	residentID := fs.String("resident-id", "", "居民ID")
	date := fs.String("date", time.Now().Format("2006-01-02"), "体检日期 (YYYY-MM-DD)")
	height := fs.Float64("height", 0, "身高 (cm)")
	weight := fs.Float64("weight", 0, "体重 (kg)")
	systolic := fs.Int("systolic", 0, "收缩压 (mmHg)")
	diastolic := fs.Int("diastolic", 0, "舒张压 (mmHg)")
	heartRate := fs.Int("heart-rate", 0, "心率 (次/分)")
	visionLeft := fs.Float64("vision-left", 0, "左眼视力")
	visionRight := fs.Float64("vision-right", 0, "右眼视力")
	isInitial := fs.Bool("initial", false, "是否为初始体检")
	help := fs.Bool("help", false, "显示帮助信息")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help || *residentID == "" || *height <= 0 || *weight <= 0 {
		fmt.Println("录入体检记录")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  client add-checkup [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		return nil
	}

	checkupDate, err := time.Parse("2006-01-02", *date)
	if err != nil {
		return fmt.Errorf("日期格式错误: %v", err)
	}

	data := map[string]interface{}{
		"residentID":  *residentID,
		"date":        checkupDate.Format(time.RFC3339),
		"height":      *height,
		"weight":      *weight,
		"systolicBP":  *systolic,
		"diastolicBP": *diastolic,
		"heartRate":   *heartRate,
		"visionLeft":  *visionLeft,
		"visionRight": *visionRight,
		"isInitial":   *isInitial,
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(c.apiBase+"/checkups", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("录入失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	fmt.Println("体检记录录入成功！")
	fmt.Println(string(body))
	return nil
}

func (c *CLI) addFollowup(args []string) error {
	fs := flag.NewFlagSet("add-followup", flag.ExitOnError)
	residentID := fs.String("resident-id", "", "居民ID")
	disease := fs.String("disease", "", "慢性病类型 (hypertension/diabetes/coronary/stroke/copd)")
	date := fs.String("date", time.Now().Format("2006-01-02"), "随访日期 (YYYY-MM-DD)")

	systolic := fs.Int("systolic", 0, "高血压-收缩压 (mmHg)")
	diastolic := fs.Int("diastolic", 0, "高血压-舒张压 (mmHg)")
	medication := fs.String("medication", "", "用药情况")

	fasting := fs.Float64("fasting", 0, "糖尿病-空腹血糖 (mmol/L)")
	postprandial := fs.Float64("postprandial", 0, "糖尿病-餐后两小时血糖 (mmol/L)")
	hba1c := fs.Float64("hba1c", 0, "糖尿病-糖化血红蛋白 (%)")

	cardiacFunction := fs.String("cardiac-function", "", "冠心病-心功能评估")

	help := fs.Bool("help", false, "显示帮助信息")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help || *residentID == "" || *disease == "" {
		fmt.Println("录入慢性病随访记录")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  client add-followup [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		return nil
	}

	followupDate, err := time.Parse("2006-01-02", *date)
	if err != nil {
		return fmt.Errorf("日期格式错误: %v", err)
	}

	data := map[string]interface{}{
		"residentID": *residentID,
		"disease":    *disease,
		"date":       followupDate.Format(time.RFC3339),
	}

	switch *disease {
	case "hypertension":
		data["hypertension"] = map[string]interface{}{
			"systolicBP":  *systolic,
			"diastolicBP": *diastolic,
			"medication":  *medication,
		}
	case "diabetes":
		data["diabetes"] = map[string]interface{}{
			"fastingBloodSugar":    *fasting,
			"postprandialBloodSugar": *postprandial,
			"hba1c":                *hba1c,
			"medication":           *medication,
		}
	case "coronary":
		data["coronary"] = map[string]interface{}{
			"cardiacFunction": *cardiacFunction,
			"medication":      *medication,
		}
	}

	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(c.apiBase+"/followups", "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("录入失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	fmt.Println("随访记录录入成功！")
	fmt.Println(string(body))
	return nil
}

func (c *CLI) export(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	exportType := fs.String("type", "", "导出类型 (community/chronic/complete)")
	community := fs.String("community", "", "社区名称 (用于community类型)")
	residentID := fs.String("resident-id", "", "居民ID (用于complete类型)")
	output := fs.String("output", "", "输出文件路径")
	help := fs.Bool("help", false, "显示帮助信息")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *help || *exportType == "" {
		fmt.Println("导出数据")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  client export [options]")
		fmt.Println("")
		fmt.Println("Options:")
		fs.PrintDefaults()
		fmt.Println("")
		fmt.Println("Export Types:")
		fmt.Println("  community   导出社区建档统计")
		fmt.Println("  chronic     导出慢性病患者名单")
		fmt.Println("  complete    导出指定居民完整档案")
		return nil
	}

	var url string
	var defaultFile string

	switch *exportType {
	case "community":
		url = fmt.Sprintf("%s/export/community-stats?community=%s", c.apiBase, *community)
		defaultFile = "community_stats.csv"
	case "chronic":
		url = fmt.Sprintf("%s/export/chronic-patients", c.apiBase)
		defaultFile = "chronic_patients.csv"
	case "complete":
		if *residentID == "" {
			return fmt.Errorf("complete类型需要指定 --resident-id")
		}
		url = fmt.Sprintf("%s/export/complete-archive?residentID=%s", c.apiBase, *residentID)
		defaultFile = "complete_archive.csv"
	default:
		return fmt.Errorf("未知的导出类型: %s", *exportType)
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("导出失败 (HTTP %d): %s", resp.StatusCode, string(body))
	}

	outputFile := *output
	if outputFile == "" {
		outputFile = defaultFile
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("创建文件失败: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	fmt.Printf("数据已导出到: %s\n", outputFile)
	return nil
}

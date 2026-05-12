package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const baseURL = "http://localhost:8080/api"

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("请求失败: %s - %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

type DonorInput struct {
	DonorNo      string      `json:"donor_no"`
	Name         string      `json:"name"`
	Gender       string      `json:"gender"`
	Age          int         `json:"age"`
	BloodType    string      `json:"blood_type"`
	Height       float64     `json:"height"`
	Weight       float64     `json:"weight"`
	DeathDate    time.Time   `json:"death_date"`
	CauseOfDeath string      `json:"cause_of_death"`
	Organs       []OrganInput `json:"organs"`
	LocationID   string      `json:"location_id"`
}

type OrganInput struct {
	OrganType string `json:"organ_type"`
}

func (c *Client) RegisterDonor(args []string) error {
	if len(args) < 7 {
		return fmt.Errorf("用法: register-donor <捐献者编号> <姓名> <性别> <年龄> <血型> <身高> <体重> <死亡日期> <死亡原因> <器官类型1,器官类型2,...> [位置ID]")
	}

	organs := strings.Split(args[9], ",")
	organInputs := make([]OrganInput, len(organs))
	for i, o := range organs {
		organInputs[i] = OrganInput{OrganType: strings.TrimSpace(o)}
	}

	deathDate, _ := time.Parse("2006-01-02", args[7])

	locationID := ""
	if len(args) > 10 {
		locationID = args[10]
	}

	age, _ := time.ParseDuration(args[3])
	height, _ := time.ParseDuration(args[5])
	weight, _ := time.ParseDuration(args[6])

	donor := DonorInput{
		DonorNo:      args[0],
		Name:         args[1],
		Gender:       args[2],
		Age:          int(age.Hours()) / 8760,
		BloodType:    args[4],
		Height:       height.Hours(),
		Weight:       weight.Hours(),
		DeathDate:    deathDate,
		CauseOfDeath: args[8],
		Organs:       organInputs,
		LocationID:   locationID,
	}

	resp, err := c.doRequest("POST", "/donors", donor)
	if err != nil {
		return err
	}

	fmt.Println("捐献者登记成功:")
	fmt.Println(string(resp))
	return nil
}

type RecipientInput struct {
	RecipientNo      string    `json:"recipient_no"`
	Name             string    `json:"name"`
	BloodType        string    `json:"blood_type"`
	OrganNeeded      string    `json:"organ_needed"`
	RegistrationDate time.Time `json:"registration_date"`
	UrgencyLevel     string    `json:"urgency_level"`
	HLA              string    `json:"hla"`
	PRA              int       `json:"pra"`
	Age              int       `json:"age"`
	LocationID       string    `json:"location_id"`
}

func (c *Client) RegisterRecipient(args []string) error {
	if len(args) < 8 {
		return fmt.Errorf("用法: register-recipient <受体编号> <姓名> <血型> <需要器官类型> <登记日期> <紧急程度> <HLA> <PRA> <年龄> [位置ID]")
	}

	regDate, _ := time.Parse("2006-01-02", args[4])
	var pra, age int
	fmt.Sscanf(args[7], "%d", &pra)
	fmt.Sscanf(args[8], "%d", &age)

	locationID := ""
	if len(args) > 9 {
		locationID = args[9]
	}

	recipient := RecipientInput{
		RecipientNo:      args[0],
		Name:             args[1],
		BloodType:        args[2],
		OrganNeeded:      args[3],
		RegistrationDate: regDate,
		UrgencyLevel:     args[5],
		HLA:              args[6],
		PRA:              pra,
		Age:              age,
		LocationID:       locationID,
	}

	resp, err := c.doRequest("POST", "/recipients", recipient)
	if err != nil {
		return err
	}

	fmt.Println("受体登记成功:")
	fmt.Println(string(resp))
	return nil
}

func (c *Client) StartMatching(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("用法: start-matching <器官ID>")
	}

	organID := args[0]
	resp, err := c.doRequest("POST", "/organs/"+organID+"/match", nil)
	if err != nil {
		return err
	}

	fmt.Println("匹配启动成功:")
	fmt.Println(string(resp))
	return nil
}

type PaginatedResponse struct {
	Data []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		RelatedID   string `json:"related_id"`
		RelatedType string `json:"related_type"`
		Assignee    string `json:"assignee"`
		Status      string `json:"status"`
		DueDate     string `json:"due_date"`
	} `json:"data"`
}

func (c *Client) ListTodos(args []string) error {
	resp, err := c.doRequest("GET", "/todos?page=1&size=50", nil)
	if err != nil {
		return err
	}

	var result PaginatedResponse
	json.Unmarshal(resp, &result)

	fmt.Println("待办事项列表:")
	fmt.Println("-" + strings.Repeat("-", 100))
	for i, todo := range result.Data {
		fmt.Printf("%d. 类型: %s\n   状态: %s\n   负责人: %s\n   截止日期: %s\n   关联ID: %s\n\n",
			i+1, todo.Type, todo.Status, todo.Assignee, todo.DueDate, todo.RelatedID)
	}

	return nil
}

func printUsage() {
	fmt.Println("器官捐献管理系统 CLI")
	fmt.Println()
	fmt.Println("可用命令:")
	fmt.Println("  register-donor    登记捐献者")
	fmt.Println("  register-recipient 登记受体")
	fmt.Println("  start-matching    触发匹配")
	fmt.Println("  list-todos        查看待办事项")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  client register-donor D001 张三 男 45 A 175 70 2024-01-15 脑死亡 heart,liver,cornea")
	fmt.Println("  client register-recipient R001 李四 A kidney 2024-01-01 紧急 A1,A2,B7,B8,DR1,DR3 50 35")
	fmt.Println("  client start-matching <organ-id>")
	fmt.Println("  client list-todos")
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(baseURL)
	command := os.Args[1]
	args := os.Args[2:]

	var err error
	switch command {
	case "register-donor":
		err = client.RegisterDonor(args)
	case "register-recipient":
		err = client.RegisterRecipient(args)
	case "start-matching":
		err = client.StartMatching(args)
	case "list-todos":
		err = client.ListTodos(args)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"housekeeping/internal/api"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	serverURL string
}

func NewClient(serverURL string) *Client {
	return &Client{serverURL: strings.TrimSuffix(serverURL, "/")}
}

func (c *Client) doRequest(method, path string, body interface{}) (*api.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.serverURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResp api.Response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	return &apiResp, nil
}

func (c *Client) registerAunt(req api.RegisterAuntRequest) error {
	resp, err := c.doRequest("POST", "/api/aunt/register", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var aunt api.Aunt
	json.Unmarshal(dataBytes, &aunt)

	fmt.Printf("阿姨入驻成功!\n")
	fmt.Printf("ID: %s\n", aunt.ID)
	fmt.Printf("姓名: %s\n", aunt.Name)
	fmt.Printf("服务类别: %s\n", aunt.ServiceCategory)
	fmt.Printf("综合评分: %.1f\n", aunt.Rating)
	return nil
}

func (c *Client) getAunt(id string) error {
	path := "/api/aunt"
	if id != "" {
		path += "?id=" + id
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)

	if id != "" {
		var aunt api.Aunt
		json.Unmarshal(dataBytes, &aunt)
		printAunt(&aunt)
	} else {
		var aunts []api.Aunt
		json.Unmarshal(dataBytes, &aunts)
		fmt.Printf("共找到 %d 位阿姨:\n", len(aunts))
		for i, aunt := range aunts {
			fmt.Printf("\n[%d]\n", i+1)
			printAunt(&aunt)
		}
	}
	return nil
}

func printAunt(aunt *api.Aunt) {
	fmt.Printf("ID: %s\n", aunt.ID)
	fmt.Printf("姓名: %s\n", aunt.Name)
	fmt.Printf("年龄: %d岁\n", aunt.Age)
	fmt.Printf("电话: %s\n", aunt.Phone)
	fmt.Printf("服务类别: %s\n", aunt.ServiceCategory)
	fmt.Printf("从业年限: %d年\n", aunt.YearsOfService)
	fmt.Printf("服务区域: %s\n", aunt.ServiceArea)
	fmt.Printf("综合评分: %.1f\n", aunt.Rating)
}

func (c *Client) registerCustomer(req api.RegisterCustomerRequest) error {
	resp, err := c.doRequest("POST", "/api/customer/register", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var customer api.Customer
	json.Unmarshal(dataBytes, &customer)

	fmt.Printf("客户注册成功!\n")
	fmt.Printf("ID: %s\n", customer.ID)
	fmt.Printf("姓名: %s\n", customer.Name)
	fmt.Printf("电话: %s\n", customer.Phone)
	return nil
}

func (c *Client) getCustomer(id string) error {
	path := "/api/customer"
	if id != "" {
		path += "?id=" + id
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)

	if id != "" {
		var customer api.Customer
		json.Unmarshal(dataBytes, &customer)
		printCustomer(&customer)
	} else {
		var customers []api.Customer
		json.Unmarshal(dataBytes, &customers)
		fmt.Printf("共找到 %d 位客户:\n", len(customers))
		for i, customer := range customers {
			fmt.Printf("\n[%d]\n", i+1)
			printCustomer(&customer)
		}
	}
	return nil
}

func printCustomer(customer *api.Customer) {
	fmt.Printf("ID: %s\n", customer.ID)
	fmt.Printf("姓名: %s\n", customer.Name)
	fmt.Printf("电话: %s\n", customer.Phone)
}

func (c *Client) createBooking(req api.CreateBookingRequest) error {
	resp, err := c.doRequest("POST", "/api/booking/create", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var booking api.Booking
	json.Unmarshal(dataBytes, &booking)

	fmt.Printf("订单创建成功!\n")
	printBooking(&booking)
	return nil
}

func (c *Client) getBooking(id, auntID, customerID string, pending bool) error {
	path := "/api/booking"
	params := []string{}
	if id != "" {
		params = append(params, "id="+id)
	}
	if auntID != "" {
		params = append(params, "aunt_id="+auntID)
	}
	if customerID != "" {
		params = append(params, "customer_id="+customerID)
	}
	if pending {
		params = append(params, "pending=true")
	}
	if len(params) > 0 {
		path += "?" + strings.Join(params, "&")
	}

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)

	if id != "" {
		var booking api.Booking
		json.Unmarshal(dataBytes, &booking)
		printBooking(&booking)
	} else {
		var bookings []api.Booking
		json.Unmarshal(dataBytes, &bookings)
		fmt.Printf("共找到 %d 个订单:\n", len(bookings))
		for i, booking := range bookings {
			fmt.Printf("\n[%d]\n", i+1)
			printBooking(&booking)
		}
	}
	return nil
}

func printBooking(booking *api.Booking) {
	fmt.Printf("订单ID: %s\n", booking.ID)
	fmt.Printf("客户ID: %s\n", booking.CustomerID)
	fmt.Printf("阿姨ID: %s\n", booking.AuntID)
	fmt.Printf("服务类别: %s\n", booking.ServiceCategory)
	fmt.Printf("服务日期: %s\n", booking.ServiceDate.Format("2006-01-02 15:04:05"))
	fmt.Printf("预估时长: %.1f小时\n", booking.EstimatedDuration)
	fmt.Printf("实际时长: %.1f小时\n", booking.ActualDuration)
	fmt.Printf("服务地址: %s\n", booking.Address)
	fmt.Printf("订单状态: %s\n", booking.Status)
	if booking.Price > 0 {
		fmt.Printf("订单金额: %d元\n", booking.Price)
	}
}

func (c *Client) startService(bookingID string) error {
	resp, err := c.doRequest("POST", "/api/service/start", api.StartServiceRequest{BookingID: bookingID})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var booking api.Booking
	json.Unmarshal(dataBytes, &booking)

	fmt.Printf("服务已开始!\n")
	printBooking(&booking)
	return nil
}

func (c *Client) completeService(bookingID string, actualDuration float64) error {
	resp, err := c.doRequest("POST", "/api/service/complete", api.CompleteServiceRequest{
		BookingID:      bookingID,
		ActualDuration: actualDuration,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var booking api.Booking
	json.Unmarshal(dataBytes, &booking)

	fmt.Printf("服务已完成!\n")
	printBooking(&booking)
	return nil
}

func (c *Client) addReview(req api.AddReviewRequest) error {
	resp, err := c.doRequest("POST", "/api/review/add", req)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var review api.Review
	json.Unmarshal(dataBytes, &review)

	fmt.Printf("评价提交成功!\n")
	fmt.Printf("阿姨ID: %s\n", review.AuntID)
	fmt.Printf("评分: %.1f\n", review.Rating)
	if review.Comment != "" {
		fmt.Printf("评论: %s\n", review.Comment)
	}
	return nil
}

func (c *Client) getStats(auntID, month string) error {
	path := "/api/stats?aunt_id=" + auntID + "&month=" + month

	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	dataBytes, _ := json.Marshal(resp.Data)
	var stats api.MonthlyStats
	json.Unmarshal(dataBytes, &stats)

	fmt.Printf("月度统计 (%s):\n", stats.Month)
	fmt.Printf("阿姨ID: %s\n", stats.AuntID)
	fmt.Printf("接单数: %d\n", stats.TotalBookings)
	fmt.Printf("总服务时长: %.1f小时\n", stats.TotalHours)
	fmt.Printf("平均评分: %.1f\n", stats.AverageRating)
	fmt.Printf("复购率: %.1f%%\n", stats.RepeatRate*100)
	return nil
}

func printUsage() {
	fmt.Println("家政服务预约平台客户端")
	fmt.Println("\n用法:")
	fmt.Println("  client [选项] <命令> [参数]")
	fmt.Println("\n选项:")
	fmt.Println("  -server <URL>    服务端地址 (默认: http://localhost:8080)")
	fmt.Println("\n命令:")
	fmt.Println("  aunt-register    阿姨入驻")
	fmt.Println("  aunt-list        查看阿姨列表")
	fmt.Println("  aunt-get <ID>    查看阿姨详情")
	fmt.Println("  customer-register 客户注册")
	fmt.Println("  customer-list    查看客户列表")
	fmt.Println("  customer-get <ID> 查看客户详情")
	fmt.Println("  booking-create   创建预约订单")
	fmt.Println("  booking-list     查看订单列表")
	fmt.Println("  booking-get <ID> 查看订单详情")
	fmt.Println("  booking-pending  查看待匹配订单")
	fmt.Println("  service-start <ID> 开始服务")
	fmt.Println("  service-complete <ID> <时长> 完成服务")
	fmt.Println("  review-add       添加评价")
	fmt.Println("  stats <阿姨ID> <月份> 查看月度统计")
	fmt.Println("\n示例:")
	fmt.Println("  client aunt-register")
	fmt.Println("  client booking-create")
	fmt.Println("  client stats A001 2026-05")
}

func promptString(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return strings.TrimSpace(input)
}

func promptInt(prompt string) int {
	var val int
	fmt.Print(prompt)
	fmt.Scanln(&val)
	return val
}

func promptFloat(prompt string) float64 {
	var val float64
	fmt.Print(prompt)
	fmt.Scanln(&val)
	return val
}

func main() {
	serverURL := flag.String("server", defaultServerURL, "服务端地址")
	flag.Usage = printUsage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*serverURL)

	cmd := args[0]
	var err error

	switch cmd {
	case "aunt-register":
		req := api.RegisterAuntRequest{
			ID:              promptString("请输入阿姨ID: "),
			Name:            promptString("请输入姓名: "),
			Age:             promptInt("请输入年龄: "),
			Phone:           promptString("请输入手机号: "),
			ServiceCategory: api.ServiceCategory(promptString("请输入服务类别(日常保洁/深度清洁/月嫂/育婴师/老人陪护): ")),
			YearsOfService:  promptInt("请输入从业年限: "),
			ServiceArea:     promptString("请输入服务区域: "),
		}
		err = client.registerAunt(req)

	case "aunt-list":
		err = client.getAunt("")

	case "aunt-get":
		if len(args) < 2 {
			fmt.Println("错误: 请指定阿姨ID")
			os.Exit(1)
		}
		err = client.getAunt(args[1])

	case "customer-register":
		req := api.RegisterCustomerRequest{
			ID:    promptString("请输入客户ID: "),
			Name:  promptString("请输入姓名: "),
			Phone: promptString("请输入手机号: "),
		}
		err = client.registerCustomer(req)

	case "customer-list":
		err = client.getCustomer("")

	case "customer-get":
		if len(args) < 2 {
			fmt.Println("错误: 请指定客户ID")
			os.Exit(1)
		}
		err = client.getCustomer(args[1])

	case "booking-create":
		categoryStr := promptString("请输入服务类别(日常保洁/深度清洁/月嫂/育婴师/老人陪护): ")
		dateStr := promptString("请输入服务日期和时间(格式: 2006-01-02 15:04): ")
		date, parseErr := time.Parse("2006-01-02 15:04", dateStr)
		if parseErr != nil {
			err = fmt.Errorf("日期格式错误: %v", parseErr)
			break
		}

		req := api.CreateBookingRequest{
			CustomerID:       promptString("请输入客户ID: "),
			ServiceCategory:  api.ServiceCategory(categoryStr),
			ServiceDate:      date,
			EstimatedDuration: promptFloat("请输入预估时长(小时, 2-8小时, 半小时为单位): "),
			Address:          promptString("请输入服务地址: "),
		}
		err = client.createBooking(req)

	case "booking-list":
		err = client.getBooking("", "", "", false)

	case "booking-get":
		if len(args) < 2 {
			fmt.Println("错误: 请指定订单ID")
			os.Exit(1)
		}
		err = client.getBooking(args[1], "", "", false)

	case "booking-pending":
		err = client.getBooking("", "", "", true)

	case "service-start":
		if len(args) < 2 {
			fmt.Println("错误: 请指定订单ID")
			os.Exit(1)
		}
		err = client.startService(args[1])

	case "service-complete":
		if len(args) < 3 {
			fmt.Println("错误: 请指定订单ID和实际时长")
			os.Exit(1)
		}
		var duration float64
		fmt.Sscanf(args[2], "%f", &duration)
		err = client.completeService(args[1], duration)

	case "review-add":
		req := api.AddReviewRequest{
			AuntID:     promptString("请输入阿姨ID: "),
			BookingID:  promptString("请输入订单ID: "),
			CustomerID: promptString("请输入客户ID: "),
			Rating:     promptFloat("请输入评分(1-5分): "),
			Comment:    promptString("请输入评论(可选): "),
		}
		err = client.addReview(req)

	case "stats":
		if len(args) < 3 {
			fmt.Println("错误: 请指定阿姨ID和月份(格式: YYYY-MM)")
			os.Exit(1)
		}
		err = client.getStats(args[1], args[2])

	default:
		fmt.Printf("错误: 未知命令 '%s'\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}

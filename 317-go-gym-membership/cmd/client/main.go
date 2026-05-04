package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gym-membership/types"
)

const defaultServerURL = "http://localhost:8080"

type APIClient struct {
	baseURL string
	client  *http.Client
}

func NewAPIClient(baseURL string) *APIClient {
	if baseURL == "" {
		baseURL = defaultServerURL
	}
	return &APIClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) doRequest(method, path string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求失败: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}

	return nil
}

func (c *APIClient) CreateMember(name, phone, password string) (*types.CreateMemberResponse, error) {
	req := types.CreateMemberRequest{
		Name:     name,
		Phone:    phone,
		Password: password,
	}

	var resp types.CreateMemberResponse
	if err := c.doRequest("POST", "/api/members", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) Login(phone, password string) (*types.LoginResponse, error) {
	req := types.LoginRequest{
		Phone:    phone,
		Password: password,
	}

	var resp types.LoginResponse
	if err := c.doRequest("POST", "/api/login", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) PurchaseCard(memberID string, cardType types.CardType) (*types.PurchaseCardResponse, error) {
	req := types.PurchaseCardRequest{
		MemberID: memberID,
		CardType: cardType,
	}

	var resp types.PurchaseCardResponse
	if err := c.doRequest("POST", "/api/cards/purchase", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) RenewCard(memberID string, cardType types.CardType) (*types.RenewCardResponse, error) {
	req := types.RenewCardRequest{
		MemberID: memberID,
		CardType: cardType,
	}

	var resp types.RenewCardResponse
	if err := c.doRequest("POST", "/api/cards/renew", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetMemberInfo(memberID string) (*types.GetMemberInfoResponse, error) {
	var resp types.GetMemberInfoResponse
	if err := c.doRequest("GET", "/api/members/info?member_id="+memberID, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) CreateClass(req *CreateClassClientRequest) (*types.CreateClassResponse, error) {
	apiReq := types.CreateClassRequest{
		Name:        req.Name,
		Weekday:     time.Weekday(req.Weekday),
		StartTime:   req.StartTime,
		Duration:    req.Duration,
		MaxCapacity: req.MaxCapacity,
		MinCapacity: req.MinCapacity,
		Instructor:  req.Instructor,
		Location:    req.Location,
	}

	var resp types.CreateClassResponse
	if err := c.doRequest("POST", "/api/classes", apiReq, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) ListClasses(weekday string) (*types.ListClassesResponse, error) {
	path := "/api/classes"
	if weekday != "" {
		path += "?weekday=" + weekday
	}

	var resp types.ListClassesResponse
	if err := c.doRequest("GET", path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) BookClass(memberID, classID, date string) (*types.BookClassResponse, error) {
	req := types.BookClassRequest{
		MemberID: memberID,
		ClassID:  classID,
		Date:     date,
	}

	var resp types.BookClassResponse
	if err := c.doRequest("POST", "/api/bookings", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) CancelBooking(bookingID, memberID string) (*types.CancelBookingResponse, error) {
	req := types.CancelBookingRequest{
		BookingID: bookingID,
		MemberID:  memberID,
	}

	var resp types.CancelBookingResponse
	if err := c.doRequest("POST", "/api/bookings/cancel", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) CheckIn(memberID, classID, date string) (*types.CheckInResponse, error) {
	req := types.CheckInRequest{
		MemberID: memberID,
		ClassID:  classID,
		Date:     date,
	}

	var resp types.CheckInResponse
	if err := c.doRequest("POST", "/api/checkin", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GetClassBookings(classID, date string) (*types.GetClassBookingsResponse, error) {
	path := fmt.Sprintf("/api/classes/bookings?class_id=%s&date=%s", classID, date)

	var resp types.GetClassBookingsResponse
	if err := c.doRequest("GET", path, nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) SetCardPrice(cardType types.CardType, price float64, validDays int) (*types.SetCardPriceResponse, error) {
	req := types.SetCardPriceRequest{
		CardType:  cardType,
		Price:     price,
		ValidDays: validDays,
	}

	var resp types.SetCardPriceResponse
	if err := c.doRequest("POST", "/api/prices", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) ListCardPrices() (*types.ListCardPricesResponse, error) {
	var resp types.ListCardPricesResponse
	if err := c.doRequest("GET", "/api/prices", nil, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (c *APIClient) GenerateClassInstances(startDate, endDate string) (*types.GenerateClassInstancesResponse, error) {
	req := types.GenerateClassInstancesRequest{
		StartDate: startDate,
		EndDate:   endDate,
	}

	var resp types.GenerateClassInstancesResponse
	if err := c.doRequest("POST", "/api/classes/instances", req, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type CreateClassClientRequest struct {
	Name        string `json:"name"`
	Weekday     int    `json:"weekday"`
	StartTime   string `json:"start_time"`
	Duration    int    `json:"duration"`
	MaxCapacity int    `json:"max_capacity"`
	MinCapacity int    `json:"min_capacity"`
	Instructor  string `json:"instructor"`
	Location    string `json:"location"`
}

var (
	apiClient   *APIClient
	currentUser *types.Member
	serverURL   string
)

func main() {
	serverURL = os.Getenv("GYM_SERVER_URL")
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	apiClient = NewAPIClient(serverURL)

	fmt.Println("========================================")
	fmt.Println("     健身房会员管理系统 - 客户端")
	fmt.Println("========================================")
	fmt.Printf("服务器地址: %s\n", serverURL)
	fmt.Println("========================================")

	mainMenu()
}

func mainMenu() {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n========== 主菜单 ==========")
		if currentUser != nil {
			fmt.Printf("当前用户: %s (%s)\n", currentUser.Name, currentUser.Phone)
		}
		fmt.Println("1. 会员注册")
		fmt.Println("2. 会员登录")
		fmt.Println("3. 办卡/续费")
		fmt.Println("4. 查看课程列表")
		fmt.Println("5. 预约课程")
		fmt.Println("6. 取消预约")
		fmt.Println("7. 签到")
		fmt.Println("8. 查看个人信息")
		fmt.Println("------ 管理员功能 ------")
		fmt.Println("9. 创建课程")
		fmt.Println("10. 查看课程预约情况")
		fmt.Println("11. 设置会员卡价格")
		fmt.Println("12. 生成课程实例")
		fmt.Println("13. 查看卡价列表")
		fmt.Println("-----------------------")
		fmt.Println("0. 退出")
		fmt.Print("请选择功能: ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		switch input {
		case "1":
			registerMember()
		case "2":
			loginMember()
		case "3":
			purchaseOrRenewCard()
		case "4":
			listClasses()
		case "5":
			bookClass()
		case "6":
			cancelBooking()
		case "7":
			checkIn()
		case "8":
			viewMemberInfo()
		case "9":
			createClass()
		case "10":
			viewClassBookings()
		case "11":
			setCardPrice()
		case "12":
			generateClassInstances()
		case "13":
			listCardPrices()
		case "0":
			fmt.Println("感谢使用，再见！")
			os.Exit(0)
		default:
			fmt.Println("无效的选择，请重新输入。")
		}
	}
}

func registerMember() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 会员注册 ==========")
	fmt.Print("请输入姓名: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("请输入手机号(11位): ")
	phone, _ := reader.ReadString('\n')
	phone = strings.TrimSpace(phone)

	fmt.Print("请输入密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	resp, err := apiClient.CreateMember(name, phone, password)
	if err != nil {
		fmt.Printf("注册失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("注册成功！")
		fmt.Printf("会员ID: %s\n", resp.Member.ID)
	} else {
		fmt.Printf("注册失败: %s\n", resp.Message)
	}
}

func loginMember() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 会员登录 ==========")
	fmt.Print("请输入手机号: ")
	phone, _ := reader.ReadString('\n')
	phone = strings.TrimSpace(phone)

	fmt.Print("请输入密码: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	resp, err := apiClient.Login(phone, password)
	if err != nil {
		fmt.Printf("登录失败: %v\n", err)
		return
	}

	if resp.Success {
		currentUser = resp.Member
		fmt.Println("登录成功！")
		fmt.Printf("欢迎, %s!\n", currentUser.Name)
	} else {
		fmt.Printf("登录失败: %s\n", resp.Message)
	}
}

func purchaseOrRenewCard() {
	if currentUser == nil {
		fmt.Println("请先登录！")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 办卡/续费 ==========")
	fmt.Println("可用卡类型:")
	fmt.Println("1. 月卡 (monthly)")
	fmt.Println("2. 季卡 (quarterly)")
	fmt.Println("3. 年卡 (yearly)")
	fmt.Print("请选择卡类型(输入数字或英文): ")
	cardTypeInput, _ := reader.ReadString('\n')
	cardTypeInput = strings.TrimSpace(cardTypeInput)

	var cardType types.CardType
	switch strings.ToLower(cardTypeInput) {
	case "1", "monthly":
		cardType = types.MonthlyCard
	case "2", "quarterly":
		cardType = types.QuarterlyCard
	case "3", "yearly":
		cardType = types.YearlyCard
	default:
		fmt.Println("无效的卡类型")
		return
	}

	fmt.Println("\n操作类型:")
	fmt.Println("1. 新办卡")
	fmt.Println("2. 续费")
	fmt.Print("请选择操作类型: ")
	actionInput, _ := reader.ReadString('\n')
	actionInput = strings.TrimSpace(actionInput)

	var respCard *types.MembershipCard
	var err error

	if actionInput == "1" {
		resp, err := apiClient.PurchaseCard(currentUser.ID, cardType)
		if err == nil && resp.Success {
			respCard = resp.Card
		} else if err == nil {
			fmt.Printf("办卡失败: %s\n", resp.Message)
			return
		}
	} else if actionInput == "2" {
		resp, err := apiClient.RenewCard(currentUser.ID, cardType)
		if err == nil && resp.Success {
			respCard = resp.Card
		} else if err == nil {
			fmt.Printf("续费失败: %s\n", resp.Message)
			return
		}
	} else {
		fmt.Println("无效的操作类型")
		return
	}

	if err != nil {
		fmt.Printf("操作失败: %v\n", err)
		return
	}

	fmt.Println("\n操作成功！")
	fmt.Printf("卡类型: %s\n", respCard.Type)
	fmt.Printf("开始日期: %s\n", respCard.StartDate.Format("2006-01-02"))
	fmt.Printf("结束日期: %s\n", respCard.EndDate.Format("2006-01-02"))
	fmt.Printf("价格: %.2f\n", respCard.Price)
	fmt.Printf("状态: %s\n", respCard.Status)
}

func listClasses() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 课程列表 ==========")
	fmt.Print("是否按星期筛选? (输入星期数字0-6，直接回车显示所有): ")
	weekdayInput, _ := reader.ReadString('\n')
	weekdayInput = strings.TrimSpace(weekdayInput)

	var weekday string
	if weekdayInput != "" {
		weekday = weekdayInput
	}

	resp, err := apiClient.ListClasses(weekday)
	if err != nil {
		fmt.Printf("获取课程列表失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取课程列表失败: %s\n", resp.Message)
		return
	}

	if len(resp.Classes) == 0 {
		fmt.Println("暂无课程")
		return
	}

	weekdays := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}

	fmt.Println("\n课程列表:")
	fmt.Println("----------------------------------------")
	for i, class := range resp.Classes {
		fmt.Printf("%d. 课程ID: %s\n", i+1, class.ID)
		fmt.Printf("   名称: %s\n", class.Name)
		fmt.Printf("   星期: %s\n", weekdays[class.Weekday])
		fmt.Printf("   时间: %s (时长%d分钟)\n", class.StartTime, class.Duration)
		fmt.Printf("   教练: %s\n", class.Instructor)
		fmt.Printf("   地点: %s\n", class.Location)
		fmt.Printf("   人数限制: 最少%d人, 最多%d人\n", class.MinCapacity, class.MaxCapacity)
		fmt.Println("----------------------------------------")
	}
}

func bookClass() {
	if currentUser == nil {
		fmt.Println("请先登录！")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 预约课程 ==========")
	fmt.Print("请输入课程ID: ")
	classID, _ := reader.ReadString('\n')
	classID = strings.TrimSpace(classID)

	fmt.Print("请输入日期 (格式: 2006-01-02): ")
	date, _ := reader.ReadString('\n')
	date = strings.TrimSpace(date)

	resp, err := apiClient.BookClass(currentUser.ID, classID, date)
	if err != nil {
		fmt.Printf("预约失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("预约成功！")
		fmt.Printf("预约ID: %s\n", resp.Booking.ID)
	} else {
		fmt.Printf("预约失败: %s\n", resp.Message)
	}
}

func cancelBooking() {
	if currentUser == nil {
		fmt.Println("请先登录！")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 取消预约 ==========")
	fmt.Print("请输入预约ID: ")
	bookingID, _ := reader.ReadString('\n')
	bookingID = strings.TrimSpace(bookingID)

	resp, err := apiClient.CancelBooking(bookingID, currentUser.ID)
	if err != nil {
		fmt.Printf("取消预约失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("取消预约成功！")
	} else {
		fmt.Printf("取消预约失败: %s\n", resp.Message)
	}
}

func checkIn() {
	if currentUser == nil {
		fmt.Println("请先登录！")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 签到 ==========")
	fmt.Print("请输入课程ID: ")
	classID, _ := reader.ReadString('\n')
	classID = strings.TrimSpace(classID)

	fmt.Print("请输入日期 (格式: 2006-01-02): ")
	date, _ := reader.ReadString('\n')
	date = strings.TrimSpace(date)

	resp, err := apiClient.CheckIn(currentUser.ID, classID, date)
	if err != nil {
		fmt.Printf("签到失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("签到成功！")
		fmt.Printf("签到时间: %s\n", resp.Attendance.CheckInTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("签到失败: %s\n", resp.Message)
	}
}

func viewMemberInfo() {
	if currentUser == nil {
		fmt.Println("请先登录！")
		return
	}

	resp, err := apiClient.GetMemberInfo(currentUser.ID)
	if err != nil {
		fmt.Printf("获取会员信息失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取会员信息失败: %s\n", resp.Message)
		return
	}

	fmt.Println("\n========== 个人信息 ==========")
	fmt.Printf("姓名: %s\n", resp.Member.Name)
	fmt.Printf("手机号: %s\n", resp.Member.Phone)
	fmt.Printf("注册时间: %s\n", resp.Member.CreatedAt.Format("2006-01-02 15:04:05"))

	if resp.CurrentCard != nil {
		fmt.Println("\n---------- 当前会员卡 ----------")
		fmt.Printf("卡类型: %s\n", resp.CurrentCard.Type)
		fmt.Printf("开始日期: %s\n", resp.CurrentCard.StartDate.Format("2006-01-02"))
		fmt.Printf("结束日期: %s\n", resp.CurrentCard.EndDate.Format("2006-01-02"))
		fmt.Printf("价格: %.2f\n", resp.CurrentCard.Price)
		fmt.Printf("状态: %s\n", resp.CurrentCard.Status)
	} else {
		fmt.Println("\n当前无激活的会员卡")
	}

	if len(resp.ActiveBookings) > 0 {
		fmt.Println("\n---------- 已预约课程 ----------")
		for i, b := range resp.ActiveBookings {
			fmt.Printf("%d. 预约ID: %s\n", i+1, b.ID)
			fmt.Printf("   课程: %s\n", b.ClassName)
			fmt.Printf("   日期: %s\n", b.Date)
			fmt.Printf("   状态: %s\n", b.Status)
		}
	} else {
		fmt.Println("\n暂无已预约课程")
	}

	if len(resp.AttendanceHistory) > 0 {
		fmt.Println("\n---------- 历史签到记录 ----------")
		for i, a := range resp.AttendanceHistory {
			fmt.Printf("%d. 课程: %s\n", i+1, a.ClassName)
			fmt.Printf("   日期: %s\n", a.Date)
			fmt.Printf("   签到时间: %s\n", a.CheckInTime.Format("2006-01-02 15:04:05"))
		}
	} else {
		fmt.Println("\n暂无签到记录")
	}
}

func createClass() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 创建课程 ==========")
	fmt.Print("请输入课程名称: ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Println("星期: 0=星期日, 1=星期一, ..., 6=星期六")
	fmt.Print("请输入星期数字: ")
	weekdayInput, _ := reader.ReadString('\n')
	weekdayInput = strings.TrimSpace(weekdayInput)
	weekday, _ := strconv.Atoi(weekdayInput)

	fmt.Print("请输入开始时间 (格式: 15:04): ")
	startTime, _ := reader.ReadString('\n')
	startTime = strings.TrimSpace(startTime)

	fmt.Print("请输入时长 (分钟): ")
	durationInput, _ := reader.ReadString('\n')
	durationInput = strings.TrimSpace(durationInput)
	duration, _ := strconv.Atoi(durationInput)

	fmt.Print("请输入最大人数: ")
	maxCapInput, _ := reader.ReadString('\n')
	maxCapInput = strings.TrimSpace(maxCapInput)
	maxCap, _ := strconv.Atoi(maxCapInput)

	fmt.Print("请输入最少开课人数: ")
	minCapInput, _ := reader.ReadString('\n')
	minCapInput = strings.TrimSpace(minCapInput)
	minCap, _ := strconv.Atoi(minCapInput)

	fmt.Print("请输入教练姓名: ")
	instructor, _ := reader.ReadString('\n')
	instructor = strings.TrimSpace(instructor)

	fmt.Print("请输入上课地点: ")
	location, _ := reader.ReadString('\n')
	location = strings.TrimSpace(location)

	req := &CreateClassClientRequest{
		Name:        name,
		Weekday:     weekday,
		StartTime:   startTime,
		Duration:    duration,
		MaxCapacity: maxCap,
		MinCapacity: minCap,
		Instructor:  instructor,
		Location:    location,
	}

	resp, err := apiClient.CreateClass(req)
	if err != nil {
		fmt.Printf("创建课程失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("创建课程成功！")
		fmt.Printf("课程ID: %s\n", resp.Class.ID)
	} else {
		fmt.Printf("创建课程失败: %s\n", resp.Message)
	}
}

func viewClassBookings() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 查看课程预约情况 ==========")
	fmt.Print("请输入课程ID: ")
	classID, _ := reader.ReadString('\n')
	classID = strings.TrimSpace(classID)

	fmt.Print("请输入日期 (格式: 2006-01-02): ")
	date, _ := reader.ReadString('\n')
	date = strings.TrimSpace(date)

	resp, err := apiClient.GetClassBookings(classID, date)
	if err != nil {
		fmt.Printf("获取课程预约情况失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取课程预约情况失败: %s\n", resp.Message)
		return
	}

	fmt.Println("\n---------- 课程信息 ----------")
	if resp.Class != nil {
		fmt.Printf("课程名称: %s\n", resp.Class.Name)
		fmt.Printf("开始时间: %s\n", resp.Class.StartTime)
		fmt.Printf("教练: %s\n", resp.Class.Instructor)
		fmt.Printf("地点: %s\n", resp.Class.Location)
	}

	if resp.ClassInstance != nil {
		fmt.Printf("日期: %s\n", resp.ClassInstance.Date.Format("2006-01-02"))
		fmt.Printf("状态: %s\n", resp.ClassInstance.Status)
	}

	fmt.Printf("\n已签到人数: %d\n", resp.AttendanceCount)

	if len(resp.Bookings) > 0 {
		fmt.Println("\n---------- 预约名单 ----------")
		for i, b := range resp.Bookings {
			fmt.Printf("%d. 会员: %s (%s)\n", i+1, b.MemberName, b.MemberPhone)
			fmt.Printf("   预约时间: %s\n", b.BookedAt.Format("2006-01-02 15:04:05"))
			fmt.Printf("   状态: %s\n", b.Status)
		}
	} else {
		fmt.Println("\n暂无预约")
	}
}

func setCardPrice() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 设置会员卡价格 ==========")
	fmt.Println("可用卡类型:")
	fmt.Println("1. 月卡 (monthly)")
	fmt.Println("2. 季卡 (quarterly)")
	fmt.Println("3. 年卡 (yearly)")
	fmt.Print("请选择卡类型(输入数字或英文): ")
	cardTypeInput, _ := reader.ReadString('\n')
	cardTypeInput = strings.TrimSpace(cardTypeInput)

	var cardType types.CardType
	switch strings.ToLower(cardTypeInput) {
	case "1", "monthly":
		cardType = types.MonthlyCard
	case "2", "quarterly":
		cardType = types.QuarterlyCard
	case "3", "yearly":
		cardType = types.YearlyCard
	default:
		fmt.Println("无效的卡类型")
		return
	}

	fmt.Print("请输入价格: ")
	priceInput, _ := reader.ReadString('\n')
	priceInput = strings.TrimSpace(priceInput)
	price, _ := strconv.ParseFloat(priceInput, 64)

	fmt.Print("请输入有效天数: ")
	daysInput, _ := reader.ReadString('\n')
	daysInput = strings.TrimSpace(daysInput)
	validDays, _ := strconv.Atoi(daysInput)

	resp, err := apiClient.SetCardPrice(cardType, price, validDays)
	if err != nil {
		fmt.Printf("设置价格失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Println("设置价格成功！")
		fmt.Printf("卡类型: %s\n", resp.Price.CardType)
		fmt.Printf("价格: %.2f\n", resp.Price.Price)
		fmt.Printf("有效天数: %d\n", resp.Price.ValidDays)
	} else {
		fmt.Printf("设置价格失败: %s\n", resp.Message)
	}
}

func generateClassInstances() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n========== 生成课程实例 ==========")
	fmt.Print("请输入开始日期 (格式: 2006-01-02): ")
	startDate, _ := reader.ReadString('\n')
	startDate = strings.TrimSpace(startDate)

	fmt.Print("请输入结束日期 (格式: 2006-01-02): ")
	endDate, _ := reader.ReadString('\n')
	endDate = strings.TrimSpace(endDate)

	resp, err := apiClient.GenerateClassInstances(startDate, endDate)
	if err != nil {
		fmt.Printf("生成课程实例失败: %v\n", err)
		return
	}

	if resp.Success {
		fmt.Printf("生成课程实例成功！共生成 %d 个课程实例\n", len(resp.Instances))
	} else {
		fmt.Printf("生成课程实例失败: %s\n", resp.Message)
	}
}

func listCardPrices() {
	resp, err := apiClient.ListCardPrices()
	if err != nil {
		fmt.Printf("获取卡价列表失败: %v\n", err)
		return
	}

	if !resp.Success {
		fmt.Printf("获取卡价列表失败: %s\n", resp.Message)
		return
	}

	if len(resp.Prices) == 0 {
		fmt.Println("暂无卡价设置")
		return
	}

	fmt.Println("\n========== 会员卡价格列表 ==========")
	fmt.Println("----------------------------------------")
	for i, p := range resp.Prices {
		fmt.Printf("%d. 卡类型: %s\n", i+1, p.CardType)
		fmt.Printf("   价格: %.2f\n", p.Price)
		fmt.Printf("   有效天数: %d\n", p.ValidDays)
		fmt.Printf("   更新时间: %s\n", p.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println("----------------------------------------")
	}
}

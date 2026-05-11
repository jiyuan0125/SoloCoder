package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"gasstation/internal/common"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultServerURL = "http://localhost:8080"

type Client struct {
	baseURL string
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = defaultServerURL
	}
	return &Client{baseURL: baseURL}
}

func (c *Client) doRequest(method, path string, body interface{}) (*common.APIResponse, error) {
	var reader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, reader)
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

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp common.APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("response: %s", string(respBody))
	}
	return &apiResp, nil
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(data))
}

func printHelp() {
	fmt.Println(`加油站管理系统客户端
用法:
  client [命令] [选项]

命令:
  stations                    列出所有加油站
  station-create              创建加油站
  station-get <id>            获取加油站详情
  station-update <id>         更新加油站信息
  
  fuels <station-id>          列出加油站油品
  fuel-add <station-id>       添加油品到加油站
  
  members                     列出所有会员
  member-register             注册会员
  member-get <id>             获取会员信息
  member-by-phone <phone>     通过手机号查询会员
  
  refuel                      加油
  
  records                     列出加油记录
  records-export              导出加油记录CSV
  
  price-requests              列出价格调整请求
  price-request               提交价格调整请求
  price-review <id>           审核价格调整请求
  price-history               查看价格变更历史
  
  restock-todos               列出补货待办
  restock-complete <id>       完成补货待办

全局选项:
  -server <url>               服务端地址 (默认: http://localhost:8080)
  -h, -help                   显示帮助

示例:
  client stations
  client station-create -name "中国石油" -address "北京朝阳区" -lat 39.9 -lon 116.4 -phone 12345678 -open=true
  client member-register -phone 13800138000
  client refuel -station 1 -fuel 92 -liters 50.5 -phone 13800138000 -points 100
  client records -station 1 -start 2026-01-01T00:00:00Z -end 2026-12-31T23:59:59Z`)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	cmd := os.Args[1]
	if cmd == "-h" || cmd == "-help" || cmd == "help" {
		printHelp()
		return
	}

	serverURL := defaultServerURL
	for i, arg := range os.Args {
		if arg == "-server" && i+1 < len(os.Args) {
			serverURL = os.Args[i+1]
			break
		}
	}

	client := NewClient(serverURL)

	switch cmd {
	case "stations":
		listStations(client)
	case "station-create":
		createStation(client)
	case "station-get":
		getStation(client)
	case "station-update":
		updateStation(client)
	case "fuels":
		listFuels(client)
	case "fuel-add":
		addFuel(client)
	case "members":
		listMembers(client)
	case "member-register":
		registerMember(client)
	case "member-get":
		getMember(client)
	case "member-by-phone":
		getMemberByPhone(client)
	case "refuel":
		doRefuel(client)
	case "records":
		listRecords(client)
	case "records-export":
		exportRecords(client)
	case "price-requests":
		listPriceRequests(client)
	case "price-request":
		submitPriceRequest(client)
	case "price-review":
		reviewPriceRequest(client)
	case "price-history":
		getPriceHistory(client)
	case "restock-todos":
		listRestockTodos(client)
	case "restock-complete":
		completeRestockTodo(client)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
		printHelp()
	}
}

func listStations(c *Client) {
	resp, err := c.doRequest(http.MethodGet, "/stations", nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func createStation(c *Client) {
	fs := flag.NewFlagSet("station-create", flag.ExitOnError)
	name := fs.String("name", "", "加油站名称")
	addr := fs.String("address", "", "地址")
	lat := fs.Float64("lat", 0, "纬度")
	lon := fs.Float64("lon", 0, "经度")
	phone := fs.String("phone", "", "联系电话")
	open := fs.Bool("open", true, "是否营业")
	fs.Parse(os.Args[2:])

	if *name == "" {
		fmt.Println("请提供 -name 参数")
		return
	}

	req := common.CreateStationRequest{
		Name:      *name,
		Address:   *addr,
		Latitude:  *lat,
		Longitude: *lon,
		Phone:     *phone,
		IsOpen:    *open,
	}

	resp, err := c.doRequest(http.MethodPost, "/stations", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func getStation(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供加油站ID")
		return
	}
	id := os.Args[2]
	resp, err := c.doRequest(http.MethodGet, "/stations/"+id, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func updateStation(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供加油站ID")
		return
	}
	id := os.Args[2]
	fs := flag.NewFlagSet("station-update", flag.ExitOnError)
	name := fs.String("name", "", "加油站名称")
	addr := fs.String("address", "", "地址")
	phone := fs.String("phone", "", "联系电话")
	open := fs.Bool("open", false, "是否营业")
	openSet := fs.Bool("set-open", false, "是否设置营业状态")
	fs.Parse(os.Args[3:])

	req := common.UpdateStationRequest{}
	if *name != "" {
		req.Name = name
	}
	if *addr != "" {
		req.Address = addr
	}
	if *phone != "" {
		req.Phone = phone
	}
	if *openSet {
		req.IsOpen = open
	}

	resp, err := c.doRequest(http.MethodPut, "/stations/"+id, req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func listFuels(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供加油站ID")
		return
	}
	stationID := os.Args[2]
	resp, err := c.doRequest(http.MethodGet, "/stations/"+stationID+"/fuels", nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func addFuel(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供加油站ID")
		return
	}
	stationID := os.Args[2]
	fs := flag.NewFlagSet("fuel-add", flag.ExitOnError)
	code := fs.String("code", "", "油品编号 (如 92, 95, 98, 0)")
	name := fs.String("name", "", "油品名称")
	price := fs.Int64("price", 0, "单价（分）")
	stock := fs.Int64("stock", 0, "库存量（升）")
	fs.Parse(os.Args[3:])

	if *code == "" || *price <= 0 {
		fmt.Println("请提供 -code 和 -price 参数")
		return
	}

	body := struct {
		FuelCode string `json:"fuel_code"`
		FuelName string `json:"fuel_name"`
		Price    int64  `json:"price"`
		Stock    int64  `json:"stock"`
	}{*code, *name, *price, *stock}

	resp, err := c.doRequest(http.MethodPost, "/stations/"+stationID+"/fuels", body)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func listMembers(c *Client) {
	resp, err := c.doRequest(http.MethodGet, "/members", nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func registerMember(c *Client) {
	fs := flag.NewFlagSet("member-register", flag.ExitOnError)
	phone := fs.String("phone", "", "手机号")
	fs.Parse(os.Args[2:])

	if *phone == "" {
		fmt.Println("请提供 -phone 参数")
		return
	}

	req := common.RegisterMemberRequest{Phone: *phone}
	resp, err := c.doRequest(http.MethodPost, "/members", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func getMember(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供会员ID")
		return
	}
	id := os.Args[2]
	resp, err := c.doRequest(http.MethodGet, "/members/"+id, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func getMemberByPhone(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供手机号")
		return
	}
	phone := os.Args[2]
	resp, err := c.doRequest(http.MethodGet, "/members/by-phone?phone="+phone, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func doRefuel(c *Client) {
	fs := flag.NewFlagSet("refuel", flag.ExitOnError)
	station := fs.String("station", "", "加油站ID")
	fuel := fs.String("fuel", "", "油品编号")
	liters := fs.String("liters", "", "加油升数")
	phone := fs.String("phone", "", "会员手机号（可选）")
	points := fs.String("points", "", "使用积分数量（可选）")
	fs.Parse(os.Args[2:])

	if *station == "" || *fuel == "" || *liters == "" {
		fmt.Println("请提供 -station, -fuel, -liters 参数")
		return
	}

	l, err := strconv.ParseFloat(*liters, 64)
	if err != nil {
		fmt.Println("升数格式错误:", err)
		return
	}

	req := common.RefuelRequest{
		StationID: *station,
		FuelCode:  *fuel,
		Liters:    l,
	}

	if *phone != "" {
		req.MemberPhone = phone
	}

	if *points != "" {
		p, _ := strconv.ParseInt(*points, 10, 64)
		req.UsePoints = &p
	}

	resp, err := c.doRequest(http.MethodPost, "/refuel", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func listRecords(c *Client) {
	fs := flag.NewFlagSet("records", flag.ExitOnError)
	station := fs.String("station", "", "加油站ID（可选）")
	start := fs.String("start", "", "开始时间 RFC3339（可选）")
	end := fs.String("end", "", "结束时间 RFC3339（可选）")
	fs.Parse(os.Args[2:])

	path := "/records?"
	params := []string{}
	if *station != "" {
		params = append(params, "station_id="+*station)
	}
	if *start != "" {
		params = append(params, "start="+*start)
	}
	if *end != "" {
		params = append(params, "end="+*end)
	}
	path += strings.Join(params, "&")

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func exportRecords(c *Client) {
	fs := flag.NewFlagSet("records-export", flag.ExitOnError)
	station := fs.String("station", "", "加油站ID（可选）")
	start := fs.String("start", "", "开始时间 RFC3339（可选）")
	end := fs.String("end", "", "结束时间 RFC3339（可选）")
	output := fs.String("output", "records.csv", "输出文件名")
	fs.Parse(os.Args[2:])

	path := "/records/export?"
	params := []string{}
	if *station != "" {
		params = append(params, "station_id="+*station)
	}
	if *start != "" {
		params = append(params, "start="+*start)
	}
	if *end != "" {
		params = append(params, "end="+*end)
	}
	path += strings.Join(params, "&")

	url := c.baseURL + path
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	if err := os.WriteFile(*output, data, 0644); err != nil {
		fmt.Println("写入文件错误:", err)
		return
	}
	fmt.Println("已导出到:", *output)
}

func listPriceRequests(c *Client) {
	fs := flag.NewFlagSet("price-requests", flag.ExitOnError)
	station := fs.String("station", "", "加油站ID（可选）")
	fs.Parse(os.Args[2:])

	path := "/price-requests"
	if *station != "" {
		path += "?station_id=" + *station
	}

	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func submitPriceRequest(c *Client) {
	fs := flag.NewFlagSet("price-request", flag.ExitOnError)
	station := fs.String("station", "", "加油站ID")
	fuel := fs.String("fuel", "", "油品编号")
	price := fs.Int64("price", 0, "新价格（分）")
	fs.Parse(os.Args[2:])

	if *station == "" || *fuel == "" || *price <= 0 {
		fmt.Println("请提供 -station, -fuel, -price 参数")
		return
	}

	req := common.SubmitPriceChangeRequest{
		StationID: *station,
		FuelCode:  *fuel,
		NewPrice:  *price,
	}

	resp, err := c.doRequest(http.MethodPost, "/price-requests", req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func reviewPriceRequest(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供价格调整请求ID")
		return
	}
	id := os.Args[2]
	fs := flag.NewFlagSet("price-review", flag.ExitOnError)
	approved := fs.Bool("approve", false, "是否通过")
	rejected := fs.Bool("reject", false, "是否拒绝")
	reviewer := fs.String("by", "admin", "审核人")
	fs.Parse(os.Args[3:])

	if !*approved && !*rejected {
		fmt.Println("请提供 -approve 或 -reject 参数")
		return
	}

	req := common.ReviewPriceChangeRequest{
		Approved:   *approved,
		ReviewedBy: *reviewer,
	}

	resp, err := c.doRequest(http.MethodPost, "/price-requests/"+id, req)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func getPriceHistory(c *Client) {
	fs := flag.NewFlagSet("price-history", flag.ExitOnError)
	station := fs.String("station", "", "加油站ID")
	fuel := fs.String("fuel", "", "油品编号")
	fs.Parse(os.Args[2:])

	if *station == "" || *fuel == "" {
		fmt.Println("请提供 -station 和 -fuel 参数")
		return
	}

	path := "/price-history?station_id=" + *station + "&fuel_code=" + *fuel
	resp, err := c.doRequest(http.MethodGet, path, nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func listRestockTodos(c *Client) {
	resp, err := c.doRequest(http.MethodGet, "/restock-todos", nil)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func completeRestockTodo(c *Client) {
	if len(os.Args) < 3 {
		fmt.Println("请提供补货待办ID")
		return
	}
	id := os.Args[2]
	fs := flag.NewFlagSet("restock-complete", flag.ExitOnError)
	stock := fs.Int64("stock", 5000, "新库存数量")
	fs.Parse(os.Args[3:])

	body := struct {
		NewStock int64 `json:"new_stock"`
	}{*stock}

	resp, err := c.doRequest(http.MethodPost, "/restock-todos/"+id, body)
	if err != nil {
		fmt.Println("错误:", err)
		return
	}
	printJSON(resp)
}

func init() {
	_ = time.Now()
}

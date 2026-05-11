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

	"locker/pkg/api"
)

const (
	defaultServer = "http://localhost:8080"
)

type Client struct {
	server string
}

func NewClient(server string) *Client {
	return &Client{server: server}
}

func (c *Client) sendRequest(method, endpoint string, body interface{}, resp interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.server+endpoint, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if len(data) > 0 {
			json.Unmarshal(data, &errResp)
		}
		if errResp.Error == "" {
			return fmt.Errorf("server returned status %d", res.StatusCode)
		}
		return fmt.Errorf("error: %s", errResp.Error)
	}

	if resp != nil && len(data) > 0 {
		return json.Unmarshal(data, resp)
	}
	return nil
}

func printUsage() {
	fmt.Println("智能快递柜管理系统客户端")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  locker-client <command> [options]")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  create-locker       创建快递柜")
	fmt.Println("  list-compartments   列出格口状态")
	fmt.Println("  create-courier      创建快递员")
	fmt.Println("  get-courier         查询快递员信息")
	fmt.Println("  recharge-courier    快递员充值")
	fmt.Println("  store-package       存件")
	fmt.Println("  pickup-package      取件")
	fmt.Println("  list-overdue        列出超时包裹")
	fmt.Println("  handle-overdue      处理超时包裹")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  locker-client create-locker --id L001 --address \"北京市朝阳区\" --small 10 --medium 5 --large 2 --start 07:00 --end 22:00")
	fmt.Println("  locker-client create-courier --id C001 --name \"张三\" --balance 100")
	fmt.Println("  locker-client store-package --locker L001 --courier C001 --phone 1234 --size small")
	fmt.Println("  locker-client pickup-package --locker L001 --code 123456")
}

func main() {
	server := flag.String("server", defaultServer, "服务端地址")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}

	client := NewClient(*server)
	command := strings.ToLower(args[0])

	switch command {
	case "create-locker":
		handleCreateLocker(client, args[1:])
	case "list-compartments":
		handleListCompartments(client, args[1:])
	case "create-courier":
		handleCreateCourier(client, args[1:])
	case "get-courier":
		handleGetCourier(client, args[1:])
	case "recharge-courier":
		handleRechargeCourier(client, args[1:])
	case "store-package":
		handleStorePackage(client, args[1:])
	case "pickup-package":
		handlePickupPackage(client, args[1:])
	case "list-overdue":
		handleListOverdue(client)
	case "handle-overdue":
		handleHandleOverdue(client, args[1:])
	default:
		fmt.Printf("未知命令: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func handleCreateLocker(c *Client, args []string) {
	fs := flag.NewFlagSet("create-locker", flag.ExitOnError)
	id := fs.String("id", "", "快递柜编号")
	address := fs.String("address", "", "安装地址")
	small := fs.Int("small", 0, "小格口数量")
	medium := fs.Int("medium", 0, "中格口数量")
	large := fs.Int("large", 0, "大格口数量")
	start := fs.String("start", "07:00", "营业时间开始")
	end := fs.String("end", "22:00", "营业时间结束")
	fs.Parse(args)

	if *id == "" || *address == "" {
		fmt.Println("错误: 必须提供 --id 和 --address 参数")
		os.Exit(1)
	}

	req := api.CreateLockerRequest{
		ID:            *id,
		Address:       *address,
		SmallCount:    *small,
		MediumCount:   *medium,
		LargeCount:    *large,
		BusinessStart: *start,
		BusinessEnd:   *end,
	}

	var resp api.SuccessResponse
	if err := c.sendRequest("POST", "/locker/create", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("快递柜创建成功: %s\n", *id)
}

func handleListCompartments(c *Client, args []string) {
	fs := flag.NewFlagSet("list-compartments", flag.ExitOnError)
	lockerID := fs.String("locker", "", "快递柜编号")
	fs.Parse(args)

	if *lockerID == "" {
		fmt.Println("错误: 必须提供 --locker 参数")
		os.Exit(1)
	}

	req := api.ListCompartmentsRequest{LockerID: *lockerID}
	var resp api.ListCompartmentsResponse
	if err := c.sendRequest("POST", "/locker/compartments", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("快递柜 %s 格口状态:\n", *lockerID)
	fmt.Println("----------------------------------------")
	for _, c := range resp.Compartments {
		fmt.Printf("格口 %s: 尺寸=%s, 状态=%s\n", c.ID, c.Size, c.Status)
		if c.Status != api.StatusFree {
			fmt.Printf("  快递员: %s, 收件人尾号: %s, 取件码: %s\n", c.CourierID, c.RecipientPhone, c.PickupCode)
			fmt.Printf("  存件时间: %s, 到期时间: %s\n", c.StoreTime, c.ExpireTime)
		}
	}
}

func handleCreateCourier(c *Client, args []string) {
	fs := flag.NewFlagSet("create-courier", flag.ExitOnError)
	id := fs.String("id", "", "快递员编号")
	name := fs.String("name", "", "姓名")
	balance := fs.Float64("balance", 0, "初始余额")
	fs.Parse(args)

	if *id == "" || *name == "" {
		fmt.Println("错误: 必须提供 --id 和 --name 参数")
		os.Exit(1)
	}

	req := api.CreateCourierRequest{
		ID:      *id,
		Name:    *name,
		Balance: *balance,
	}

	var resp api.SuccessResponse
	if err := c.sendRequest("POST", "/courier/create", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("快递员创建成功: %s (%s), 余额: %.2f 元\n", *id, *name, *balance)
}

func handleGetCourier(c *Client, args []string) {
	fs := flag.NewFlagSet("get-courier", flag.ExitOnError)
	id := fs.String("id", "", "快递员编号")
	fs.Parse(args)

	if *id == "" {
		fmt.Println("错误: 必须提供 --id 参数")
		os.Exit(1)
	}

	req := api.GetCourierRequest{ID: *id}
	var resp api.CourierInfo
	if err := c.sendRequest("POST", "/courier/get", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("快递员信息:\n")
	fmt.Printf("  编号: %s\n", resp.ID)
	fmt.Printf("  姓名: %s\n", resp.Name)
	fmt.Printf("  余额: %.2f 元\n", resp.Balance)
}

func handleRechargeCourier(c *Client, args []string) {
	fs := flag.NewFlagSet("recharge-courier", flag.ExitOnError)
	id := fs.String("id", "", "快递员编号")
	amount := fs.Float64("amount", 0, "充值金额")
	fs.Parse(args)

	if *id == "" || *amount <= 0 {
		fmt.Println("错误: 必须提供 --id 和 --amount 参数")
		os.Exit(1)
	}

	req := api.RechargeCourierRequest{ID: *id, Amount: *amount}
	var resp api.CourierInfo
	if err := c.sendRequest("POST", "/courier/recharge", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("充值成功! 当前余额: %.2f 元\n", resp.Balance)
}

func handleStorePackage(c *Client, args []string) {
	fs := flag.NewFlagSet("store-package", flag.ExitOnError)
	lockerID := fs.String("locker", "", "快递柜编号")
	courierID := fs.String("courier", "", "快递员编号")
	phone := fs.String("phone", "", "收件人手机号后四位")
	size := fs.String("size", "", "格口尺寸 (small/medium/large)")
	fs.Parse(args)

	if *lockerID == "" || *courierID == "" || *phone == "" || *size == "" {
		fmt.Println("错误: 必须提供 --locker, --courier, --phone, --size 参数")
		os.Exit(1)
	}

	req := api.StorePackageRequest{
		LockerID:       *lockerID,
		CourierID:      *courierID,
		RecipientPhone: *phone,
		Size:           api.LockerSize(*size),
	}

	var resp api.StorePackageResponse
	if err := c.sendRequest("POST", "/package/store", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("存件成功!\n")
	fmt.Printf("  格口编号: %s\n", resp.CompartmentID)
	fmt.Printf("  取件码: %s\n", resp.PickupCode)
}

func handlePickupPackage(c *Client, args []string) {
	fs := flag.NewFlagSet("pickup-package", flag.ExitOnError)
	lockerID := fs.String("locker", "", "快递柜编号")
	code := fs.String("code", "", "取件码")
	fs.Parse(args)

	if *lockerID == "" || *code == "" {
		fmt.Println("错误: 必须提供 --locker 和 --code 参数")
		os.Exit(1)
	}

	req := api.PickupPackageRequest{
		LockerID:   *lockerID,
		PickupCode: *code,
	}

	var resp api.PickupPackageResponse
	if err := c.sendRequest("POST", "/package/pickup", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("取件成功! 格口 %s 已打开\n", resp.CompartmentID)
}

func handleListOverdue(c *Client) {
	var resp api.ListOverdueResponse
	if err := c.sendRequest("POST", "/overdue/list", nil, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if len(resp.OverduePackages) == 0 {
		fmt.Println("当前没有超时包裹")
		return
	}

	fmt.Println("超时包裹列表:")
	fmt.Println("----------------------------------------")
	for _, p := range resp.OverduePackages {
		fmt.Printf("快递柜: %s, 格口: %s\n", p.LockerID, p.CompartmentID)
		fmt.Printf("  快递员: %s, 收件人尾号: %s\n", p.CourierID, p.RecipientPhone)
		fmt.Printf("  存件时间: %s, 到期时间: %s\n", p.StoreTime, p.ExpireTime)
	}
}

func handleHandleOverdue(c *Client, args []string) {
	fs := flag.NewFlagSet("handle-overdue", flag.ExitOnError)
	lockerID := fs.String("locker", "", "快递柜编号")
	compartmentID := fs.String("compartment", "", "格口编号")
	courierID := fs.String("courier", "", "快递员编号")
	fs.Parse(args)

	if *lockerID == "" || *compartmentID == "" || *courierID == "" {
		fmt.Println("错误: 必须提供 --locker, --compartment, --courier 参数")
		os.Exit(1)
	}

	req := api.HandleOverdueRequest{
		LockerID:      *lockerID,
		CompartmentID: *compartmentID,
		CourierID:     *courierID,
	}

	var resp api.SuccessResponse
	if err := c.sendRequest("POST", "/overdue/handle", req, &resp); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("超时包裹处理成功!")
}

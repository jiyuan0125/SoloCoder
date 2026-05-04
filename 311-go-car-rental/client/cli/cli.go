package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"carrental/client/api"
	"carrental/common"
)

type CLI struct {
	client *api.Client
	reader *bufio.Reader
}

func NewCLI(client *api.Client) *CLI {
	return &CLI{
		client: client,
		reader: bufio.NewReader(os.Stdin),
	}
}

func (c *CLI) prompt(prompt string) string {
	fmt.Print(prompt)
	input, _ := c.reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func (c *CLI) promptFloat(prompt string) (float64, error) {
	input := c.prompt(prompt)
	return strconv.ParseFloat(input, 64)
}

func (c *CLI) Run() {
	fmt.Println("===== 汽车租赁管理系统客户端 =====")
	fmt.Println("1. 工作人员功能")
	fmt.Println("2. 用户功能")
	fmt.Println("0. 退出")

	choice := c.prompt("请选择: ")

	switch choice {
	case "1":
		c.runStaffMenu()
	case "2":
		c.runUserMenu()
	case "0":
		return
	default:
		fmt.Println("无效选择")
	}
}

func (c *CLI) runStaffMenu() {
	for {
		fmt.Println("\n----- 工作人员功能 -----")
		fmt.Println("1. 录入车辆")
		fmt.Println("2. 查看车辆列表")
		fmt.Println("3. 更新车辆状态")
		fmt.Println("4. 处理还车")
		fmt.Println("5. 查看所有订单")
		fmt.Println("0. 返回主菜单")

		choice := c.prompt("请选择: ")

		switch choice {
		case "1":
			c.createCar()
		case "2":
			c.listCars()
		case "3":
			c.updateCarStatus()
		case "4":
			c.returnCar()
		case "5":
			c.listOrders("")
		case "0":
			return
		default:
			fmt.Println("无效选择")
		}
	}
}

func (c *CLI) runUserMenu() {
	for {
		fmt.Println("\n----- 用户功能 -----")
		fmt.Println("1. 查看可用车辆")
		fmt.Println("2. 检查车辆可用性")
		fmt.Println("3. 提交租车订单")
		fmt.Println("4. 查看我的订单")
		fmt.Println("5. 查看订单详情")
		fmt.Println("0. 返回主菜单")

		choice := c.prompt("请选择: ")

		switch choice {
		case "1":
			c.listCars()
		case "2":
			c.checkAvailability()
		case "3":
			c.createOrder()
		case "4":
			userPhone := c.prompt("请输入您的手机号: ")
			c.listOrders(userPhone)
		case "5":
			c.getOrder()
		case "0":
			return
		default:
			fmt.Println("无效选择")
		}
	}
}

func (c *CLI) createCar() {
	fmt.Println("\n----- 录入车辆 -----")
	brand := c.prompt("品牌: ")
	if brand == "" {
		fmt.Println("品牌不能为空")
		return
	}

	model := c.prompt("车型: ")
	if model == "" {
		fmt.Println("车型不能为空")
		return
	}

	dailyRate, err := c.promptFloat("日租金: ")
	if err != nil {
		fmt.Println("无效的日租金")
		return
	}

	deposit, err := c.promptFloat("押金金额: ")
	if err != nil {
		fmt.Println("无效的押金金额")
		return
	}

	req := &common.CreateCarRequest{
		Brand:     brand,
		Model:     model,
		DailyRate: dailyRate,
		Deposit:   deposit,
	}

	car, err := c.client.CreateCar(req)
	if err != nil {
		fmt.Printf("录入失败: %v\n", err)
		return
	}

	fmt.Printf("车辆录入成功!\n")
	c.printCar(car)
}

func (c *CLI) listCars() {
	fmt.Println("\n----- 车辆列表 -----")
	cars, err := c.client.ListCars()
	if err != nil {
		fmt.Printf("获取车辆列表失败: %v\n", err)
		return
	}

	if len(cars) == 0 {
		fmt.Println("暂无车辆")
		return
	}

	for _, car := range cars {
		c.printCar(car)
		fmt.Println("------------------")
	}
}

func (c *CLI) printCar(car *common.Car) {
	fmt.Printf("ID: %s\n", car.ID)
	fmt.Printf("品牌: %s\n", car.Brand)
	fmt.Printf("车型: %s\n", car.Model)
	fmt.Printf("日租金: %.2f\n", car.DailyRate)
	fmt.Printf("押金: %.2f\n", car.Deposit)
	fmt.Printf("状态: %s\n", car.Status)
}

func (c *CLI) updateCarStatus() {
	fmt.Println("\n----- 更新车辆状态 -----")
	carID := c.prompt("车辆ID: ")

	fmt.Println("可选状态:")
	fmt.Println("1. available (可用)")
	fmt.Println("2. maintenance (维护中)")
	statusChoice := c.prompt("请选择状态: ")

	var status common.CarStatus
	switch statusChoice {
	case "1":
		status = common.CarStatusAvailable
	case "2":
		status = common.CarStatusMaintenance
	default:
		fmt.Println("无效选择")
		return
	}

	car, err := c.client.UpdateCarStatus(carID, status)
	if err != nil {
		fmt.Printf("更新失败: %v\n", err)
		return
	}

	fmt.Println("状态更新成功!")
	c.printCar(car)
}

func (c *CLI) checkAvailability() {
	fmt.Println("\n----- 检查车辆可用性 -----")
	carID := c.prompt("车辆ID: ")
	pickupDate := c.prompt("取车日期 (YYYY-MM-DD): ")
	returnDate := c.prompt("还车日期 (YYYY-MM-DD): ")

	available, err := c.client.CheckAvailability(carID, pickupDate, returnDate)
	if err != nil {
		fmt.Printf("检查失败: %v\n", err)
		return
	}

	if available {
		fmt.Println("该车辆在所选日期可用")
	} else {
		fmt.Println("该车辆在所选日期不可用")
	}
}

func (c *CLI) createOrder() {
	fmt.Println("\n----- 提交租车订单 -----")
	carID := c.prompt("车辆ID: ")
	userName := c.prompt("您的姓名: ")
	userPhone := c.prompt("您的手机号: ")
	pickupDateStr := c.prompt("取车日期 (YYYY-MM-DD): ")
	returnDateStr := c.prompt("还车日期 (YYYY-MM-DD): ")

	if userName == "" || userPhone == "" {
		fmt.Println("姓名和手机号不能为空")
		return
	}

	pickupDate, err := common.ParseDate(pickupDateStr)
	if err != nil {
		fmt.Println("无效的取车日期")
		return
	}

	returnDate, err := common.ParseDate(returnDateStr)
	if err != nil {
		fmt.Println("无效的还车日期")
		return
	}

	req := &common.CreateOrderRequest{
		CarID:      carID,
		UserName:   userName,
		UserPhone:  userPhone,
		PickupDate: pickupDate,
		ReturnDate: returnDate,
	}

	order, err := c.client.CreateOrder(req)
	if err != nil {
		fmt.Printf("订单提交失败: %v\n", err)
		return
	}

	fmt.Println("订单提交成功!")
	c.printOrder(order)
}

func (c *CLI) getOrder() {
	orderID := c.prompt("订单ID: ")

	order, err := c.client.GetOrder(orderID)
	if err != nil {
		fmt.Printf("获取订单失败: %v\n", err)
		return
	}

	c.printOrderDetail(order)
}

func (c *CLI) listOrders(userPhone string) {
	fmt.Println("\n----- 订单列表 -----")
	orders, err := c.client.ListOrders(userPhone)
	if err != nil {
		fmt.Printf("获取订单列表失败: %v\n", err)
		return
	}

	if len(orders) == 0 {
		fmt.Println("暂无订单")
		return
	}

	for _, order := range orders {
		c.printOrder(order)
		fmt.Println("------------------")
	}
}

func (c *CLI) printOrder(order *common.Order) {
	fmt.Printf("订单ID: %s\n", order.ID)
	fmt.Printf("车辆ID: %s\n", order.CarID)
	fmt.Printf("用户: %s (%s)\n", order.UserName, order.UserPhone)
	fmt.Printf("取车日期: %s\n", common.FormatDate(order.PickupDate))
	fmt.Printf("还车日期: %s\n", common.FormatDate(order.ReturnDate))
	fmt.Printf("原始总费用: %.2f\n", order.OriginalTotalCost)
	fmt.Printf("实际总费用: %.2f\n", order.TotalCost)
	fmt.Printf("押金: %.2f\n", order.Deposit)
	fmt.Printf("状态: %s\n", order.Status)
}

func (c *CLI) printOrderDetail(order *common.Order) {
	fmt.Println("\n===== 订单详情 =====")
	c.printOrder(order)
	if order.ActualReturnDate != nil {
		fmt.Printf("实际还车日期: %s\n", common.FormatDate(*order.ActualReturnDate))
	}
	fmt.Printf("押金扣款: %.2f\n", order.DepositDeduction)
	fmt.Printf("行驶里程: %.2f\n", order.Mileage)
	fmt.Printf("车况: %s\n", order.Condition)
	if order.ViolationOrDamage {
		fmt.Printf("违章/损坏金额: %.2f\n", order.ViolationAmount)
	}
}

func (c *CLI) returnCar() {
	fmt.Println("\n----- 处理还车 -----")
	orderID := c.prompt("订单ID: ")
	actualReturnDateStr := c.prompt("实际还车日期 (YYYY-MM-DD): ")
	mileage, _ := c.promptFloat("行驶里程: ")
	condition := c.prompt("车况描述: ")

	hasViolation := c.prompt("是否有违章或损坏? (y/n): ")
	var violationAmount float64
	if strings.ToLower(hasViolation) == "y" || strings.ToLower(hasViolation) == "yes" {
		violationAmount, _ = c.promptFloat("违章/损坏金额: ")
	}

	actualReturnDate, err := common.ParseDate(actualReturnDateStr)
	if err != nil {
		fmt.Println("无效的日期格式")
		return
	}

	req := &common.ReturnCarRequest{
		OrderID:           orderID,
		ActualReturnDate:  actualReturnDate,
		Mileage:           mileage,
		Condition:         condition,
		ViolationOrDamage: violationAmount > 0 || strings.ToLower(hasViolation) == "y",
		ViolationAmount:   violationAmount,
	}

	order, err := c.client.ReturnCar(req)
	if err != nil {
		fmt.Printf("还车处理失败: %v\n", err)
		return
	}

	fmt.Println("还车处理成功!")
	c.printOrderDetail(order)
}

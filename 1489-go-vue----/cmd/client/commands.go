package main

import (
	"flag"
	"fmt"

	"lockservice/pkg/common"
)

func cmdCreateOrder(args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	userID := fs.String("user", "u1", "user id")
	address := fs.String("address", "北京市朝阳区望京", "service address")
	lockType := fs.String("lock", "security", "lock type: security|interior|password|fingerprint|car")
	urgency := fs.String("urgency", "normal", "urgency: normal|urgent")
	slot := fs.String("slot", "morning", "time slot: morning|afternoon|evening")
	date := fs.String("date", "2026-05-11", "date YYYY-MM-DD")
	fs.Parse(args)

	client := NewClient(getServerURL())
	req := common.CreateOrderReq{
		UserID:       *userID,
		Address:      *address,
		LockType:     *lockType,
		Urgency:      *urgency,
		TimeSlot:     *slot,
		TimeSlotDate: *date,
	}

	var resp common.CreateOrderResp
	if err := client.post("/order/create", req, &resp); err != nil {
		return err
	}

	if !resp.Success {
		fmt.Printf("创建订单失败: %s\n", resp.Message)
		if resp.Alternative != nil {
			if len(resp.Alternative.OtherSlots) > 0 {
				fmt.Printf("可选时段: %v\n", resp.Alternative.OtherSlots)
			}
			if len(resp.Alternative.OtherMasters) > 0 {
				fmt.Printf("可选师傅:\n")
				for _, m := range resp.Alternative.OtherMasters {
					fmt.Printf("  - %s (%s) 可用时段: %v\n", m.Name, m.ID, m.AvailableSlots)
				}
			}
		}
		return nil
	}

	fmt.Printf("订单创建成功!\n")
	fmt.Printf("订单ID: %s\n", resp.OrderID)
	printOrder(resp.Order)
	return nil
}

func cmdAcceptOrder(args []string) error {
	fs := flag.NewFlagSet("accept", flag.ExitOnError)
	orderID := fs.String("order", "", "order id")
	fs.Parse(args)

	if *orderID == "" {
		return fmt.Errorf("order id required")
	}

	client := NewClient(getServerURL())
	req := common.AcceptOrderReq{OrderID: *orderID}
	var resp common.AcceptOrderResp
	if err := client.post("/order/accept", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Println("接单成功")
	} else {
		fmt.Printf("接单失败: %s\n", resp.Message)
	}
	return nil
}

func cmdStartService(args []string) error {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	orderID := fs.String("order", "", "order id")
	fs.Parse(args)

	if *orderID == "" {
		return fmt.Errorf("order id required")
	}

	client := NewClient(getServerURL())
	req := common.StartServiceReq{OrderID: *orderID}
	var resp common.StartServiceResp
	if err := client.post("/order/start", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Println("开始服务")
	} else {
		fmt.Printf("开始服务失败: %s\n", resp.Message)
	}
	return nil
}

func cmdCompleteService(args []string) error {
	fs := flag.NewFlagSet("complete", flag.ExitOnError)
	orderID := fs.String("order", "", "order id")
	needReplace := fs.Bool("replace", false, "need replace lock")
	brand := fs.String("brand", "", "lock brand")
	model := fs.String("model", "", "lock model")
	level := fs.String("level", "", "lock level")
	partsCost := fs.Float64("parts", 0, "parts cost")
	fs.Parse(args)

	if *orderID == "" {
		return fmt.Errorf("order id required")
	}

	client := NewClient(getServerURL())
	req := common.CompleteServiceReq{
		OrderID:         *orderID,
		NeedReplaceLock: *needReplace,
		LockBrand:       *brand,
		LockModel:       *model,
		LockLevel:       *level,
		PartsCost:       *partsCost,
	}
	var resp common.CompleteServiceResp
	if err := client.post("/order/complete", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Println("服务完成")
		printOrder(resp.Order)
	} else {
		fmt.Printf("完成服务失败: %s\n", resp.Message)
	}
	return nil
}

func cmdPayOrder(args []string) error {
	fs := flag.NewFlagSet("pay", flag.ExitOnError)
	orderID := fs.String("order", "", "order id")
	fs.Parse(args)

	if *orderID == "" {
		return fmt.Errorf("order id required")
	}

	client := NewClient(getServerURL())
	req := common.PayOrderReq{OrderID: *orderID}
	var resp common.PayOrderResp
	if err := client.post("/order/pay", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Println("支付成功")
	} else {
		fmt.Printf("支付失败: %s\n", resp.Message)
	}
	return nil
}

func cmdRateOrder(args []string) error {
	fs := flag.NewFlagSet("rate", flag.ExitOnError)
	orderID := fs.String("order", "", "order id")
	rating := fs.Int("rating", 5, "rating 1-5")
	comment := fs.String("comment", "", "comment")
	fs.Parse(args)

	if *orderID == "" {
		return fmt.Errorf("order id required")
	}

	client := NewClient(getServerURL())
	req := common.RateOrderReq{
		OrderID: *orderID,
		Rating:  *rating,
		Comment: *comment,
	}
	var resp common.RateOrderResp
	if err := client.post("/order/rate", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Printf("评价成功 (评分: %d)\n", *rating)
		if resp.HasComplaint {
			fmt.Println("已自动生成投诉工单")
		}
	} else {
		fmt.Printf("评价失败: %s\n", resp.Message)
	}
	return nil
}

func cmdGetOrder(args []string) error {
	fs := flag.NewFlagSet("get", flag.ExitOnError)
	orderID := fs.String("order", "", "order id")
	fs.Parse(args)

	if *orderID == "" {
		return fmt.Errorf("order id required")
	}

	client := NewClient(getServerURL())
	var resp common.GetOrderResp
	if err := client.get("/order/get?id="+*orderID, &resp); err != nil {
		return err
	}

	if resp.Success {
		printOrder(resp.Order)
	} else {
		fmt.Printf("获取订单失败: %s\n", resp.Message)
	}
	return nil
}

func cmdListOrders(args []string) error {
	client := NewClient(getServerURL())
	var resp common.ListOrdersResp
	if err := client.get("/order/list", &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	fmt.Printf("共 %d 个订单\n", len(resp.Orders))
	for i, o := range resp.Orders {
		fmt.Printf("--- 订单 %d ---\n", i+1)
		printOrder(o)
	}
	return nil
}

func cmdListMasters(args []string) error {
	client := NewClient(getServerURL())
	var resp common.ListMastersResp
	if err := client.get("/master/list", &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Message)
	}

	fmt.Printf("共 %d 位师傅\n", len(resp.Masters))
	for i, m := range resp.Masters {
		fmt.Printf("--- 师傅 %d ---\n", i+1)
		fmt.Printf("ID: %s\n", m.ID)
		fmt.Printf("姓名: %s\n", m.Name)
		fmt.Printf("电话: %s\n", m.Phone)
		fmt.Printf("服务区域: %s\n", m.ServiceArea)
		fmt.Printf("可开锁型: %v\n", m.SupportedLocks)
		fmt.Printf("状态: %s\n", m.Status)
		fmt.Printf("位置: %s\n", m.Location)
	}
	return nil
}

func cmdAddMaster(args []string) error {
	fs := flag.NewFlagSet("add-master", flag.ExitOnError)
	id := fs.String("id", "", "master id")
	name := fs.String("name", "", "master name")
	phone := fs.String("phone", "", "master phone")
	area := fs.String("area", "", "service area")
	locks := fs.String("locks", "", "supported locks (comma separated)")
	location := fs.String("location", "", "location")
	fs.Parse(args)

	if *id == "" || *name == "" {
		return fmt.Errorf("id and name required")
	}

	var supportedLocks []string
	if *locks != "" {
		supportedLocks = splitLocks(*locks)
	}

	client := NewClient(getServerURL())
	req := common.AddMasterReq{
		ID:             *id,
		Name:           *name,
		Phone:          *phone,
		ServiceArea:    *area,
		SupportedLocks: supportedLocks,
		Location:       *location,
	}
	var resp common.AddMasterResp
	if err := client.post("/master/add", req, &resp); err != nil {
		return err
	}

	if resp.Success {
		fmt.Println("师傅添加成功")
	} else {
		fmt.Printf("添加失败: %s\n", resp.Message)
	}
	return nil
}

func splitLocks(s string) []string {
	var result []string
	start := 0
	for i, c := range s {
		if c == ',' {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

func printOrder(o *common.OrderView) {
	if o == nil {
		return
	}
	fmt.Printf("订单ID: %s\n", o.ID)
	fmt.Printf("用户: %s\n", o.UserID)
	fmt.Printf("地址: %s\n", o.Address)
	fmt.Printf("锁型: %s\n", o.LockType)
	fmt.Printf("紧急程度: %s\n", o.Urgency)
	fmt.Printf("时段: %s (%s)\n", o.TimeSlot, o.TimeSlotDate)
	fmt.Printf("师傅ID: %s\n", o.MasterID)
	fmt.Printf("状态: %s\n", o.Status)
	fmt.Printf("基础费用: %.2f\n", o.BaseCost)
	if o.UrgentFee > 0 {
		fmt.Printf("加急费: %.2f\n", o.UrgentFee)
	}
	if o.Detail.NeedReplaceLock {
		fmt.Printf("换锁配件费: %.2f\n", o.Detail.PartsCost)
		fmt.Printf("安装费: %.2f\n", o.Detail.LaborCost)
	}
	fmt.Printf("总费用: %.2f\n", o.TotalCost)
	fmt.Printf("创建时间: %s\n", o.CreatedAt)
	if o.Rating > 0 {
		fmt.Printf("评分: %d\n", o.Rating)
		if o.Comment != "" {
			fmt.Printf("评价: %s\n", o.Comment)
		}
		if o.HasComplaint {
			fmt.Println("已投诉")
		}
	}
}

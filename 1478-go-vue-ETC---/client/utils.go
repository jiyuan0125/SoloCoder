package main

import (
	"etc-system/common"
	"fmt"
	"os"
)

func writeFile(filename string, data []byte) error {
	return os.WriteFile(filename, data, 0644)
}

func PrintAccount(account *common.Account) {
	fmt.Println("=== 账户信息 ===")
	fmt.Printf("车牌号: %s\n", account.LicensePlate)
	fmt.Printf("银行卡: %s\n", account.BankCard)
	fmt.Printf("余额: %.2f 元\n", account.Balance)
	fmt.Printf("状态: %s\n", accountStatusText(account.Status))
	if account.OweAmount > 0 {
		fmt.Printf("欠费金额: %.2f 元\n", account.OweAmount)
	}
	fmt.Printf("车辆类型: %s\n", vehicleTypeText(account.VehicleType))
	if account.VehicleType == common.Passenger {
		fmt.Printf("座位数: %d\n", account.Seats)
	} else {
		fmt.Printf("载重量: %.1f 吨\n", account.LoadWeight)
	}
	fmt.Printf("创建时间: %s\n", account.CreatedAt.Format("2006-01-02 15:04:05"))
}

func PrintPassRecord(record *common.PassRecord) {
	fmt.Println("=== 通行记录 ===")
	fmt.Printf("车牌号: %s\n", record.LicensePlate)
	fmt.Printf("入口站: %s\n", record.EntryStation)
	fmt.Printf("出口站: %s\n", record.ExitStation)
	fmt.Printf("通行时间: %s\n", record.PassDate.Format("2006-01-02 15:04:05"))
	fmt.Printf("行驶里程: %.2f 公里\n", record.Mileage)
	fmt.Printf("通行费用: %.2f 元\n", record.Fee)
	fmt.Printf("扣费状态: %s\n", paymentStatusText(record.PaymentStatus))
}

func accountStatusText(status common.AccountStatus) string {
	switch status {
	case common.StatusNormal:
		return "正常"
	case common.StatusOwe:
		return "欠费"
	case common.StatusBlacklist:
		return "黑名单"
	default:
		return string(status)
	}
}

func vehicleTypeText(typ common.VehicleType) string {
	switch typ {
	case common.Passenger:
		return "客车"
	case common.Truck:
		return "货车"
	default:
		return "未知"
	}
}

func paymentStatusText(status common.PaymentStatus) string {
	switch status {
	case common.PaymentSuccess:
		return "已扣费"
	case common.PaymentOwe:
		return "欠费"
	case common.PaymentFree:
		return "免费"
	default:
		return string(status)
	}
}

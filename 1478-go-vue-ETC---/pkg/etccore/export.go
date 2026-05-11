package etccore

import (
	"etc-system/common"
	"fmt"
	"strings"
)

const (
	UTF8BOM = "\xEF\xBB\xBF"
)

func ExportToCSV(records []common.PassRecord) string {
	var builder strings.Builder

	builder.WriteString(UTF8BOM)
	builder.WriteString("车牌号,入口站,出口站,通行日期,里程(公里),费用(元),扣费状态\n")

	for _, record := range records {
		feeStr := fmt.Sprintf("%.2f", record.Fee)
		mileageStr := fmt.Sprintf("%.2f", record.Mileage)
		passDateStr := record.PassDate.Format("2006-01-02 15:04:05")

		statusText := ""
		switch record.PaymentStatus {
		case common.PaymentSuccess:
			statusText = "已扣费"
		case common.PaymentOwe:
			statusText = "欠费"
		case common.PaymentFree:
			statusText = "免费"
		}

		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s\n",
			record.LicensePlate,
			record.EntryStation,
			record.ExitStation,
			passDateStr,
			mileageStr,
			feeStr,
			statusText,
		)
		builder.WriteString(line)
	}

	return builder.String()
}

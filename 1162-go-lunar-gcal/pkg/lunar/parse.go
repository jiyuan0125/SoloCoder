package lunar

import (
	"fmt"
	"strconv"
	"strings"
)

var monthNameToNumber = map[string]int{
	"正月": 1, "一月": 1, "1月": 1,
	"二月": 2, "2月": 2,
	"三月": 3, "3月": 3,
	"四月": 4, "4月": 4,
	"五月": 5, "5月": 5,
	"六月": 6, "6月": 6,
	"七月": 7, "7月": 7,
	"八月": 8, "8月": 8,
	"九月": 9, "9月": 9,
	"十月": 10, "10月": 10,
	"冬月": 11, "十一月": 11, "11月": 11,
	"腊月": 12, "十二月": 12, "12月": 12,
}

var dayNameToNumber = map[string]int{
	"初一": 1, "初二": 2, "初三": 3, "初四": 4, "初五": 5,
	"初六": 6, "初七": 7, "初八": 8, "初九": 9, "初十": 10,
	"十一": 11, "十二": 12, "十三": 13, "十四": 14, "十五": 15,
	"十六": 16, "十七": 17, "十八": 18, "十九": 19, "二十": 20,
	"廿一": 21, "廿二": 22, "廿三": 23, "廿四": 24, "廿五": 25,
	"廿六": 26, "廿七": 27, "廿八": 28, "廿九": 29, "三十": 30,
}

func ParseLunarDate(yearStr, monthStr, dayStr string) (*LunarDate, error) {
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return nil, fmt.Errorf("年份格式错误: %s", yearStr)
	}

	isLeap := false
	monthStr = strings.TrimSpace(monthStr)
	originalMonthStr := monthStr

	if strings.HasPrefix(monthStr, "闰") {
		isLeap = true
		monthStr = strings.TrimPrefix(monthStr, "闰")
	}

	month, ok := monthNameToNumber[monthStr]
	if !ok {
		if n, err := strconv.Atoi(strings.TrimSuffix(monthStr, "月")); err == nil {
			if n >= 1 && n <= 12 {
				month = n
			} else {
				return nil, fmt.Errorf("月份超出范围: %s", originalMonthStr)
			}
		} else {
			return nil, fmt.Errorf("无法识别的月份: %s", originalMonthStr)
		}
	}

	dayStr = strings.TrimSpace(dayStr)
	day, ok := dayNameToNumber[dayStr]
	if !ok {
		if n, err := strconv.Atoi(dayStr); err == nil {
			if n >= 1 && n <= 30 {
				day = n
			} else {
				return nil, fmt.Errorf("日期超出范围: %s", dayStr)
			}
		} else {
			return nil, fmt.Errorf("无法识别的日期: %s", dayStr)
		}
	}

	return &LunarDate{
		Year:        year,
		Month:       month,
		Day:         day,
		IsLeapMonth: isLeap,
	}, nil
}

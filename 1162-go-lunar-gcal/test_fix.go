//go:build ignore

package main

import (
	"fmt"
	"time"

	"go-lunar-gcal/pkg/lunar"
)

func main() {
	fmt.Println("=== 测试1: 公历2024-02-10 转农历 ===")
	date := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	l, err := lunar.SolarToLunar(date)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("公历 2024-02-10 -> 农历: %s\n", l.String())
		fmt.Printf("  年: %d, 月: %d, 日: %d, 闰月: %v\n", l.Year, l.Month, l.Day, l.IsLeapMonth)
		fmt.Printf("  预期: 甲辰年(2024) 正月 初一\n")
	}

	fmt.Println()
	fmt.Println("=== 测试2: 2023年闰月检查 ===")
	lunar2023 := lunar.LunarDate{Year: 2023, Month: 2, Day: 1, IsLeapMonth: true}
	_, err = lunar.LunarToSolar(lunar2023)
	if err != nil {
		fmt.Println("错误:", err)
		fmt.Println("预期: 2023年(癸卯年)应该有闰二月")
	} else {
		fmt.Println("2023年有闰二月 ✓")
	}

	fmt.Println()
	fmt.Println("=== 测试3: 2024年闰月检查 ===")
	lunar2024 := lunar.LunarDate{Year: 2024, Month: 2, Day: 1, IsLeapMonth: true}
	_, err = lunar.LunarToSolar(lunar2024)
	if err != nil {
		fmt.Println("错误:", err)
		fmt.Println("预期: 2024年(甲辰年)不应该有闰月 ✓")
	} else {
		fmt.Println("2024年有闰二月 (预期是没有) ✗")
	}

	fmt.Println()
	fmt.Println("=== 测试4: 农历2024年正月初一转公历 ===")
	lunarDate := lunar.LunarDate{
		Year:        2024,
		Month:       1,
		Day:         1,
		IsLeapMonth: false,
	}
	solar, err := lunar.LunarToSolar(lunarDate)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("农历 2024年正月初一 -> 公历: %s\n", solar.Format("2006-01-02"))
		fmt.Printf("预期: 2024-02-10\n")
	}

	fmt.Println()
	fmt.Println("=== 测试5: 公历2023-03-22 转农历 (2023年闰二月初一) ===")
	date2 := time.Date(2023, 3, 22, 0, 0, 0, 0, time.UTC)
	l2, err := lunar.SolarToLunar(date2)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("公历 2023-03-22 -> 农历: %s\n", l2.String())
		fmt.Printf("  年: %d, 月: %d, 日: %d, 闰月: %v\n", l2.Year, l2.Month, l2.Day, l2.IsLeapMonth)
		fmt.Printf("预期: 癸卯年 闰二月 初一\n")
	}

	fmt.Println()
	fmt.Println("=== 测试6: 农历2023年闰二月初一转公历 ===")
	lunarDate2 := lunar.LunarDate{
		Year:        2023,
		Month:       2,
		Day:         1,
		IsLeapMonth: true,
	}
	solar2, err := lunar.LunarToSolar(lunarDate2)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("农历 2023年闰二月初一 -> 公历: %s\n", solar2.Format("2006-01-02"))
		fmt.Printf("预期: 2023-03-22\n")
	}
}

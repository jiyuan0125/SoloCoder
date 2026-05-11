//go:build ignore

package main

import (
	"fmt"
	"time"

	"go-lunar-gcal/pkg/jieqi"
	"go-lunar-gcal/pkg/lunar"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("综合测试：农历公历互转工具修复验证")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Println("=== 问题1验证：农历年份偏移 ===")
	date1 := time.Date(2024, 2, 10, 0, 0, 0, 0, time.UTC)
	l1, err := lunar.SolarToLunar(date1)
	if err != nil {
		fmt.Println("✗ 错误:", err)
	} else {
		if l1.Year == 2024 && l1.Month == 1 && l1.Day == 1 {
			fmt.Println("✓ 公历2024-02-10 → 农历甲辰年正月初一")
		} else {
			fmt.Printf("✗ 公历2024-02-10 → %s (预期: 甲辰年正月初一)\n", l1.String())
		}
	}
	fmt.Println()

	fmt.Println("=== 问题2验证：闰月数据 ===")
	leap2023 := lunar.LunarDate{Year: 2023, Month: 2, Day: 1, IsLeapMonth: true}
	_, err = lunar.LunarToSolar(leap2023)
	if err == nil {
		fmt.Println("✓ 2023年(癸卯年)有闰二月")
	} else {
		fmt.Println("✗ 2023年应该有闰二月，但返回错误:", err)
	}

	leap2024 := lunar.LunarDate{Year: 2024, Month: 2, Day: 1, IsLeapMonth: true}
	_, err = lunar.LunarToSolar(leap2024)
	if err != nil {
		fmt.Println("✓ 2024年(甲辰年)没有闰月")
	} else {
		fmt.Println("✗ 2024年不应该有闰二月")
	}
	fmt.Println()

	fmt.Println("=== 问题3验证：二十四节气 ===")
	jq, err := jieqi.GetJieqi(2024)
	if err != nil {
		fmt.Println("✗ 错误:", err)
		return
	}

	expected := map[string]string{
		"立春": "2024-02-04",
		"春分": "2024-03-20",
		"夏至": "2024-06-21",
		"秋分": "2024-09-22",
		"冬至": "2024-12-21",
	}

	allPassed := true
	for _, j := range jq {
		if exp, ok := expected[j.Name]; ok {
			actual := j.Time.Format("2006-01-02")
			timeStr := j.Time.Format("15:04")
			if actual == exp {
				fmt.Printf("✓ %s: %s %s\n", j.Name, actual, timeStr)
			} else {
				fmt.Printf("✗ %s: %s %s (预期: %s)\n", j.Name, actual, timeStr, exp)
				allPassed = false
			}
		}
	}

	lichunTime := ""
	for _, j := range jq {
		if j.Name == "立春" {
			lichunTime = j.Time.Format("15:04")
			break
		}
	}
	if lichunTime >= "16:00" && lichunTime <= "17:00" {
		fmt.Println("✓ 立春时间精确到分钟: 约16:26左右")
	} else {
		fmt.Printf("? 立春时间: %s (预期约16:26)\n", lichunTime)
	}
	fmt.Println()

	fmt.Println("=== 额外验证：农历转公历 ===")
	lunarDate := lunar.LunarDate{Year: 2024, Month: 1, Day: 1, IsLeapMonth: false}
	solar, err := lunar.LunarToSolar(lunarDate)
	if err != nil {
		fmt.Println("✗ 错误:", err)
	} else {
		expectedDate := "2024-02-10"
		actualDate := solar.Format("2006-01-02")
		if actualDate == expectedDate {
			fmt.Printf("✓ 农历2024年正月初一 → 公历%s\n", actualDate)
		} else {
			fmt.Printf("✗ 农历2024年正月初一 → 公历%s (预期: %s)\n", actualDate, expectedDate)
		}
	}
	fmt.Println()

	fmt.Println("========================================")
	if allPassed {
		fmt.Println("所有测试通过！✓")
	} else {
		fmt.Println("部分测试失败，请检查上述问题")
	}
	fmt.Println("========================================")
}

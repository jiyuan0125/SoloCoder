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

	allPassed := true

	fmt.Println("=== 问题1验证：闰月边界roundtrip ===")

	leapFeb1 := lunar.LunarDate{Year: 2023, Month: 2, Day: 1, IsLeapMonth: true}
	solar, err := lunar.LunarToSolar(leapFeb1)
	if err != nil {
		fmt.Println("✗ 农历2023年闰二月初一转公历失败:", err)
		allPassed = false
	} else {
		expectedDate := "2023-03-22"
		actualDate := solar.Format("2006-01-02")
		if actualDate == expectedDate {
			fmt.Printf("✓ 农历2023年闰二月初一 → 公历%s\n", actualDate)
		} else {
			fmt.Printf("✗ 农历2023年闰二月初一 → 公历%s (预期: %s)\n", actualDate, expectedDate)
			allPassed = false
		}

		l, err := lunar.SolarToLunar(solar)
		if err != nil {
			fmt.Println("✗ 公历转农历失败:", err)
			allPassed = false
		} else {
			if l.Year == 2023 && l.Month == 2 && l.Day == 1 && l.IsLeapMonth {
				fmt.Printf("✓ 公历%s → 农历闰二月初一\n", actualDate)
			} else {
				fmt.Printf("✗ 公历%s → %s (预期: 闰二月初一)\n", actualDate, l.String())
				allPassed = false
			}
		}
	}
	fmt.Println()

	fmt.Println("=== 问题2验证：所有年份节气时间精确 ===")

	years := []int{2023, 2024, 2025}
	for _, y := range years {
		jq, err := jieqi.GetJieqi(y)
		if err != nil {
			fmt.Printf("✗ 获取%d年节气失败: %v\n", y, err)
			allPassed = false
			continue
		}

		hasNoonTime := false
		for _, j := range jq {
			if j.Time.Hour() == 12 && j.Time.Minute() == 0 {
				hasNoonTime = true
				break
			}
		}

		if hasNoonTime {
			fmt.Printf("✗ %d年仍有节气时间为12:00\n", y)
			allPassed = false
		} else {
			fmt.Printf("✓ %d年所有节气时间都不是12:00\n", y)
		}
	}
	fmt.Println()

	fmt.Println("=== 问题3验证：2024年节气日期 ===")

	jq2024, err := jieqi.GetJieqi(2024)
	if err != nil {
		fmt.Println("✗ 获取2024年节气失败:", err)
		allPassed = false
	} else {
		expectedDates := map[string]string{
			"立春": "2024-02-04",
			"惊蛰": "2024-03-05",
			"春分": "2024-03-20",
			"清明": "2024-04-04",
			"夏至": "2024-06-21",
			"秋分": "2024-09-22",
			"冬至": "2024-12-21",
		}

		for _, j := range jq2024 {
			if expected, ok := expectedDates[j.Name]; ok {
				actual := j.Time.Format("2006-01-02")
				if actual == expected {
					fmt.Printf("✓ %s: %s %s\n", j.Name, actual, j.Time.Format("15:04"))
				} else {
					fmt.Printf("✗ %s: %s (预期: %s)\n", j.Name, actual, expected)
					allPassed = false
				}
			}
		}
	}
	fmt.Println()

	fmt.Println("=== 额外验证：更多roundtrip测试 ===")

	testDates := []string{
		"2023-03-22",
		"2023-04-19",
		"2023-04-20",
		"2024-02-10",
	}

	for _, d := range testDates {
		t, _ := time.Parse("2006-01-02", d)
		l, err := lunar.SolarToLunar(t)
		if err != nil {
			fmt.Printf("✗ %s 转农历失败: %v\n", d, err)
			allPassed = false
			continue
		}
		s, err := lunar.LunarToSolar(*l)
		if err != nil {
			fmt.Printf("✗ 农历转公历失败: %v\n", err)
			allPassed = false
			continue
		}
		roundtrip := s.Format("2006-01-02")
		if roundtrip == d {
			fmt.Printf("✓ %s ↔ %s\n", d, l.String())
		} else {
			fmt.Printf("✗ %s → %s → %s (roundtrip不一致)\n", d, l.String(), roundtrip)
			allPassed = false
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

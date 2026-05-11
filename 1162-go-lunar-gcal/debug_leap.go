//go:build ignore

package main

import (
	"fmt"
	"time"

	"go-lunar-gcal/pkg/lunar"
)

func main() {
	fmt.Println("=== 测试农历转公历 ===")

	feb29 := lunar.LunarDate{Year: 2023, Month: 2, Day: 29, IsLeapMonth: false}
	solarFeb29, err := lunar.LunarToSolar(feb29)
	if err != nil {
		fmt.Println("农历2023年二月廿九 错误:", err)
	} else {
		fmt.Printf("农历2023年二月廿九 → 公历: %s\n", solarFeb29.Format("2006-01-02"))
	}

	leapFeb1 := lunar.LunarDate{Year: 2023, Month: 2, Day: 1, IsLeapMonth: true}
	solarLeapFeb1, err := lunar.LunarToSolar(leapFeb1)
	if err != nil {
		fmt.Println("农历2023年闰二月初一 错误:", err)
	} else {
		fmt.Printf("农历2023年闰二月初一 → 公历: %s\n", solarLeapFeb1.Format("2006-01-02"))
	}

	leapFeb29 := lunar.LunarDate{Year: 2023, Month: 2, Day: 29, IsLeapMonth: true}
	solarLeapFeb29, err := lunar.LunarToSolar(leapFeb29)
	if err != nil {
		fmt.Println("农历2023年闰二月廿九 错误:", err)
	} else {
		fmt.Printf("农历2023年闰二月廿九 → 公历: %s\n", solarLeapFeb29.Format("2006-01-02"))
	}

	mar1 := lunar.LunarDate{Year: 2023, Month: 3, Day: 1, IsLeapMonth: false}
	solarMar1, err := lunar.LunarToSolar(mar1)
	if err != nil {
		fmt.Println("农历2023年三月初一 错误:", err)
	} else {
		fmt.Printf("农历2023年三月初一 → 公历: %s\n", solarMar1.Format("2006-01-02"))
	}

	fmt.Println()
	fmt.Println("=== 测试公历转农历 (反向) ===")

	dates := []string{
		"2023-03-21",
		"2023-03-22",
		"2023-03-23",
		"2023-04-19",
		"2023-04-20",
	}

	for _, d := range dates {
		t, _ := time.Parse("2006-01-02", d)
		l, err := lunar.SolarToLunar(t)
		if err != nil {
			fmt.Printf("%s → 错误: %v\n", d, err)
		} else {
			fmt.Printf("%s → %s (年:%d,月:%d,日:%d,闰:%v)\n",
				d, l.String(), l.Year, l.Month, l.Day, l.IsLeapMonth)
		}
	}
}

package main

import (
	"fmt"
	"id-number-parser/common"
	"os"
	"time"
)

func printParseResult(result common.ParseResponse) {
	if !result.Success {
		fmt.Printf("解析失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("=== 身份证号解析结果 ===")
	fmt.Printf("原始身份证号: %s\n", result.OriginalID)
	if result.StandardizedID != result.OriginalID {
		fmt.Printf("标准化身份证号: %s\n", result.StandardizedID)
	}

	birthDate, err := time.Parse(time.RFC3339, result.BirthDate)
	if err == nil {
		fmt.Printf("出生日期: %s\n", birthDate.Format("2006年01月02日"))
	} else {
		fmt.Printf("出生日期: %s\n", result.BirthDate)
	}

	fmt.Printf("性别: %s\n", result.Gender)
	fmt.Printf("籍贯省份: %s (%s)\n", result.ProvinceName, result.ProvinceCode)
	fmt.Printf("当前年龄: %d岁\n", result.Age)
}

func printValidateResult(result common.ValidateResponse) {
	if !result.Success {
		fmt.Printf("校验请求失败: %s\n", result.Message)
		os.Exit(1)
	}

	fmt.Println("=== 身份证号校验结果 ===")
	if result.Valid {
		fmt.Println("✓ 身份证号有效")
		fmt.Printf("详情: %s\n", result.Message)
	} else {
		fmt.Println("✗ 身份证号无效")
		fmt.Printf("原因: %s\n", result.Message)
		os.Exit(1)
	}
}

package client

import (
	"fmt"
	"os"
	"sort"
	
	"disk-usage/internal/common"
	"disk-usage/internal/protocol"
)

type Formatter struct {
	output string
}

func NewFormatter(output string) *Formatter {
	return &Formatter{
		output: output,
	}
}

func (f *Formatter) FormatScanResult(result *protocol.ScanResult, summary bool) error {
	if summary {
		return f.printSummary(result)
	}
	
	if f.output == "csv" {
		return f.exportCSV(result)
	}
	
	return f.printHumanReadable(result)
}

func (f *Formatter) FormatCompareResult(result *protocol.CompareResult) error {
	if f.output == "csv" {
		return f.exportCompareCSV(result)
	}
	
	return f.printCompareHumanReadable(result)
}

func (f *Formatter) printSummary(result *protocol.ScanResult) error {
	fmt.Printf("扫描完成，耗时: %d 毫秒\n", result.ScanTime)
	fmt.Println("========================================")
	fmt.Printf("总大小: %s\n", common.FormatSize(result.TotalSize))
	fmt.Printf("文件总数: %d\n", result.TotalFiles)
	fmt.Printf("目录总数: %d\n", result.TotalDirs)
	fmt.Println("========================================")
	
	if len(result.Warnings) > 0 {
		fmt.Printf("\n警告信息 (%d 条):\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}
	
	return nil
}

func (f *Formatter) printHumanReadable(result *protocol.ScanResult) error {
	fmt.Printf("扫描完成，耗时: %d 毫秒\n", result.ScanTime)
	fmt.Println("========================================")
	fmt.Printf("总大小: %s | 文件数: %d | 目录数: %d\n", 
		common.FormatSize(result.TotalSize), 
		result.TotalFiles, 
		result.TotalDirs)
	fmt.Println("========================================")
	
	if len(result.Directories) > 0 {
		fmt.Println("\n子目录大小统计 (按大小降序):")
		fmt.Println("----------------------------------------")
		fmt.Printf("%-8s %-8s %-6s %s\n", "大小", "文件数", "目录数", "路径")
		fmt.Println("----------------------------------------")
		
		for _, dir := range result.Directories {
			fmt.Printf("%-8s %-8d %-6d %s\n", 
				common.FormatSize(dir.Size), 
				dir.Files, 
				dir.Dirs, 
				dir.Path)
		}
	}
	
	if len(result.Warnings) > 0 {
		fmt.Printf("\n警告信息 (%d 条):\n", len(result.Warnings))
		for _, warn := range result.Warnings {
			fmt.Printf("  - %s\n", warn)
		}
	}
	
	return nil
}

func (f *Formatter) exportCSV(result *protocol.ScanResult) error {
	file, err := os.Create("disk_usage.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	
	file.WriteString("路径,大小(字节),大小(可读),文件数,目录数,深度\n")
	
	for _, dir := range result.Directories {
		file.WriteString(fmt.Sprintf("%s,%d,%s,%d,%d,%d\n",
			dir.Path,
			dir.Size,
			common.FormatSize(dir.Size),
			dir.Files,
			dir.Dirs,
			dir.Depth))
	}
	
	fmt.Printf("结果已导出到: disk_usage.csv\n")
	return nil
}

func (f *Formatter) printCompareHumanReadable(result *protocol.CompareResult) error {
	fmt.Println("扫描对比结果")
	fmt.Println("========================================")
	fmt.Printf("上一次扫描总大小: %s | 文件数: %d\n", 
		common.FormatSize(result.OldScan.TotalSize), 
		result.OldScan.TotalFiles)
	fmt.Printf("当前扫描总大小: %s | 文件数: %d\n", 
		common.FormatSize(result.NewScan.TotalSize), 
		result.NewScan.TotalFiles)
	
	totalGrowth := result.NewScan.TotalSize - result.OldScan.TotalSize
	filesGrowth := result.NewScan.TotalFiles - result.OldScan.TotalFiles
	
	if totalGrowth > 0 {
		fmt.Printf("总增长: +%s | 文件增长: +%d\n", 
			common.FormatSize(totalGrowth), 
			filesGrowth)
	} else {
		fmt.Printf("总变化: %s | 文件变化: %d\n", 
			common.FormatSize(totalGrowth), 
			filesGrowth)
	}
	
	fmt.Println("========================================")
	
	if len(result.GrowthDirs) > 0 {
		sort.Slice(result.GrowthDirs, func(i, j int) bool {
			return result.GrowthDirs[i].Growth > result.GrowthDirs[j].Growth
		})
		
		fmt.Println("\n增长的目录 (按增长量降序):")
		fmt.Println("----------------------------------------")
		fmt.Printf("%-8s %-8s %-8s %s\n", "增长量", "原大小", "新大小", "路径")
		fmt.Println("----------------------------------------")
		
		for _, dir := range result.GrowthDirs {
			fmt.Printf("%+8s %-8s %-8s %s\n", 
				"+"+common.FormatSize(dir.Growth), 
				common.FormatSize(dir.OldSize), 
				common.FormatSize(dir.NewSize), 
				dir.Path)
		}
	} else {
		fmt.Println("\n没有发现增长的目录")
	}
	
	return nil
}

func (f *Formatter) exportCompareCSV(result *protocol.CompareResult) error {
	file, err := os.Create("disk_usage_compare.csv")
	if err != nil {
		return err
	}
	defer file.Close()
	
	file.WriteString("路径,原大小(字节),原大小(可读),新大小(字节),新大小(可读),增长量(字节),增长量(可读),原文件数,新文件数\n")
	
	for _, dir := range result.GrowthDirs {
		file.WriteString(fmt.Sprintf("%s,%d,%s,%d,%s,%d,%s,%d,%d\n",
			dir.Path,
			dir.OldSize,
			common.FormatSize(dir.OldSize),
			dir.NewSize,
			common.FormatSize(dir.NewSize),
			dir.Growth,
			common.FormatSize(dir.Growth),
			dir.OldFiles,
			dir.NewFiles))
	}
	
	fmt.Printf("对比结果已导出到: disk_usage_compare.csv\n")
	return nil
}

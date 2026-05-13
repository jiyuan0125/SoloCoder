package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"reportgen/internal/model"
	"reportgen/internal/output"
	"reportgen/internal/report"
	"reportgen/internal/template"
)

var rootCmd = &cobra.Command{
	Use:   "reportgen",
	Short: "报表生成工具 - 从CSV或数据库读取数据，按模板生成报表",
	Long: `reportgen 是一个强大的报表生成命令行工具。
支持从CSV文件或MySQL/PostgreSQL数据库读取数据，
按YAML模板定义的结构、筛选条件和汇总方式生成报表。
输出格式支持文本表格和HTML。`,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "生成单个报表",
	RunE:  runGenerate,
}

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量生成多个报表",
	RunE:  runBatch,
}

var (
	configFlag     string
	dataSourceFlag string
	templateFlag   string
	outputFlag     string
	formatFlag     string
	configDirFlag  string
)

func init() {
	generateCmd.Flags().StringVarP(&configFlag, "config", "c", "", "报表配置文件路径 (YAML格式)")
	generateCmd.Flags().StringVarP(&dataSourceFlag, "data-source", "d", "", "数据源路径 (CSV文件路径)")
	generateCmd.Flags().StringVarP(&templateFlag, "template", "t", "", "模板文件路径")
	generateCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "输出文件路径")
	generateCmd.Flags().StringVarP(&formatFlag, "format", "f", "", "输出格式: text 或 html (覆盖配置文件中的设置)")

	rootCmd.AddCommand(generateCmd)

	batchCmd.Flags().StringVarP(&configDirFlag, "config-dir", "C", "", "报表配置文件目录")
	rootCmd.AddCommand(batchCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	engine := report.NewEngine()

	if configFlag != "" {
		result, err := engine.GenerateFromConfig(configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		cfg, err := template.ParseReportConfig(configFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		if outputFlag != "" {
			cfg.OutputPath = outputFlag
		}
		if formatFlag != "" {
			cfg.OutputFormat = model.OutputFormat(formatFlag)
		}

		if err := outputReport(result, cfg.OutputFormat, cfg.OutputPath); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("报表 '%s' 生成成功\n", result.Name)
		return nil
	}

	if dataSourceFlag == "" || templateFlag == "" {
		return fmt.Errorf("必须提供 --config 参数，或者同时提供 --data-source 和 --template 参数")
	}

	outFormat := model.OutputFormat(formatFlag)
	if outFormat == "" {
		outFormat = model.OutputFormatText
	}

	cfg := &model.ReportConfig{
		Name: "临时报表",
		DataSource: model.DataSource{
			Type: model.DataSourceCSV,
			Path: dataSourceFlag,
		},
		OutputFormat: outFormat,
		OutputPath:   outputFlag,
	}

	tplCfg, err := template.ParseReportConfig(templateFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
	cfg.Template = tplCfg.Template

	result, err := engine.Generate(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	if err := outputReport(result, cfg.OutputFormat, cfg.OutputPath); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("报表 '%s' 生成成功\n", result.Name)
	return nil
}

func runBatch(cmd *cobra.Command, args []string) error {
	if configDirFlag == "" {
		return fmt.Errorf("必须提供 --config-dir 参数指定配置文件目录")
	}

	info, err := os.Stat(configDirFlag)
	if os.IsNotExist(err) {
		return fmt.Errorf("配置目录不存在: %s", configDirFlag)
	}
	if err != nil {
		return fmt.Errorf("访问配置目录失败: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("路径不是目录: %s", configDirFlag)
	}

	configFiles, err := findConfigFiles(configDirFlag)
	if err != nil {
		return fmt.Errorf("查找配置文件失败: %w", err)
	}

	if len(configFiles) == 0 {
		fmt.Println("目录中未找到配置文件 (.yaml 或 .yml)")
		return nil
	}

	fmt.Printf("找到 %d 个配置文件，开始批量生成...\n\n", len(configFiles))

	successCount := 0
	failCount := 0

	for _, configFile := range configFiles {
		fmt.Printf("正在处理: %s\n", configFile)

		engine := report.NewEngine()
		result, err := engine.GenerateFromConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  错误: %v\n", err)
			failCount++
			continue
		}

		cfg, err := template.ParseReportConfig(configFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  错误: %v\n", err)
			failCount++
			continue
		}

		if err := outputReport(result, cfg.OutputFormat, cfg.OutputPath); err != nil {
			fmt.Fprintf(os.Stderr, "  错误: %v\n", err)
			failCount++
			continue
		}

		fmt.Printf("  ✓ 报表 '%s' 生成成功\n", result.Name)
		successCount++
	}

	fmt.Printf("\n批量生成完成: 成功 %d, 失败 %d\n", successCount, failCount)

	if failCount > 0 {
		os.Exit(2)
	}

	return nil
}

func findConfigFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func outputReport(result *model.ReportResult, format model.OutputFormat, outputPath string) error {
	switch format {
	case model.OutputFormatHTML:
		return output.GenerateHTMLReport(result, outputPath)
	case model.OutputFormatText:
		fallthrough
	default:
		return output.GenerateTextReport(result, outputPath)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

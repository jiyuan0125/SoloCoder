package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"timeseries-agg/internal"
)

type aggregateFlags struct {
	inputFile       string
	outputFile      string
	granularity     string
	timeColumn      int
	columns         []string
	fillMissing     bool
	fillMethod      string
	verify          bool
}

func NewAggregateCmd() *cobra.Command {
	flags := &aggregateFlags{}

	cmd := &cobra.Command{
		Use:   "aggregate",
		Short: "聚合时序数据",
		Long: `从 CSV 文件读取时序数据，按指定的时间粒度和聚合方法进行聚合。

支持的时间粒度: minute, hour, day
支持的聚合方式: avg, max, min, sum, count

列指定格式: 列索引:聚合方式 或 列名:聚合方式
例如: 1:avg 2:max 或 "temperature:avg" "humidity:max"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAggregate(flags)
		},
	}

	cmd.Flags().StringVarP(&flags.inputFile, "input", "i", "", "输入 CSV 文件路径 (必填)")
	cmd.Flags().StringVarP(&flags.outputFile, "output", "o", "output.csv", "输出 CSV 文件路径")
	cmd.Flags().StringVarP(&flags.granularity, "granularity", "g", "hour", "时间粒度: minute|hour|day")
	cmd.Flags().IntVarP(&flags.timeColumn, "time-column", "t", 0, "时间戳列的索引 (从 0 开始)")
	cmd.Flags().StringSliceVarP(&flags.columns, "columns", "c", []string{}, "指标列配置，格式: 索引:方式 或 名称:方式 (可多次指定)")
	cmd.Flags().BoolVarP(&flags.fillMissing, "fill-missing", "f", true, "是否填充缺失的时间段")
	cmd.Flags().StringVar(&flags.fillMethod, "fill-method", "null", "缺失值填充方法: null")
	cmd.Flags().BoolVar(&flags.verify, "verify", true, "是否验证数据一致性")

	cmd.MarkFlagRequired("input")

	return cmd
}

func runAggregate(flags *aggregateFlags) error {
	granularity := internal.Granularity(flags.granularity)
	switch granularity {
	case internal.GranularityMinute, internal.GranularityHour, internal.GranularityDay:
	default:
		exitWithError(fmt.Sprintf("无效的时间粒度: %s，可选值: minute, hour, day", flags.granularity))
	}

	reader, err := internal.NewCSVReader(flags.inputFile)
	if err != nil {
		exitWithError(err.Error())
	}
	defer reader.Close()

	header, err := reader.ReadHeader()
	if err != nil {
		exitWithError(err.Error())
	}

	columnConfigs, err := parseColumnConfigs(flags.columns, header)
	if err != nil {
		exitWithError(err.Error())
	}

	if len(columnConfigs) == 0 {
		for i, h := range header {
			if i == flags.timeColumn {
				continue
			}
			columnConfigs = append(columnConfigs, internal.ColumnConfig{
				Index:  i,
				Name:   h,
				Method: internal.AggAvg,
			})
		}
	}

	config := &internal.Config{
		InputFile:       flags.inputFile,
		OutputFile:      flags.outputFile,
		Granularity:     granularity,
		TimeColumnIndex: flags.timeColumn,
		Columns:         columnConfigs,
		FillMissing:     flags.fillMissing,
		FillMethod:      flags.fillMethod,
	}

	aggregator := internal.NewAggregator(config)
	aggregator.SetHeader(header)

	valueColIdxs := make([]int, 0, len(columnConfigs))
	for _, col := range columnConfigs {
		valueColIdxs = append(valueColIdxs, col.Index)
	}

	fmt.Fprintf(os.Stderr, "开始处理文件: %s\n", flags.inputFile)
	fmt.Fprintf(os.Stderr, "时间粒度: %s\n", flags.granularity)
	fmt.Fprintf(os.Stderr, "时区: %s\n", reader.GetTimezone().String())

	err = reader.ReadAll(flags.timeColumn, valueColIdxs, func(row internal.DataRow, valid bool, warning string) {
		if !valid {
			aggregator.AddInvalidRow()
			if warning != "" {
				fmt.Fprintln(os.Stderr, "警告:", warning)
			}
			return
		}
		aggregator.AddRow(row.Time, row.Values)
	})
	if err != nil {
		exitWithError(err.Error())
	}

	err = aggregator.WriteOutput(flags.outputFile, reader.GetTimezone())
	if err != nil {
		exitWithError(err.Error())
	}

	stats := aggregator.GetStats()
	fmt.Fprintf(os.Stderr, "处理完成！\n")
	fmt.Fprintf(os.Stderr, "  总行数: %d\n", stats["total_rows"])
	fmt.Fprintf(os.Stderr, "  有效行: %d\n", stats["valid_rows"])
	fmt.Fprintf(os.Stderr, "  无效行: %d\n", stats["invalid_rows"])
	fmt.Fprintf(os.Stderr, "  聚合时段数: %d\n", stats["bucket_count"])
	fmt.Fprintf(os.Stderr, "输出文件: %s\n", flags.outputFile)

	if flags.verify {
		fmt.Fprintln(os.Stderr, "正在验证数据一致性...")
		bucketCounts := make(map[string]int)

		verifier := internal.NewVerifier(config)
		report, err := verifier.VerifyAndGenerateReport(
			flags.inputFile,
			stats["valid_rows"].(int),
			bucketCounts,
			reader.GetTimezone(),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "验证失败: %v\n", err)
		} else {
			report.PrintReport()
		}
	}

	return nil
}

func parseColumnConfigs(colSpecs []string, header []string) ([]internal.ColumnConfig, error) {
	var configs []internal.ColumnConfig

	nameToIndex := make(map[string]int)
	for i, h := range header {
		nameToIndex[h] = i
		nameToIndex[strings.ToLower(h)] = i
	}

	for _, spec := range colSpecs {
		parts := strings.SplitN(spec, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("无效的列配置格式: %s，应为: 索引:方式 或 名称:方式", spec)
		}

		colRef := strings.TrimSpace(parts[0])
		methodStr := strings.TrimSpace(strings.ToLower(parts[1]))

		method := internal.AggregationMethod(methodStr)
		switch method {
		case internal.AggAvg, internal.AggMax, internal.AggMin, internal.AggSum, internal.AggCount:
		default:
			return nil, fmt.Errorf("无效的聚合方式: %s，可选值: avg, max, min, sum, count", methodStr)
		}

		var colIndex int
		var colName string

		if idx, err := strconv.Atoi(colRef); err == nil {
			colIndex = idx
			if idx >= 0 && idx < len(header) {
				colName = header[idx]
			}
		} else {
			if idx, ok := nameToIndex[colRef]; ok {
				colIndex = idx
				colName = colRef
			} else if idx, ok := nameToIndex[strings.ToLower(colRef)]; ok {
				colIndex = idx
				colName = colRef
			} else {
				return nil, fmt.Errorf("找不到列: %s", colRef)
			}
		}

		configs = append(configs, internal.ColumnConfig{
			Index:  colIndex,
			Name:   colName,
			Method: method,
		})
	}

	return configs, nil
}

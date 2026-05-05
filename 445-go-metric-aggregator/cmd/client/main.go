package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"metric-aggregator/internal/client"
	"metric-aggregator/pkg/common"
	"os"
	"time"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "服务端地址")
	flag.Usage = printUsage
	
	flag.Parse()
	
	if flag.NArg() == 0 {
		printUsage()
		os.Exit(1)
	}
	
	cmd := flag.Arg(0)
	c := client.NewClient(*serverURL)
	
	switch cmd {
	case "health":
		handleHealth(c)
	case "create":
		handleCreate(c)
	case "delete":
		handleDelete(c)
	case "list":
		handleList(c)
	case "info":
		handleInfo(c)
	case "report":
		handleReport(c)
	case "query":
		handleQuery(c)
	default:
		fmt.Fprintf(os.Stderr, "未知命令: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`指标聚合客户端

用法:
  metric-client [选项] <命令> [参数]

选项:
  -server string    服务端地址 (默认 "http://localhost:8080")
  -h, -help         显示帮助信息

命令:
  health                       检查服务健康状态
  create <metric_name>         创建新指标
  delete <metric_name>         删除指标
  list                         列出所有指标
  info <metric_name>           查看指标详细信息
  report <metric> <timestamp> <value>
                               上报单个数据点
                               timestamp: Unix时间戳(秒) 或 "now"
  query [选项] <metrics...>    查询聚合数据
    查询选项:
      -start int      开始时间戳 (默认 1小时前)
      -end int        结束时间戳 (默认 当前时间)
      -g string       聚合粒度: 1m|5m|1h (默认 1m)
      -a string       聚合方式: avg|max|min|sum (默认 avg)

示例:
  metric-client create cpu.usage
  metric-client report cpu.usage now 45.5
  metric-client query -g 1m -a avg cpu.usage memory.usage
  metric-client list
  metric-client info cpu.usage
  metric-client delete cpu.usage`)
}

func handleHealth(c *client.Client) {
	if err := c.Health(); err != nil {
		fmt.Fprintf(os.Stderr, "服务不可用: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("服务状态: 正常")
}

func handleCreate(c *client.Client) {
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "用法: metric-client create <metric_name>")
		os.Exit(1)
	}
	
	name := flag.Arg(1)
	if err := c.CreateMetric(name); err != nil {
		fmt.Fprintf(os.Stderr, "创建指标失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("指标创建成功: %s\n", name)
}

func handleDelete(c *client.Client) {
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "用法: metric-client delete <metric_name>")
		os.Exit(1)
	}
	
	name := flag.Arg(1)
	if err := c.DeleteMetric(name); err != nil {
		fmt.Fprintf(os.Stderr, "删除指标失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("指标删除成功: %s\n", name)
}

func handleList(c *client.Client) {
	metrics, err := c.ListMetrics()
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取指标列表失败: %v\n", err)
		os.Exit(1)
	}
	
	if len(metrics) == 0 {
		fmt.Println("没有指标")
		return
	}
	
	fmt.Println("指标列表:")
	for _, m := range metrics {
		fmt.Printf("  - %s\n", m)
	}
}

func handleInfo(c *client.Client) {
	if flag.NArg() < 2 {
		fmt.Fprintln(os.Stderr, "用法: metric-client info <metric_name>")
		os.Exit(1)
	}
	
	name := flag.Arg(1)
	info, err := c.GetMetricInfo(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取指标信息失败: %v\n", err)
		os.Exit(1)
	}
	
	data, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(data))
}

func handleReport(c *client.Client) {
	args := flag.Args()
	if len(args) < 4 {
		fmt.Fprintln(os.Stderr, "用法: metric-client report <metric> <timestamp> <value>")
		os.Exit(1)
	}
	
	metric := args[1]
	tsStr := args[2]
	valStr := args[3]
	
	var ts int64
	if tsStr == "now" {
		ts = time.Now().Unix()
	} else {
		var err error
		ts, err = parseTimestamp(tsStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无效的时间戳: %v\n", err)
			os.Exit(1)
		}
	}
	
	value, err := parseValue(valStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无效的数值: %v\n", err)
		os.Exit(1)
	}
	
	points := []common.DataPoint{{
		Metric:    metric,
		Timestamp: ts,
		Value:     value,
	}}
	
	resp, err := c.Report(points)
	if err != nil {
		fmt.Fprintf(os.Stderr, "上报数据失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("上报结果: 成功=%d, 异常值=%d\n", resp.Success, resp.Outliers)
}

func handleQuery(c *client.Client) {
	queryFlag := flag.NewFlagSet("query", flag.ExitOnError)
	startFlag := queryFlag.String("start", "", "开始时间戳 (默认 1小时前)")
	endFlag := queryFlag.String("end", "", "结束时间戳 (默认 当前时间)")
	gFlag := queryFlag.String("g", "1m", "聚合粒度: 1m|5m|1h")
	aFlag := queryFlag.String("a", "avg", "聚合方式: avg|max|min|sum")
	
	args := flag.Args()[1:]
	if err := queryFlag.Parse(args); err != nil {
		os.Exit(1)
	}
	
	metrics := queryFlag.Args()
	if len(metrics) == 0 {
		fmt.Fprintln(os.Stderr, "请指定要查询的指标名称")
		os.Exit(1)
	}
	
	now := time.Now()
	var start, end int64
	
	if *startFlag == "" {
		start = now.Add(-1 * time.Hour).Unix()
	} else {
		var err error
		start, err = parseTimestamp(*startFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无效的开始时间戳: %v\n", err)
			os.Exit(1)
		}
	}
	
	if *endFlag == "" {
		end = now.Unix()
	} else {
		var err error
		end, err = parseTimestamp(*endFlag)
		if err != nil {
			fmt.Fprintf(os.Stderr, "无效的结束时间戳: %v\n", err)
			os.Exit(1)
		}
	}
	
	granularity := common.Granularity(*gFlag)
	if !common.ValidGranularity(granularity) {
		fmt.Fprintf(os.Stderr, "无效的聚合粒度: %s (支持: 1m, 5m, 1h)\n", *gFlag)
		os.Exit(1)
	}
	
	aggregation := common.AggregationType(*aFlag)
	if !common.ValidAggregation(aggregation) {
		fmt.Fprintf(os.Stderr, "无效的聚合方式: %s (支持: avg, max, min, sum)\n", *aFlag)
		os.Exit(1)
	}
	
	resp, err := c.Query(metrics, start, end, granularity, aggregation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询失败: %v\n", err)
		os.Exit(1)
	}
	
	for i, result := range resp.Results {
		if i > 0 {
			fmt.Println("---")
		}
		printQueryResult(result)
	}
}

func printQueryResult(result common.AggregationResult) {
	fmt.Printf("指标: %s\n", result.Metric)
	fmt.Printf("范围: %s -> %s\n", 
		time.Unix(result.Start, 0).Format("2006-01-02 15:04:05"),
		time.Unix(result.End, 0).Format("2006-01-02 15:04:05"))
	fmt.Printf("粒度: %s\n", result.Granularity)
	fmt.Println()
	
	if len(result.Points) == 0 {
		fmt.Println("没有数据点")
		return
	}
	
	fmt.Println("数据点:")
	for _, p := range result.Points {
		filled := ""
		if p.IsFilled {
			filled = " (填充)"
		}
		fmt.Printf("  %s: %.6f%s\n",
			time.Unix(p.Timestamp, 0).Format("2006-01-02 15:04:05"),
			p.Value,
			filled)
	}
}

func parseTimestamp(s string) (int64, error) {
	val, err := parseLong(s)
	if err == nil {
		return common.TruncateTimestamp(val), nil
	}
	return 0, err
}

func parseLong(s string) (int64, error) {
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return val, nil
}

func parseValue(s string) (float64, error) {
	var val float64
	_, err := fmt.Sscanf(s, "%f", &val)
	return val, err
}

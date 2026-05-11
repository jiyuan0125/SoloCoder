package main

import (
	"fmt"
	"os"
)

const usage = `Quality Trace CLI - 制造业质量追溯系统命令行工具

用法:
  qt-cli [--server URL] <command> [arguments]

服务器地址:
  --server URL    服务端地址 (默认: http://localhost:8080)

命令:
  health                          检查服务端健康状态

批次管理:
  create-batch <product> <factory> <plan_qty> [start_time]
                                  创建生产批次
  get-batch <batch_id>            获取批次详情
  complete-batch <batch_id> <actual_qty> [end_time]
                                  完成批次
  list-batches [--status status] [--product name]
                                  列出批次

工序管理:
  add-process-flow <product> <seq:name>...
                                  为产品添加工序流程
  start-process <batch_id> <operator> [start_time]
                                  开始工序
  complete-process <batch_id> <pass|fail> [end_time]
                                  完成工序
  rework-decision <batch_id> <rework|scrap>
                                  做出返工或报废决定
  list-processes [batch_id]       列出工序记录

检验管理:
  add-inspection-spec <product> <incoming|process|final> <metric> <lower> <upper>
                                  添加检验规格
  add-inspection-record <batch_id> <type> <metric> <value> <inspector> [time]
                                  添加检验记录
  list-inspections [--batch id] [--type type]
                                  列出检验记录

数据导出:
  export-batches <file> [--status status] [--product name]
                                  导出批次CSV
  export-processes <file> [--batch id]
                                  导出工序记录CSV
  export-inspections <file> [--batch id] [--type type]
                                  导出检验记录CSV

指标统计:
  metrics [days]                  获取指标统计 (默认30天)
`

var commands = map[string]func(*Client, []string){
	"health":               cmdHealth,
	"create-batch":         cmdCreateBatch,
	"get-batch":            cmdGetBatch,
	"complete-batch":       cmdCompleteBatch,
	"list-batches":         cmdListBatches,
	"add-process-flow":     cmdAddProcessFlow,
	"start-process":        cmdStartProcess,
	"complete-process":     cmdCompleteProcess,
	"rework-decision":      cmdReworkDecision,
	"list-processes":       cmdListProcessRecords,
	"add-inspection-spec":  cmdAddInspectionSpec,
	"add-inspection-record": cmdAddInspectionRecord,
	"list-inspections":     cmdListInspectionRecords,
	"export-batches":       cmdExportBatches,
	"export-processes":     cmdExportProcessRecords,
	"export-inspections":   cmdExportInspectionRecords,
	"metrics":              cmdGetMetrics,
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(1)
	}

	serverURL := "http://localhost:8080"
	args := os.Args[1:]

	if args[0] == "--server" {
		if len(args) < 3 {
			fmt.Println(usage)
			os.Exit(1)
		}
		serverURL = args[1]
		args = args[2:]
	}

	if len(args) < 1 {
		fmt.Println(usage)
		os.Exit(1)
	}

	cmd := args[0]
	cmdArgs := args[1:]

	fn, exists := commands[cmd]
	if !exists {
		fmt.Printf("Unknown command: %s\n\n", cmd)
		fmt.Println(usage)
		os.Exit(1)
	}

	client := NewClient(serverURL)
	fn(client, cmdArgs)
}

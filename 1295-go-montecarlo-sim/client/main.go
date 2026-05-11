package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"montecarlo-sim/common"
)

func main() {
	batchFile := flag.String("batch", "", "批量任务配置文件路径（JSON格式）")
	serverURL := flag.String("server", "http://localhost:8104", "服务端URL")
	flag.Parse()

	if *batchFile != "" {
		runBatch(*batchFile, *serverURL)
		return
	}

	runInteractive(*serverURL)
}

func runBatch(configPath string, serverURL string) {
	config, err := common.LoadBatchConfig(configPath)
	if err != nil {
		fmt.Printf("加载配置文件失败: %v\n", err)
		os.Exit(1)
	}

	effectiveURL := serverURL
	if config.ServerURL != "" && serverURL == "http://localhost:8104" {
		effectiveURL = config.ServerURL
	}

	fmt.Printf("执行批量任务，服务端: %s\n\n", effectiveURL)

	for _, task := range config.Tasks {
		var req common.SimulationRequest
		req.Type = task.Type
		if task.Pi != nil {
			req.Pi = task.Pi
		}
		if task.Option != nil {
			req.Option = task.Option
		}
		if task.Probability != nil {
			req.Probability = task.Probability
		}

		fmt.Printf("=== 任务: %s ===\n", task.Name)
		resp, err := executeRequest(effectiveURL, &req)
		if err != nil {
			fmt.Printf("执行失败: %v\n", err)
			fmt.Println()
			continue
		}
		printResponse(resp)
		fmt.Println()
	}
}

func runInteractive(serverURL string) {
	fmt.Println("=== 蒙特卡洛模拟客户端 ===")
	fmt.Println("选择模拟类型:")
	fmt.Println("1) π值估算")
	fmt.Println("2) 欧式期权定价")
	fmt.Println("3) 概率估算")
	fmt.Println("4) 退出")

	choice := readInt("请输入选项 (1-4): ", 1, 4)
	if choice == 4 {
		return
	}

	var req common.SimulationRequest

	switch choice {
	case 1:
		req.Type = common.TypePi
		samples := readInt("采样次数 (建议10000+): ", 1, 100000000)
		confidence := readFloat("置信水平 (0.0-1.0, 建议0.95): ", 0.0, 1.0)
		req.Pi = &common.PiRequest{
			Samples:    samples,
			Confidence: confidence,
		}
	case 2:
		req.Type = common.TypeOption
		samples := readInt("采样次数 (建议10000+): ", 1, 100000000)
		confidence := readFloat("置信水平 (0.0-1.0, 建议0.95): ", 0.0, 1.0)
		s0 := readFloat("当前股价 S0: ", 0.0, 1000000.0)
		k := readFloat("执行价 K: ", 0.0, 1000000.0)
		t := readFloat("到期时间 T (年): ", 0.001, 100.0)
		r := readFloat("无风险利率 r (如0.05表示5%): ", -1.0, 1.0)
		sigma := readFloat("波动率 σ (如0.2表示20%): ", 0.0, 5.0)
		mu := readFloat("漂移率 μ (如0.1表示10%): ", -1.0, 1.0)
		optionType := readString("期权类型 (call/put): ", []string{"call", "put"})
		steps := readInt("时间步数 (建议1或252): ", 1, 1000)
		req.Option = &common.OptionRequest{
			Samples:    samples,
			Confidence: confidence,
			S0:         s0,
			K:          k,
			T:          t,
			R:          r,
			Sigma:      sigma,
			Mu:         mu,
			Type:       optionType,
			Steps:      steps,
		}
	case 3:
		req.Type = common.TypeProbability
		samples := readInt("采样次数 (建议10000+): ", 1, 100000000)
		confidence := readFloat("置信水平 (0.0-1.0, 建议0.95): ", 0.0, 1.0)
		fmt.Println("选择概率事件:")
		fmt.Println("1) 抛硬币正面朝上")
		fmt.Println("2) 两个骰子点数之和大于某阈值")
		eventChoice := readInt("请输入选项 (1-2): ", 1, 2)
		if eventChoice == 1 {
			req.Probability = &common.ProbabilityRequest{
				Samples:    samples,
				Confidence: confidence,
				Event:      common.EventCoinHeads,
			}
		} else {
			threshold := readFloat("阈值 (两个骰子点数之和范围2-12): ", 1.0, 12.0)
			req.Probability = &common.ProbabilityRequest{
				Samples:    samples,
				Confidence: confidence,
				Event:      common.EventDiceSumGreater,
				EventParams: map[string]interface{}{
					"threshold": threshold,
				},
			}
		}
	}

	fmt.Printf("\n连接服务端: %s\n", serverURL)
	resp, err := executeRequest(serverURL, &req)
	if err != nil {
		fmt.Printf("执行失败: %v\n", err)
		os.Exit(1)
	}
	printResponse(resp)
}

func executeRequest(serverURL string, req *common.SimulationRequest) (*common.SimulationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.Post(serverURL+"/simulate", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("连接服务端失败: %v", err)
	}
	defer httpResp.Body.Close()

	var resp common.SimulationResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}
	return &resp, nil
}

func printResponse(resp *common.SimulationResponse) {
	if !resp.Success {
		fmt.Printf("❌ 错误: %s\n", resp.Message)
		return
	}

	fmt.Printf("✓ 模拟完成\n")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-18s %s\n", "指标", "值")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("%-18s %.10f\n", "估算值", resp.Estimate)
	fmt.Printf("%-18s %.10f\n", "标准差", resp.StdDev)
	fmt.Printf("%-18s %.10f\n", "标准误", resp.StdErr)
	fmt.Printf("%-18s %.10f\n", "置信区间下界", resp.CILower)
	fmt.Printf("%-18s %.10f\n", "置信区间上界", resp.CIUpper)
	fmt.Printf("%-18s %d\n", "实际采样数", resp.Samples)
	fmt.Println(strings.Repeat("-", 60))

	if len(resp.Warnings) > 0 {
		fmt.Println("\n⚠️  警告:")
		for _, w := range resp.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	fmt.Println("\n💡 提示: 蒙特卡洛方法收敛速度为O(1/√n)，采样数增加100倍，精度才提升10倍")
}

func readInt(prompt string, min int, max int) int {
	for {
		fmt.Print(prompt)
		var input string
		fmt.Scanln(&input)
		val, err := strconv.Atoi(strings.TrimSpace(input))
		if err != nil || val < min || val > max {
			fmt.Printf("请输入 %d 到 %d 之间的整数\n", min, max)
			continue
		}
		return val
	}
}

func readFloat(prompt string, min float64, max float64) float64 {
	for {
		fmt.Print(prompt)
		var input string
		fmt.Scanln(&input)
		val, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
		if err != nil || val < min || val > max {
			fmt.Printf("请输入 %.4f 到 %.4f 之间的数\n", min, max)
			continue
		}
		return val
	}
}

func readString(prompt string, validOptions []string) string {
	for {
		fmt.Print(prompt)
		var input string
		fmt.Scanln(&input)
		val := strings.TrimSpace(strings.ToLower(input))
		for _, opt := range validOptions {
			if val == opt {
				return val
			}
		}
		fmt.Printf("请输入以下选项之一: %v\n", validOptions)
	}
}

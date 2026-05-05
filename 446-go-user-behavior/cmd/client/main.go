package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"userbehavior/internal/client"
	"userbehavior/internal/shared"
)

const (
	DefaultServerURL = "http://localhost:8080"
)

var (
	apiClient *client.APIClient
	sessionID string
	userID    string
)

func main() {
	if len(os.Args) > 1 {
		apiClient = client.NewAPIClient(os.Args[1])
	} else {
		apiClient = client.NewAPIClient(DefaultServerURL)
	}

	fmt.Println("=")
	fmt.Println("用户行为追踪客户端")
	fmt.Println("=")
	fmt.Printf("服务器地址: %s\n", DefaultServerURL)
	fmt.Println()

	if err := checkServer(); err != nil {
		fmt.Printf("警告: 无法连接到服务器: %v\n", err)
		fmt.Println("请确保服务端已启动 (go run cmd/server/main.go)")
	}

	showHelp()
	runInteractive()
}

func checkServer() error {
	_, err := apiClient.Health()
	return err
}

func showHelp() {
	fmt.Println("\n可用命令:")
	fmt.Println("  health                      - 检查服务器健康状态")
	fmt.Println("  set-user <user_id>          - 设置当前用户ID")
	fmt.Println("  track-page <url> <page_key> [duration_ms]")
	fmt.Println("                              - 记录页面浏览")
	fmt.Println("  track-click <button_id>     - 记录按钮点击")
	fmt.Println("  track-feature <feature_name> [duration_ms]")
	fmt.Println("                              - 记录功能使用")
	fmt.Println("  funnel                      - 交互式漏斗分析")
	fmt.Println("  retention <start_date> <end_date>")
	fmt.Println("                              - 留存分析 (日期格式: YYYY-MM-DD)")
	fmt.Println("  paths <from_page> <to_page> - 路径分析")
	fmt.Println("  profile <user_id>           - 获取用户画像")
	fmt.Println("  realtime                    - 获取实时统计")
	fmt.Println("  help                        - 显示帮助信息")
	fmt.Println("  exit                        - 退出程序")
	fmt.Println()
}

func runInteractive() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\n> ")

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)

		if len(parts) == 0 {
			fmt.Print("> ")
			continue
		}

		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "exit", "quit":
			fmt.Println("再见!")
			return

		case "help":
			showHelp()

		case "health":
			cmdHealth()

		case "set-user":
			if len(parts) < 2 {
				fmt.Println("用法: set-user <user_id>")
			} else {
				userID = parts[1]
				sessionID = ""
				fmt.Printf("已设置用户ID: %s\n", userID)
			}

		case "track-page":
			if len(parts) < 3 {
				fmt.Println("用法: track-page <url> <page_key> [duration_ms]")
			} else {
				duration := int64(0)
				if len(parts) >= 4 {
					duration, _ = strconv.ParseInt(parts[3], 10, 64)
				}
				cmdTrackPage(parts[1], parts[2], duration)
			}

		case "track-click":
			if len(parts) < 2 {
				fmt.Println("用法: track-click <button_id>")
			} else {
				cmdTrackClick(parts[1])
			}

		case "track-feature":
			if len(parts) < 2 {
				fmt.Println("用法: track-feature <feature_name> [duration_ms]")
			} else {
				duration := int64(0)
				if len(parts) >= 3 {
					duration, _ = strconv.ParseInt(parts[2], 10, 64)
				}
				cmdTrackFeature(parts[1], duration)
			}

		case "funnel":
			cmdFunnel()

		case "retention":
			if len(parts) < 3 {
				fmt.Println("用法: retention <start_date> <end_date> (日期格式: YYYY-MM-DD)")
			} else {
				cmdRetention(parts[1], parts[2])
			}

		case "paths":
			if len(parts) < 3 {
				fmt.Println("用法: paths <from_page> <to_page>")
			} else {
				cmdPaths(parts[1], parts[2])
			}

		case "profile":
			if len(parts) < 2 {
				fmt.Println("用法: profile <user_id>")
			} else {
				cmdProfile(parts[1])
			}

		case "realtime":
			cmdRealtime()

		default:
			fmt.Printf("未知命令: %s\n", cmd)
			fmt.Println("输入 'help' 查看可用命令")
		}

		fmt.Print("\n> ")
	}
}

func cmdHealth() {
	ok, err := apiClient.Health()
	if err != nil {
		fmt.Printf("服务器不健康: %v\n", err)
		return
	}
	if ok {
		fmt.Println("服务器状态: 健康 ✓")
	}
}

func requireUserID() bool {
	if userID == "" {
		fmt.Println("请先使用 'set-user <user_id>' 设置用户ID")
		return false
	}
	return true
}

func cmdTrackPage(url, pageKey string, durationMs int64) {
	if !requireUserID() {
		return
	}

	req := &shared.TrackRequest{
		UserID:    userID,
		SessionID: sessionID,
		Type:      shared.BehaviorTypePageView,
		URL:       url,
		PageKey:   pageKey,
		Duration:  durationMs,
	}

	resp, err := apiClient.Track(req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	sessionID = resp.SessionID
	fmt.Printf("成功! 会话ID: %s\n", resp.SessionID)
	fmt.Printf("消息: %s\n", resp.Message)
}

func cmdTrackClick(buttonID string) {
	if !requireUserID() {
		return
	}

	req := &shared.TrackRequest{
		UserID:    userID,
		SessionID: sessionID,
		Type:      shared.BehaviorTypeButtonClick,
		ButtonID:  buttonID,
	}

	resp, err := apiClient.Track(req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	sessionID = resp.SessionID
	fmt.Printf("成功! 会话ID: %s\n", resp.SessionID)
	fmt.Printf("消息: %s\n", resp.Message)
}

func cmdTrackFeature(featureName string, durationMs int64) {
	if !requireUserID() {
		return
	}

	req := &shared.TrackRequest{
		UserID:      userID,
		SessionID:   sessionID,
		Type:        shared.BehaviorTypeFeatureUse,
		FeatureName: featureName,
		Duration:    durationMs,
	}

	resp, err := apiClient.Track(req)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	sessionID = resp.SessionID
	fmt.Printf("成功! 会话ID: %s\n", resp.SessionID)
	fmt.Printf("消息: %s\n", resp.Message)
}

func cmdFunnel() {
	scanner := bufio.NewScanner(os.Stdin)
	steps := make([]shared.FunnelStep, 0)

	fmt.Println("\n漏斗分析向导 (最多10步)")
	fmt.Println("输入 'done' 完成步骤定义")

	for i := 0; i < shared.MaxFunnelSteps; i++ {
		fmt.Printf("\n步骤 %d:\n", i+1)
		fmt.Print("  步骤名称 (如: 浏览商品): ")
		scanner.Scan()
		name := scanner.Text()
		if name == "done" {
			break
		}

		fmt.Print("  行为类型 (page_view/button_click/feature_use): ")
		scanner.Scan()
		behaviorTypeStr := scanner.Text()

		var behaviorType shared.BehaviorType
		switch behaviorTypeStr {
		case "page_view":
			behaviorType = shared.BehaviorTypePageView
		case "button_click":
			behaviorType = shared.BehaviorTypeButtonClick
		case "feature_use":
			behaviorType = shared.BehaviorTypeFeatureUse
		default:
			fmt.Println("无效的行为类型, 跳过此步骤")
			i--
			continue
		}

		step := shared.FunnelStep{
			Name:         name,
			BehaviorType: behaviorType,
		}

		switch behaviorType {
		case shared.BehaviorTypePageView:
			fmt.Print("  页面标识 (page_key): ")
			scanner.Scan()
			step.PageKey = scanner.Text()
		case shared.BehaviorTypeButtonClick:
			fmt.Print("  按钮标识 (button_id): ")
			scanner.Scan()
			step.ButtonID = scanner.Text()
		case shared.BehaviorTypeFeatureUse:
			fmt.Print("  功能名称 (feature_name): ")
			scanner.Scan()
			step.FeatureName = scanner.Text()
		}

		steps = append(steps, step)

		fmt.Print("  继续添加步骤? (y/n/done): ")
		scanner.Scan()
		cont := scanner.Text()
		if cont == "n" || cont == "done" {
			break
		}
	}

	if len(steps) == 0 {
		fmt.Println("未定义任何步骤")
		return
	}

	fmt.Println("\n正在分析漏斗...")
	result, err := apiClient.FunnelAnalysis(steps)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("漏斗分析结果")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-4s %-20s %-12s %-12s %-12s %s\n", "步骤", "名称", "用户数", "转化率", "流失率", "状态")
	fmt.Println(strings.Repeat("-", 80))

	for i, step := range result.Steps {
		status := "可用"
		if !step.IsAvailable {
			status = "不可用"
		}

		convRate := "-"
		churnRate := "-"
		if step.IsAvailable {
			convRate = fmt.Sprintf("%.1f%%", step.ConversionRate)
			churnRate = fmt.Sprintf("%.1f%%", step.ChurnRate)
		}

		fmt.Printf("%-4d %-20s %-12d %-12s %-12s %s\n",
			i+1, step.Name, step.UserCount, convRate, churnRate, status)
	}
}

func cmdRetention(startDateStr, endDateStr string) {
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		fmt.Printf("无效的开始日期格式: %v\n", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		fmt.Printf("无效的结束日期格式: %v\n", err)
		return
	}

	result, err := apiClient.RetentionAnalysis(startDate, endDate)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("留存分析结果")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("分析周期: %s 至 %s\n", startDateStr, endDateStr)
	fmt.Println()
	fmt.Printf("次日留存率:  %.1f%%\n", result.Day1Retention)
	fmt.Printf("7日留存率:   %.1f%%\n", result.Day7Retention)
	fmt.Printf("30日留存率:  %.1f%%\n", result.Day30Retention)
}

func cmdPaths(fromPage, toPage string) {
	result, err := apiClient.PathAnalysis(fromPage, toPage)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if len(result.Paths) == 0 {
		fmt.Printf("\n未找到从 '%s' 到 '%s' 的路径\n", fromPage, toPage)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("路径分析: %s -> %s (Top 10)\n", fromPage, toPage)
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("%-6s %-50s %-12s %s\n", "排名", "路径", "用户数", "占比")
	fmt.Println(strings.Repeat("-", 80))

	for i, path := range result.Paths {
		pathStr := strings.Join(path.Path, " → ")
		fmt.Printf("%-6d %-50s %-12d %.1f%%\n",
			i+1, pathStr, path.UserCount, path.Percentage)
	}
}

func cmdProfile(userID string) {
	profile, err := apiClient.GetUserProfile(userID)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Printf("用户画像: %s\n", userID)
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("活跃等级:    %s\n", profile.ActivityLevel)
	fmt.Printf("首次活跃:    %s\n", profile.FirstActive.Format("2006-01-02 15:04:05"))
	fmt.Printf("最后活跃:    %s\n", profile.LastActive.Format("2006-01-02 15:04:05"))
	fmt.Printf("总会话数:    %d\n", profile.TotalSessions)
	fmt.Println()
	fmt.Println("兴趣标签:")
	for tag, count := range profile.InterestTags {
		fmt.Printf("  %s: %d次\n", tag, count)
	}
}

func cmdRealtime() {
	stats, err := apiClient.GetRealtimeStats()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("实时统计")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("当前在线用户数: %d\n", stats.OnlineUserCount)
	fmt.Println()

	if len(stats.ActivePages) > 0 {
		fmt.Println("活跃页面排行 (近5分钟):")
		fmt.Printf("%-6s %-30s %s\n", "排名", "页面", "访问量")
		fmt.Println(strings.Repeat("-", 50))
		for _, page := range stats.ActivePages {
			fmt.Printf("%-6d %-30s %d\n", page.Rank, page.PageKey, page.UserCount)
		}
	} else {
		fmt.Println("暂无活跃页面数据")
	}
}

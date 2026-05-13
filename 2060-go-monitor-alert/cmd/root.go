package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monitor-alert/internal/alert"
	"monitor-alert/internal/collector"
	"monitor-alert/internal/config"
	"monitor-alert/internal/engine"
	"monitor-alert/internal/notifier"
	"monitor-alert/internal/storage"
	"monitor-alert/internal/types"

	"github.com/spf13/cobra"
)

var (
	cfgFile  string
	dataDir  string
)

var rootCmd = &cobra.Command{
	Use:   "monitor-alert",
	Short: "监控告警系统",
	Long:  `一个基于 Go + Cobra 的监控告警系统，支持系统指标采集、告警规则评估和多渠道通知。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "config.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVar(&dataDir, "data", "./data", "数据目录")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(rulesCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(alertsCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "启动监控告警系统",
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}
	
	ruleEngine, err := engine.NewRuleEngine(cfg.Rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "规则引擎初始化失败: %v\n", err)
		os.Exit(1)
	}
	
	channels, err := config.ParseChannels(cfg.Channels)
	if err != nil {
		return err
	}
	
	notifManager := notifier.NewNotificationManager()
	for _, ch := range channels {
		switch ch.Type {
		case types.ChannelStdout:
			notifManager.Register(ch.Name, notifier.NewStdoutNotifier(ch.Name))
		case types.ChannelEmail:
			emailCfg, ok := ch.Config.(types.EmailConfig)
			if ok {
				notifManager.Register(ch.Name, notifier.NewEmailNotifier(ch.Name, emailCfg))
			}
		case types.ChannelWebhook:
			webhookCfg, ok := ch.Config.(types.WebhookConfig)
			if ok {
				notifManager.Register(ch.Name, notifier.NewWebhookNotifier(ch.Name, webhookCfg))
			}
		}
	}
	
	store, err := storage.NewStorage(cfg.Storage.DataDir)
	if err != nil {
		return err
	}
	
	alertMgr := alert.NewAlertManager()
	sysCollector := collector.NewSystemCollector(cfg.Global.CollectInterval)
	
	ruleStartTimes := make(map[string]time.Time)
	for _, rule := range cfg.Rules {
		ruleStartTimes[rule.Name] = time.Time{}
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	collectTicker := time.NewTicker(cfg.Global.CollectInterval)
	defer collectTicker.Stop()
	
	evalTicker := time.NewTicker(cfg.Global.EvaluateInterval)
	defer evalTicker.Stop()
	
	log.Printf("监控告警系统已启动")
	log.Printf("采集间隔: %v, 评估间隔: %v", cfg.Global.CollectInterval, cfg.Global.EvaluateInterval)
	
	var lastMetrics []types.Metric
	
	for {
		select {
		case <-sigChan:
			log.Printf("收到终止信号，正在退出...")
			return nil
		case <-collectTicker.C:
			metrics, err := sysCollector.Collect(ctx)
			if err != nil {
				log.Printf("WARN: 指标采集出错: %v", err)
			}
			if len(metrics) > 0 {
				lastMetrics = metrics
				if err := store.Save(metrics); err != nil {
					log.Printf("WARN: 保存指标失败: %v", err)
				}
			}
		case <-evalTicker.C:
			if len(lastMetrics) == 0 {
				continue
			}
			
			for _, rule := range ruleEngine.Rules() {
				startTime := ruleStartTimes[rule.Name]
				firing, message, err := ruleEngine.Evaluate(rule, lastMetrics, startTime)
				if err != nil {
					log.Printf("WARN: 评估规则 '%s' 失败: %v", rule.Name, err)
					continue
				}
				
				now := time.Now()
				
				if firing {
					if startTime.IsZero() {
						ruleStartTimes[rule.Name] = now
						continue
					}
					
					if alertMgr.IsActive(rule.Name) {
						if rule.ReminderInterval > 0 && alertMgr.ShouldRemind(rule.Name, rule.ReminderInterval) {
							alert, _ := alertMgr.GetActive(rule.Name)
							alertMgr.UpdateLastNotified(rule.Name)
							notifManager.Send(ctx, *alert, types.AlertStatusFiring, rule.Channels)
						}
					} else {
						newAlert := alertMgr.Trigger(rule, message)
						notifManager.Send(ctx, *newAlert, types.AlertStatusFiring, rule.Channels)
					}
				} else {
					ruleStartTimes[rule.Name] = time.Time{}
					if alertMgr.IsActive(rule.Name) {
						resolvedAlert := alertMgr.Resolve(rule.Name)
						if resolvedAlert != nil {
							notifManager.Send(ctx, *resolvedAlert, types.AlertStatusResolved, rule.Channels)
						}
					}
				}
			}
		}
	}
}

var queryCmd = &cobra.Command{
	Use:   "query [metric]",
	Short: "查询历史指标数据",
	Args:  cobra.ExactArgs(1),
	RunE:  runQuery,
}

var (
	queryDuration time.Duration
	queryTrend    bool
)

func init() {
	queryCmd.Flags().DurationVarP(&queryDuration, "duration", "d", 1*time.Hour, "查询时间范围")
	queryCmd.Flags().BoolVarP(&queryTrend, "trend", "t", false, "显示趋势分析")
}

func runQuery(cmd *cobra.Command, args []string) error {
	metricName := args[0]
	
	store, err := storage.NewStorage(dataDir)
	if err != nil {
		return err
	}
	
	if queryTrend {
		result, err := store.TrendAnalysis(metricName, queryDuration)
		if err != nil {
			return err
		}
		fmt.Printf("趋势分析: %s\n", result.MetricName)
		fmt.Printf("  数据点数量: %d\n", result.Count)
		fmt.Printf("  平均值: %.2f\n", result.Average)
		fmt.Printf("  最小值: %.2f\n", result.Min)
		fmt.Printf("  最大值: %.2f\n", result.Max)
		fmt.Printf("  最新值: %.2f\n", result.Latest)
		fmt.Printf("  趋势: %s\n", result.Trend)
	} else {
		metrics, err := store.Query(metricName, time.Now().Add(-queryDuration), time.Now())
		if err != nil {
			return err
		}
		fmt.Printf("查询 %s 最近 %v 的数据:\n", metricName, queryDuration)
		for i, m := range metrics {
			if i >= 10 {
				fmt.Printf("  ... (还有 %d 条记录)\n", len(metrics)-i)
				break
			}
			fmt.Printf("  %s: %.2f %s\n", m.Timestamp.Format(time.RFC3339), m.Value, m.Unit)
		}
		if len(metrics) == 0 {
			fmt.Println("  无数据")
		}
	}
	
	return nil
}

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "列出所有告警规则",
	RunE:  runRules,
}

func runRules(cmd *cobra.Command, args []string) error {
	cfg, err := config.LoadConfig(cfgFile)
	if err != nil {
		return err
	}
	
	ruleEngine, err := engine.NewRuleEngine(cfg.Rules)
	if err != nil {
		fmt.Fprintf(os.Stderr, "规则引擎初始化失败: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("告警规则列表 (共 %d 条):\n\n", len(ruleEngine.Rules()))
	for _, rule := range ruleEngine.Rules() {
		fmt.Printf("名称: %s\n", rule.Name)
		if rule.Description != "" {
			fmt.Printf("描述: %s\n", rule.Description)
		}
		fmt.Printf("条件: %s\n", rule.Condition)
		fmt.Printf("持续时间: %v\n", rule.Duration)
		fmt.Printf("严重级别: %s\n", rule.Severity)
		fmt.Printf("提醒间隔: %v\n", rule.ReminderInterval)
		fmt.Printf("通知渠道: %v\n\n", rule.Channels)
	}
	
	return nil
}

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "显示统计信息",
	RunE:  runStats,
}

var (
	adjustTotal int64
)

func init() {
	statsCmd.Flags().Int64Var(&adjustTotal, "adjust-total", 0, "调整总数并按比例分配配额")
}

func runStats(cmd *cobra.Command, args []string) error {
	store, err := storage.NewStorage(dataDir)
	if err != nil {
		return err
	}
	
	if adjustTotal > 0 {
		if err := store.AdjustTotal(adjustTotal); err != nil {
			return err
		}
		fmt.Printf("已将总数调整为 %d\n\n", adjustTotal)
	}
	
	stats := store.Stats()
	fmt.Println("统计信息:")
	fmt.Printf("  总记录数: %d\n", stats.TotalRecords)
	fmt.Printf("  指标计数:\n")
	for name, count := range stats.MetricsCount {
		fmt.Printf("    %s: %d\n", name, count)
	}
	fmt.Printf("  告警数: %d\n", stats.AlertsCount)
	fmt.Printf("  最后更新: %s\n", stats.LastUpdated.Format(time.RFC3339))
	
	return nil
}

var alertsCmd = &cobra.Command{
	Use:   "alerts",
	Short: "查看活动告警",
	RunE:  runAlerts,
}

func runAlerts(cmd *cobra.Command, args []string) error {
	store, err := storage.NewStorage(dataDir)
	if err != nil {
		return err
	}
	
	_ = store
	fmt.Println("注意: 活动告警存储在内存中，需要运行 'start' 命令查看。")
	fmt.Println("历史告警记录请查看数据目录。")
	
	return nil
}

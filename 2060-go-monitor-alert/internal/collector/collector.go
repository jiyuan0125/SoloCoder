package collector

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"time"

	"monitor-alert/internal/types"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Collector interface {
	Collect(ctx context.Context) ([]types.Metric, error)
	Name() string
}

type SystemCollector struct {
	interval time.Duration
}

func NewSystemCollector(interval time.Duration) *SystemCollector {
	return &SystemCollector{interval: interval}
}

func (c *SystemCollector) Collect(ctx context.Context) ([]types.Metric, error) {
	var metrics []types.Metric
	now := time.Now()

	if cpuPercent, err := collectCPU(); err == nil {
		metrics = append(metrics, types.Metric{
			Name:      "cpu_usage",
			Type:      types.MetricCPU,
			Value:     cpuPercent,
			Unit:      "percent",
			Timestamp: now,
		})
	} else {
		log.Printf("WARN: 采集 CPU 指标失败: %v，跳过该指标", err)
	}

	if memPercent, err := collectMemory(); err == nil {
		metrics = append(metrics, types.Metric{
			Name:      "memory_usage",
			Type:      types.MetricMemory,
			Value:     memPercent,
			Unit:      "percent",
			Timestamp: now,
		})
	} else {
		log.Printf("WARN: 采集内存指标失败: %v，跳过该指标", err)
	}

	if diskMetrics, err := collectDisk(); err == nil {
		for _, m := range diskMetrics {
			m.Timestamp = now
			metrics = append(metrics, m)
		}
	} else {
		log.Printf("WARN: 采集磁盘指标失败: %v，跳过该指标", err)
	}

	if netMetrics, err := collectNetwork(); err == nil {
		for _, m := range netMetrics {
			m.Timestamp = now
			metrics = append(metrics, m)
		}
	} else {
		log.Printf("WARN: 采集网络指标失败: %v，跳过该指标", err)
	}

	return metrics, nil
}

func (c *SystemCollector) Name() string {
	return "system_collector"
}

func collectCPU() (float64, error) {
	percents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, fmt.Errorf("获取 CPU 使用率失败: %w", err)
	}
	if len(percents) == 0 {
		return 0, fmt.Errorf("未获取到 CPU 使用率数据")
	}
	return percents[0], nil
}

func collectMemory() (float64, error) {
	vm, err := mem.VirtualMemory()
	if err != nil {
		return 0, fmt.Errorf("获取内存信息失败: %w", err)
	}
	return vm.UsedPercent, nil
}

func collectDisk() ([]types.Metric, error) {
	var metrics []types.Metric
	
	if runtime.GOOS == "windows" {
		partitions, err := disk.Partitions(false)
		if err != nil {
			return nil, fmt.Errorf("获取磁盘分区失败: %w", err)
		}
		for _, p := range partitions {
			usage, err := disk.Usage(p.Mountpoint)
			if err != nil {
				log.Printf("WARN: 获取分区 %s 使用信息失败: %v", p.Mountpoint, err)
				continue
			}
			metrics = append(metrics, types.Metric{
				Name:  "disk_usage",
				Type:  types.MetricDisk,
				Value: usage.UsedPercent,
				Unit:  "percent",
				Labels: map[string]string{
					"device":    p.Device,
					"mountpoint": p.Mountpoint,
				},
			})
		}
	} else {
		usage, err := disk.Usage("/")
		if err != nil {
			return nil, fmt.Errorf("获取磁盘使用信息失败: %w", err)
		}
		metrics = append(metrics, types.Metric{
			Name:  "disk_usage",
			Type:  types.MetricDisk,
			Value: usage.UsedPercent,
			Unit:  "percent",
			Labels: map[string]string{
				"mountpoint": "/",
			},
		})
	}
	
	return metrics, nil
}

func collectNetwork() ([]types.Metric, error) {
	io, err := net.IOCounters(false)
	if err != nil {
		return nil, fmt.Errorf("获取网络 IO 信息失败: %w", err)
	}
	
	if len(io) == 0 {
		return nil, fmt.Errorf("未获取到网络 IO 数据")
	}
	
	now := time.Now()
	
	metrics := []types.Metric{
		{
			Name:      "network_bytes_sent",
			Type:      types.MetricNetwork,
			Value:     float64(io[0].BytesSent),
			Unit:      "bytes",
			Timestamp: now,
		},
		{
			Name:      "network_bytes_recv",
			Type:      types.MetricNetwork,
			Value:     float64(io[0].BytesRecv),
			Unit:      "bytes",
			Timestamp: now,
		},
	}
	
	return metrics, nil
}

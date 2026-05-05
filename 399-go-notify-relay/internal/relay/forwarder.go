package relay

import (
	"log"
	"sync"
	"time"

	"github.com/notify-relay/internal/protocol"
)

type Forwarder struct {
	channels     map[string]Channel
	channelStats map[string]int64
	totalAlerts  int64
	forwardLogs  []protocol.ForwardLog
	maxLogs      int
	mu           sync.RWMutex
}

func NewForwarder() *Forwarder {
	allChannels := GetAllChannels()
	channelMap := make(map[string]Channel)
	channelStats := make(map[string]int64)

	for _, ch := range allChannels {
		channelMap[ch.Name()] = ch
		channelStats[ch.Name()] = 0
	}

	return &Forwarder{
		channels:     channelMap,
		channelStats: channelStats,
		forwardLogs:  make([]protocol.ForwardLog, 0),
		maxLogs:      100,
	}
}

func (f *Forwarder) Forward(alert *Alert) {
	f.mu.Lock()
	f.totalAlerts++
	f.mu.Unlock()

	channelsToUse := GetChannelsForLevel(alert.Level)

	var wg sync.WaitGroup

	for _, channelName := range channelsToUse {
		wg.Add(1)
		go func(chName string) {
			defer wg.Done()
			f.sendToChannel(alert, chName)
		}(channelName)
	}

	wg.Wait()
}

func (f *Forwarder) sendToChannel(alert *Alert, channelName string) {
	ch, exists := f.channels[channelName]
	if !exists {
		log.Printf("未知渠道: %s", channelName)
		return
	}

	startTime := time.Now()
	err := ch.Send(alert)
	duration := time.Since(startTime)

	f.mu.Lock()
	defer f.mu.Unlock()

	if err == nil {
		f.channelStats[channelName]++
	}

	logEntry := protocol.ForwardLog{
		Timestamp: startTime.Format(time.RFC3339),
		Level:     string(alert.Level),
		Message:   alert.Message,
		Channel:   channelName,
		Status:    "success",
	}

	if err != nil {
		logEntry.Status = "failed"
		logEntry.Error = err.Error()
		log.Printf("渠道 %s 转发失败: %v (耗时: %v)", channelName, err, duration)
	} else {
		log.Printf("渠道 %s 转发成功 (耗时: %v)", channelName, duration)
	}

	f.forwardLogs = append(f.forwardLogs, logEntry)
	if len(f.forwardLogs) > f.maxLogs {
		f.forwardLogs = f.forwardLogs[1:]
	}
}

func (f *Forwarder) GetTotalAlerts() int64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.totalAlerts
}

func (f *Forwarder) GetChannelStats() map[string]int64 {
	f.mu.RLock()
	defer f.mu.RUnlock()
	stats := make(map[string]int64)
	for k, v := range f.channelStats {
		stats[k] = v
	}
	return stats
}

func (f *Forwarder) GetForwardLogs() []protocol.ForwardLog {
	f.mu.RLock()
	defer f.mu.RUnlock()
	logs := make([]protocol.ForwardLog, len(f.forwardLogs))
	copy(logs, f.forwardLogs)
	return logs
}

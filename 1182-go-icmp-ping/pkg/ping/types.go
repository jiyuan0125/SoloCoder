package ping

import (
	"net"
	"time"
)

const (
	ICMPTypeEchoReply       = 0
	ICMPTypeEchoRequest     = 8
	ICMPTypeTimeExceeded    = 11
	ICMPCodeTTLExceeded     = 0
	ICMPCodeFragmentReassemblyTimeExceeded = 1
	DefaultTTL              = 128
	DefaultInterval         = 1 * time.Second
	DefaultTimeout          = 5 * time.Second
	DefaultCount            = -1
	IPv4HeaderLen           = 20
	ICMPHeaderLen           = 8
	TimestampSize           = 8
)

type Config struct {
	Target   string
	Count    int
	Interval time.Duration
	Timeout  time.Duration
	TTL      int
}

type Result struct {
	RTT        time.Duration
	TargetIP   net.IP
	Seq        uint16
	TTL        uint8
	IsTimeout  bool
	TimeExceeded *TimeExceededInfo
}

type TimeExceededInfo struct {
	RouterIP   net.IP
	OriginalTarget net.IP
	Seq        uint16
}

type Stats struct {
	PacketsSent     int
	PacketsReceived int
	PacketLoss      float64
	MinRTT          time.Duration
	MaxRTT          time.Duration
	AvgRTT          time.Duration
	StdDevRTT       time.Duration
	TimeExceeded    []*TimeExceededInfo
}

type Task struct {
	ID         string
	Config     *Config
	Stats      *Stats
	IsRunning  bool
	CreatedAt  time.Time
	FinishedAt *time.Time
}

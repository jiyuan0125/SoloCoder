package core

import (
	"time"

	"github.com/health-aggregator/pkg/common"
)

const (
	DefaultInterval           = 10 * time.Second
	DefaultTimeout            = 3 * time.Second
	DefaultFailureThreshold   = 3
	DefaultRecoveryThreshold  = 2
	DefaultHistorySize        = 50
	DefaultUnhealthyThreshold = 0.5
	DefaultConcurrency        = 10
)

type GlobalConfig struct {
	Interval           time.Duration `yaml:"interval,omitempty"`
	Timeout            time.Duration `yaml:"timeout,omitempty"`
	FailureThreshold   int           `yaml:"failure_threshold,omitempty"`
	RecoveryThreshold  int           `yaml:"recovery_threshold,omitempty"`
	HistorySize        int           `yaml:"history_size,omitempty"`
	UnhealthyThreshold float64       `yaml:"unhealthy_threshold,omitempty"`
	Concurrency        int           `yaml:"concurrency,omitempty"`
}

func DefaultGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Interval:           DefaultInterval,
		Timeout:            DefaultTimeout,
		FailureThreshold:   DefaultFailureThreshold,
		RecoveryThreshold:  DefaultRecoveryThreshold,
		HistorySize:        DefaultHistorySize,
		UnhealthyThreshold: DefaultUnhealthyThreshold,
		Concurrency:        DefaultConcurrency,
	}
}

func ApplyDefaults(svc *common.ServiceConfig, global *GlobalConfig) {
	if svc.Interval <= 0 {
		svc.Interval = global.Interval
	}
	if svc.Timeout <= 0 {
		svc.Timeout = global.Timeout
	}
	if svc.FailureThreshold <= 0 {
		svc.FailureThreshold = global.FailureThreshold
	}
	if svc.RecoveryThreshold <= 0 {
		svc.RecoveryThreshold = global.RecoveryThreshold
	}
	if svc.ProbeType == common.ProbeTypeHTTP && len(svc.ExpectedStatusCodes) == 0 {
		svc.ExpectedStatusCodes = []int{200, 201, 202, 203, 204, 205, 206, 207, 208, 226}
	}
}

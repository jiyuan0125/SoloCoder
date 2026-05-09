package leaf

import "time"

type Config struct {
	BusinessKey           string
	CenterURL             string
	CenterTimeout         time.Duration
	MaxRetries            int
	RetryInterval         time.Duration
	PreloadThresholdRatio float64
	WaitTimeout           time.Duration
}

func DefaultConfig(businessKey, centerURL string) *Config {
	return &Config{
		BusinessKey:           businessKey,
		CenterURL:             centerURL,
		CenterTimeout:         5 * time.Second,
		MaxRetries:            3,
		RetryInterval:         1 * time.Second,
		PreloadThresholdRatio: 0.1,
		WaitTimeout:           30 * time.Second,
	}
}

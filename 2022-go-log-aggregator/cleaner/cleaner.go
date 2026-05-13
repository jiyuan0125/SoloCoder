package cleaner

import (
	"log"
	"time"

	"log-aggregator/db"
)

type Cleaner struct {
	interval time.Duration
	stop     chan struct{}
}

func NewCleaner() *Cleaner {
	return &Cleaner{
		interval: 1 * time.Hour,
		stop:     make(chan struct{}),
	}
}

func (c *Cleaner) Start() {
	ticker := time.NewTicker(c.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.clean()
			case <-c.stop:
				return
			}
		}
	}()
}

func (c *Cleaner) Stop() {
	close(c.stop)
}

func (c *Cleaner) clean() {
	retentionDays, err := db.GetRetentionDays()
	if err != nil {
		log.Printf("Failed to get retention days: %v", err)
		return
	}

	if err := db.CleanOldLogs(retentionDays); err != nil {
		log.Printf("Failed to clean old logs: %v", err)
	}
}

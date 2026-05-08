package watcher

import "time"

const (
	DefaultDebounce = 500 * time.Millisecond
	DefaultMaxWait  = 5 * time.Second
)

type WatchConfig struct {
	Path     string
	Debounce time.Duration
	MaxWait  time.Duration
	Callback func(event interface{})
}

func (c *WatchConfig) Validate() error {
	if c.Path == "" {
		return ErrPathRequired
	}
	if c.Debounce <= 0 {
		c.Debounce = DefaultDebounce
	}
	if c.MaxWait <= 0 {
		c.MaxWait = DefaultMaxWait
	}
	if c.Callback == nil {
		return ErrCallbackRequired
	}
	return nil
}

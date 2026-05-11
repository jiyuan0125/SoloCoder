package ping

import (
	"context"
	"net"
	"sync"
	"time"
)

type Pinger struct {
	config *Config
	stats  *Statistics
	mu     sync.RWMutex
}

func NewPinger(config *Config) *Pinger {
	if config.TTL == 0 {
		config.TTL = DefaultTTL
	}
	if config.Interval == 0 {
		config.Interval = DefaultInterval
	}
	if config.Timeout == 0 {
		config.Timeout = DefaultTimeout
	}
	if config.Count == 0 {
		config.Count = DefaultCount
	}

	return &Pinger{
		config: config,
		stats:  NewStatistics(),
	}
}

func (p *Pinger) Run(ctx context.Context, resultChan chan<- *Result) error {
	targetIP, err := ResolveTarget(p.config.Target)
	if err != nil {
		return err
	}

	sock, err := NewRawSocket(p.config.TTL)
	if err != nil {
		return err
	}
	defer sock.Close()

	identifier := GetProcessID()
	var seq uint16 = 0

	recvCtx, recvCancel := context.WithCancel(ctx)
	defer recvCancel()

	recvErrChan := make(chan error, 1)

	go p.receiveLoop(recvCtx, sock, targetIP, identifier, resultChan, recvErrChan)

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	count := 0
	for {
		if p.config.Count > 0 && count >= p.config.Count {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-recvErrChan:
			return err
		case <-ticker.C:
			now := time.Now()
			packet := BuildICMPRequest(identifier, seq, now)

			if err := sock.SendTo(targetIP, packet); err != nil {
				return err
			}

			p.mu.Lock()
			p.stats.AddSent()
			p.mu.Unlock()

			seq++
			count++
		}
	}

	time.Sleep(p.config.Timeout)
	return nil
}

func (p *Pinger) receiveLoop(ctx context.Context, sock *RawSocket, targetIP net.IP, identifier uint16, resultChan chan<- *Result, errChan chan<- error) {
	buf := make([]byte, 1500)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			sock.SetReadTimeout(100 * time.Millisecond)

			n, _, err := sock.RecvFrom(buf)
			if err != nil {
				if isTimeoutError(err) {
					continue
				}
				errChan <- err
				return
			}

			if n < IPv4HeaderLen+ICMPHeaderLen {
				continue
			}

			result, err := ParseReceivedPacket(buf[:n], identifier, targetIP)
			if err != nil {
				continue
			}

			if result == nil {
				continue
			}

			p.mu.Lock()
			if result.TimeExceeded != nil {
				p.stats.AddTimeExceeded(result.TimeExceeded)
			} else if !result.IsTimeout {
				p.stats.AddReceived(result.RTT)
			}
			p.mu.Unlock()

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "resource temporarily unavailable" || 
		errStr == "connection timed out" || 
		errStr == "i/o timeout"
}

func (p *Pinger) GetStats() *Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats.GetStats()
}

func (p *Pinger) GetConfig() *Config {
	return p.config
}

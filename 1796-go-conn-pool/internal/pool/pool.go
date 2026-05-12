package pool

import (
	"conn-pool/internal/config"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type PoolStats struct {
	Active          int64
	Idle            int64
	TotalBorrows    int64
	TotalRecreations int64
}

type Pool struct {
	config atomic.Pointer[config.PoolConfig]
	address string
	
	idleConns   chan *PooledConnection
	inUseCount  atomic.Int64
	totalCount  atomic.Int64
	targetMaxOpen atomic.Int64
	
	stats struct {
		borrows      atomic.Int64
		recreations  atomic.Int64
	}
	
	mu sync.Mutex
	
	closed atomic.Bool
}

func NewPool(cfg config.PoolConfig) *Pool {
	p := &Pool{
		address: cfg.Address,
	}
	c := cfg
	p.config.Store(&c)
	p.targetMaxOpen.Store(int64(cfg.MaxOpen))
	p.idleConns = make(chan *PooledConnection, cfg.MaxOpen)
	p.warmUp()
	return p
}

func (p *Pool) warmUp() {
	cfg := p.config.Load()
	minIdle := cfg.MinIdle
	if minIdle <= 0 {
		return
	}
	
	log.Printf("[WARMUP] 开始预热连接池: %s, 目标: %d 个连接", p.address, minIdle)
	
	created := 0
	for i := 0; i < minIdle; i++ {
		conn, err := p.createConnection()
		if err != nil {
			log.Printf("[WARN] 预热失败 (第 %d 个连接): %v", i+1, err)
			continue
		}
		p.idleConns <- conn
		p.totalCount.Add(1)
		created++
	}
	
	log.Printf("[WARMUP] 预热完成: %s, 成功创建: %d/%d 个连接", p.address, created, minIdle)
}

func (p *Pool) createConnection() (*PooledConnection, error) {
	cfg := p.config.Load()
	conn, err := net.DialTimeout("tcp", cfg.Address, 5*time.Second)
	if err != nil {
		log.Printf("[ERROR] 创建连接失败: 地址=%s, 错误=%v", cfg.Address, err)
		return nil, err
	}
	return newPooledConnection(conn), nil
}

func (p *Pool) Get() (*PooledConnection, error) {
	if p.closed.Load() {
		return nil, nil
	}
	
	cfg := p.config.Load()
	
	for {
		select {
		case conn := <-p.idleConns:
			if conn.IsClosed() {
				p.totalCount.Add(-1)
				continue
			}
			
			uses := conn.GetUses()
			if uses >= int64(cfg.MaxUsesPerConn) {
				go p.recreateAndReplace(conn)
				continue
			}
			
			p.inUseCount.Add(1)
			p.stats.borrows.Add(1)
			return conn, nil
		default:
			if p.totalCount.Load() < p.targetMaxOpen.Load() {
				conn, err := p.createConnection()
				if err != nil {
					return nil, err
				}
				p.totalCount.Add(1)
				p.inUseCount.Add(1)
				p.stats.borrows.Add(1)
				return conn, nil
			}
			
			conn := <-p.idleConns
			if conn.IsClosed() {
				p.totalCount.Add(-1)
				continue
			}
			
			uses := conn.GetUses()
			if uses >= int64(cfg.MaxUsesPerConn) {
				go p.recreateAndReplace(conn)
				continue
			}
			
			p.inUseCount.Add(1)
			p.stats.borrows.Add(1)
			return conn, nil
		}
	}
}

func (p *Pool) recreateAndReplace(old *PooledConnection) {
	p.stats.recreations.Add(1)
	log.Printf("[INFO] 连接达到最大使用次数，准备重建: %s, 使用次数=%d", p.address, old.GetUses())
	
	old.Close()
	
	newConn, err := p.createConnection()
	if err != nil {
		p.totalCount.Add(-1)
		log.Printf("[ERROR] 重建连接失败: %v", err)
		return
	}
	
	select {
	case p.idleConns <- newConn:
		log.Printf("[INFO] 连接重建完成并归还: %s", p.address)
	default:
		newConn.Close()
		p.totalCount.Add(-1)
	}
}

func (p *Pool) Put(conn *PooledConnection) {
	if conn == nil || conn.IsClosed() {
		return
	}
	
	cfg := p.config.Load()
	
	conn.IncrementUses()
	p.inUseCount.Add(-1)
	
	if conn.GetUses() >= int64(cfg.MaxUsesPerConn) {
		go p.recreateAndReplace(conn)
		return
	}
	
	if p.totalCount.Load() > p.targetMaxOpen.Load() {
		p.mu.Lock()
		if p.totalCount.Load() > p.targetMaxOpen.Load() {
			conn.Close()
			p.totalCount.Add(-1)
			p.mu.Unlock()
			log.Printf("[INFO] 连接归还时关闭 (缩容中): %s", p.address)
			return
		}
		p.mu.Unlock()
	}
	
	select {
	case p.idleConns <- conn:
	default:
		conn.Close()
		p.totalCount.Add(-1)
	}
}

func (p *Pool) SetMaxOpen(newMax int) {
	if newMax < 1 {
		newMax = 1
	}
	
	oldMax := p.targetMaxOpen.Swap(int64(newMax))
	cfg := *p.config.Load()
	cfg.MaxOpen = newMax
	p.config.Store(&cfg)
	
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if newMax > int(oldMax) {
		log.Printf("[INFO] 扩容连接池: %s, %d -> %d", p.address, oldMax, newMax)
		needed := int64(newMax) - p.totalCount.Load()
		for i := int64(0); i < needed; i++ {
			conn, err := p.createConnection()
			if err != nil {
				continue
			}
			select {
			case p.idleConns <- conn:
				p.totalCount.Add(1)
			default:
				conn.Close()
			}
		}
	} else if newMax < int(oldMax) {
		log.Printf("[INFO] 缩容连接池: %s, %d -> %d", p.address, oldMax, newMax)
		p.trimIdleLocked()
	}
}

func (p *Pool) trimIdleLocked() {
	target := p.targetMaxOpen.Load()
	for p.totalCount.Load() > target {
		select {
		case conn := <-p.idleConns:
			conn.Close()
			p.totalCount.Add(-1)
			log.Printf("[INFO] 缩容关闭空闲连接: %s", p.address)
		default:
			return
		}
	}
}

func (p *Pool) GetStats() PoolStats {
	return PoolStats{
		Active:          p.inUseCount.Load(),
		Idle:            p.totalCount.Load() - p.inUseCount.Load(),
		TotalBorrows:    p.stats.borrows.Load(),
		TotalRecreations: p.stats.recreations.Load(),
	}
}

func (p *Pool) GetConfig() config.PoolConfig {
	return *p.config.Load()
}

func (p *Pool) Close() {
	if !p.closed.CompareAndSwap(false, true) {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for {
		select {
		case conn := <-p.idleConns:
			conn.Close()
		default:
			return
		}
	}
}

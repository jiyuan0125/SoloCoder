package dhcp

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync"
	"time"
)

const DefaultLeaseTime uint32 = 86400

type Lease struct {
	MAC         net.HardwareAddr
	IP          net.IP
	StartTime   time.Time
	Duration    uint32
	ExpireTime  time.Time
	T1Time      time.Time
	T2Time      time.Time
	ServerIP    net.IP
}

type IPPool struct {
	mu          sync.RWMutex
	startIP     uint32
	endIP       uint32
	available   map[uint32]bool
	reserved    map[uint32]string
	leases      map[string]*Lease
	ipToLease   map[uint32]*Lease
	serverIP    net.IP
	subnetMask  net.IPMask
	routers     []net.IP
	dnsServers  []net.IP
	leaseTime   uint32
}

func ipToUint32(ip net.IP) uint32 {
	ip4 := ip.To4()
	if ip4 == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip4)
}

func uint32ToIP(n uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	return ip
}

func NewIPPool(startIP, endIP net.IP) (*IPPool, error) {
	start := ipToUint32(startIP)
	end := ipToUint32(endIP)
	if start == 0 || end == 0 {
		return nil, fmt.Errorf("invalid IP addresses")
	}
	if start > end {
		return nil, fmt.Errorf("start IP must be less than or equal to end IP")
	}
	pool := &IPPool{
		startIP:   start,
		endIP:     end,
		available: make(map[uint32]bool),
		reserved:  make(map[uint32]string),
		leases:    make(map[string]*Lease),
		ipToLease: make(map[uint32]*Lease),
		leaseTime: DefaultLeaseTime,
	}
	for ip := start; ip <= end; ip++ {
		pool.available[ip] = true
	}
	return pool, nil
}

func (p *IPPool) SetServerIP(ip net.IP) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.serverIP = ip
}

func (p *IPPool) SetSubnetMask(mask net.IPMask) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subnetMask = mask
}

func (p *IPPool) SetRouters(ips []net.IP) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.routers = make([]net.IP, len(ips))
	copy(p.routers, ips)
}

func (p *IPPool) SetDNSServers(ips []net.IP) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.dnsServers = make([]net.IP, len(ips))
	copy(p.dnsServers, ips)
}

func (p *IPPool) SetLeaseTime(seconds uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.leaseTime = seconds
}

func (p *IPPool) ServerIP() net.IP {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.serverIP
}

func (p *IPPool) SubnetMask() net.IPMask {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.subnetMask
}

func (p *IPPool) Routers() []net.IP {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.routers
}

func (p *IPPool) DNSServers() []net.IP {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.dnsServers
}

func (p *IPPool) LeaseTime() uint32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.leaseTime
}

func (p *IPPool) AvailableCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	count := 0
	for _, avail := range p.available {
		if avail {
			count++
		}
	}
	return count
}

func (p *IPPool) TotalCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return int(p.endIP - p.startIP + 1)
}

func (p *IPPool) AddRange(startIP, endIP net.IP) error {
	start := ipToUint32(startIP)
	end := ipToUint32(endIP)
	if start == 0 || end == 0 {
		return fmt.Errorf("invalid IP addresses")
	}
	if start > end {
		return fmt.Errorf("start IP must be less than or equal to end IP")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for ip := start; ip <= end; ip++ {
		if ip >= p.startIP && ip <= p.endIP {
			continue
		}
		p.available[ip] = true
	}
	if start < p.startIP {
		p.startIP = start
	}
	if end > p.endIP {
		p.endIP = end
	}
	return nil
}

func (p *IPPool) RemoveRange(startIP, endIP net.IP) error {
	start := ipToUint32(startIP)
	end := ipToUint32(endIP)
	if start == 0 || end == 0 {
		return fmt.Errorf("invalid IP addresses")
	}
	if start > end {
		return fmt.Errorf("start IP must be less than or equal to end IP")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for ip := start; ip <= end; ip++ {
		if _, ok := p.ipToLease[ip]; ok {
			return fmt.Errorf("IP %s is in use", uint32ToIP(ip))
		}
		delete(p.available, ip)
	}
	return nil
}

func (p *IPPool) Allocate(mac net.HardwareAddr) (net.IP, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	macStr := mac.String()
	if lease, ok := p.leases[macStr]; ok {
		if !lease.ExpireTime.Before(time.Now()) {
			return nil, fmt.Errorf("MAC %s already has an active lease", macStr)
		}
	}
	var allocatedIP uint32
	found := false
	for ip := p.startIP; ip <= p.endIP; ip++ {
		if p.available[ip] {
			allocatedIP = ip
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("no available IP addresses in pool")
	}
	delete(p.available, allocatedIP)
	p.reserved[allocatedIP] = macStr
	return uint32ToIP(allocatedIP), nil
}

func (p *IPPool) Reserve(mac net.HardwareAddr, ip net.IP) error {
	ipn := ipToUint32(ip)
	if ipn == 0 {
		return fmt.Errorf("invalid IP address")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.ipToLease[ipn]; ok {
		return fmt.Errorf("IP %s is already in use", ip)
	}
	if existingMAC, ok := p.reserved[ipn]; ok {
		if existingMAC != mac.String() {
			return fmt.Errorf("IP %s is already reserved by another client", ip)
		}
		return nil
	}
	if !p.available[ipn] {
		return fmt.Errorf("IP %s is not in pool", ip)
	}
	delete(p.available, ipn)
	p.reserved[ipn] = mac.String()
	return nil
}

func (p *IPPool) GrantLease(mac net.HardwareAddr, ip net.IP) *Lease {
	p.mu.Lock()
	defer p.mu.Unlock()
	ipn := ipToUint32(ip)
	if ipn == 0 {
		return nil
	}
	macStr := mac.String()
	if _, ok := p.ipToLease[ipn]; ok {
		return nil
	}
	if reservedMAC, ok := p.reserved[ipn]; ok {
		if reservedMAC != macStr {
			return nil
		}
		delete(p.reserved, ipn)
	} else if p.available[ipn] {
		delete(p.available, ipn)
	} else {
		return nil
	}
	now := time.Now()
	duration := p.leaseTime
	lease := &Lease{
		MAC:        mac,
		IP:         ip,
		StartTime:  now,
		Duration:   duration,
		ExpireTime: now.Add(time.Duration(duration) * time.Second),
		T1Time:     now.Add(time.Duration(duration/2) * time.Second),
		T2Time:     now.Add(time.Duration(duration*7/8) * time.Second),
		ServerIP:   p.serverIP,
	}
	if oldLease, ok := p.leases[macStr]; ok {
		oldIP := ipToUint32(oldLease.IP)
		delete(p.ipToLease, oldIP)
		p.available[oldIP] = true
	}
	p.leases[macStr] = lease
	p.ipToLease[ipn] = lease
	return lease
}

func (p *IPPool) RenewLease(mac net.HardwareAddr) (*Lease, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	macStr := mac.String()
	lease, ok := p.leases[macStr]
	if !ok {
		return nil, fmt.Errorf("no active lease for MAC %s", macStr)
	}
	now := time.Now()
	duration := p.leaseTime
	lease.StartTime = now
	lease.ExpireTime = now.Add(time.Duration(duration) * time.Second)
	lease.T1Time = now.Add(time.Duration(duration/2) * time.Second)
	lease.T2Time = now.Add(time.Duration(duration*7/8) * time.Second)
	return lease, nil
}

func (p *IPPool) Release(mac net.HardwareAddr) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	macStr := mac.String()
	lease, ok := p.leases[macStr]
	if !ok {
		return fmt.Errorf("no active lease for MAC %s", macStr)
	}
	ipn := ipToUint32(lease.IP)
	delete(p.leases, macStr)
	delete(p.ipToLease, ipn)
	p.available[ipn] = true
	return nil
}

func (p *IPPool) GetLeaseByMAC(mac net.HardwareAddr) *Lease {
	p.mu.RLock()
	defer p.mu.RUnlock()
	lease, ok := p.leases[mac.String()]
	if !ok {
		return nil
	}
	return p.copyLease(lease)
}

func (p *IPPool) GetLeaseByIP(ip net.IP) *Lease {
	p.mu.RLock()
	defer p.mu.RUnlock()
	ipn := ipToUint32(ip)
	lease, ok := p.ipToLease[ipn]
	if !ok {
		return nil
	}
	return p.copyLease(lease)
}

func (p *IPPool) GetAllLeases() []*Lease {
	p.mu.RLock()
	defer p.mu.RUnlock()
	leases := make([]*Lease, 0, len(p.leases))
	for _, lease := range p.leases {
		leases = append(leases, p.copyLease(lease))
	}
	return leases
}

func (p *IPPool) copyLease(l *Lease) *Lease {
	return &Lease{
		MAC:        append(net.HardwareAddr(nil), l.MAC...),
		IP:         append(net.IP(nil), l.IP...),
		StartTime:  l.StartTime,
		Duration:   l.Duration,
		ExpireTime: l.ExpireTime,
		T1Time:     l.T1Time,
		T2Time:     l.T2Time,
		ServerIP:   append(net.IP(nil), l.ServerIP...),
	}
}

func (p *IPPool) CleanExpired() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	expired := 0
	for macStr, lease := range p.leases {
		if lease.ExpireTime.Before(now) {
			ipn := ipToUint32(lease.IP)
			delete(p.leases, macStr)
			delete(p.ipToLease, ipn)
			p.available[ipn] = true
			expired++
		}
	}
	return expired
}

func (l *Lease) IsExpired() bool {
	return time.Now().After(l.ExpireTime)
}

func (l *Lease) IsT1Reached() bool {
	return time.Now().After(l.T1Time)
}

func (l *Lease) IsT2Reached() bool {
	return time.Now().After(l.T2Time)
}

func (l *Lease) RemainingTime() time.Duration {
	return time.Until(l.ExpireTime)
}

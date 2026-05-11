package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dhcp-parser/api"
	"github.com/dhcp-parser/dhcp"
)

var (
	pool   *dhcp.IPPool
	server *dhcp.Server
)

func main() {
	var port int
	flag.IntVar(&port, "port", 0, "HTTP server port")
	flag.Parse()

	if port == 0 {
		envPort := os.Getenv("DHCP_SERVER_PORT")
		if envPort != "" {
			p, err := strconv.Atoi(envPort)
			if err == nil {
				port = p
			}
		}
	}
	if port == 0 {
		port = 8205
	}

	var err error
	pool, err = dhcp.NewIPPool(
		net.ParseIP("192.168.1.100"),
		net.ParseIP("192.168.1.200"),
	)
	if err != nil {
		fmt.Printf("Failed to create IP pool: %v\n", err)
		os.Exit(1)
	}
	pool.SetServerIP(net.ParseIP("192.168.1.1"))
	pool.SetSubnetMask(net.IPv4Mask(255, 255, 255, 0))
	pool.SetRouters([]net.IP{net.ParseIP("192.168.1.1")})
	pool.SetDNSServers([]net.IP{net.ParseIP("8.8.8.8"), net.ParseIP("8.8.4.4")})

	server = dhcp.NewServer(pool)

	http.HandleFunc("/parse", handleParse)
	http.HandleFunc("/build", handleBuild)
	http.HandleFunc("/pool", handlePool)
	http.HandleFunc("/pool/add", handlePoolAdd)
	http.HandleFunc("/pool/remove", handlePoolRemove)
	http.HandleFunc("/leases", handleLeases)
	http.HandleFunc("/leases/by-mac", handleLeaseByMAC)
	http.HandleFunc("/leases/by-ip", handleLeaseByIP)
	http.HandleFunc("/dora", handleDORA)
	http.HandleFunc("/config", handleConfig)

	fmt.Printf("DHCP server listening on port %d...\n", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
		os.Exit(1)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ParseResponse{Success: false, Error: "Method not allowed"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{Success: false, Error: "Failed to read body"})
		return
	}
	var req api.ParseRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{Success: false, Error: "Invalid JSON"})
		return
	}
	data, err := hex.DecodeString(strings.TrimSpace(req.Hex))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{Success: false, Error: "Invalid hex string"})
		return
	}
	packet, err := dhcp.Parse(data)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ParseResponse{Success: false, Error: err.Error()})
		return
	}
	info := packetToInfo(packet)
	writeJSON(w, http.StatusOK, api.ParseResponse{Success: true, Packet: info})
}

func packetToInfo(p *dhcp.Packet) api.PacketInfo {
	info := api.PacketInfo{
		Op:        p.Op,
		OpName:    dhcp.OpToString(p.Op),
		Htype:     p.Htype,
		Hlen:      p.Hlen,
		Hops:      p.Hops,
		Xid:       p.Xid,
		Secs:      p.Secs,
		Flags:     p.Flags,
		Broadcast: p.IsBroadcast(),
		Ciaddr:    p.Ciaddr.String(),
		Yiaddr:    p.Yiaddr.String(),
		Siaddr:    p.Siaddr.String(),
		Giaddr:    p.Giaddr.String(),
		Chaddr:    p.Chaddr.String(),
		Options:   make([]api.OptionInfo, 0),
	}
	if msgType, err := p.MessageType(); err == nil {
		info.MessageType = msgType
		info.MessageName = dhcp.MessageTypeToString(msgType)
	}
	for _, opt := range p.Options.All() {
		info.Options = append(info.Options, api.OptionInfo{
			Code:   opt.Code,
			Length: opt.Length,
			Value:  hex.EncodeToString(opt.Data),
		})
	}
	return info
}

func handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.BuildResponse{Success: false, Error: "Method not allowed"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.BuildResponse{Success: false, Error: "Failed to read body"})
		return
	}
	var req api.BuildRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.BuildResponse{Success: false, Error: "Invalid JSON"})
		return
	}
	var chaddr net.HardwareAddr
	if req.Chaddr != "" {
		chaddr, err = net.ParseMAC(req.Chaddr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, api.BuildResponse{Success: false, Error: "Invalid MAC address"})
			return
		}
	}
	builder := dhcp.NewBuilder(req.Op).
		WithXid(req.Xid).
		WithHops(req.Hops).
		WithSecs(req.Secs).
		WithBroadcast(req.Broadcast).
		WithMessageType(req.MessageType)
	if req.Htype > 0 {
		builder.WithHtype(req.Htype)
	}
	if req.Hlen > 0 {
		builder.WithHlen(req.Hlen)
	}
	if chaddr != nil {
		builder.WithChaddr(chaddr)
	}
	if req.Ciaddr != "" {
		if ip := net.ParseIP(req.Ciaddr); ip != nil {
			builder.WithCiaddr(ip)
		}
	}
	if req.Yiaddr != "" {
		if ip := net.ParseIP(req.Yiaddr); ip != nil {
			builder.WithYiaddr(ip)
		}
	}
	if req.Siaddr != "" {
		if ip := net.ParseIP(req.Siaddr); ip != nil {
			builder.WithSiaddr(ip)
		}
	}
	if req.Giaddr != "" {
		if ip := net.ParseIP(req.Giaddr); ip != nil {
			builder.WithGiaddr(ip)
		}
	}
	if req.ServerIP != "" {
		if ip := net.ParseIP(req.ServerIP); ip != nil {
			builder.WithServerIdentifier(ip)
		}
	}
	if req.LeaseTime > 0 {
		builder.WithLeaseTime(req.LeaseTime)
	}
	if req.RequestedIP != "" {
		if ip := net.ParseIP(req.RequestedIP); ip != nil {
			builder.WithRequestedIP(ip)
		}
	}
	if req.SubnetMask != "" {
		if mask := parseIPMask(req.SubnetMask); mask != nil {
			builder.WithSubnetMask(mask)
		}
	}
	if len(req.Routers) > 0 {
		ips := make([]net.IP, 0, len(req.Routers))
		for _, s := range req.Routers {
			if ip := net.ParseIP(s); ip != nil {
				ips = append(ips, ip)
			}
		}
		builder.WithRouters(ips)
	}
	if len(req.DNSServers) > 0 {
		ips := make([]net.IP, 0, len(req.DNSServers))
		for _, s := range req.DNSServers {
			if ip := net.ParseIP(s); ip != nil {
				ips = append(ips, ip)
			}
		}
		builder.WithDNSServers(ips)
	}
	if req.HostName != "" {
		builder.WithHostName(req.HostName)
	}
	for code, data := range req.CustomOptions {
		builder.WithOption(code, data)
	}
	packet := builder.Build()
	bytes := packet.Bytes()
	writeJSON(w, http.StatusOK, api.BuildResponse{
		Success: true,
		Hex:     hex.EncodeToString(bytes),
		Bytes:   bytes,
	})
}

func parseIPMask(s string) net.IPMask {
	parts := strings.Split(s, ".")
	if len(parts) == 4 {
		bytes := make([]byte, 4)
		for i, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil {
				return nil
			}
			bytes[i] = byte(n)
		}
		return net.IPMask(bytes)
	}
	ones, err := strconv.Atoi(s)
	if err == nil && ones >= 0 && ones <= 32 {
		return net.CIDRMask(ones, 32)
	}
	return nil
}

func handlePool(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, api.PoolStatusResponse{Success: false, Error: "Method not allowed"})
		return
	}
	total := pool.TotalCount()
	available := pool.AvailableCount()
	writeJSON(w, http.StatusOK, api.PoolStatusResponse{
		Success:   true,
		Total:     total,
		Available: available,
		InUse:     total - available,
	})
}

func handlePoolAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.IPRangeResponse{Success: false, Error: "Method not allowed"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: "Failed to read body"})
		return
	}
	var req api.IPRangeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: "Invalid JSON"})
		return
	}
	start := net.ParseIP(req.StartIP)
	end := net.ParseIP(req.EndIP)
	if start == nil || end == nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: "Invalid IP address"})
		return
	}
	if err := pool.AddRange(start, end); err != nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.IPRangeResponse{Success: true})
}

func handlePoolRemove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.IPRangeResponse{Success: false, Error: "Method not allowed"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: "Failed to read body"})
		return
	}
	var req api.IPRangeRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: "Invalid JSON"})
		return
	}
	start := net.ParseIP(req.StartIP)
	end := net.ParseIP(req.EndIP)
	if start == nil || end == nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: "Invalid IP address"})
		return
	}
	if err := pool.RemoveRange(start, end); err != nil {
		writeJSON(w, http.StatusBadRequest, api.IPRangeResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, api.IPRangeResponse{Success: true})
}

func handleLeases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, api.LeaseListResponse{Success: false, Error: "Method not allowed"})
		return
	}
	leases := pool.GetAllLeases()
	pool.CleanExpired()
	leaseInfos := make([]api.LeaseInfo, 0, len(leases))
	for _, l := range leases {
		leaseInfos = append(leaseInfos, leaseToInfo(l))
	}
	writeJSON(w, http.StatusOK, api.LeaseListResponse{Success: true, Leases: leaseInfos})
}

func handleLeaseByMAC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, api.LeaseResponse{Success: false, Error: "Method not allowed"})
		return
	}
	macStr := r.URL.Query().Get("mac")
	if macStr == "" {
		writeJSON(w, http.StatusBadRequest, api.LeaseResponse{Success: false, Error: "Missing 'mac' parameter"})
		return
	}
	mac, err := net.ParseMAC(macStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.LeaseResponse{Success: false, Error: "Invalid MAC address"})
		return
	}
	lease := pool.GetLeaseByMAC(mac)
	if lease == nil {
		writeJSON(w, http.StatusNotFound, api.LeaseResponse{Success: false, Error: "Lease not found"})
		return
	}
	writeJSON(w, http.StatusOK, api.LeaseResponse{Success: true, Lease: leaseToInfo(lease)})
}

func handleLeaseByIP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, api.LeaseResponse{Success: false, Error: "Method not allowed"})
		return
	}
	ipStr := r.URL.Query().Get("ip")
	if ipStr == "" {
		writeJSON(w, http.StatusBadRequest, api.LeaseResponse{Success: false, Error: "Missing 'ip' parameter"})
		return
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		writeJSON(w, http.StatusBadRequest, api.LeaseResponse{Success: false, Error: "Invalid IP address"})
		return
	}
	lease := pool.GetLeaseByIP(ip)
	if lease == nil {
		writeJSON(w, http.StatusNotFound, api.LeaseResponse{Success: false, Error: "Lease not found"})
		return
	}
	writeJSON(w, http.StatusOK, api.LeaseResponse{Success: true, Lease: leaseToInfo(lease)})
}

func leaseToInfo(l *dhcp.Lease) api.LeaseInfo {
	return api.LeaseInfo{
		MAC:         l.MAC.String(),
		IP:          l.IP.String(),
		StartTime:   l.StartTime,
		Duration:    l.Duration,
		ExpireTime:  l.ExpireTime,
		T1Time:      l.T1Time,
		T2Time:      l.T2Time,
		ServerIP:    l.ServerIP.String(),
		IsExpired:   l.IsExpired(),
		IsT1Reached: l.IsT1Reached(),
		IsT2Reached: l.IsT2Reached(),
	}
}

func handleDORA(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.DORAResponse{Success: false, Error: "Method not allowed"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.DORAResponse{Success: false, Error: "Failed to read body"})
		return
	}
	var req api.DORARequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.DORAResponse{Success: false, Error: "Invalid JSON"})
		return
	}
	mac, err := net.ParseMAC(req.MAC)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.DORAResponse{Success: false, Error: "Invalid MAC address"})
		return
	}
	result, err := server.SimulateDORA(mac)
	if err != nil || !result.Success {
		resp := api.DORAResponse{Success: false, Error: result.Error}
		if err != nil {
			resp.Error = err.Error()
		}
		writeJSON(w, http.StatusBadRequest, resp)
		return
	}
	leaseInfo := leaseToInfo(result.Lease)
	writeJSON(w, http.StatusOK, api.DORAResponse{
		Success:    true,
		AssignedIP: result.AssignedIP.String(),
		Lease:      &leaseInfo,
	})
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, api.ServerConfigResponse{Success: false, Error: "Method not allowed"})
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, api.ServerConfigResponse{Success: false, Error: "Failed to read body"})
		return
	}
	var req api.ServerConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeJSON(w, http.StatusBadRequest, api.ServerConfigResponse{Success: false, Error: "Invalid JSON"})
		return
	}
	if req.ServerIP != "" {
		if ip := net.ParseIP(req.ServerIP); ip != nil {
			pool.SetServerIP(ip)
		}
	}
	if req.SubnetMask != "" {
		if mask := parseIPMask(req.SubnetMask); mask != nil {
			pool.SetSubnetMask(mask)
		}
	}
	if req.Routers != nil {
		ips := make([]net.IP, 0, len(req.Routers))
		for _, s := range req.Routers {
			if ip := net.ParseIP(s); ip != nil {
				ips = append(ips, ip)
			}
		}
		pool.SetRouters(ips)
	}
	if req.DNSServers != nil {
		ips := make([]net.IP, 0, len(req.DNSServers))
		for _, s := range req.DNSServers {
			if ip := net.ParseIP(s); ip != nil {
				ips = append(ips, ip)
			}
		}
		pool.SetDNSServers(ips)
	}
	if req.LeaseTime > 0 {
		pool.SetLeaseTime(req.LeaseTime)
	}
	_ = time.Now()
	writeJSON(w, http.StatusOK, api.ServerConfigResponse{Success: true})
}

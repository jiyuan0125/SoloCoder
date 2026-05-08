package snowflake

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

func GenerateMachineID() (int64, error) {
	hash := sha256.New()

	hostname, err := os.Hostname()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrMachineIDGenFailed, err)
	}
	hash.Write([]byte(hostname))

	interfaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range interfaces {
			if iface.Flags&net.FlagUp != 0 && len(iface.HardwareAddr) > 0 {
				hash.Write(iface.HardwareAddr)
			}
		}
	}

	ips := getLocalIPs()
	for _, ip := range ips {
		hash.Write([]byte(ip))
	}

	sum := hash.Sum(nil)
	id := binary.BigEndian.Uint64(sum[:8])
	return int64(id & uint64(MaxMachineID)), nil
}

func getLocalIPs() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip != nil && !ip.IsLoopback() {
			if ip4 := ip.To4(); ip4 != nil {
				ips = append(ips, ip4.String())
			}
		}
	}
	return ips
}

func getMachineIDFromIP() (int64, error) {
	ips := getLocalIPs()
	if len(ips) == 0 {
		return 0, fmt.Errorf("%w: no IP addresses found", ErrMachineIDGenFailed)
	}

	hash := sha256.New()
	for _, ip := range ips {
		hash.Write([]byte(ip))
	}
	sum := hash.Sum(nil)
	id := binary.BigEndian.Uint64(sum[:8])
	return int64(id & uint64(MaxMachineID)), nil
}

func getMachineIDFromHostname() (int64, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return 0, fmt.Errorf("%w: %v", ErrMachineIDGenFailed, err)
	}

	hostname = strings.ToLower(strings.TrimSpace(hostname))
	hash := sha256.New()
	hash.Write([]byte(hostname))
	sum := hash.Sum(nil)
	id := binary.BigEndian.Uint64(sum[:8])
	return int64(id & uint64(MaxMachineID)), nil
}

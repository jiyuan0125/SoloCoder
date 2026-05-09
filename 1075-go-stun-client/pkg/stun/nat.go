package stun

import (
	"errors"
	"net"
	"time"
)

type NATType string

const (
	NATTypeUnknown            NATType = "Unknown"
	NATTypeNone               NATType = "None"
	NATTypeFullCone           NATType = "Full Cone"
	NATTypeRestrictedCone     NATType = "Restricted Cone"
	NATTypePortRestrictedCone NATType = "Port Restricted Cone"
	NATTypeSymmetric          NATType = "Symmetric"
	NATTypeBlocked            NATType = "Blocked"
	NATTypeSymmetricUDPFirewall NATType = "Symmetric UDP Firewall"
)

type StunServer struct {
	Host string
	Port int
}

var DefaultServers = []StunServer{
	{Host: "stun.l.google.com", Port: 19302},
	{Host: "stun1.l.google.com", Port: 19302},
	{Host: "stun2.l.google.com", Port: 19302},
	{Host: "stun3.l.google.com", Port: 19302},
	{Host: "stun4.l.google.com", Port: 19302},
}

type NATResult struct {
	NATType        NATType
	MappingAddress *MappedAddress
	LocalIP        string
	PublicIP       string
}

type NATDetector struct {
	Client        *Client
	PrimaryServer StunServer
	AltServer     StunServer
	Credentials   *Credentials
}

func NewNATDetector(options ...DetectorOption) *NATDetector {
	d := &NATDetector{
		Client: NewClient(),
		PrimaryServer: DefaultServers[0],
		AltServer:     DefaultServers[1],
	}
	for _, opt := range options {
		opt(d)
	}
	return d
}

type DetectorOption func(*NATDetector)

func WithServers(primary, alt StunServer) DetectorOption {
	return func(d *NATDetector) {
		d.PrimaryServer = primary
		d.AltServer = alt
	}
}

func WithCredentials(creds *Credentials) DetectorOption {
	return func(d *NATDetector) {
		d.Credentials = creds
	}
}

func (d *NATDetector) Detect() (*NATResult, error) {
	result := &NATResult{
		NATType: NATTypeUnknown,
	}

	localIP, err := getLocalIP()
	if err == nil {
		result.LocalIP = localIP
	}

	server1Addr := formatServer(d.PrimaryServer)

	test1Result, err := d.Client.PerformBinding(server1Addr, d.Credentials)
	if err != nil {
		result.NATType = NATTypeBlocked
		return result, err
	}

	result.MappingAddress = test1Result.MappingAddress
	result.PublicIP = test1Result.MappingAddress.IP.String()

	if test1Result.MappingAddress.IP.Equal(net.ParseIP(localIP)) {
		test2Result, err := d.runTestII()
		if err != nil {
			result.NATType = NATTypeSymmetricUDPFirewall
			return result, nil
		}
		if test2Result.MappingAddress.IP.Equal(test1Result.MappingAddress.IP) &&
		   test2Result.MappingAddress.Port == test1Result.MappingAddress.Port {
			result.NATType = NATTypeNone
			return result, nil
		}
		result.NATType = NATTypeSymmetric
		return result, nil
	}

	testIResult, err := d.runTestI()
	if err != nil {
		result.NATType = NATTypeUnknown
		return result, err
	}

	if testIResult.MappingAddress.IP.Equal(test1Result.MappingAddress.IP) &&
	   testIResult.MappingAddress.Port == test1Result.MappingAddress.Port {
		test2Result, err := d.runTestII()
		if err != nil {
			result.NATType = NATTypePortRestrictedCone
			return result, nil
		}
		if test2Result.MappingAddress.IP.Equal(test1Result.MappingAddress.IP) &&
		   test2Result.MappingAddress.Port == test1Result.MappingAddress.Port {
			result.NATType = NATTypeFullCone
		} else {
			result.NATType = NATTypeRestrictedCone
		}
		return result, nil
	}

	test3Result, err := d.runTestIII()
	if err != nil {
		result.NATType = NATTypePortRestrictedCone
		return result, nil
	}

	if test3Result.MappingAddress.IP.Equal(testIResult.MappingAddress.IP) &&
	   test3Result.MappingAddress.Port == testIResult.MappingAddress.Port {
		result.NATType = NATTypeRestrictedCone
		return result, nil
	}

	result.NATType = NATTypeSymmetric
	return result, nil
}

func (d *NATDetector) runTestI() (*BindingResult, error) {
	serverAddr := formatServer(d.PrimaryServer)
	return d.Client.PerformBinding(serverAddr, d.Credentials)
}

func (d *NATDetector) runTestII() (*BindingResult, error) {
	serverAddr := formatServer(d.AltServer)
	return d.Client.PerformBinding(serverAddr, d.Credentials)
}

func (d *NATDetector) runTestIII() (*BindingResult, error) {
	altPort := d.PrimaryServer.Port + 1
	if altPort > 65535 {
		altPort = d.PrimaryServer.Port - 1
	}
	serverAddr := formatServer(StunServer{Host: d.PrimaryServer.Host, Port: altPort})
	return d.Client.PerformBinding(serverAddr, d.Credentials)
}

func formatServer(server StunServer) string {
	return net.JoinHostPort(server.Host, itoa(server.Port))
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[pos:])
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		interfaces, err := net.Interfaces()
		if err != nil {
			return "", err
		}
		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if ip != nil && !ip.IsLoopback() && ip.IsPrivate() {
					return ip.String(), nil
				}
			}
		}
		return "", errors.New("no local IP found")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func (d *NATDetector) DetectWithServer(server StunServer) (*NATResult, error) {
	d.PrimaryServer = server
	return d.Detect()
}

type BindingOptions struct {
	Server      StunServer
	Credentials *Credentials
	Timeout     time.Duration
	MaxRetries  int
}

func SimpleBinding(opts *BindingOptions) (*MappedAddress, error) {
	clientOpts := make([]ClientOption, 0)
	if opts.Timeout > 0 {
		clientOpts = append(clientOpts, WithTimeout(opts.Timeout))
	}
	if opts.MaxRetries > 0 {
		clientOpts = append(clientOpts, WithMaxRetries(opts.MaxRetries))
	}

	client := NewClient(clientOpts...)
	serverAddr := formatServer(opts.Server)
	result, err := client.PerformBinding(serverAddr, opts.Credentials)
	if err != nil {
		return nil, err
	}
	return result.MappingAddress, nil
}

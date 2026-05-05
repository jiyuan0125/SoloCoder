package certchecker

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"cert-checker/internal/protocol"
)

const defaultPort = "443"
const defaultTimeout = 5 * time.Second

func parseDomainAndPort(input string) (string, string) {
	parts := strings.SplitN(input, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return input, defaultPort
}

func matchDomain(domain, certDomain string) bool {
	if certDomain == domain {
		return true
	}

	if strings.HasPrefix(certDomain, "*.") {
		suffix := certDomain[2:]
		parts := strings.SplitN(domain, ".", 2)
		if len(parts) == 2 && parts[1] == suffix {
			return true
		}
	}

	return false
}

func CheckCertificate(domain string) protocol.CertInfo {
	result := protocol.CertInfo{
		Domain:    domain,
		CheckedAt: time.Now(),
	}

	host, port := parseDomainAndPort(domain)
	addr := net.JoinHostPort(host, port)

	dialer := &net.Dialer{
		Timeout: defaultTimeout,
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
	if err != nil {
		result.Error = fmt.Sprintf("连接失败: %v", err)
		return result
	}
	defer conn.Close()

	state := conn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		result.Error = "未获取到证书"
		return result
	}

	cert := state.PeerCertificates[0]
	leafCert := cert

	for _, c := range state.PeerCertificates {
		if !c.IsCA {
			leafCert = c
			break
		}
	}

	result.CommonName = leafCert.Subject.CommonName
	result.Issuer = leafCert.Issuer.CommonName
	result.ValidFrom = leafCert.NotBefore
	result.ValidTo = leafCert.NotAfter

	now := time.Now()
	if now.After(leafCert.NotAfter) {
		result.IsExpired = true
		result.RemainingDays = 0
	} else {
		duration := leafCert.NotAfter.Sub(now)
		result.RemainingDays = int(duration.Hours() / 24)
		result.IsExpired = false
	}

	result.SANs = append([]string{}, leafCert.DNSNames...)

	result.IsWildcard = false
	for _, san := range leafCert.DNSNames {
		if strings.HasPrefix(san, "*.") {
			result.IsWildcard = true
			break
		}
	}
	if strings.HasPrefix(result.CommonName, "*.") {
		result.IsWildcard = true
	}

	result.RequestDomainOK = false
	if len(result.SANs) > 0 {
		for _, san := range result.SANs {
			if matchDomain(host, san) {
				result.RequestDomainOK = true
				break
			}
		}
	} else {
		if matchDomain(host, result.CommonName) {
			result.RequestDomainOK = true
		}
	}

	return result
}

func CheckDomainFromFile(filepath string) ([]protocol.CertInfo, error) {
	return nil, fmt.Errorf("未实现")
}

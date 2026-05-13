package parser

import (
	"bytes"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

type CertificateInfo struct {
	Subject         string
	Issuer          string
	SerialNumber    string
	NotBefore       time.Time
	NotAfter        time.Time
	PublicKeyAlgorithm string
	PublicKey       interface{}
	SignatureAlgorithm string
	IsExpired       bool
	IsNotYetValid   bool
	KeyUsage        []string
	ExtKeyUsage     []string
	DNSNames        []string
	EmailAddresses  []string
	IPAddresses     []string
	Raw             []byte
}

func ParseCertificate(data []byte) (*CertificateInfo, error) {
	cert, err := x509.ParseCertificate(data)
	if err != nil {
		return nil, fmt.Errorf("解析证书失败: %v", err)
	}

	now := time.Now()
	info := &CertificateInfo{
		Subject:         formatName(cert.Subject),
		Issuer:          formatName(cert.Issuer),
		SerialNumber:    cert.SerialNumber.String(),
		NotBefore:       cert.NotBefore,
		NotAfter:        cert.NotAfter,
		PublicKeyAlgorithm: cert.PublicKeyAlgorithm.String(),
		PublicKey:       cert.PublicKey,
		SignatureAlgorithm: cert.SignatureAlgorithm.String(),
		IsExpired:       now.After(cert.NotAfter),
		IsNotYetValid:   now.Before(cert.NotBefore),
		KeyUsage:        keyUsageToString(cert.KeyUsage),
		ExtKeyUsage:     extKeyUsageToString(cert.ExtKeyUsage),
		DNSNames:        cert.DNSNames,
		EmailAddresses:  cert.EmailAddresses,
		IPAddresses:     ipAddressesToString(cert.IPAddresses),
		Raw:             data,
	}

	return info, nil
}

func formatName(name pkix.Name) string {
	var parts []string

	if len(name.CommonName) > 0 {
		parts = append(parts, fmt.Sprintf("CN=%s", name.CommonName))
	}

	for _, org := range name.Organization {
		parts = append(parts, fmt.Sprintf("O=%s", org))
	}

	for _, ou := range name.OrganizationalUnit {
		parts = append(parts, fmt.Sprintf("OU=%s", ou))
	}

	for _, locality := range name.Locality {
		parts = append(parts, fmt.Sprintf("L=%s", locality))
	}

	for _, province := range name.Province {
		parts = append(parts, fmt.Sprintf("ST=%s", province))
	}

	for _, country := range name.Country {
		parts = append(parts, fmt.Sprintf("C=%s", country))
	}

	return strings.Join(parts, ", ")
}

func keyUsageToString(usage x509.KeyUsage) []string {
	var result []string

	if usage&x509.KeyUsageDigitalSignature != 0 {
		result = append(result, "Digital Signature")
	}
	if usage&x509.KeyUsageContentCommitment != 0 {
		result = append(result, "Content Commitment")
	}
	if usage&x509.KeyUsageKeyEncipherment != 0 {
		result = append(result, "Key Encipherment")
	}
	if usage&x509.KeyUsageDataEncipherment != 0 {
		result = append(result, "Data Encipherment")
	}
	if usage&x509.KeyUsageKeyAgreement != 0 {
		result = append(result, "Key Agreement")
	}
	if usage&x509.KeyUsageCertSign != 0 {
		result = append(result, "Certificate Sign")
	}
	if usage&x509.KeyUsageCRLSign != 0 {
		result = append(result, "CRL Sign")
	}
	if usage&x509.KeyUsageEncipherOnly != 0 {
		result = append(result, "Encipher Only")
	}
	if usage&x509.KeyUsageDecipherOnly != 0 {
		result = append(result, "Decipher Only")
	}

	return result
}

func extKeyUsageToString(usages []x509.ExtKeyUsage) []string {
	usageMap := map[x509.ExtKeyUsage]string{
		x509.ExtKeyUsageAny:                            "Any",
		x509.ExtKeyUsageServerAuth:                     "Server Authentication",
		x509.ExtKeyUsageClientAuth:                     "Client Authentication",
		x509.ExtKeyUsageCodeSigning:                    "Code Signing",
		x509.ExtKeyUsageEmailProtection:                "Email Protection",
		x509.ExtKeyUsageIPSECEndSystem:                 "IPSec End System",
		x509.ExtKeyUsageIPSECTunnel:                    "IPSec Tunnel",
		x509.ExtKeyUsageIPSECUser:                      "IPSec User",
		x509.ExtKeyUsageTimeStamping:                   "Time Stamping",
		x509.ExtKeyUsageOCSPSigning:                    "OCSP Signing",
		x509.ExtKeyUsageMicrosoftServerGatedCrypto:     "Microsoft Server Gated Crypto",
		x509.ExtKeyUsageNetscapeServerGatedCrypto:      "Netscape Server Gated Crypto",
	}

	var result []string
	for _, usage := range usages {
		if name, ok := usageMap[usage]; ok {
			result = append(result, name)
		} else {
			result = append(result, fmt.Sprintf("Unknown (%d)", usage))
		}
	}

	return result
}

func ipAddressesToString(ips []net.IP) []string {
	var result []string
	for _, ip := range ips {
		result = append(result, ip.String())
	}
	return result
}

func (c *CertificateInfo) ToString() string {
	var buf bytes.Buffer

	buf.WriteString("X.509 Certificate Information\n")
	buf.WriteString("=============================\n\n")

	buf.WriteString(fmt.Sprintf("Subject: %s\n", c.Subject))
	buf.WriteString(fmt.Sprintf("Issuer:  %s\n", c.Issuer))
	buf.WriteString(fmt.Sprintf("Serial Number: %s\n", c.SerialNumber))
	buf.WriteString(fmt.Sprintf("Validity Period:\n"))
	buf.WriteString(fmt.Sprintf("  Not Before: %s\n", c.NotBefore.Format(time.RFC3339)))
	buf.WriteString(fmt.Sprintf("  Not After:  %s\n", c.NotAfter.Format(time.RFC3339)))

	if c.IsExpired {
		buf.WriteString("\n⚠️  证书已过期\n")
	}
	if c.IsNotYetValid {
		buf.WriteString("\n⚠️  证书尚未生效\n")
	}

	buf.WriteString(fmt.Sprintf("\nPublic Key Algorithm: %s\n", c.PublicKeyAlgorithm))
	buf.WriteString(fmt.Sprintf("Signature Algorithm:  %s\n", c.SignatureAlgorithm))

	if len(c.KeyUsage) > 0 {
		buf.WriteString(fmt.Sprintf("\nKey Usage: %s\n", strings.Join(c.KeyUsage, ", ")))
	}

	if len(c.ExtKeyUsage) > 0 {
		buf.WriteString(fmt.Sprintf("Extended Key Usage: %s\n", strings.Join(c.ExtKeyUsage, ", ")))
	}

	if len(c.DNSNames) > 0 {
		buf.WriteString(fmt.Sprintf("DNS Names: %s\n", strings.Join(c.DNSNames, ", ")))
	}

	if len(c.EmailAddresses) > 0 {
		buf.WriteString(fmt.Sprintf("Email Addresses: %s\n", strings.Join(c.EmailAddresses, ", ")))
	}

	if len(c.IPAddresses) > 0 {
		buf.WriteString(fmt.Sprintf("IP Addresses: %s\n", strings.Join(c.IPAddresses, ", ")))
	}

	buf.WriteString(fmt.Sprintf("\nRaw Data (hex):\n%s\n", hex.EncodeToString(c.Raw)))

	return buf.String()
}

func (c *CertificateInfo) ToJSON() (string, error) {
	jsonData := map[string]interface{}{
		"subject":           c.Subject,
		"issuer":            c.Issuer,
		"serial_number":     c.SerialNumber,
		"not_before":        c.NotBefore.Format(time.RFC3339),
		"not_after":         c.NotAfter.Format(time.RFC3339),
		"public_key_algorithm": c.PublicKeyAlgorithm,
		"signature_algorithm":  c.SignatureAlgorithm,
		"is_expired":        c.IsExpired,
		"is_not_yet_valid":  c.IsNotYetValid,
		"key_usage":         c.KeyUsage,
		"extended_key_usage": c.ExtKeyUsage,
		"dns_names":         c.DNSNames,
		"email_addresses":   c.EmailAddresses,
		"ip_addresses":      c.IPAddresses,
		"raw_data_hex":      hex.EncodeToString(c.Raw),
	}

	jsonBytes, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

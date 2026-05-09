package certparse

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"x509cert/pkg/common"
)

func ParseCertificateFromPEM(pemStr string) (*common.CertificateInfo, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing certificate")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %v", err)
	}

	return parseCertificate(cert), nil
}

func parseCertificate(cert *x509.Certificate) *common.CertificateInfo {
	subject := common.SubjectInfo{
		CommonName:         cert.Subject.CommonName,
		Organization:       cert.Subject.Organization,
		OrganizationalUnit: cert.Subject.OrganizationalUnit,
		Country:            cert.Subject.Country,
		Locality:           cert.Subject.Locality,
		Province:           cert.Subject.Province,
		StreetAddress:      cert.Subject.StreetAddress,
		PostalCode:         cert.Subject.PostalCode,
	}

	issuer := common.IssuerInfo{
		CommonName:         cert.Issuer.CommonName,
		Organization:       cert.Issuer.Organization,
		OrganizationalUnit: cert.Issuer.OrganizationalUnit,
		Country:            cert.Issuer.Country,
		Locality:           cert.Issuer.Locality,
		Province:           cert.Issuer.Province,
		StreetAddress:      cert.Issuer.StreetAddress,
		PostalCode:         cert.Issuer.PostalCode,
	}

	serialNumber := serialNumberToString(cert.SerialNumber)

	sans := common.SANInfo{
		DNSNames:    cert.DNSNames,
		IPAddresses: make([]string, 0, len(cert.IPAddresses)),
	}

	for _, ip := range cert.IPAddresses {
		sans.IPAddresses = append(sans.IPAddresses, ip.String())
	}

	return &common.CertificateInfo{
		Subject:      subject,
		Issuer:       issuer,
		SerialNumber: serialNumber,
		NotBefore:    cert.NotBefore,
		NotAfter:     cert.NotAfter,
		SANs:         sans,
	}
}

func serialNumberToString(serial *big.Int) string {
	return serial.String()
}

func IsExpired(cert *x509.Certificate) bool {
	return time.Now().After(cert.NotAfter)
}

func IsNotYetValid(cert *x509.Certificate) bool {
	return time.Now().Before(cert.NotBefore)
}

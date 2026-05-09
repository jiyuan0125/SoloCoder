package certparse

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"x509cert/pkg/common"
)

func ParseCertificateChainFromPEM(pemStr string) ([]*x509.Certificate, error) {
	var certificates []*x509.Certificate
	pemData := []byte(pemStr)

	for len(pemData) > 0 {
		block, rest := pem.Decode(pemData)
		if block == nil {
			if len(certificates) == 0 {
				return nil, fmt.Errorf("failed to decode PEM data")
			}
			break
		}

		if block.Type == "CERTIFICATE" {
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate: %v", err)
			}
			certificates = append(certificates, cert)
		}

		pemData = rest
	}

	if len(certificates) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM chain")
	}

	return certificates, nil
}

func VerifyCertificateChain(pemChain string) (*common.VerifyResponse, error) {
	certs, err := ParseCertificateChainFromPEM(pemChain)
	if err != nil {
		return nil, err
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found")
	}

	response := &common.VerifyResponse{
		Success:     true,
		Valid:       true,
		Expired:     false,
		NotYetValid: false,
		Certificates: make([]*common.CertificateVerifyInfo, 0, len(certs)),
	}

	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	leafCert := certs[0]
	isSelfSigned := IsSelfSigned(leafCert)

	if isSelfSigned && len(certs) == 1 {
		roots.AddCert(leafCert)
	} else {
		for i := 1; i < len(certs); i++ {
			if i == len(certs)-1 && IsSelfSigned(certs[i]) {
				roots.AddCert(certs[i])
			} else {
				intermediates.AddCert(certs[i])
			}
		}
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	_, verifyErr := leafCert.Verify(opts)

	for i, cert := range certs {
		certInfo := parseCertificate(cert)
		isExpired := IsExpired(cert)
		isNotYetValid := IsNotYetValid(cert)

		verifyInfo := &common.CertificateVerifyInfo{
			Index:          i,
			Info:           certInfo,
			IsExpired:      isExpired,
			IsNotYetValid:  isNotYetValid,
			SignatureValid: false,
		}

		if i == 0 {
			verifyInfo.SignatureValid = verifyErr == nil
		} else if i == len(certs)-1 && isSelfSigned {
			verifyInfo.SignatureValid = verifyErr == nil
		} else {
			prevCert := certs[i-1]
			sigErr := cert.CheckSignatureFrom(cert)
			if sigErr == nil {
				verifyInfo.SignatureValid = true
			} else {
				for j := i; j < len(certs); j++ {
					if err := prevCert.CheckSignatureFrom(certs[j]); err == nil {
						verifyInfo.SignatureValid = true
						break
					}
				}
			}
		}

		if isExpired {
			response.Expired = true
			response.Valid = false
		}
		if isNotYetValid {
			response.NotYetValid = true
			response.Valid = false
		}

		response.Certificates = append(response.Certificates, verifyInfo)
	}

	if verifyErr != nil {
		response.Valid = false
	}

	return response, nil
}

func IsSelfSigned(cert *x509.Certificate) bool {
	return cert.CheckSignatureFrom(cert) == nil
}

package samlparser

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"time"

	"saml-parser/pkg/model"
)

type ValidationOptions struct {
	Audience  string
	Recipient string
	PublicKey *rsa.PublicKey
}

func Validate(xmlContent string, options ValidationOptions) (*model.ValidateResponse, error) {
	parseResp, err := Parse(xmlContent)
	if err != nil {
		return &model.ValidateResponse{
			Success: false,
			Errors:  []string{err.Error()},
			Parse:   *parseResp,
		}, err
	}

	errorsList := []string{}
	valid := true

	if parseErr := validateConditions(parseResp, options.Audience); parseErr != nil {
		errorsList = append(errorsList, parseErr.Error())
		valid = false
	}

	if scErr := validateSubjectConfirmation(parseResp, options.Recipient); scErr != nil {
		errorsList = append(errorsList, scErr.Error())
		valid = false
	}

	if sigErr := validateSignatureBasic(parseResp); sigErr != nil {
		errorsList = append(errorsList, sigErr.Error())
		valid = false
	}

	return &model.ValidateResponse{
		Success:   true,
		Valid:     valid,
		Validated: true,
		Errors:    errorsList,
		Parse:     *parseResp,
	}, nil
}

func validateConditions(parseResp *model.ParseResponse, expectedAudience string) error {
	if parseResp.Conditions.NotBefore != "" {
		notBefore, err := parseSAMLTime(parseResp.Conditions.NotBefore)
		if err == nil && !notBefore.IsZero() {
			adjustedBefore := notBefore.Add(-time.Duration(ClockSkewMinutes) * time.Minute)
			if time.Now().Before(adjustedBefore) {
				return errors.New("assertion is not yet valid (NotBefore)")
			}
		}
	}

	if parseResp.Conditions.NotOnOrAfter != "" {
		notOnOrAfter, err := parseSAMLTime(parseResp.Conditions.NotOnOrAfter)
		if err == nil && !notOnOrAfter.IsZero() {
			adjustedAfter := notOnOrAfter.Add(time.Duration(ClockSkewMinutes) * time.Minute)
			if time.Now().After(adjustedAfter) {
				return errors.New("assertion has expired (NotOnOrAfter)")
			}
		}
	}

	if expectedAudience != "" {
		audiences := parseResp.Conditions.AudienceRestriction.Audiences
		if len(audiences) == 0 {
			return errors.New("no AudienceRestriction found in assertion")
		}

		found := false
		for _, audience := range audiences {
			if audience == expectedAudience {
				found = true
				break
			}
		}

		if !found {
			return errors.New("audience mismatch: assertion not intended for this service provider")
		}
	}

	return nil
}

func validateSubjectConfirmation(parseResp *model.ParseResponse, expectedRecipient string) error {
	sc := parseResp.Subject.SubjectConfirmation

	if sc.Method == "" {
		return errors.New("SubjectConfirmation Method is required but missing")
	}

	if sc.Method != BearerMethod {
		return errors.New("SubjectConfirmation Method must be bearer")
	}

	if sc.Data.Recipient == "" {
		return errors.New("SubjectConfirmationData Recipient is required but missing")
	}

	if expectedRecipient != "" && sc.Data.Recipient != expectedRecipient {
		return errors.New("SubjectConfirmationData Recipient mismatch")
	}

	if sc.Data.NotOnOrAfter != "" {
		notOnOrAfter, err := parseSAMLTime(sc.Data.NotOnOrAfter)
		if err == nil && !notOnOrAfter.IsZero() {
			adjustedAfter := notOnOrAfter.Add(time.Duration(ClockSkewMinutes) * time.Minute)
			if time.Now().After(adjustedAfter) {
				return errors.New("SubjectConfirmation has expired (NotOnOrAfter)")
			}
		}
	}

	return nil
}

func validateSignatureBasic(parseResp *model.ParseResponse) error {
	if !parseResp.Signature.HasSignature {
		return errors.New("assertion is not signed")
	}

	if parseResp.Signature.DigestValue == "" {
		return errors.New("signature missing DigestValue")
	}

	if parseResp.Signature.SignatureValue == "" {
		return errors.New("signature missing SignatureValue")
	}

	return nil
}

func ParseRSAPublicKeyFromPEM(pemData string) (*rsa.PublicKey, error) {
	block, _ := base64.StdEncoding.DecodeString(pemData)
	if block == nil {
		return nil, errors.New("failed to decode base64 public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block)
	if err != nil {
		return nil, errors.New("failed to parse public key: " + err.Error())
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("public key is not RSA")
	}

	return rsaPub, nil
}

func ParseX509Certificate(certData string) (*x509.Certificate, error) {
	derBytes, err := base64.StdEncoding.DecodeString(certData)
	if err != nil {
		return nil, errors.New("failed to decode base64 certificate: " + err.Error())
	}

	cert, err := x509.ParseCertificate(derBytes)
	if err != nil {
		return nil, errors.New("failed to parse X.509 certificate: " + err.Error())
	}

	return cert, nil
}

func PublicKeyFromCertificate(cert *x509.Certificate) (*rsa.PublicKey, error) {
	pubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("certificate does not contain an RSA public key")
	}
	return pubKey, nil
}

func computeHash(data []byte, algorithm string) ([]byte, error) {
	var h hash.Hash
	
	switch algorithm {
	case "http://www.w3.org/2001/04/xmlenc#sha256", "http://www.w3.org/2001/04/xmldsig-more#sha256":
		h = sha256.New()
	case "http://www.w3.org/2000/09/xmldsig#sha1":
		h = sha1.New()
	default:
		return nil, fmt.Errorf("unsupported digest algorithm: %s", algorithm)
	}
	
	h.Write(data)
	return h.Sum(nil), nil
}

func _unused() crypto.Hash {
	return crypto.SHA256
}

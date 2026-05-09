package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"x509cert/pkg/common"
)

const (
	defaultServerURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "parse":
		handleParse()
	case "verify":
		handleVerify()
	case "help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("X.509 Certificate Parser and Verifier Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client parse <certificate-file.pem>")
	fmt.Println("  client verify <certificate-chain.pem>")
	fmt.Println()
	fmt.Println("Environment Variables:")
	fmt.Println("  SERVER_URL  Server URL (default: http://localhost:8080)")
	fmt.Println()
}

func handleParse() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Certificate file path required")
		fmt.Println("Usage: client parse <certificate-file.pem>")
		os.Exit(1)
	}

	filePath := os.Args[2]
	pemData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	serverURL := getServerURL()
	request := common.ParseRequest{
		PEM: string(pemData),
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/parse", "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var response common.ParseResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		os.Exit(1)
	}

	if !response.Success {
		fmt.Printf("Error: %s\n", response.Error)
		os.Exit(1)
	}

	printCertificateInfo(response.Certificate)
}

func handleVerify() {
	if len(os.Args) < 3 {
		fmt.Println("Error: Certificate chain file path required")
		fmt.Println("Usage: client verify <certificate-chain.pem>")
		os.Exit(1)
	}

	filePath := os.Args[2]
	pemData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	serverURL := getServerURL()
	request := common.VerifyRequest{
		PEMChain: string(pemData),
	}

	requestBytes, err := json.Marshal(request)
	if err != nil {
		fmt.Printf("Error marshaling request: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(serverURL+"/api/verify", "application/json", bytes.NewBuffer(requestBytes))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		os.Exit(1)
	}

	var response common.VerifyResponse
	if err := json.Unmarshal(body, &response); err != nil {
		fmt.Printf("Error unmarshaling response: %v\n", err)
		os.Exit(1)
	}

	if !response.Success {
		fmt.Printf("Error: %s\n", response.Error)
		os.Exit(1)
	}

	printVerifyResult(&response)

	if !response.Valid {
		os.Exit(2)
	}
}

func getServerURL() string {
	serverURL := os.Getenv("SERVER_URL")
	if serverURL == "" {
		serverURL = defaultServerURL
	}
	return serverURL
}

func printCertificateInfo(cert *common.CertificateInfo) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	fmt.Println("Certificate Information")
	fmt.Println("========================")
	fmt.Fprintln(w, "\nSubject:")
	if cert.Subject.CommonName != "" {
		fmt.Fprintf(w, "  Common Name (CN):\t%s\n", cert.Subject.CommonName)
	}
	if len(cert.Subject.Organization) > 0 {
		fmt.Fprintf(w, "  Organization (O):\t%v\n", cert.Subject.Organization)
	}
	if len(cert.Subject.OrganizationalUnit) > 0 {
		fmt.Fprintf(w, "  Organizational Unit (OU):\t%v\n", cert.Subject.OrganizationalUnit)
	}
	if len(cert.Subject.Country) > 0 {
		fmt.Fprintf(w, "  Country (C):\t%v\n", cert.Subject.Country)
	}
	if len(cert.Subject.Locality) > 0 {
		fmt.Fprintf(w, "  Locality (L):\t%v\n", cert.Subject.Locality)
	}
	if len(cert.Subject.Province) > 0 {
		fmt.Fprintf(w, "  Province (ST):\t%v\n", cert.Subject.Province)
	}

	fmt.Fprintln(w, "\nIssuer:")
	if cert.Issuer.CommonName != "" {
		fmt.Fprintf(w, "  Common Name (CN):\t%s\n", cert.Issuer.CommonName)
	}
	if len(cert.Issuer.Organization) > 0 {
		fmt.Fprintf(w, "  Organization (O):\t%v\n", cert.Issuer.Organization)
	}
	if len(cert.Issuer.OrganizationalUnit) > 0 {
		fmt.Fprintf(w, "  Organizational Unit (OU):\t%v\n", cert.Issuer.OrganizationalUnit)
	}
	if len(cert.Issuer.Country) > 0 {
		fmt.Fprintf(w, "  Country (C):\t%v\n", cert.Issuer.Country)
	}

	fmt.Fprintln(w, "\nValidity:")
	fmt.Fprintf(w, "  Not Before:\t%s\n", formatTime(cert.NotBefore))
	fmt.Fprintf(w, "  Not After:\t%s\n", formatTime(cert.NotAfter))

	fmt.Fprintf(w, "\nSerial Number:\t%s\n", cert.SerialNumber)

	if len(cert.SANs.DNSNames) > 0 || len(cert.SANs.IPAddresses) > 0 {
		fmt.Fprintln(w, "\nSubject Alternative Names (SANs):")
		if len(cert.SANs.DNSNames) > 0 {
			fmt.Fprintln(w, "  DNS Names:")
			for _, dns := range cert.SANs.DNSNames {
				fmt.Fprintf(w, "    - %s\n", dns)
			}
		}
		if len(cert.SANs.IPAddresses) > 0 {
			fmt.Fprintln(w, "  IP Addresses:")
			for _, ip := range cert.SANs.IPAddresses {
				fmt.Fprintf(w, "    - %s\n", ip)
			}
		}
	}
}

func printVerifyResult(response *common.VerifyResponse) {
	fmt.Println("Certificate Chain Verification Result")
	fmt.Println("=======================================")
	fmt.Println()

	if response.Valid {
		fmt.Println("✓ Certificate chain is valid")
	} else {
		fmt.Println("✗ Certificate chain is NOT valid")
	}

	if response.Expired {
		fmt.Println("⚠ One or more certificates have expired")
	}

	if response.NotYetValid {
		fmt.Println("⚠ One or more certificates are not yet valid")
	}

	fmt.Println()
	fmt.Printf("Number of certificates in chain: %d\n\n", len(response.Certificates))

	for i, certVerify := range response.Certificates {
		fmt.Printf("Certificate %d:\n", i+1)
		fmt.Println("---------------")

		fmt.Printf("  Subject CN: %s\n", certVerify.Info.Subject.CommonName)
		fmt.Printf("  Issuer CN: %s\n", certVerify.Info.Issuer.CommonName)
		fmt.Printf("  Valid from: %s\n", formatTime(certVerify.Info.NotBefore))
		fmt.Printf("  Valid to: %s\n", formatTime(certVerify.Info.NotAfter))
		fmt.Printf("  Serial Number: %s\n", certVerify.Info.SerialNumber)

		fmt.Println()
		fmt.Printf("  Status:\n")
		if certVerify.IsExpired {
			fmt.Println("    ✗ EXPIRED")
		} else if certVerify.IsNotYetValid {
			fmt.Println("    ✗ NOT YET VALID")
		} else {
			fmt.Println("    ✓ Valid period")
		}

		if certVerify.SignatureValid {
			fmt.Println("    ✓ Signature valid")
		} else {
			fmt.Println("    ✗ Signature INVALID")
		}

		if len(certVerify.Info.SANs.DNSNames) > 0 || len(certVerify.Info.SANs.IPAddresses) > 0 {
			fmt.Printf("\n  SANs:\n")
			if len(certVerify.Info.SANs.DNSNames) > 0 {
				fmt.Printf("    DNS: %v\n", certVerify.Info.SANs.DNSNames)
			}
			if len(certVerify.Info.SANs.IPAddresses) > 0 {
				fmt.Printf("    IPs: %v\n", certVerify.Info.SANs.IPAddresses)
			}
		}

		fmt.Println()
	}
}

func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

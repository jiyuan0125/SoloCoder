package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"ecdh-demo/internal/common"
	"ecdh-demo/pkg/ecdh"
)

const serverURL = "http://localhost:8080"

func toECDH(c string) ecdh.Curve {
	switch c {
	case common.CurveP256:
		return ecdh.CurveP256
	case common.CurveP384:
		return ecdh.CurveP384
	case common.CurveP521:
		return ecdh.CurveP521
	default:
		return ""
	}
}

func postJSON(url string, body interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil
}

func doExchange(curve ecdh.Curve) error {
	if err := ecdh.ValidateCurve(curve); err != nil {
		return fmt.Errorf("unsupported curve")
	}

	startReq := common.StartNegotiationRequest{
		Curve: curveName(curve),
	}

	resp, err := postJSON(serverURL+"/negotiate/start", startReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("start negotiation failed: %s", errResp.Error)
	}

	var startResp common.StartNegotiationResponse
	if err := json.NewDecoder(resp.Body).Decode(&startResp); err != nil {
		return fmt.Errorf("failed to decode start response: %w", err)
	}

	clientKP, err := ecdh.GenerateKeyPair(curve)
	if err != nil {
		return fmt.Errorf("failed to generate client key pair: %w", err)
	}

	serverPubBytes, err := base64.StdEncoding.DecodeString(startResp.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to decode server public key: %w", err)
	}

	completeReq := common.CompleteNegotiationRequest{
		SessionID: startResp.SessionID,
		Curve:     curveName(curve),
		PublicKey: base64.StdEncoding.EncodeToString(clientKP.PublicKey),
	}

	resp2, err := postJSON(serverURL+"/negotiate/complete", completeReq)
	if err != nil {
		return err
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusOK {
		var errResp common.ErrorResponse
		json.NewDecoder(resp2.Body).Decode(&errResp)
		return fmt.Errorf("complete negotiation failed: %s", errResp.Error)
	}

	var completeResp common.CompleteNegotiationResponse
	if err := json.NewDecoder(resp2.Body).Decode(&completeResp); err != nil {
		return fmt.Errorf("failed to decode complete response: %w", err)
	}

	clientSharedSecret, err := ecdh.ComputeSharedSecretFromBytes(clientKP, curve, serverPubBytes)
	if err != nil {
		return fmt.Errorf("failed to compute client shared secret: %w", err)
	}

	fmt.Printf("Session ID: %s\n", completeResp.SessionID)
	fmt.Printf("Server Shared Secret: %s\n", completeResp.SharedSecret)
	fmt.Printf("Client Shared Secret: %s\n", clientSharedSecret)

	if completeResp.SharedSecret == clientSharedSecret {
		fmt.Println("Success: Keys match!")
	} else {
		fmt.Println("Error: Keys do not match!")
	}

	return nil
}

func curveName(c ecdh.Curve) string {
	switch c {
	case ecdh.CurveP256:
		return common.CurveP256
	case ecdh.CurveP384:
		return common.CurveP384
	case ecdh.CurveP521:
		return common.CurveP521
	default:
		return ""
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: client <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  exchange  Perform ECDH key exchange")
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "exchange":
		fs := flag.NewFlagSet("exchange", flag.ExitOnError)
		curveStr := fs.String("curve", common.CurveP256, "Curve to use (P-256, P-384, P-521)")
		fs.Parse(args)

		curve := toECDH(*curveStr)
		if err := doExchange(curve); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

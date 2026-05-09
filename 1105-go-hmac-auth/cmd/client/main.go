package main

import (
	"bytes"
	"encoding/json"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"hmac-auth/pkg/hmacauth"
	"hmac-auth/pkg/types"
)

var (
	serverURL = flag.String("server", "http://localhost:8080", "server URL")
	method    = flag.String("method", "POST", "HTTP method for signing")
	path      = flag.String("path", "/api/test", "request path for signing")
)

func readStdin() ([]byte, error) {
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, fmt.Errorf("no input provided. Please pipe message content through stdin")
	}
	return io.ReadAll(os.Stdin)
}

func localSign(keyHex string, body []byte) (string, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return "", fmt.Errorf("invalid key: %v", err)
	}
	ts := time.Now().Unix()
	signature := hmacauth.Sign(key, *method, *path, ts, body)
	fmt.Printf("Timestamp: %d\n", ts)
	return signature, nil
}

func remoteSign(body []byte) error {
	ts := time.Now().Unix()
	reqBody := types.SignRequest{
		Method:    *method,
		Path:      *path,
		Timestamp: ts,
		Body:      string(body),
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(*serverURL+"/sign", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result types.SignResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("invalid response: %v", err)
	}

	fmt.Printf("Timestamp: %d\n", ts)
	fmt.Printf("Signature: %s\n", result.Signature)
	fmt.Printf("Key Version: %d\n", result.Version)
	return nil
}

func localVerify(keyHex, signature string, body []byte) (bool, error) {
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		return false, fmt.Errorf("invalid key: %v", err)
	}

	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	timestamp := flagSet.Int64("timestamp", 0, "timestamp for verification")
	flagSet.Parse(os.Args[2:])

	if *timestamp == 0 {
		return false, fmt.Errorf("timestamp is required for verification (use -timestamp)")
	}

	valid := hmacauth.Verify(key, *method, *path, *timestamp, body, signature)
	return valid, nil
}

func remoteVerify(signature string, body []byte) error {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	timestamp := flagSet.Int64("timestamp", 0, "timestamp for verification")
	version := flagSet.Int("version", 0, "optional key version")
	flagSet.Parse(os.Args[2:])

	if *timestamp == 0 {
		return fmt.Errorf("timestamp is required for verification (use -timestamp)")
	}

	reqBody := types.VerifyRequest{
		Method:    *method,
		Path:      *path,
		Timestamp: *timestamp,
		Body:      string(body),
		Signature: signature,
	}

	if *version != 0 {
		v := *version
		reqBody.Version = &v
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(*serverURL+"/verify", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result types.VerifyResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("invalid response: %v", err)
	}

	fmt.Printf("Valid: %v\n", result.Valid)
	fmt.Printf("Message: %s\n", result.Message)
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s sign [options]        Sign message via server\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s local-sign [options]  Sign message locally with key\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s verify [options]      Verify signature via server\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s local-verify [options] Verify signature locally with key\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nGlobal Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nLocal Sign/Verify Options:\n")
	fmt.Fprintf(os.Stderr, "  -key string     HMAC key (hex encoded)\n")
	fmt.Fprintf(os.Stderr, "  -signature string  Signature to verify\n")
	fmt.Fprintf(os.Stderr, "  -timestamp int  Timestamp for verification\n")
	fmt.Fprintf(os.Stderr, "  -version int    Optional key version for remote verify\n")
	fmt.Fprintf(os.Stderr, "\nExample:\n")
	fmt.Fprintf(os.Stderr, "  echo 'hello world' | %s sign\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  echo 'hello world' | %s verify -timestamp 1234567890 -signature abc123\n", os.Args[0])
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	cmd := strings.ToLower(flag.Arg(0))

	switch cmd {
	case "sign":
		body, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if err := remoteSign(body); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "local-sign":
		flagSet := flag.NewFlagSet("local-sign", flag.ContinueOnError)
		keyFlag := flagSet.String("key", "", "HMAC key (hex encoded)")
		flagSet.Parse(os.Args[2:])

		if *keyFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -key is required for local-sign\n")
			os.Exit(1)
		}

		body, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		sig, err := localSign(*keyFlag, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Signature: %s\n", sig)

	case "verify":
		flagSet := flag.NewFlagSet("verify", flag.ContinueOnError)
		sigFlag := flagSet.String("signature", "", "signature to verify")
		flagSet.Parse(os.Args[2:])

		if *sigFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -signature is required for verify\n")
			os.Exit(1)
		}

		body, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := remoteVerify(*sigFlag, body); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "local-verify":
		flagSet := flag.NewFlagSet("local-verify", flag.ContinueOnError)
		keyFlag := flagSet.String("key", "", "HMAC key (hex encoded)")
		sigFlag := flagSet.String("signature", "", "signature to verify")
		flagSet.Parse(os.Args[2:])

		if *keyFlag == "" || *sigFlag == "" {
			fmt.Fprintf(os.Stderr, "Error: -key and -signature are required for local-verify\n")
			os.Exit(1)
		}

		body, err := readStdin()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		valid, err := localVerify(*keyFlag, *sigFlag, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Valid: %v\n", valid)
		if valid {
			os.Exit(0)
		}
		os.Exit(1)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/example/ed25519-sign/common"
	"github.com/example/ed25519-sign/cryptolib"
)

const defaultServer = "http://localhost:8080"

type clientMode int

const (
	modeLocal clientMode = iota
	modeRemote
)

var mode clientMode = modeLocal
var serverURL string

func printUsage() {
	fmt.Println(`Ed25519 Digital Signature Tool

Usage:
  ed25519-tool <command> [options]

Commands:
  generate-key              Generate Ed25519 key pair
  sign                      Sign a message
  sign-file                 Sign a file
  verify                    Verify a signature
  verify-file               Verify a file signature
  convert-public-key        Convert Ed25519 public key to X25519
  convert-private-key       Convert Ed25519 private key to X25519
  inspect-key               Inspect a key (public or private)

Global Options:
  --server <url>            Use remote server instead of local mode (e.g., http://localhost:8080)
  --help                    Show this help message

Examples:
  # Local mode
  ed25519-tool generate-key
  ed25519-tool sign --private-key <base64-key> --message "hello"
  ed25519-tool sign-file --private-key <base64-key> --file /path/to/file
  ed25519-tool verify --public-key <base64-key> --message "hello" --signature <base64-sig>
  ed25519-tool verify-file --public-key <base64-key> --file /path/to/file --signature <base64-sig>
  ed25519-tool convert-public-key --public-key <base64-key>
  ed25519-tool convert-private-key --private-key <base64-key>
  ed25519-tool inspect-key --key <base64-key>

  # Remote mode (requires running server)
  ed25519-tool --server http://localhost:8080 generate-key
`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	var help bool
	flag.StringVar(&serverURL, "server", "", "Remote server URL")
	flag.BoolVar(&help, "help", false, "Show help")
	flag.CommandLine.Parse(os.Args[1:])
	if help {
		printUsage()
		return
	}
	if serverURL != "" {
		mode = modeRemote
		if !strings.HasPrefix(serverURL, "http") {
			serverURL = "http://" + serverURL
		}
	}
	args := flag.Args()
	if len(args) < 1 {
		printUsage()
		os.Exit(1)
	}
	cmd := args[0]
	var err error
	switch cmd {
	case "generate-key":
		err = cmdGenerateKey()
	case "sign":
		err = cmdSign()
	case "sign-file":
		err = cmdSignFile()
	case "verify":
		err = cmdVerify()
	case "verify-file":
		err = cmdVerifyFile()
	case "convert-public-key":
		err = cmdConvertPublicKey()
	case "convert-private-key":
		err = cmdConvertPrivateKey()
	case "inspect-key":
		err = cmdInspectKey()
	default:
		printUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func cmdGenerateKey() error {
	if mode == modeLocal {
		pub, priv, err := cryptolib.GenerateKey()
		if err != nil {
			return err
		}
		fmt.Printf("Public Key:  %s\n", pub)
		fmt.Printf("Private Key: %s\n", priv)
		return nil
	}
	var resp common.GenerateKeyResponse
	if err := httpGet("/generate-key", &resp); err != nil {
		return err
	}
	fmt.Printf("Public Key:  %s\n", resp.PublicKey)
	fmt.Printf("Private Key: %s\n", resp.PrivateKey)
	return nil
}

func cmdSign() error {
	fs := flag.NewFlagSet("sign", flag.ExitOnError)
	privateKey := fs.String("private-key", "", "Base64-encoded Ed25519 private key")
	message := fs.String("message", "", "Message to sign")
	fs.Parse(flag.Args()[1:])
	if *privateKey == "" || *message == "" {
		return fmt.Errorf("usage: sign --private-key <key> --message <msg>")
	}
	if mode == modeLocal {
		sig, err := cryptolib.Sign(*privateKey, *message)
		if err != nil {
			return err
		}
		fmt.Printf("Signature: %s\n", sig)
		return nil
	}
	req := common.SignRequest{PrivateKey: *privateKey, Message: *message}
	var resp common.SignResponse
	if err := httpPost("/sign", req, &resp); err != nil {
		return err
	}
	fmt.Printf("Signature: %s\n", resp.Signature)
	return nil
}

func cmdSignFile() error {
	fs := flag.NewFlagSet("sign-file", flag.ExitOnError)
	privateKey := fs.String("private-key", "", "Base64-encoded Ed25519 private key")
	filePath := fs.String("file", "", "Path to file to sign")
	fs.Parse(flag.Args()[1:])
	if *privateKey == "" || *filePath == "" {
		return fmt.Errorf("usage: sign-file --private-key <key> --file <path>")
	}
	sig, err := cryptolib.SignFile(*privateKey, *filePath)
	if err != nil {
		return err
	}
	fmt.Printf("Signature: %s\n", sig)
	return nil
}

func cmdVerify() error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	publicKey := fs.String("public-key", "", "Base64-encoded Ed25519 public key")
	message := fs.String("message", "", "Message to verify")
	signature := fs.String("signature", "", "Base64-encoded signature")
	fs.Parse(flag.Args()[1:])
	if *publicKey == "" || *message == "" || *signature == "" {
		return fmt.Errorf("usage: verify --public-key <key> --message <msg> --signature <sig>")
	}
	if mode == modeLocal {
		err := cryptolib.Verify(*publicKey, *message, *signature)
		if err != nil {
			fmt.Printf("Valid: false\n")
			return nil
		}
		fmt.Printf("Valid: true\n")
		return nil
	}
	req := common.VerifyRequest{
		PublicKey: *publicKey,
		Message:   *message,
		Signature: *signature,
	}
	var resp common.VerifyResponse
	if err := httpPost("/verify", req, &resp); err != nil {
		return err
	}
	fmt.Printf("Valid: %v\n", resp.Valid)
	return nil
}

func cmdVerifyFile() error {
	fs := flag.NewFlagSet("verify-file", flag.ExitOnError)
	publicKey := fs.String("public-key", "", "Base64-encoded Ed25519 public key")
	filePath := fs.String("file", "", "Path to file to verify")
	signature := fs.String("signature", "", "Base64-encoded signature")
	fs.Parse(flag.Args()[1:])
	if *publicKey == "" || *filePath == "" || *signature == "" {
		return fmt.Errorf("usage: verify-file --public-key <key> --file <path> --signature <sig>")
	}
	err := cryptolib.VerifyFile(*publicKey, *filePath, *signature)
	if err != nil {
		fmt.Printf("Valid: false\n")
		return nil
	}
	fmt.Printf("Valid: true\n")
	return nil
}

func cmdConvertPublicKey() error {
	fs := flag.NewFlagSet("convert-public-key", flag.ExitOnError)
	publicKey := fs.String("public-key", "", "Base64-encoded Ed25519 public key")
	fs.Parse(flag.Args()[1:])
	if *publicKey == "" {
		return fmt.Errorf("usage: convert-public-key --public-key <key>")
	}
	if mode == modeLocal {
		xPub, err := cryptolib.PublicKeyToX25519(*publicKey)
		if err != nil {
			return err
		}
		fmt.Printf("X25519 Public Key: %s\n", xPub)
		return nil
	}
	req := common.ConvertPublicKeyRequest{PublicKey: *publicKey}
	var resp common.ConvertPublicKeyResponse
	if err := httpPost("/convert-public-key", req, &resp); err != nil {
		return err
	}
	fmt.Printf("X25519 Public Key: %s\n", resp.X25519PublicKey)
	return nil
}

func cmdConvertPrivateKey() error {
	fs := flag.NewFlagSet("convert-private-key", flag.ExitOnError)
	privateKey := fs.String("private-key", "", "Base64-encoded Ed25519 private key")
	fs.Parse(flag.Args()[1:])
	if *privateKey == "" {
		return fmt.Errorf("usage: convert-private-key --private-key <key>")
	}
	if mode == modeLocal {
		xPriv, err := cryptolib.PrivateKeyToX25519(*privateKey)
		if err != nil {
			return err
		}
		fmt.Printf("X25519 Private Key: %s\n", xPriv)
		return nil
	}
	req := common.ConvertPrivateKeyRequest{PrivateKey: *privateKey}
	var resp common.ConvertPrivateKeyResponse
	if err := httpPost("/convert-private-key", req, &resp); err != nil {
		return err
	}
	fmt.Printf("X25519 Private Key: %s\n", resp.X25519PrivateKey)
	return nil
}

func cmdInspectKey() error {
	fs := flag.NewFlagSet("inspect-key", flag.ExitOnError)
	key := fs.String("key", "", "Base64-encoded key (public or private)")
	fs.Parse(flag.Args()[1:])
	if *key == "" {
		return fmt.Errorf("usage: inspect-key --key <key>")
	}
	if mode == modeLocal {
		info, err := cryptolib.InspectKey(*key)
		if err != nil {
			return err
		}
		fmt.Printf("Type:        %s\n", info.Type)
		fmt.Printf("EncodedLen:  %d\n", info.EncodedLen)
		fmt.Printf("DecodedLen:  %d\n", info.DecodedLen)
		fmt.Printf("Fingerprint: %s\n", info.Fingerprint)
		return nil
	}
	req := common.InspectKeyRequest{Key: *key}
	var resp common.InspectKeyResponse
	if err := httpPost("/inspect-key", req, &resp); err != nil {
		return err
	}
	fmt.Printf("Type:        %s\n", resp.Type)
	fmt.Printf("EncodedLen:  %d\n", resp.EncodedLen)
	fmt.Printf("DecodedLen:  %d\n", resp.DecodedLen)
	fmt.Printf("Fingerprint: %s\n", resp.Fingerprint)
	return nil
}

func httpGet(path string, out interface{}) error {
	url := serverURL + path
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(body))
	}
	return json.Unmarshal(body, out)
}

func httpPost(path string, req interface{}, out interface{}) error {
	url := serverURL + path
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var errResp common.ErrorResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return fmt.Errorf("%s", errResp.Error)
		}
		return fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}
	return json.Unmarshal(respBody, out)
}

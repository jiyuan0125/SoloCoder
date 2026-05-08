package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"totp-generator/internal/common"
	"totp-generator/internal/totp"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "register":
		handleRegister(os.Args[2:])
	case "generate":
		handleGenerate(os.Args[2:])
	case "validate":
		handleValidate(os.Args[2:])
	case "local-generate":
		handleLocalGenerate(os.Args[2:])
	case "local-validate":
		handleLocalValidate(os.Args[2:])
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  client register --account <account> [--issuer <issuer>] [--server <url>]")
	fmt.Println("  client generate --secret <secret> [--server <url>]")
	fmt.Println("  client validate --secret <secret> --totp <code> [--window <n>] [--server <url>]")
	fmt.Println("  client local-generate --secret <secret>")
	fmt.Println("  client local-validate --secret <secret> --totp <code> [--window <n>]")
	fmt.Println()
	fmt.Println("Environment variables:")
	fmt.Println("  TOTP_SECRET  - default secret")
	fmt.Println("  TOTP_SERVER  - default server URL")
}

func getEnvSecret() string {
	return os.Getenv("TOTP_SECRET")
}

func getEnvServer() string {
	if s := os.Getenv("TOTP_SERVER"); s != "" {
		return s
	}
	return defaultServerURL
}

func handleRegister(args []string) {
	fs := flag.NewFlagSet("register", flag.ExitOnError)
	account := fs.String("account", "", "account name (required)")
	issuer := fs.String("issuer", "", "issuer name (optional)")
	server := fs.String("server", getEnvServer(), "server URL")
	fs.Parse(args)

	if *account == "" {
		fmt.Println("error: --account is required")
		os.Exit(1)
	}

	req := common.RegisterRequest{
		Account: *account,
		Issuer:  *issuer,
	}

	var resp common.RegisterResponse
	if err := postJSON(*server+"/register", req, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Secret: %s\n", resp.Secret)
	fmt.Printf("QR Code URL: %s\n", resp.QRCodeURL)
}

func handleGenerate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	secret := fs.String("secret", getEnvSecret(), "base32 secret")
	server := fs.String("server", getEnvServer(), "server URL")
	fs.Parse(args)

	if *secret == "" {
		fmt.Println("error: --secret or TOTP_SECRET is required")
		os.Exit(1)
	}

	req := common.GenerateRequest{Secret: *secret}

	var resp common.GenerateResponse
	if err := postJSON(*server+"/generate", req, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(resp.TOTP)
}

func handleValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	secret := fs.String("secret", getEnvSecret(), "base32 secret")
	totpCode := fs.String("totp", "", "6-digit code (required)")
	window := fs.Int("window", 0, "window size")
	server := fs.String("server", getEnvServer(), "server URL")
	fs.Parse(args)

	if *secret == "" {
		fmt.Println("error: --secret or TOTP_SECRET is required")
		os.Exit(1)
	}
	if *totpCode == "" {
		fmt.Println("error: --totp is required")
		os.Exit(1)
	}

	req := common.ValidateRequest{
		Secret: *secret,
		TOTP:   *totpCode,
		Window: *window,
	}

	var resp common.ValidateResponse
	if err := postJSON(*server+"/validate", req, &resp); err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	if resp.Valid {
		fmt.Println("valid")
		os.Exit(0)
	} else {
		fmt.Println("invalid")
		os.Exit(2)
	}
}

func handleLocalGenerate(args []string) {
	fs := flag.NewFlagSet("local-generate", flag.ExitOnError)
	secret := fs.String("secret", getEnvSecret(), "base32 secret")
	fs.Parse(args)

	if *secret == "" {
		fmt.Println("error: --secret or TOTP_SECRET is required")
		os.Exit(1)
	}

	code, err := totp.Generate(*secret)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(code)
}

func handleLocalValidate(args []string) {
	fs := flag.NewFlagSet("local-validate", flag.ExitOnError)
	secret := fs.String("secret", getEnvSecret(), "base32 secret")
	totpCode := fs.String("totp", "", "6-digit code (required)")
	window := fs.Int("window", totp.DefaultWindow, "window size")
	fs.Parse(args)

	if *secret == "" {
		fmt.Println("error: --secret or TOTP_SECRET is required")
		os.Exit(1)
	}
	if *totpCode == "" {
		fmt.Println("error: --totp is required")
		os.Exit(1)
	}

	valid := totp.ValidateWithWindow(*secret, *totpCode, *window)
	if valid {
		fmt.Println("valid")
		os.Exit(0)
	} else {
		fmt.Println("invalid")
		os.Exit(2)
	}
}

func postJSON(url string, req, resp interface{}) error {
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 400 {
		return fmt.Errorf("server returned %s", res.Status)
	}

	return json.NewDecoder(res.Body).Decode(resp)
}

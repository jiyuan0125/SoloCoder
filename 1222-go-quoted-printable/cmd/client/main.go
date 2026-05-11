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
	"quoted-printable/pkg/api"
)

const defaultServer = "http://localhost:8304"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	server := defaultServer
	if envServer := os.Getenv("QP_SERVER"); envServer != "" {
		server = envServer
	}

	flagSet := flag.NewFlagSet("qp", flag.ExitOnError)
	serverFlag := flagSet.String("server", server, "server URL")

	subCmd := os.Args[1]
	if subCmd == "help" {
		printUsage()
		return
	}

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}

	args := flagSet.Args()

	var input string
	if len(args) > 0 {
		input = strings.Join(args, " ")
	} else {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "read stdin error: %v\n", err)
				os.Exit(1)
			}
			input = string(data)
		}
	}

	if input == "" {
		fmt.Fprintf(os.Stderr, "no input provided\n")
		printUsage()
		os.Exit(1)
	}

	var result string
	var err error

	switch subCmd {
	case "encode":
		result, err = callEncode(*serverFlag, input)
	case "decode":
		result, err = callDecode(*serverFlag, input)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", subCmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Print(result)
}

func callEncode(server, text string) (string, error) {
	reqBody := api.EncodeRequest{Text: text}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(server+"/encode", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiResp api.EncodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if apiResp.Error != "" {
		return "", fmt.Errorf("server error: %s", apiResp.Error)
	}

	return apiResp.Encoded, nil
}

func callDecode(server, encoded string) (string, error) {
	reqBody := api.DecodeRequest{Encoded: encoded}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(server+"/decode", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apiResp api.DecodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return "", err
	}

	if apiResp.Error != "" {
		return "", fmt.Errorf("server error: %s", apiResp.Error)
	}

	return apiResp.Decoded, nil
}

func printUsage() {
	fmt.Println("Usage: qp <command> [options] [input]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  encode    Encode text to quoted-printable")
	fmt.Println("  decode    Decode quoted-printable to text")
	fmt.Println("  help      Show this help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -server URL   Server URL (default: http://localhost:8304)")
	fmt.Println("                Can also be set via QP_SERVER environment variable")
	fmt.Println()
	fmt.Println("Input can be provided as arguments or via stdin.")
}

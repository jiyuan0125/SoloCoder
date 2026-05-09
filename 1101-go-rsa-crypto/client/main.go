package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"rsa-crypto/common"
)

const (
	serverURL = "http://localhost:8080"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "genkey":
		cmdGenKey(os.Args[2:])
	case "encrypt":
		cmdEncrypt(os.Args[2:])
	case "decrypt":
		cmdDecrypt(os.Args[2:])
	case "-h", "--help", "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("RSA Crypto Client")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client <command> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  genkey  [--size 2048|4096] [--async]  Generate RSA key pair")
	fmt.Println("  encrypt [--padding pkcs1v15|oaep] [--hash sha1|sha256|sha512] [--label <text>]")
	fmt.Println("          Read plaintext from stdin, send to server for encryption")
	fmt.Println("  decrypt [--padding pkcs1v15|oaep] [--hash sha1|sha256|sha512] [--label <text>]")
	fmt.Println("          Read ciphertext from stdin, send to server for decryption")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  client genkey --size 2048")
	fmt.Println("  echo 'Hello World' | client encrypt --padding oaep")
	fmt.Println("  echo 'base64_ciphertext' | client decrypt --padding oaep")
}

func cmdGenKey(args []string) {
	keySize := 2048
	async := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--size":
			if i+1 < len(args) {
				switch args[i+1] {
				case "2048":
					keySize = 2048
				case "4096":
					keySize = 4096
				default:
					fmt.Fprintf(os.Stderr, "Invalid key size: %s\n", args[i+1])
					os.Exit(1)
				}
				i++
			}
		case "--async":
			async = true
		}
	}

	if async {
		generateAsync(keySize)
	} else {
		generateSync(keySize)
	}
}

func generateSync(keySize int) {
	fmt.Fprintf(os.Stderr, "Generating %d-bit RSA key pair...\n", keySize)

	reqBody, _ := json.Marshal(common.KeyPairRequest{KeySize: keySize})
	resp, err := http.Post(serverURL+"/api/keys/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		printErrorResponse(resp)
		os.Exit(1)
	}

	var keyResp common.KeyPairResponse
	json.NewDecoder(resp.Body).Decode(&keyResp)

	fmt.Println("Public Key:")
	fmt.Println(keyResp.PublicKey)
	fmt.Println("Private Key:")
	fmt.Println(keyResp.PrivateKey)
}

func generateAsync(keySize int) {
	fmt.Fprintf(os.Stderr, "Starting async %d-bit RSA key pair generation...\n", keySize)

	reqBody, _ := json.Marshal(common.AsyncKeyGenRequest{KeySize: keySize})
	resp, err := http.Post(serverURL+"/api/keys/async-generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		printErrorResponse(resp)
		os.Exit(1)
	}

	var asyncResp common.AsyncKeyGenResponse
	json.NewDecoder(resp.Body).Decode(&asyncResp)

	fmt.Fprintf(os.Stderr, "Task ID: %s. Polling for completion...\n", asyncResp.TaskID)

	for {
		time.Sleep(500 * time.Millisecond)
		resp, err := http.Get(serverURL + "/api/keys/task/" + asyncResp.TaskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to check task status: %v\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			printErrorResponse(resp)
			resp.Body.Close()
			os.Exit(1)
		}

		var task common.KeyGenTask
		json.NewDecoder(resp.Body).Decode(&task)
		resp.Body.Close()

		switch task.Status {
		case common.TaskStatusCompleted:
			fmt.Fprintf(os.Stderr, "Task completed!\n\n")
			fmt.Println("Public Key:")
			fmt.Println(task.PublicKey)
			fmt.Println("Private Key:")
			fmt.Println(task.PrivateKey)
			return
		case common.TaskStatusFailed:
			fmt.Fprintf(os.Stderr, "Task failed: %s\n", task.Error)
			os.Exit(1)
		case common.TaskStatusRunning:
			fmt.Fprintf(os.Stderr, "Still generating...\n")
		}
	}
}

func cmdEncrypt(args []string) {
	padding := common.PaddingOAEP
	hashAlgo := common.HashSHA256
	label := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--padding":
			if i+1 < len(args) {
				switch args[i+1] {
				case "pkcs1v15":
					padding = common.PaddingPKCS1v15
				case "oaep":
					padding = common.PaddingOAEP
				default:
					fmt.Fprintf(os.Stderr, "Invalid padding: %s\n", args[i+1])
					os.Exit(1)
				}
				i++
			}
		case "--hash":
			if i+1 < len(args) {
				hashAlgo = common.HashAlgorithm(args[i+1])
				i++
			}
		case "--label":
			if i+1 < len(args) {
				label = args[i+1]
				i++
			}
		}
	}

	plaintext, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	plaintext = strings.TrimSuffix(plaintext, "\n")

	req := common.EncryptRequest{
		Plaintext: plaintext,
		Padding:   padding,
		HashAlgo:  hashAlgo,
		Label:     label,
	}

	reqBody, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/encrypt", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		printErrorResponse(resp)
		os.Exit(1)
	}

	var encResp common.EncryptResponse
	json.NewDecoder(resp.Body).Decode(&encResp)

	fmt.Println(encResp.Ciphertext)
}

func cmdDecrypt(args []string) {
	padding := common.PaddingOAEP
	hashAlgo := common.HashSHA256
	label := ""

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--padding":
			if i+1 < len(args) {
				switch args[i+1] {
				case "pkcs1v15":
					padding = common.PaddingPKCS1v15
				case "oaep":
					padding = common.PaddingOAEP
				default:
					fmt.Fprintf(os.Stderr, "Invalid padding: %s\n", args[i+1])
					os.Exit(1)
				}
				i++
			}
		case "--hash":
			if i+1 < len(args) {
				hashAlgo = common.HashAlgorithm(args[i+1])
				i++
			}
		case "--label":
			if i+1 < len(args) {
				label = args[i+1]
				i++
			}
		}
	}

	ciphertext, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read stdin: %v\n", err)
		os.Exit(1)
	}

	ciphertext = strings.TrimSpace(ciphertext)

	req := common.DecryptRequest{
		Ciphertext: ciphertext,
		Padding:    padding,
		HashAlgo:   hashAlgo,
		Label:      label,
	}

	reqBody, _ := json.Marshal(req)
	resp, err := http.Post(serverURL+"/api/decrypt", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to server: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		printErrorResponse(resp)
		os.Exit(1)
	}

	var decResp common.DecryptResponse
	json.NewDecoder(resp.Body).Decode(&decResp)

	fmt.Println(decResp.Plaintext)
}

func readStdin() (string, error) {
	data, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func printErrorResponse(resp *http.Response) {
	body, _ := io.ReadAll(resp.Body)
	var errResp common.ErrorResponse
	if json.Unmarshal(body, &errResp) == nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", errResp.Error)
	} else {
		fmt.Fprintf(os.Stderr, "HTTP %d: %s\n", resp.StatusCode, string(body))
	}
}

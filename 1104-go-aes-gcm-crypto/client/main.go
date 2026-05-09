package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"aes-gcm-crypto/common"
)

const serverURL = "http://localhost:8080"

func readStdin() (string, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func doEncrypt(data, password string) error {
	reqBody := common.EncryptRequest{
		Plaintext: data,
		Password:  password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/encrypt", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBody common.EncryptResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	if respBody.Error != "" {
		return fmt.Errorf("server error: %s", respBody.Error)
	}

	fmt.Print(respBody.Ciphertext)
	return nil
}

func doDecrypt(data, password string) error {
	reqBody := common.DecryptRequest{
		Ciphertext: data,
		Password:   password,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	resp, err := http.Post(serverURL+"/decrypt", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var respBody common.DecryptResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return err
	}

	if respBody.Error != "" {
		return fmt.Errorf("server error: %s", respBody.Error)
	}

	fmt.Print(respBody.Plaintext)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: client <encrypt|decrypt> --password <password>")
		os.Exit(1)
	}

	command := os.Args[1]

	flagSet := flag.NewFlagSet(command, flag.ExitOnError)
	password := flagSet.String("password", "", "Password for encryption/decryption")

	if err := flagSet.Parse(os.Args[2:]); err != nil {
		os.Exit(1)
	}

	if *password == "" {
		fmt.Fprintln(os.Stderr, "Error: --password is required")
		os.Exit(1)
	}

	data, err := readStdin()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
		os.Exit(1)
	}

	switch command {
	case "encrypt":
		if err := doEncrypt(data, *password); err != nil {
			fmt.Fprintf(os.Stderr, "Encryption error: %v\n", err)
			os.Exit(1)
		}
	case "decrypt":
		if err := doDecrypt(data, *password); err != nil {
			fmt.Fprintf(os.Stderr, "Decryption error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		fmt.Fprintln(os.Stderr, "Usage: client <encrypt|decrypt> --password <password>")
		os.Exit(1)
	}
}

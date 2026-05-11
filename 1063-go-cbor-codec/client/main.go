package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/example/cbor-codec/api"
)

const defaultServer = "http://localhost:8110"

func main() {
	var (
		serverAddr string
		encodeCmd  string
		decodeCmd  string
		jsonToCBOR string
		cborToJSON string
	)

	flag.StringVar(&serverAddr, "server", defaultServer, "CBOR server address")
	flag.StringVar(&encodeCmd, "encode", "", "JSON input to encode to CBOR")
	flag.StringVar(&decodeCmd, "decode", "", "Hex CBOR input to decode")
	flag.StringVar(&jsonToCBOR, "json-to-cbor", "", "Convert JSON to CBOR")
	flag.StringVar(&cborToJSON, "cbor-to-json", "", "Convert CBOR to JSON")
	flag.Parse()

	if encodeCmd != "" {
		encode(serverAddr, encodeCmd)
		return
	}
	if decodeCmd != "" {
		decode(serverAddr, decodeCmd)
		return
	}
	if jsonToCBOR != "" {
		if jsonToCBOR == "-" {
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			jsonToCBORFunc(serverAddr, string(data))
		} else {
			jsonToCBORFunc(serverAddr, jsonToCBOR)
		}
		return
	}
	if cborToJSON != "" {
		if cborToJSON == "-" {
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading stdin: %v\n", err)
				os.Exit(1)
			}
			cborToJSONFunc(serverAddr, string(data))
		} else {
			cborToJSONFunc(serverAddr, cborToJSON)
		}
		return
	}

	fmt.Println("CBOR Client - Command line tool for CBOR encoding/decoding")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  client -server <url> -encode '<json>'")
	fmt.Println("  client -server <url> -decode '<hex-cbor>'")
	fmt.Println("  client -server <url> -json-to-cbor '<json>'")
	fmt.Println("  client -server <url> -cbor-to-json '<hex-cbor>'")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println(`  client -encode '{"name":"test","value":42}'`)
	fmt.Println(`  client -decode "a1646e616d656474657374"`)
	fmt.Println(`  echo '{"hello":"world"}' | client -json-to-cbor -`)
	flag.Usage()
}

func encode(server, jsonInput string) {
	req := api.EncodeRequest{InputJSON: jsonInput}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(server+"/encode", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result api.EncodeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nBody: %s\n", err, respBody)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.ErrorMessage)
		os.Exit(1)
	}

	fmt.Println(result.OutputCBOR)
}

func decode(server, cborHex string) {
	req := api.DecodeRequest{InputCBOR: cborHex}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(server+"/decode", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result api.DecodeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nBody: %s\n", err, respBody)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.ErrorMessage)
		os.Exit(1)
	}

	fmt.Println(result.OutputJSON)
}

func jsonToCBORFunc(server, jsonInput string) {
	req := api.ConvertJSONToCBORRequest{InputJSON: jsonInput}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(server+"/json-to-cbor", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result api.ConvertJSONToCBORResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nBody: %s\n", err, respBody)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.ErrorMessage)
		os.Exit(1)
	}

	fmt.Println(result.OutputCBOR)
}

func cborToJSONFunc(server, cborHex string) {
	req := api.ConvertCBORToJSONRequest{InputCBOR: cborHex}
	body, err := json.Marshal(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	resp, err := http.Post(server+"/cbor-to-json", "application/json", bytes.NewBuffer(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	var result api.ConvertCBORToJSONResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\nBody: %s\n", err, respBody)
		os.Exit(1)
	}

	if !result.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", result.ErrorMessage)
		os.Exit(1)
	}

	fmt.Println(result.OutputJSON)
}

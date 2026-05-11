package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/matrix-ops/matrix-service/api"
	"github.com/matrix-ops/matrix-service/matrix"
)

const defaultServerURL = "http://localhost:8080"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	serverURL := defaultServerURL
	if envURL := os.Getenv("MATRIX_SERVER_URL"); envURL != "" {
		serverURL = envURL
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	compress := false
	for i, arg := range args {
		if arg == "--compress" {
			compress = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	var err error
	switch strings.ToLower(cmd) {
	case "multiply":
		err = handleMultiply(serverURL, args, compress)
	case "transpose":
		err = handleTranspose(serverURL, args, compress)
	case "determinant":
		err = handleDeterminant(serverURL, args, compress)
	case "help", "--help", "-h":
		printUsage()
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Matrix Operations Client")
	fmt.Println("Usage: matmul <command> [options] <arguments>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  multiply <matrixA> <matrixB>  Multiply two matrices")
	fmt.Println("  transpose <matrix>            Transpose a matrix")
	fmt.Println("  determinant <matrix>          Calculate determinant of a square matrix")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --compress                    Request sparse format response")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  matmul multiply \"[[1,2],[3,4]]\" \"[[5,6],[7,8]]\"")
	fmt.Println("  matmul transpose \"[[1,2,3],[4,5,6]]\"")
	fmt.Println("  matmul determinant \"[[1,2],[3,4]]\"")
	fmt.Println()
	fmt.Println("Environment:")
	fmt.Println("  MATRIX_SERVER_URL  Set to override default server URL (default: http://localhost:8080)")
}

func parseMatrix(arg string) ([][]float64, error) {
	var m [][]float64
	err := json.Unmarshal([]byte(arg), &m)
	if err != nil {
		return nil, fmt.Errorf("invalid matrix format: %v", err)
	}
	return m, nil
}

func handleMultiply(serverURL string, args []string, compress bool) error {
	if len(args) != 2 {
		return fmt.Errorf("multiply requires two matrix arguments")
	}

	matrixA, err := parseMatrix(args[0])
	if err != nil {
		return err
	}
	matrixB, err := parseMatrix(args[1])
	if err != nil {
		return err
	}

	req := api.MultiplyRequest{
		MatrixA:  matrixA,
		MatrixB:  matrixB,
		Compress: &compress,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	respBody, err := sendRequest(serverURL+"/multiply", body)
	if err != nil {
		return err
	}

	var resp api.MatrixResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return printMatrixResult(resp.Result)
}

func handleTranspose(serverURL string, args []string, compress bool) error {
	if len(args) != 1 {
		return fmt.Errorf("transpose requires one matrix argument")
	}

	matrixData, err := parseMatrix(args[0])
	if err != nil {
		return err
	}

	req := api.UnaryRequest{
		Matrix:   matrixData,
		Compress: &compress,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	respBody, err := sendRequest(serverURL+"/transpose", body)
	if err != nil {
		return err
	}

	var resp api.MatrixResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return printMatrixResult(resp.Result)
}

func handleDeterminant(serverURL string, args []string, compress bool) error {
	if len(args) != 1 {
		return fmt.Errorf("determinant requires one matrix argument")
	}

	matrixData, err := parseMatrix(args[0])
	if err != nil {
		return err
	}

	req := api.UnaryRequest{
		Matrix:   matrixData,
		Compress: &compress,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	respBody, err := sendRequest(serverURL+"/determinant", body)
	if err != nil {
		return err
	}

	var resp api.DeterminantResponse
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return err
	}

	if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	fmt.Println(formatNumber(resp.Result))
	return nil
}

func sendRequest(url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connection error: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp api.MatrixResponse
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf(errResp.Error)
		}
		return nil, fmt.Errorf("server returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func printMatrixResult(rawResult json.RawMessage) error {
	var dense [][]float64
	if err := json.Unmarshal(rawResult, &dense); err == nil {
		fmt.Println(api.FormatMatrixCompact(dense))
		return nil
	}

	var sparse matrix.SparseMatrix
	if err := json.Unmarshal(rawResult, &sparse); err == nil {
		m, err := sparse.ToDense()
		if err != nil {
			return err
		}
		fmt.Println(api.FormatMatrixCompact(m.To2D()))
		return nil
	}

	fmt.Println(string(rawResult))
	return nil
}

func formatNumber(n float64) string {
	if n == float64(int64(n)) {
		return strconv.FormatInt(int64(n), 10)
	}
	return strconv.FormatFloat(n, 'f', -1, 64)
}

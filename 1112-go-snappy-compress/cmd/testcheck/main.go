package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"

	"snappy-compress/internal/api"
	"snappy-compress/internal/snappy"
)

func main() {
	passed := 0
	failed := 0

	check := func(name string, ok bool, detail string) {
		if ok {
			passed++
			fmt.Printf("  PASS: %s\n", name)
		} else {
			failed++
			fmt.Printf("  FAIL: %s — %s\n", name, detail)
		}
	}

	fmt.Println("=== Round-trip tests ===")

	testCases := [][]byte{
		{},
		{0x41},
		bytes.Repeat([]byte("A"), 100),
		[]byte("abcabcabcabc"),
		[]byte("The quick brown fox jumps over the lazy dog"),
		{0x00, 0x01, 0x02, 0xff, 0xfe, 0x80, 0x7f},
	}

	for i, tc := range testCases {
		encoded := snappy.Encode(tc)
		decoded, err := snappy.Decode(encoded)
		check(fmt.Sprintf("Round-trip case %d (len=%d)", i, len(tc)),
			err == nil && bytes.Equal(tc, decoded),
			fmt.Sprintf("err=%v, equal=%v", err, bytes.Equal(tc, decoded)))
	}

	fmt.Println("=== Stream format tests ===")

	encoded := snappy.Encode([]byte("test"))
	check("Stream starts with 0xff",
		encoded[0] == 0xff,
		fmt.Sprintf("got 0x%02x", encoded[0]))

	idLen := int(encoded[1]) | int(encoded[2])<<8 | int(encoded[3])<<16
	idStr := string(encoded[4 : 4+idLen])
	check("Stream identifier is sNaPpY",
		idStr == "sNaPpY",
		fmt.Sprintf("got %q", idStr))

	fmt.Println("=== Boundary/corruption tests ===")

	corrupt := append([]byte{}, snappy.Encode([]byte("hello"))...)
	if len(corrupt) > 10 {
		corrupt = corrupt[:len(corrupt)-5]
	}
	_, err := snappy.Decode(corrupt)
	check("Truncated data returns error",
		err != nil,
		fmt.Sprintf("expected error, got nil"))

	_, err = snappy.Decode([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	check("Invalid stream identifier returns error",
		err != nil,
		fmt.Sprintf("expected error, got nil"))

	fmt.Println("=== HTTP API tests ===")

	compressReq := api.CompressRequest{Data: base64.StdEncoding.EncodeToString([]byte("hello world"))}
	body, _ := json.Marshal(compressReq)
	resp, err := http.Post("http://localhost:8080/compress", "application/json", bytes.NewReader(body))
	if err != nil {
		check("HTTP compress", false, err.Error())
	} else {
		defer resp.Body.Close()
		var compressResp api.CompressResponse
		json.NewDecoder(resp.Body).Decode(&compressResp)
		compressedData, _ := base64.StdEncoding.DecodeString(compressResp.Data)
		check("HTTP compress valid stream",
			len(compressedData) > 0 && compressedData[0] == 0xff,
			fmt.Sprintf("len=%d, first=0x%02x", len(compressedData), compressedData[0]))

		decompressReq := api.DecompressRequest{Data: compressResp.Data}
		body2, _ := json.Marshal(decompressReq)
		resp2, _ := http.Post("http://localhost:8080/decompress", "application/json", bytes.NewReader(body2))
		defer resp2.Body.Close()
		var decompressResp api.DecompressResponse
		json.NewDecoder(resp2.Body).Decode(&decompressResp)
		decompressedData, _ := base64.StdEncoding.DecodeString(decompressResp.Data)
		check("HTTP round-trip hello world",
			bytes.Equal(decompressedData, []byte("hello world")),
			fmt.Sprintf("got %q", decompressedData))
	}

	singleReq := api.CompressRequest{Data: base64.StdEncoding.EncodeToString([]byte{0x41})}
	body3, _ := json.Marshal(singleReq)
	resp3, _ := http.Post("http://localhost:8080/compress", "application/json", bytes.NewReader(body3))
	defer resp3.Body.Close()
	var cr api.CompressResponse
	json.NewDecoder(resp3.Body).Decode(&cr)
	decReq := api.DecompressRequest{Data: cr.Data}
	body4, _ := json.Marshal(decReq)
	resp4, _ := http.Post("http://localhost:8080/decompress", "application/json", bytes.NewReader(body4))
	defer resp4.Body.Close()
	var dr api.DecompressResponse
	json.NewDecoder(resp4.Body).Decode(&dr)
	dd, _ := base64.StdEncoding.DecodeString(dr.Data)
	check("HTTP single byte round-trip",
		bytes.Equal(dd, []byte{0x41}),
		fmt.Sprintf("got %v", dd))

	repeated := bytes.Repeat([]byte("A"), 100)
	repReq := api.CompressRequest{Data: base64.StdEncoding.EncodeToString(repeated)}
	body5, _ := json.Marshal(repReq)
	resp5, _ := http.Post("http://localhost:8080/compress", "application/json", bytes.NewReader(body5))
	defer resp5.Body.Close()
	var cr2 api.CompressResponse
	json.NewDecoder(resp5.Body).Decode(&cr2)
	compressedBin, _ := base64.StdEncoding.DecodeString(cr2.Data)
	decReq2 := api.DecompressRequest{Data: cr2.Data}
	body6, _ := json.Marshal(decReq2)
	resp6, _ := http.Post("http://localhost:8080/decompress", "application/json", bytes.NewReader(body6))
	defer resp6.Body.Close()
	var dr2 api.DecompressResponse
	json.NewDecoder(resp6.Body).Decode(&dr2)
	dd2, _ := base64.StdEncoding.DecodeString(dr2.Data)
	check("HTTP repeated bytes round-trip (100 As)",
		bytes.Equal(dd2, repeated),
		fmt.Sprintf("got len=%d", len(dd2)))
	check("Repeated bytes compresses well",
		len(compressedBin) < len(repeated),
		fmt.Sprintf("compressed=%d, original=%d", len(compressedBin), len(repeated)))

	fmt.Println("=== Client CLI tests ===")

	var stdout bytes.Buffer
	cmd := exec.Command("/tmp/snappy-client")
	cmd.Stdin = bytes.NewReader([]byte("hello"))
	cmd.Stdout = &stdout
	cmd.Stderr = &bytes.Buffer{}
	err = cmd.Run()
	check("Client compress runs",
		err == nil,
		fmt.Sprintf("err=%v", err))

	if err == nil && len(stdout.Bytes()) > 0 {
		var stdout2 bytes.Buffer
		cmd3 := exec.Command("/tmp/snappy-client", "--decompress")
		cmd3.Stdin = &stdout
		cmd3.Stdout = &stdout2
		cmd3.Stderr = &bytes.Buffer{}
		err = cmd3.Run()
		check("Client decompress runs",
			err == nil,
			fmt.Sprintf("err=%v", err))
		check("Client round-trip hello",
			err == nil && bytes.Equal(stdout2.Bytes(), []byte("hello")),
			fmt.Sprintf("got %q, err=%v", stdout2.Bytes(), err))
	}

	tmpFile := "/tmp/snappy_test_input.txt"
	os.WriteFile(tmpFile, []byte("file content test"), 0644)
	var stdout3 bytes.Buffer
	cmd4 := exec.Command("/tmp/snappy-client", "--file", tmpFile)
	cmd4.Stdout = &stdout3
	cmd4.Stderr = &bytes.Buffer{}
	err = cmd4.Run()
	check("Client --file flag runs",
		err == nil,
		fmt.Sprintf("err=%v", err))

	if err == nil && len(stdout3.Bytes()) > 0 {
		var stdout4 bytes.Buffer
		cmd5 := exec.Command("/tmp/snappy-client", "--decompress")
		cmd5.Stdin = &stdout3
		cmd5.Stdout = &stdout4
		cmd5.Stderr = &bytes.Buffer{}
		err = cmd5.Run()
		check("Client --file round-trip",
			err == nil && bytes.Equal(stdout4.Bytes(), []byte("file content test")),
			fmt.Sprintf("got %q, err=%v", stdout4.Bytes(), err))
	}

	fmt.Println("=== Compression quality ===")

	repeatedEncoded := snappy.Encode(bytes.Repeat([]byte("A"), 100))
	check("100 As compressed < 50 bytes",
		len(repeatedEncoded) < 50,
		fmt.Sprintf("size=%d", len(repeatedEncoded)))

	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Passed: %d, Failed: %d\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

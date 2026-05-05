package protocol

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

func Encode(w io.Writer, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	length := len(data)
	lengthBytes := []byte(fmt.Sprintf("%08d", length))
	if _, err := w.Write(lengthBytes); err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func Decode(r io.Reader, v interface{}) error {
	lengthBuf := make([]byte, 8)
	if _, err := io.ReadFull(r, lengthBuf); err != nil {
		return err
	}
	var length int
	fmt.Sscanf(string(lengthBuf), "%08d", &length)
	dataBuf := make([]byte, length)
	if _, err := io.ReadFull(r, dataBuf); err != nil {
		return err
	}
	return json.Unmarshal(dataBuf, v)
}

func SendRequest(conn net.Conn, req *Request) (*Response, error) {
	writer := bufio.NewWriter(conn)
	if err := Encode(writer, req); err != nil {
		return nil, err
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}
	
	var resp Response
	if err := Decode(bufio.NewReader(conn), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

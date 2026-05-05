package common

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
)

type notFoundMarker struct{}

var NotFound = &notFoundMarker{}

func IsNotFound(value interface{}) bool {
	_, ok := value.(*notFoundMarker)
	return ok
}

type Request struct {
	Filename     string   `json:"filename"`
	Fields       []string `json:"fields"`
	OutputJSON   bool     `json:"output_json"`
	Compact      bool     `json:"compact"`
	HasStdinData bool     `json:"has_stdin_data"`
}

type Response struct {
	Results []string `json:"results"`
	Error   string   `json:"error,omitempty"`
}

func SendRequest(conn net.Conn, req *Request) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	if _, err := conn.Write(lengthBytes); err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func ReceiveRequest(conn net.Conn) (*Request, error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err
	}

	length := uint32(lengthBytes[0])<<24 |
		uint32(lengthBytes[1])<<16 |
		uint32(lengthBytes[2])<<8 |
		uint32(lengthBytes[3])

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, err
	}

	return &req, nil
}

func SendResponse(conn net.Conn, resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	length := uint32(len(data))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}

	if _, err := conn.Write(lengthBytes); err != nil {
		return err
	}

	_, err = conn.Write(data)
	return err
}

func ReceiveResponse(conn net.Conn) (*Response, error) {
	lengthBytes := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBytes); err != nil {
		return nil, err
	}

	length := uint32(lengthBytes[0])<<24 |
		uint32(lengthBytes[1])<<16 |
		uint32(lengthBytes[2])<<8 |
		uint32(lengthBytes[3])

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	var resp Response
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

func SendStdinData(conn net.Conn, stdin io.Reader) error {
	scanner := bufio.NewScanner(stdin)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, bufio.MaxScanTokenSize)

	for scanner.Scan() {
		line := scanner.Text()
		data, err := json.Marshal(line + "\n")
		if err != nil {
			return err
		}

		length := uint32(len(data))
		lengthBytes := []byte{
			byte(length >> 24),
			byte(length >> 16),
			byte(length >> 8),
			byte(length),
		}

		if _, err := conn.Write(lengthBytes); err != nil {
			return err
		}
		if _, err := conn.Write(data); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	eofMarker, _ := json.Marshal("__EOF__")
	length := uint32(len(eofMarker))
	lengthBytes := []byte{
		byte(length >> 24),
		byte(length >> 16),
		byte(length >> 8),
		byte(length),
	}
	if _, err := conn.Write(lengthBytes); err != nil {
		return err
	}
	_, err := conn.Write(eofMarker)
	return err
}

func ReceiveStdinData(conn net.Conn) (io.Reader, error) {
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		for {
			lengthBytes := make([]byte, 4)
			if _, err := io.ReadFull(conn, lengthBytes); err != nil {
				return
			}

			length := uint32(lengthBytes[0])<<24 |
				uint32(lengthBytes[1])<<16 |
				uint32(lengthBytes[2])<<8 |
				uint32(lengthBytes[3])

			data := make([]byte, length)
			if _, err := io.ReadFull(conn, data); err != nil {
				return
			}

			var line string
			if err := json.Unmarshal(data, &line); err != nil {
				return
			}

			if line == "__EOF__" {
				return
			}

			if _, err := pw.Write([]byte(line)); err != nil {
				return
			}
		}
	}()

	return pr, nil
}

func FormatValue(value interface{}, compact bool) string {
	if IsNotFound(value) {
		return ""
	}

	switch v := value.(type) {
	case nil:
		if compact {
			return "null"
		}
		return "null"
	case string:
		if compact {
			return v
		}
		return v
	case float64:
		if compact {
			return fmt.Sprintf("%v", v)
		}
		return fmt.Sprintf("%v", v)
	case bool:
		if compact {
			return fmt.Sprintf("%v", v)
		}
		return fmt.Sprintf("%v", v)
	default:
		if compact {
			bytes, _ := json.Marshal(v)
			return string(bytes)
		}
		bytes, _ := json.Marshal(v)
		return string(bytes)
	}
}

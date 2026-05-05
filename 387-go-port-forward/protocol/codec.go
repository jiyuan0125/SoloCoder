package protocol

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

func EncodeMessage(msg Message) ([]byte, error) {
	payloadLen := len(msg.Payload)
	buf := make([]byte, 5+payloadLen)
	
	buf[0] = byte(msg.Type)
	binary.BigEndian.PutUint32(buf[1:5], uint32(payloadLen))
	
	if payloadLen > 0 {
		copy(buf[5:], msg.Payload)
	}
	
	return buf, nil
}

func DecodeMessage(conn net.Conn) (Message, error) {
	header := make([]byte, 5)
	_, err := io.ReadFull(conn, header)
	if err != nil {
		return Message{}, err
	}
	
	msgType := MessageType(header[0])
	payloadLen := binary.BigEndian.Uint32(header[1:5])
	
	var payload []byte
	if payloadLen > 0 {
		payload = make([]byte, payloadLen)
		_, err = io.ReadFull(conn, payload)
		if err != nil {
			return Message{}, err
		}
	}
	
	return Message{
		Type:    msgType,
		Payload: payload,
	}, nil
}

func EncodeJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func DecodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func SendMessage(conn net.Conn, msg Message) error {
	encoded, err := EncodeMessage(msg)
	if err != nil {
		return err
	}
	_, err = conn.Write(encoded)
	return err
}

func SendResponse(conn net.Conn, success bool, message string, data interface{}) error {
	resp := Response{
		Success: success,
		Message: message,
	}
	
	if data != nil {
		encodedData, err := EncodeJSON(data)
		if err != nil {
			return err
		}
		resp.Data = encodedData
	}
	
	payload, err := EncodeJSON(resp)
	if err != nil {
		return err
	}
	
	msg := Message{
		Type:    MsgTypeResponse,
		Payload: payload,
	}
	
	return SendMessage(conn, msg)
}

func ReceiveResponse(conn net.Conn) (Response, error) {
	msg, err := DecodeMessage(conn)
	if err != nil {
		return Response{}, err
	}
	
	var resp Response
	err = DecodeJSON(msg.Payload, &resp)
	return resp, err
}

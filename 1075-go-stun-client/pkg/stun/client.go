package stun

import (
	"bytes"
	"crypto/rand"
	"errors"
	"net"
	"time"
)

type Client struct {
	Timeout    time.Duration
	MaxRetries int
}

type ClientOption func(*Client)

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.Timeout = timeout
	}
}

func WithMaxRetries(maxRetries int) ClientOption {
	return func(c *Client) {
		c.MaxRetries = maxRetries
	}
}

func NewClient(options ...ClientOption) *Client {
	c := &Client{
		Timeout:    3 * time.Second,
		MaxRetries: 3,
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

type BindingResult struct {
	MappingAddress *MappedAddress
	Response       *Message
}

func (c *Client) PerformBinding(serverAddr string, credentials *Credentials) (*BindingResult, error) {
	return c.PerformBindingFromAddr(serverAddr, credentials, "")
}

type Credentials struct {
	Username string
	Password string
	Realm    string
	Nonce    string
}

func (c *Client) PerformBindingFromAddr(serverAddr string, credentials *Credentials, localAddr string) (*BindingResult, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", serverAddr)
	if err != nil {
		return nil, err
	}

	var conn *net.UDPConn
	if localAddr == "" {
		conn, err = net.DialUDP("udp", nil, udpAddr)
	} else {
		lAddr, err := net.ResolveUDPAddr("udp", localAddr)
		if err != nil {
			return nil, err
		}
		conn, err = net.DialUDP("udp", lAddr, udpAddr)
	}
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return c.doBinding(conn, udpAddr, credentials)
}

func (c *Client) doBinding(conn *net.UDPConn, serverAddr *net.UDPAddr, credentials *Credentials) (*BindingResult, error) {
	req, err := NewMessage(BindingRequest)
	if err != nil {
		return nil, err
	}

	var authReq *Message
	if credentials != nil {
		authReq, err = NewMessage(BindingRequest)
		if err != nil {
			return nil, err
		}
	}

	var lastError error
	var response *Message

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		var sendReq *Message
		if authReq != nil && credentials.Nonce != "" {
			sendReq = authReq
			sendReq.TransactionID, _ = generateTransactionID()
		} else {
			sendReq = req
		}

		var data []byte
		data, err = buildAuthenticatedRequest(sendReq, credentials)
		if err != nil {
			return nil, err
		}

		if _, err = conn.Write(data); err != nil {
			lastError = err
			continue
		}

		response, _, err = c.receiveResponse(conn, sendReq.TransactionID)
		if err != nil {
			lastError = err
			continue
		}

		if response.Type == BindingErrorResponse {
			if authReq != nil && checkErrorRequiresAuth(response) && credentials.Username != "" && credentials.Password != "" {
				updateCredentialsFromError(response, credentials)
				continue
			}
			return nil, errors.New("received binding error response")
		}

		if response.Type == BindingResponse {
			break
		}
	}

	if response == nil {
		if lastError != nil {
			return nil, lastError
		}
		return nil, errors.New("no response received after retries")
	}

	mappingAddr, err := extractMappedAddress(response)
	if err != nil {
		return nil, err
	}

	return &BindingResult{
		MappingAddress: mappingAddr,
		Response:       response,
	}, nil
}

func (c *Client) receiveResponse(conn *net.UDPConn, expectedTID []byte) (*Message, []byte, error) {
	buf := make([]byte, MaxMessageSize)

	deadline := time.Now().Add(c.Timeout)
	for time.Now().Before(deadline) {
		conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))

		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return nil, nil, err
		}

		originalData := make([]byte, n)
		copy(originalData, buf[:n])

		msg, err := DecodeMessage(originalData)
		if err != nil {
			continue
		}

		if bytes.Equal(msg.TransactionID, expectedTID) {
			return msg, originalData, nil
		}
	}

	return nil, nil, errors.New("timeout waiting for response")
}

func generateTransactionID() ([]byte, error) {
	tid := make([]byte, TransactionIDLen)
	_, err := rand.Read(tid)
	return tid, err
}

func buildAuthenticatedRequest(req *Message, credentials *Credentials) ([]byte, error) {
	if credentials == nil || credentials.Username == "" || credentials.Nonce == "" {
		return req.Encode()
	}

	newReq, err := NewMessage(req.Type)
	if err != nil {
		return nil, err
	}
	copy(newReq.TransactionID, req.TransactionID)

	usernameAttr := NewAttribute(AttrUsername, []byte(credentials.Username))
	newReq.AddAttribute(usernameAttr)

	if credentials.Realm != "" {
		realmAttr := NewAttribute(AttrRealm, []byte(credentials.Realm))
		newReq.AddAttribute(realmAttr)
	}

	nonceAttr := NewAttribute(AttrNonce, []byte(credentials.Nonce))
	newReq.AddAttribute(nonceAttr)

	mi, err := NewMessageIntegrity(newReq, []byte(credentials.Password))
	if err != nil {
		return nil, err
	}
	newReq.AddAttribute(mi)

	fp, err := NewFingerprint(newReq)
	if err != nil {
		return nil, err
	}
	newReq.AddAttribute(fp)

	return newReq.Encode()
}

func checkErrorRequiresAuth(msg *Message) bool {
	errAttr := msg.GetAttribute(AttrErrorCode)
	if errAttr == nil {
		return false
	}

	code := extractErrorCode(errAttr)
	hasNonce := msg.GetAttribute(AttrNonce) != nil

	return (code == 401 || code == 438) && hasNonce
}

func extractErrorCode(attr *Attribute) int {
	if len(attr.Value) < 4 {
		return 0
	}
	class := int(attr.Value[2])
	number := int(attr.Value[3])
	return class*100 + number
}

func updateCredentialsFromError(msg *Message, credentials *Credentials) {
	if nonceAttr := msg.GetAttribute(AttrNonce); nonceAttr != nil {
		credentials.Nonce = string(nonceAttr.Value)
	}
	if realmAttr := msg.GetAttribute(AttrRealm); realmAttr != nil {
		credentials.Realm = string(realmAttr.Value)
	}
}

func extractMappedAddress(msg *Message) (*MappedAddress, error) {
	if xorAttr := msg.GetAttribute(AttrXORMappedAddress); xorAttr != nil {
		return DecodeXORMappedAddress(xorAttr, msg.MagicCookie, msg.TransactionID)
	}

	if mappedAttr := msg.GetAttribute(AttrMappedAddress); mappedAttr != nil {
		return decodeMappedAddress(mappedAttr)
	}

	return nil, errors.New("no mapped address found in response")
}

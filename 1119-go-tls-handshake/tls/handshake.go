package tls

import (
	"crypto/rand"
	"fmt"
)

type SimulationConfig struct {
	ClientCipherSuites []CipherSuiteID
	ServerCipherSuites []CipherSuiteID
	ClientExtensions   []Extension
	ServerExtensions   []Extension
	SessionCache       map[string]struct{}
	ResumeSessionID    SessionID
	Certificate        []byte
}

func DefaultSimulationConfig() *SimulationConfig {
	return &SimulationConfig{
		ClientCipherSuites: DefaultClientCipherSuites,
		ServerCipherSuites: DefaultServerCipherSuites,
		ClientExtensions: []Extension{
			&SNIExtension{HostName: "example.com"},
			&SupportedGroupsExtension{Groups: []NamedGroup{GroupX25519, GroupSecP256R1, GroupSecP384R1}},
			&SignatureAlgorithmsExtension{Algorithms: []SignatureAlgorithm{
				SignatureRSA_PSS_RSAE_SHA256,
				SignatureRSA_PKCS1_SHA256,
				SignatureECDSA_P256_SHA256,
			}},
			&ALPNExtension{Protocols: []string{"h2", "http/1.1"}},
		},
		ServerExtensions: []Extension{},
		SessionCache:     make(map[string]struct{}),
		Certificate:      []byte("simulated-certificate"),
	}
}

func RunHandshakeSimulation(config *SimulationConfig) (*HandshakeResult, error) {
	if config == nil {
		config = DefaultSimulationConfig()
	}
	ctx := NewHandshakeContext(config.ServerCipherSuites, config.SessionCache)
	var sessionID SessionID
	if config.ResumeSessionID != nil && len(config.ResumeSessionID) > 0 {
		sessionID = config.ResumeSessionID
	} else {
		sessionID = generateSessionID()
	}
	clientHello := NewClientHello(TLSVersion1_2, sessionID, config.ClientCipherSuites, config.ClientExtensions)
	if err := ctx.HandleClientHello(clientHello); err != nil {
		return ctx.GetResult(), err
	}
	var serverSessionID SessionID
	if ctx.IsSessionResume() {
		serverSessionID = clientHello.SessionID
	} else {
		serverSessionID = generateSessionID()
	}
	serverHello := NewServerHello(TLSVersion1_2, serverSessionID, ctx.NegotiatedSuite(), config.ServerExtensions)
	if err := ctx.HandleServerHello(serverHello); err != nil {
		return ctx.GetResult(), err
	}
	if ctx.IsSessionResume() {
		clientCCS := NewChangeCipherSpecMessage()
		if err := ctx.HandleClientChangeCipher(clientCCS); err != nil {
			return ctx.GetResult(), err
		}
		clientFinished := NewFinishedMessage([]byte("simulated-client-finished"))
		if err := ctx.HandleClientFinished(clientFinished); err != nil {
			return ctx.GetResult(), err
		}
	} else {
		certificate := NewCertificateMessage([][]byte{config.Certificate})
		if err := ctx.HandleCertificate(certificate); err != nil {
			return ctx.GetResult(), err
		}
		if ctx.State() == StateWaitServerKeyExchange {
			serverKeyExchange := NewServerKeyExchangeMessage()
			if err := ctx.HandleServerKeyExchange(serverKeyExchange); err != nil {
				return ctx.GetResult(), err
			}
		}
		serverHelloDone := NewServerHelloDoneMessage()
		if err := ctx.HandleServerHelloDone(serverHelloDone); err != nil {
			return ctx.GetResult(), err
		}
		clientKeyExchange := NewClientKeyExchangeMessage([]byte("simulated-premaster-secret"))
		if err := ctx.HandleClientKeyExchange(clientKeyExchange); err != nil {
			return ctx.GetResult(), err
		}
		clientCCS := NewChangeCipherSpecMessage()
		if err := ctx.HandleClientChangeCipher(clientCCS); err != nil {
			return ctx.GetResult(), err
		}
		clientFinished := NewFinishedMessage([]byte("simulated-client-finished"))
		if err := ctx.HandleClientFinished(clientFinished); err != nil {
			return ctx.GetResult(), err
		}
		serverCCS := NewChangeCipherSpecMessage()
		if err := ctx.HandleServerChangeCipher(serverCCS); err != nil {
			return ctx.GetResult(), err
		}
		serverFinished := NewFinishedMessage([]byte("simulated-server-finished"))
		if err := ctx.HandleServerFinished(serverFinished); err != nil {
			return ctx.GetResult(), err
		}
	}
	return ctx.GetResult(), nil
}

func generateSessionID() SessionID {
	id := make([]byte, 32)
	rand.Read(id)
	return SessionID(id)
}

func ParseCipherSuiteIDs(ids []string) ([]CipherSuiteID, error) {
	var result []CipherSuiteID
	for _, idStr := range ids {
		var id uint16
		if len(idStr) > 2 && (idStr[:2] == "0x" || idStr[:2] == "0X") {
			_, err := fmt.Sscanf(idStr, "0x%x", &id)
			if err != nil {
				return nil, fmt.Errorf("invalid cipher suite ID: %s", idStr)
			}
		} else {
			_, err := fmt.Sscanf(idStr, "%d", &id)
			if err != nil {
				return nil, fmt.Errorf("invalid cipher suite ID: %s", idStr)
			}
		}
		result = append(result, CipherSuiteID(id))
	}
	return result, nil
}

func GetAllCipherSuites() []CipherSuite {
	var result []CipherSuite
	for _, cs := range CipherSuiteRegistry {
		result = append(result, cs)
	}
	return result
}

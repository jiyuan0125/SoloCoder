package tls

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type HandshakeState int

const (
	StateInitial HandshakeState = iota
	StateWaitClientHello
	StateWaitServerHello
	StateWaitCertificate
	StateWaitServerKeyExchange
	StateWaitServerHelloDone
	StateWaitClientKeyExchange
	StateWaitClientChangeCipher
	StateWaitClientFinished
	StateWaitServerChangeCipher
	StateWaitServerFinished
	StateSessionResumeWaitClientChangeCipher
	StateSessionResumeWaitClientFinished
	StateSessionResumeWaitServerChangeCipher
	StateSessionResumeWaitServerFinished
	StateFinished
	StateError
)

func (s HandshakeState) String() string {
	switch s {
	case StateInitial:
		return "Initial"
	case StateWaitClientHello:
		return "WaitClientHello"
	case StateWaitServerHello:
		return "WaitServerHello"
	case StateWaitCertificate:
		return "WaitCertificate"
	case StateWaitServerKeyExchange:
		return "WaitServerKeyExchange"
	case StateWaitServerHelloDone:
		return "WaitServerHelloDone"
	case StateWaitClientKeyExchange:
		return "WaitClientKeyExchange"
	case StateWaitClientChangeCipher:
		return "WaitClientChangeCipher"
	case StateWaitClientFinished:
		return "WaitClientFinished"
	case StateWaitServerChangeCipher:
		return "WaitServerChangeCipher"
	case StateWaitServerFinished:
		return "WaitServerFinished"
	case StateSessionResumeWaitClientChangeCipher:
		return "SessionResumeWaitClientChangeCipher"
	case StateSessionResumeWaitClientFinished:
		return "SessionResumeWaitClientFinished"
	case StateSessionResumeWaitServerChangeCipher:
		return "SessionResumeWaitServerChangeCipher"
	case StateSessionResumeWaitServerFinished:
		return "SessionResumeWaitServerFinished"
	case StateFinished:
		return "Finished"
	case StateError:
		return "Error"
	default:
		return "Unknown"
	}
}

type HandshakeStep struct {
	FromState HandshakeState
	ToState   HandshakeState
	Message   fmt.Stringer
	IsError   bool
	ErrorMsg  string
}

type CipherSuiteInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	KeyExchange string `json:"key_exchange"`
	Encryption  string `json:"encryption"`
	MAC         string `json:"mac"`
}

type HandshakeResult struct {
	Success           bool              `json:"success"`
	IsSessionResume   bool              `json:"is_session_resume"`
	Error             string            `json:"error,omitempty"`
	NegotiatedSuite   *CipherSuiteInfo  `json:"negotiated_suite,omitempty"`
	ClientRandom      string            `json:"client_random,omitempty"`
	ServerRandom      string            `json:"server_random,omitempty"`
	SessionID         string            `json:"session_id,omitempty"`
	ClientExtensions  []string          `json:"client_extensions,omitempty"`
	ServerExtensions  []string          `json:"server_extensions,omitempty"`
	Steps             []HandshakeStep   `json:"-"`
	StepsText         []string          `json:"steps"`
	FullMessageDump   string            `json:"full_message_dump,omitempty"`
}

type HandshakeContext struct {
	state             HandshakeState
	serverCipherSuites []CipherSuiteID
	sessionCache      map[string]struct{}
	sessionID         SessionID
	clientHello       *ClientHelloMessage
	serverHello       *ServerHelloMessage
	certificate       *CertificateMessage
	serverKeyExchange *ServerKeyExchangeMessage
	serverHelloDone   *ServerHelloDoneMessage
	clientKeyExchange *ClientKeyExchangeMessage
	finished1         *FinishedMessage
	finished2         *FinishedMessage
	isSessionResume   bool
	negotiatedSuite   CipherSuiteID
	steps             []HandshakeStep
	errorMsg          string
}

func NewHandshakeContext(serverCipherSuites []CipherSuiteID, sessionCache map[string]struct{}) *HandshakeContext {
	if serverCipherSuites == nil {
		serverCipherSuites = DefaultServerCipherSuites
	}
	if sessionCache == nil {
		sessionCache = make(map[string]struct{})
	}
	return &HandshakeContext{
		state:              StateWaitClientHello,
		serverCipherSuites: serverCipherSuites,
		sessionCache:       sessionCache,
		steps:              []HandshakeStep{},
	}
}

func (ctx *HandshakeContext) State() HandshakeState {
	return ctx.state
}

func (ctx *HandshakeContext) IsSessionResume() bool {
	return ctx.isSessionResume
}

func (ctx *HandshakeContext) NegotiatedSuite() CipherSuiteID {
	return ctx.negotiatedSuite
}

func (ctx *HandshakeContext) Steps() []HandshakeStep {
	return ctx.steps
}

func (ctx *HandshakeContext) Error() string {
	return ctx.errorMsg
}

func (ctx *HandshakeContext) addStep(from, to HandshakeState, msg fmt.Stringer, isErr bool, errMsg string) {
	ctx.steps = append(ctx.steps, HandshakeStep{
		FromState: from,
		ToState:   to,
		Message:   msg,
		IsError:   isErr,
		ErrorMsg:  errMsg,
	})
}

func (ctx *HandshakeContext) transitionToError(errMsg string, expectedType HandshakeType, actualType HandshakeType) {
	ctx.errorMsg = errMsg
	ctx.addStep(ctx.state, StateError, nil, true, errMsg)
	ctx.state = StateError
}

func (ctx *HandshakeContext) HandleClientHello(ch *ClientHelloMessage) error {
	if ctx.state != StateWaitClientHello {
		ctx.transitionToError(fmt.Sprintf("unexpected ClientHello in state %s", ctx.state), HandshakeTypeClientHello, HandshakeTypeClientHello)
		return fmt.Errorf(ctx.errorMsg)
	}
	if ch.Version != TLSVersion1_2 {
		ctx.transitionToError(fmt.Sprintf("unsupported TLS version: 0x%04X, expected 0x%04X (TLS 1.2)", ch.Version, TLSVersion1_2), HandshakeTypeClientHello, HandshakeTypeClientHello)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.clientHello = ch
	if len(ch.SessionID) > 0 {
		if _, ok := ctx.sessionCache[string(ch.SessionID)]; ok {
			ctx.isSessionResume = true
			ctx.sessionID = ch.SessionID
		}
	}
	suite, err := NegotiateCipherSuite(ch.CipherSuites, ctx.serverCipherSuites)
	if err != nil {
		ctx.transitionToError(err.Error(), HandshakeTypeClientHello, HandshakeTypeClientHello)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.negotiatedSuite = suite
	ctx.addStep(ctx.state, StateWaitServerHello, ch, false, "")
	ctx.state = StateWaitServerHello
	return nil
}

func (ctx *HandshakeContext) HandleServerHello(sh *ServerHelloMessage) error {
	if ctx.state != StateWaitServerHello {
		ctx.transitionToError(fmt.Sprintf("unexpected ServerHello in state %s", ctx.state), HandshakeTypeServerHello, HandshakeTypeServerHello)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.serverHello = sh
	if ctx.isSessionResume {
		ctx.addStep(ctx.state, StateSessionResumeWaitClientChangeCipher, sh, false, "")
		ctx.state = StateSessionResumeWaitClientChangeCipher
	} else {
		ctx.addStep(ctx.state, StateWaitCertificate, sh, false, "")
		ctx.state = StateWaitCertificate
	}
	return nil
}

func (ctx *HandshakeContext) HandleCertificate(cert *CertificateMessage) error {
	if ctx.state != StateWaitCertificate {
		ctx.transitionToError(fmt.Sprintf("unexpected Certificate in state %s", ctx.state), HandshakeTypeCertificate, HandshakeTypeCertificate)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.certificate = cert
	if info, ok := ctx.negotiatedSuite.GetInfo(); ok {
		if strings.HasPrefix(info.KeyExchange, "ECDHE") {
			ctx.addStep(ctx.state, StateWaitServerKeyExchange, cert, false, "")
			ctx.state = StateWaitServerKeyExchange
			return nil
		}
	}
	ctx.addStep(ctx.state, StateWaitServerHelloDone, cert, false, "")
	ctx.state = StateWaitServerHelloDone
	return nil
}

func (ctx *HandshakeContext) HandleServerKeyExchange(ske *ServerKeyExchangeMessage) error {
	if ctx.state != StateWaitServerKeyExchange {
		ctx.transitionToError(fmt.Sprintf("unexpected ServerKeyExchange in state %s", ctx.state), HandshakeTypeServerKeyExchange, HandshakeTypeServerKeyExchange)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.serverKeyExchange = ske
	ctx.addStep(ctx.state, StateWaitServerHelloDone, ske, false, "")
	ctx.state = StateWaitServerHelloDone
	return nil
}

func (ctx *HandshakeContext) HandleServerHelloDone(shd *ServerHelloDoneMessage) error {
	expectedState := ctx.state == StateWaitServerHelloDone
	if !expectedState {
		ctx.transitionToError(fmt.Sprintf("unexpected ServerHelloDone in state %s", ctx.state), HandshakeTypeServerHelloDone, HandshakeTypeServerHelloDone)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.serverHelloDone = shd
	ctx.addStep(ctx.state, StateWaitClientKeyExchange, shd, false, "")
	ctx.state = StateWaitClientKeyExchange
	return nil
}

func (ctx *HandshakeContext) HandleClientKeyExchange(cke *ClientKeyExchangeMessage) error {
	if ctx.state != StateWaitClientKeyExchange {
		ctx.transitionToError(fmt.Sprintf("unexpected ClientKeyExchange in state %s", ctx.state), HandshakeTypeClientKeyExchange, HandshakeTypeClientKeyExchange)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.clientKeyExchange = cke
	ctx.addStep(ctx.state, StateWaitClientChangeCipher, cke, false, "")
	ctx.state = StateWaitClientChangeCipher
	return nil
}

func (ctx *HandshakeContext) HandleClientChangeCipher(ccs *ChangeCipherSpecMessage) error {
	expectedState := ctx.state == StateWaitClientChangeCipher || ctx.state == StateSessionResumeWaitClientChangeCipher
	if !expectedState {
		ctx.transitionToError(fmt.Sprintf("unexpected ChangeCipherSpec in state %s", ctx.state), 0, 0)
		return fmt.Errorf(ctx.errorMsg)
	}
	if ctx.isSessionResume {
		ctx.addStep(ctx.state, StateSessionResumeWaitClientFinished, ccs, false, "")
		ctx.state = StateSessionResumeWaitClientFinished
	} else {
		ctx.addStep(ctx.state, StateWaitClientFinished, ccs, false, "")
		ctx.state = StateWaitClientFinished
	}
	return nil
}

func (ctx *HandshakeContext) HandleClientFinished(fin *FinishedMessage) error {
	expectedState := ctx.state == StateWaitClientFinished || ctx.state == StateSessionResumeWaitClientFinished
	if !expectedState {
		ctx.transitionToError(fmt.Sprintf("unexpected Finished in state %s", ctx.state), HandshakeTypeFinished, HandshakeTypeFinished)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.finished1 = fin
	if ctx.isSessionResume {
		ctx.addStep(ctx.state, StateFinished, fin, false, "")
		ctx.state = StateFinished
	} else {
		ctx.addStep(ctx.state, StateWaitServerChangeCipher, fin, false, "")
		ctx.state = StateWaitServerChangeCipher
	}
	return nil
}

func (ctx *HandshakeContext) HandleServerChangeCipher(ccs *ChangeCipherSpecMessage) error {
	if ctx.state != StateWaitServerChangeCipher {
		ctx.transitionToError(fmt.Sprintf("unexpected ChangeCipherSpec in state %s", ctx.state), 0, 0)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.addStep(ctx.state, StateWaitServerFinished, ccs, false, "")
	ctx.state = StateWaitServerFinished
	return nil
}

func (ctx *HandshakeContext) HandleServerFinished(fin *FinishedMessage) error {
	if ctx.state != StateWaitServerFinished {
		ctx.transitionToError(fmt.Sprintf("unexpected Finished in state %s", ctx.state), HandshakeTypeFinished, HandshakeTypeFinished)
		return fmt.Errorf(ctx.errorMsg)
	}
	ctx.finished2 = fin
	ctx.addStep(ctx.state, StateFinished, fin, false, "")
	ctx.state = StateFinished
	return nil
}

func (ctx *HandshakeContext) GetResult() *HandshakeResult {
	result := &HandshakeResult{
		Success:         ctx.state == StateFinished,
		IsSessionResume: ctx.isSessionResume,
	}
	if ctx.state == StateError {
		result.Error = ctx.errorMsg
	}
	if ctx.clientHello != nil {
		result.ClientRandom = hex.EncodeToString(ctx.clientHello.Random.Bytes())
		result.SessionID = hex.EncodeToString(ctx.clientHello.SessionID)
		for _, ext := range ctx.clientHello.Extensions {
			result.ClientExtensions = append(result.ClientExtensions, ext.String())
		}
	}
	if ctx.serverHello != nil {
		result.ServerRandom = hex.EncodeToString(ctx.serverHello.Random.Bytes())
		for _, ext := range ctx.serverHello.Extensions {
			result.ServerExtensions = append(result.ServerExtensions, ext.String())
		}
	}
	if ctx.negotiatedSuite != 0 {
		if info, ok := ctx.negotiatedSuite.GetInfo(); ok {
			result.NegotiatedSuite = &CipherSuiteInfo{
				ID:          fmt.Sprintf("0x%04X", ctx.negotiatedSuite),
				Name:        info.Name,
				KeyExchange: info.KeyExchange,
				Encryption:  info.Encryption,
				MAC:         info.MAC,
			}
		}
	}
	result.StepsText = make([]string, 0, len(ctx.steps))
	for i, step := range ctx.steps {
		var stepText string
		if step.IsError {
			stepText = fmt.Sprintf("[%d] %s -> %s: ERROR: %s", i+1, step.FromState, step.ToState, step.ErrorMsg)
		} else if step.Message != nil {
			stepText = fmt.Sprintf("[%d] %s -> %s: %s", i+1, step.FromState, step.ToState, getMessageTypeName(step.Message))
		} else {
			stepText = fmt.Sprintf("[%d] %s -> %s", i+1, step.FromState, step.ToState)
		}
		result.StepsText = append(result.StepsText, stepText)
	}
	result.Steps = ctx.steps
	var fullDump strings.Builder
	for _, step := range ctx.steps {
		if step.Message != nil {
			fullDump.WriteString("===== ")
			fullDump.WriteString(getMessageTypeName(step.Message))
			fullDump.WriteString(" =====\n")
			fullDump.WriteString(step.Message.String())
			fullDump.WriteString("\n\n")
		}
	}
	result.FullMessageDump = fullDump.String()
	return result
}

func getMessageTypeName(msg fmt.Stringer) string {
	switch m := msg.(type) {
	case *ClientHelloMessage:
		return "ClientHello"
	case *ServerHelloMessage:
		return "ServerHello"
	case *CertificateMessage:
		return "Certificate"
	case *ServerKeyExchangeMessage:
		return "ServerKeyExchange"
	case *ServerHelloDoneMessage:
		return "ServerHelloDone"
	case *ClientKeyExchangeMessage:
		return "ClientKeyExchange"
	case *ChangeCipherSpecMessage:
		return "ChangeCipherSpec"
	case *FinishedMessage:
		return "Finished"
	default:
		_ = m
		return "Unknown"
	}
}

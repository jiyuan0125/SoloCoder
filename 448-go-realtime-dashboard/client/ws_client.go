package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"realtime-dashboard/common"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	serverURL      string
	conn           *websocket.Conn
	clientID       string
	disconnectTime *time.Time
	subscriptions  map[string]bool
	refreshRate    int

	metricHandler  func(*common.Metric)
	alertHandler   func(*common.Alert)
	errorHandler   func(string)

	mu             sync.Mutex
	connected      bool
}

func NewWSClient(serverURL string) *WSClient {
	return &WSClient{
		serverURL:     serverURL,
		clientID:      common.GenerateClientID(),
		subscriptions: make(map[string]bool),
		refreshRate:   int(common.MinRefreshRate.Seconds()),
		connected:     false,
	}
}

func (w *WSClient) SetClientID(clientID string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.clientID = clientID
}

func (w *WSClient) SetMetricHandler(handler func(*common.Metric)) {
	w.metricHandler = handler
}

func (w *WSClient) SetAlertHandler(handler func(*common.Alert)) {
	w.alertHandler = handler
}

func (w *WSClient) SetErrorHandler(handler func(string)) {
	w.errorHandler = handler
}

func (w *WSClient) Connect() error {
	w.mu.Lock()
	if w.connected {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	wsURL := fmt.Sprintf("%s/ws", w.serverURL)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.conn = conn
	w.connected = true
	w.mu.Unlock()

	go w.readLoop()
	go w.heartbeatLoop()

	return nil
}

func (w *WSClient) Disconnect() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.connected {
		return
	}

	now := time.Now()
	w.disconnectTime = &now
	w.connected = false

	if w.conn != nil {
		w.conn.Close()
	}
}

func (w *WSClient) Reconnect() error {
	if w.disconnectTime == nil {
		return w.Connect()
	}

	if err := w.Connect(); err != nil {
		return err
	}

	reconnectReq := common.ReconnectRequest{
		ClientID:       w.clientID,
		DisconnectTime: *w.disconnectTime,
	}

	msg, err := common.EncodeMessage(common.MsgTypeReconnect, reconnectReq)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WSClient) Subscribe(metricKeys []string, refreshRate int) error {
	w.mu.Lock()
	for _, key := range metricKeys {
		w.subscriptions[key] = true
	}
	if refreshRate > w.refreshRate {
		w.refreshRate = refreshRate
	}
	w.mu.Unlock()

	req := common.SubscribeRequest{
		ClientID:    w.clientID,
		MetricKeys:  metricKeys,
		RefreshRate: refreshRate,
	}

	msg, err := common.EncodeMessage(common.MsgTypeSubscribe, req)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.connected {
		return fmt.Errorf("not connected")
	}
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WSClient) Unsubscribe(metricKeys []string) error {
	w.mu.Lock()
	for _, key := range metricKeys {
		delete(w.subscriptions, key)
	}
	w.mu.Unlock()

	req := common.UnsubscribeRequest{
		ClientID:   w.clientID,
		MetricKeys: metricKeys,
	}

	msg, err := common.EncodeMessage(common.MsgTypeUnsubscribe, req)
	if err != nil {
		return err
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	if !w.connected {
		return fmt.Errorf("not connected")
	}
	return w.conn.WriteMessage(websocket.TextMessage, msg)
}

func (w *WSClient) readLoop() {
	for {
		w.mu.Lock()
		if !w.connected {
			w.mu.Unlock()
			return
		}
		conn := w.conn
		w.mu.Unlock()

		_, message, err := conn.ReadMessage()
		if err != nil {
			w.Disconnect()
			if w.errorHandler != nil {
				w.errorHandler(fmt.Sprintf("connection error: %v", err))
			}
			return
		}

		w.handleMessage(message)
	}
}

func (w *WSClient) handleMessage(rawMsg []byte) {
	msg, err := common.DecodeMessage(rawMsg)
	if err != nil {
		if w.errorHandler != nil {
			w.errorHandler(fmt.Sprintf("decode error: %v", err))
		}
		return
	}

	switch msg.Type {
	case common.MsgTypeMetricUpdate:
		var metric common.Metric
		if err := json.Unmarshal(msg.Payload, &metric); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(fmt.Sprintf("metric decode error: %v", err))
			}
			return
		}
		if w.metricHandler != nil {
			w.metricHandler(&metric)
		}

	case common.MsgTypeAlert:
		var alert common.Alert
		if err := json.Unmarshal(msg.Payload, &alert); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(fmt.Sprintf("alert decode error: %v", err))
			}
			return
		}
		if w.alertHandler != nil {
			w.alertHandler(&alert)
		}

	case common.MsgTypeReconnectResponse:
		var resp common.ReconnectResponse
		if err := json.Unmarshal(msg.Payload, &resp); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(fmt.Sprintf("reconnect response decode error: %v", err))
			}
			return
		}
		if resp.LatestOnly {
			fmt.Printf("[Reconnect] Note: Disconnect exceeded 5 minutes, only latest values available\n")
		}
		for _, metric := range resp.MissedData {
			if w.metricHandler != nil {
				w.metricHandler(&metric)
			}
		}

	case common.MsgTypeSubscribe:
		var resp common.SubscribeResponse
		if err := json.Unmarshal(msg.Payload, &resp); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(fmt.Sprintf("subscribe response decode error: %v", err))
			}
			return
		}
		if resp.QueuePosition != nil {
			fmt.Printf("[Info] In queue, position: %d\n", *resp.QueuePosition)
		}

	case common.MsgTypeError:
		var errMsg map[string]string
		if err := json.Unmarshal(msg.Payload, &errMsg); err != nil {
			if w.errorHandler != nil {
				w.errorHandler(fmt.Sprintf("error decode error: %v", err))
			}
			return
		}
		if w.errorHandler != nil {
			w.errorHandler(errMsg["error"])
		}

	case common.MsgTypeHeartbeat:
	}
}

func (w *WSClient) heartbeatLoop() {
	ticker := time.NewTicker(common.HeartbeatInterval)
	defer ticker.Stop()

	for range ticker.C {
		w.mu.Lock()
		if !w.connected {
			w.mu.Unlock()
			return
		}
		conn := w.conn
		w.mu.Unlock()

		heartbeatMsg := map[string]string{
			"type": "heartbeat",
		}
		msg, err := common.EncodeMessage(common.MsgTypeHeartbeat, heartbeatMsg)
		if err != nil {
			continue
		}

		w.mu.Lock()
		conn.WriteMessage(websocket.TextMessage, msg)
		w.mu.Unlock()
	}
}

func (w *WSClient) IsConnected() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.connected
}

func (w *WSClient) GetClientID() string {
	return w.clientID
}

func (w *WSClient) GetSubscriptions() []string {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([]string, 0, len(w.subscriptions))
	for key := range w.subscriptions {
		result = append(result, key)
	}
	return result
}

func (w *WSClient) GetDisconnectTime() *time.Time {
	return w.disconnectTime
}

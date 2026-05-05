package main

import (
	"sync"
	"time"

	"realtime-dashboard/common"

	"github.com/gorilla/websocket"
)

type ClientConnection struct {
	ClientID        string
	Conn            *websocket.Conn
	ConnectedAt     time.Time
	LastHeartbeat   time.Time
	SubscribedMetrics map[string]bool
	RefreshRate     time.Duration
	LastPushTime    map[string]time.Time
	DisconnectTime  *time.Time
	sendChan        chan []byte
	mu              sync.Mutex
}

type ConnectionManager struct {
	clients        map[string]*ClientConnection
	waitingQueue   []chan *ClientConnection
	maxClients     int
	mu             sync.RWMutex
	metricManager  *MetricManager
	alertManager   *AlertManager
}

func NewConnectionManager(mm *MetricManager, am *AlertManager) *ConnectionManager {
	return &ConnectionManager{
		clients:       make(map[string]*ClientConnection),
		waitingQueue:  make([]chan *ClientConnection, 0),
		maxClients:    common.MaxClients,
		metricManager: mm,
		alertManager:  am,
	}
}

func (cm *ConnectionManager) AddClient(clientID string, conn *websocket.Conn) (int, bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if len(cm.clients) >= cm.maxClients {
		queuePos := len(cm.waitingQueue) + 1
		waitChan := make(chan *ClientConnection, 1)
		cm.waitingQueue = append(cm.waitingQueue, waitChan)
		return queuePos, false
	}

	client := &ClientConnection{
		ClientID:          clientID,
		Conn:              conn,
		ConnectedAt:       time.Now(),
		LastHeartbeat:     time.Now(),
		SubscribedMetrics: make(map[string]bool),
		RefreshRate:       common.MinRefreshRate,
		LastPushTime:      make(map[string]time.Time),
		sendChan:          make(chan []byte, 100),
	}

	cm.clients[clientID] = client
	go cm.handleClientSend(client)

	return 0, true
}

func (cm *ConnectionManager) GetClient(clientID string) (*ClientConnection, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[clientID]
	return client, exists
}

func (cm *ConnectionManager) RemoveClient(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[clientID]
	if !exists {
		return
	}

	now := time.Now()
	client.DisconnectTime = &now

	delete(cm.clients, clientID)

	if len(cm.waitingQueue) > 0 {
		waitChan := cm.waitingQueue[0]
		cm.waitingQueue = cm.waitingQueue[1:]
		close(waitChan)
	}
}

func (cm *ConnectionManager) GetOnlineClients() []common.ClientInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]common.ClientInfo, 0, len(cm.clients))

	for _, client := range cm.clients {
		metrics := make([]string, 0, len(client.SubscribedMetrics))
		for m := range client.SubscribedMetrics {
			metrics = append(metrics, m)
		}

		result = append(result, common.ClientInfo{
			ClientID:          client.ClientID,
			ConnectedAt:       client.ConnectedAt,
			LastHeartbeat:     client.LastHeartbeat,
			SubscribedMetrics: metrics,
			RefreshRate:       int(client.RefreshRate.Seconds()),
		})
	}

	return result
}

func (cm *ConnectionManager) GetQueueSize() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.waitingQueue)
}

func (cm *ConnectionManager) GetOnlineCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.clients)
}

func (cm *ConnectionManager) Subscribe(clientID string, metricKeys []string, refreshRate int) []string {
	client, exists := cm.GetClient(clientID)
	if !exists {
		return nil
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	if refreshRate > int(common.MinRefreshRate.Seconds()) {
		client.RefreshRate = time.Duration(refreshRate) * time.Second
	} else {
		client.RefreshRate = common.MinRefreshRate
	}

	allMetrics := cm.metricManager.GetAllMetrics()
	subscribed := make([]string, 0)

	for _, key := range metricKeys {
		if _, exists := allMetrics[key]; exists {
			client.SubscribedMetrics[key] = true
			subscribed = append(subscribed, key)
		}
	}

	return subscribed
}

func (cm *ConnectionManager) Unsubscribe(clientID string, metricKeys []string) {
	client, exists := cm.GetClient(clientID)
	if !exists {
		return
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	if len(metricKeys) == 0 {
		client.SubscribedMetrics = make(map[string]bool)
		return
	}

	for _, key := range metricKeys {
		delete(client.SubscribedMetrics, key)
	}
}

func (cm *ConnectionManager) UpdateHeartbeat(clientID string) {
	client, exists := cm.GetClient(clientID)
	if !exists {
		return
	}

	client.mu.Lock()
	defer client.mu.Unlock()

	client.LastHeartbeat = time.Now()
}

func (cm *ConnectionManager) PushMetricUpdate(metricKey string, metric *common.Metric) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, client := range cm.clients {
		client.mu.Lock()
		if !client.SubscribedMetrics[metricKey] {
			client.mu.Unlock()
			continue
		}

		now := time.Now()
		lastPush, exists := client.LastPushTime[metricKey]
		if exists && now.Sub(lastPush) < client.RefreshRate {
			client.mu.Unlock()
			continue
		}

		client.LastPushTime[metricKey] = now
		client.mu.Unlock()

		msg, err := common.EncodeMessage(common.MsgTypeMetricUpdate, metric)
		if err != nil {
			continue
		}

		select {
		case client.sendChan <- msg:
		default:
		}
	}
}

func (cm *ConnectionManager) PushAlert(alert common.Alert) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	for _, client := range cm.clients {
		client.mu.Lock()
		if !client.SubscribedMetrics[alert.MetricKey] {
			client.mu.Unlock()
			continue
		}
		client.mu.Unlock()

		msg, err := common.EncodeMessage(common.MsgTypeAlert, alert)
		if err != nil {
			continue
		}

		select {
		case client.sendChan <- msg:
		default:
		}
	}
}

func (cm *ConnectionManager) handleClientSend(client *ClientConnection) {
	for msg := range client.sendChan {
		client.mu.Lock()
		err := client.Conn.WriteMessage(websocket.TextMessage, msg)
		client.mu.Unlock()

		if err != nil {
			cm.RemoveClient(client.ClientID)
			return
		}
	}
}

func (cm *ConnectionManager) HandleReconnect(clientID string, disconnectTime time.Time) *common.ReconnectResponse {
	client, exists := cm.GetClient(clientID)
	if !exists {
		return &common.ReconnectResponse{
			ClientID:    clientID,
			LatestOnly:  true,
		}
	}

	now := time.Now()
	response := &common.ReconnectResponse{
		ClientID:      clientID,
		LatestOnly:    false,
		Subscriptions: make([]string, 0),
	}

	if !common.IsWithinReconnectWindow(disconnectTime, now) {
		response.LatestOnly = true
		metrics := cm.metricManager.GetLatestValues()
		for _, mv := range metrics {
			if client.SubscribedMetrics[mv.Key] {
				metric, _ := cm.metricManager.GetMetric(mv.Key)
				if metric != nil {
					response.MissedData = append(response.MissedData, *metric)
				}
				response.Subscriptions = append(response.Subscriptions, mv.Key)
			}
		}
		return response
	}

	allMetrics := cm.metricManager.GetAllMetrics()
	for key := range client.SubscribedMetrics {
		metric, exists := allMetrics[key]
		if !exists {
			continue
		}

		if metric.Timestamp.After(disconnectTime) {
			response.MissedData = append(response.MissedData, *metric)
		}
		response.Subscriptions = append(response.Subscriptions, key)
	}

	return response
}

func (cm *ConnectionManager) SendError(clientID string, errMsg string) {
	client, exists := cm.GetClient(clientID)
	if !exists {
		return
	}

	errorMsg := map[string]string{
		"error": errMsg,
	}

	msg, err := common.EncodeMessage(common.MsgTypeError, errorMsg)
	if err != nil {
		return
	}

	select {
	case client.sendChan <- msg:
	default:
	}
}

func (cc *ClientConnection) SendMessage(msgType common.MessageType, payload interface{}) error {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	msg, err := common.EncodeMessage(msgType, payload)
	if err != nil {
		return err
	}

	select {
	case cc.sendChan <- msg:
		return nil
	default:
		return nil
	}
}

func (cc *ClientConnection) ReadMessage() ([]byte, error) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	_, msg, err := cc.Conn.ReadMessage()
	return msg, err
}

func (cc *ClientConnection) Close() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	close(cc.sendChan)
	cc.Conn.Close()
}

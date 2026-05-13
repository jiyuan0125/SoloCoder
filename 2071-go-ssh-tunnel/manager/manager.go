package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"ssh-tunnel-manager/database"
	"ssh-tunnel-manager/models"
	"ssh-tunnel-manager/tunnel"
)

type ManagedTunnel struct {
	ID            int64
	Config        *models.TunnelConfig
	Connection    *tunnel.SSHConnection
	State         *models.TunnelState
	CurrentLogID  int64
	ReconnectCount int
	ctx           context.Context
	cancel        context.CancelFunc
	mu            sync.Mutex
	stopped       bool
}

type Manager struct {
	db      *database.DB
	tunnels map[int64]*ManagedTunnel
	mu      sync.RWMutex
}

func New(db *database.DB) *Manager {
	return &Manager{
		db:      db,
		tunnels: make(map[int64]*ManagedTunnel),
	}
}

func (m *Manager) StartTunnel(configID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if mt, exists := m.tunnels[configID]; exists {
		mt.mu.Lock()
		if !mt.stopped {
			mt.mu.Unlock()
			return fmt.Errorf("tunnel already running")
		}
		mt.mu.Unlock()
	}

	config, err := m.db.GetTunnelConfig(configID)
	if err != nil {
		return fmt.Errorf("failed to get tunnel config: %w", err)
	}
	if config == nil {
		return fmt.Errorf("tunnel config not found")
	}

	inUse, _, err := tunnel.IsPortInUse(config.LocalPort)
	if err != nil {
		return fmt.Errorf("failed to check port: %w", err)
	}
	if inUse {
		return fmt.Errorf("port %d already in use", config.LocalPort)
	}

	ctx, cancel := context.WithCancel(context.Background())
	conn := tunnel.NewSSHConnection(config)

	mt := &ManagedTunnel{
		ID:         configID,
		Config:     config,
		Connection: conn,
		State: &models.TunnelState{
			TunnelID:       configID,
			Status:         models.TunnelStatusRunning,
			ReconnectCount: 0,
		},
		ctx:           ctx,
		cancel:        cancel,
		stopped:       false,
		ReconnectCount: 0,
	}

	m.tunnels[configID] = mt

	now := time.Now()
	log := &models.ConnectionLog{
		TunnelID:    configID,
		ConnectedAt: now,
	}
	logID, err := m.db.CreateConnectionLog(log)
	if err != nil {
		return fmt.Errorf("failed to create connection log: %w", err)
	}
	mt.CurrentLogID = logID

	mt.State.LastConnectedAt = &now
	if err := m.db.UpsertTunnelState(mt.State); err != nil {
		return fmt.Errorf("failed to update tunnel state: %w", err)
	}

	go m.runTunnel(mt)

	return nil
}

func (m *Manager) runTunnel(mt *ManagedTunnel) {
	reconnectDelay := 1 * time.Second
	maxDelay := 60 * time.Second

	for {
		mt.mu.Lock()
		if mt.stopped {
			mt.mu.Unlock()
			return
		}
		mt.mu.Unlock()

		err := mt.Connection.Start(mt.ctx)
		if err != nil {
			mt.mu.Lock()
			if mt.stopped {
				mt.mu.Unlock()
				return
			}
			mt.mu.Unlock()

			now := time.Now()
			disconnectReason := err.Error()

			if mt.CurrentLogID > 0 {
				m.db.UpdateConnectionLog(mt.CurrentLogID, now, disconnectReason)
			}

			mt.mu.Lock()
			mt.State.Status = models.TunnelStatusDisconnected
			mt.State.LastDisconnectedAt = &now
			mt.State.LastDisconnectReason = disconnectReason
			mt.ReconnectCount++
			mt.State.ReconnectCount = mt.ReconnectCount
			mt.mu.Unlock()

			m.db.UpsertTunnelState(mt.State)

			mt.mu.Lock()
			mt.State.Status = models.TunnelStatusReconnecting
			mt.mu.Unlock()
			m.db.UpsertTunnelState(mt.State)

			select {
			case <-mt.ctx.Done():
				return
			case <-time.After(reconnectDelay):
			}

			reconnectDelay *= 2
			if reconnectDelay > maxDelay {
				reconnectDelay = maxDelay
			}

			mt.mu.Lock()
			mt.Connection = tunnel.NewSSHConnection(mt.Config)
			mt.mu.Unlock()

			log := &models.ConnectionLog{
				TunnelID:    mt.ID,
				ConnectedAt: time.Now(),
			}
			logID, err := m.db.CreateConnectionLog(log)
			if err == nil {
				mt.mu.Lock()
				mt.CurrentLogID = logID
				now2 := time.Now()
				mt.State.LastConnectedAt = &now2
				mt.State.Status = models.TunnelStatusRunning
				mt.mu.Unlock()
				m.db.UpsertTunnelState(mt.State)
			}
		}
	}
}

func (m *Manager) StopTunnel(configID int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	mt, exists := m.tunnels[configID]
	if !exists {
		return false, fmt.Errorf("tunnel not found")
	}

	mt.mu.Lock()
	if mt.stopped {
		mt.mu.Unlock()
		return true, nil
	}
	mt.stopped = true
	mt.mu.Unlock()

	if mt.cancel != nil {
		mt.cancel()
	}

	if mt.Connection != nil {
		mt.Connection.Close()
	}

	now := time.Now()
	mt.mu.Lock()
	mt.State.Status = models.TunnelStatusStopped
	mt.State.LastDisconnectedAt = &now
	mt.State.LastDisconnectReason = "user stopped"
	mt.mu.Unlock()

	if mt.CurrentLogID > 0 {
		m.db.UpdateConnectionLog(mt.CurrentLogID, now, "user stopped")
	}

	m.db.UpsertTunnelState(mt.State)

	return false, nil
}

func (m *Manager) GetTunnelStatus(configID int64) (*models.TunnelState, *tunnel.Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	mt, exists := m.tunnels[configID]
	if !exists {
		state, err := m.db.GetTunnelState(configID)
		if err != nil {
			return nil, nil, err
		}
		if state == nil {
			return nil, nil, fmt.Errorf("tunnel not found")
		}
		return state, &tunnel.Stats{}, nil
	}

	mt.mu.Lock()
	state := *mt.State
	stats := *mt.Connection.GetStats()
	state.BytesUp = stats.BytesUp
	state.BytesDown = stats.BytesDown
	mt.mu.Unlock()

	return &state, &stats, nil
}

func (m *Manager) ListTunnels() ([]*models.TunnelConfig, error) {
	return m.db.ListTunnelConfigs()
}

func (m *Manager) GetConnectionLogs(tunnelID int64) ([]*models.ConnectionLog, error) {
	return m.db.ListConnectionLogs(tunnelID)
}

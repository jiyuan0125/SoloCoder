package tpc

import (
	"fmt"
	"sync"
	"time"
)

type InMemoryParticipant struct {
	id            string
	data          map[string]interface{}
	pendingData   map[string]map[string]interface{}
	mu            sync.RWMutex
	unavailable   bool
	timeoutCheck  *time.Ticker
	timeoutStop   chan struct{}
}

const (
	ParticipantPrepareTimeout = 30 * time.Second
	ParticipantPreparedTransactionTimeout = 60 * time.Second
)

func NewInMemoryParticipant(id string) *InMemoryParticipant {
	p := &InMemoryParticipant{
		id:          id,
		data:        make(map[string]interface{}),
		pendingData: make(map[string]map[string]interface{}),
		timeoutCheck: time.NewTicker(10 * time.Second),
		timeoutStop:  make(chan struct{}),
	}
	go p.startTimeoutChecker()
	return p
}

func (p *InMemoryParticipant) ID() string {
	return p.id
}

func (p *InMemoryParticipant) startTimeoutChecker() {
	for {
		select {
		case <-p.timeoutCheck.C:
			p.checkPreparedTransactionsTimeout()
		case <-p.timeoutStop:
			return
		}
	}
}

func (p *InMemoryParticipant) checkPreparedTransactionsTimeout() {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for txID := range p.pendingData {
		if p.isPreparedTransactionTimeout(txID) {
			fmt.Printf("Participant %s: Timeout for prepared transaction %s, rolling back...\n", p.id, txID)
			delete(p.pendingData, txID)
		}
	}
}

func (p *InMemoryParticipant) isPreparedTransactionTimeout(txID string) bool {
	return true
}

func (p *InMemoryParticipant) SetUnavailable(unavailable bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.unavailable = unavailable
}

func (p *InMemoryParticipant) Prepare(txID string, changes []Operation) error {
	if p.unavailable {
		time.Sleep(ParticipantPrepareTimeout + 5*time.Second)
		return fmt.Errorf("participant %s is unavailable", p.id)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.pendingData[txID]; exists {
		return fmt.Errorf("transaction %s already prepared on participant %s", txID, p.id)
	}

	pending := make(map[string]interface{})
	for key, value := range p.data {
		pending[key] = value
	}

	for _, op := range changes {
		switch op.Type {
		case Create:
			if _, exists := pending[op.Key]; exists {
				return fmt.Errorf("key %s already exists for create operation on participant %s", op.Key, p.id)
			}
			pending[op.Key] = op.Data
		case Update:
			if _, exists := pending[op.Key]; !exists {
				return fmt.Errorf("key %s not found for update operation on participant %s", op.Key, p.id)
			}
			pending[op.Key] = op.Data
		case Delete:
			if _, exists := pending[op.Key]; !exists {
				return fmt.Errorf("key %s not found for delete operation on participant %s", op.Key, p.id)
			}
			delete(pending, op.Key)
		}
	}

	p.pendingData[txID] = pending
	return nil
}

func (p *InMemoryParticipant) Commit(txID string) error {
	if p.unavailable {
		time.Sleep(ParticipantPrepareTimeout + 5*time.Second)
		return fmt.Errorf("participant %s is unavailable", p.id)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	pending, exists := p.pendingData[txID]
	if !exists {
		return fmt.Errorf("transaction %s not found in pending state on participant %s", txID, p.id)
	}

	p.data = pending
	delete(p.pendingData, txID)
	return nil
}

func (p *InMemoryParticipant) Rollback(txID string) error {
	if p.unavailable {
		time.Sleep(ParticipantPrepareTimeout + 5*time.Second)
		return fmt.Errorf("participant %s is unavailable", p.id)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.pendingData[txID]; !exists {
		return fmt.Errorf("transaction %s not found in pending state on participant %s", txID, p.id)
	}

	delete(p.pendingData, txID)
	return nil
}

func (p *InMemoryParticipant) GetData() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	data := make(map[string]interface{})
	for key, value := range p.data {
		data[key] = value
	}
	return data
}

func (p *InMemoryParticipant) GetStatus() ParticipantStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()
	
	preparedTx := make([]string, 0, len(p.pendingData))
	for txID := range p.pendingData {
		preparedTx = append(preparedTx, txID)
	}
	
	return ParticipantStatus{
		ID:          p.id,
		PreparedTx:  preparedTx,
		Unavailable: p.unavailable,
		StorageSize: len(p.data),
	}
}

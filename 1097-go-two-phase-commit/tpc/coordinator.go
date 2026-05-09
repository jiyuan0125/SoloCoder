package tpc

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type InMemoryCoordinator struct {
	transactions    map[string]*TransactionState
	participants    map[string]ITransactionParticipant
	transactionLogs map[string][]TransactionLogEntry
	mu              sync.RWMutex
}

const (
	CoordinatorPrepareTimeout = 10 * time.Second
	CoordinatorCommitTimeout  = 10 * time.Second
)

func NewInMemoryCoordinator() *InMemoryCoordinator {
	return &InMemoryCoordinator{
		transactions:    make(map[string]*TransactionState),
		participants:    make(map[string]ITransactionParticipant),
		transactionLogs: make(map[string][]TransactionLogEntry),
	}
}

func (c *InMemoryCoordinator) Begin() (*TransactionState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	txID := uuid.New().String()
	now := time.Now()

	txState := &TransactionState{
		ID:         txID,
		Status:     Pending,
		Operations: make(map[string][]Operation),
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	c.transactions[txID] = txState
	c.logTransaction(txID, Pending)

	return txState, nil
}

func (c *InMemoryCoordinator) Prepare(txID string, operations map[string][]Operation) (*TransactionState, error) {
	c.mu.Lock()
	txState, exists := c.transactions[txID]
	if !exists {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s not found", txID)
	}

	if txState.Status != Pending {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s is not in pending state", txID)
	}

	if len(operations) == 0 {
		txState.Status = Prepared
		txState.PreparedAt = time.Now()
		c.logTransaction(txID, Prepared)
		c.mu.Unlock()
		return txState, nil
	}

	for participantID := range operations {
		if _, exists := c.participants[participantID]; !exists {
			c.mu.Unlock()
			return nil, fmt.Errorf("participant %s not found", participantID)
		}
	}

	txState.Operations = operations
	participantsToPrepare := make([]ITransactionParticipant, 0, len(operations))
	for participantID := range operations {
		participantsToPrepare = append(participantsToPrepare, c.participants[participantID])
	}

	c.mu.Unlock()

	type prepareResult struct {
		participantID string
		err           error
	}

	resultsChan := make(chan prepareResult, len(participantsToPrepare))
	var wg sync.WaitGroup

	for _, participant := range participantsToPrepare {
		wg.Add(1)
		go func(p ITransactionParticipant) {
			defer wg.Done()
			err := p.Prepare(txID, operations[p.ID()])
			resultsChan <- prepareResult{participantID: p.ID(), err: err}
		}(participant)
	}

	timeoutChan := time.After(CoordinatorPrepareTimeout)
	results := make([]prepareResult, 0, len(participantsToPrepare))

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
		for i := 0; i < len(participantsToPrepare); i++ {
			results = append(results, <-resultsChan)
		}
	case <-timeoutChan:
		return c.abortTransaction(txID, "prepare timeout")
	}

	allSuccess := true
	for _, result := range results {
		if result.err != nil {
			allSuccess = false
			break
		}
	}

	if !allSuccess {
		return c.abortTransaction(txID, "some participants failed prepare")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	txState = c.transactions[txID]
	txState.Status = Prepared
	txState.PreparedAt = time.Now()
	txState.UpdatedAt = time.Now()
	c.logTransaction(txID, Prepared)

	return txState, nil
}

func (c *InMemoryCoordinator) Commit(txID string) (*TransactionState, error) {
	c.mu.Lock()
	txState, exists := c.transactions[txID]
	if !exists {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s not found", txID)
	}

	if txState.Status == Committed {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s already committed", txID)
	}

	if txState.Status == RolledBack || txState.Status == Aborted {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s already aborted or rolled back", txID)
	}

	if txState.Status != Prepared {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s is not in prepared state", txID)
	}

	if len(txState.Operations) == 0 {
		txState.Status = Committed
		txState.UpdatedAt = time.Now()
		c.logTransaction(txID, Committed)
		c.mu.Unlock()
		return txState, nil
	}

	participantsToCommit := make([]ITransactionParticipant, 0, len(txState.Operations))
	for participantID := range txState.Operations {
		if participant, exists := c.participants[participantID]; exists {
			participantsToCommit = append(participantsToCommit, participant)
		}
	}

	c.mu.Unlock()

	type commitResult struct {
		participantID string
		err           error
	}

	resultsChan := make(chan commitResult, len(participantsToCommit))
	var wg sync.WaitGroup

	for _, participant := range participantsToCommit {
		wg.Add(1)
		go func(p ITransactionParticipant) {
			defer wg.Done()
			err := p.Commit(txID)
			resultsChan <- commitResult{participantID: p.ID(), err: err}
		}(participant)
	}

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(CoordinatorCommitTimeout):
	}

	results := make([]commitResult, 0, len(participantsToCommit))
	timeout := false
	for i := 0; i < len(participantsToCommit); i++ {
		select {
		case result := <-resultsChan:
			results = append(results, result)
		case <-time.After(time.Second):
			timeout = true
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	txState = c.transactions[txID]
	txState.CommitResult = make([]ParticipantResult, len(results))
	for i, result := range results {
		txState.CommitResult[i] = ParticipantResult{
			ParticipantID: result.participantID,
			Success:       result.err == nil,
		}
		if result.err != nil {
			txState.CommitResult[i].Error = result.err.Error()
		}
	}

	if timeout {
		txState.CommitResult = append(txState.CommitResult, ParticipantResult{
			Success: false,
			Error:   "timeout",
		})
	}

	txState.Status = Committed
	txState.UpdatedAt = time.Now()
	c.logTransaction(txID, Committed)

	return txState, nil
}

func (c *InMemoryCoordinator) Rollback(txID string) (*TransactionState, error) {
	c.mu.Lock()
	txState, exists := c.transactions[txID]
	if !exists {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s not found", txID)
	}

	if txState.Status == RolledBack {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s already rolled back", txID)
	}

	if txState.Status == Committed {
		c.mu.Unlock()
		return nil, fmt.Errorf("transaction %s already committed", txID)
	}

	if txState.Status == Pending {
		txState.Status = Aborted
		txState.UpdatedAt = time.Now()
		c.logTransaction(txID, Aborted)
		c.mu.Unlock()
		return txState, nil
	}

	participantsToRollback := make([]ITransactionParticipant, 0, len(txState.Operations))
	for participantID := range txState.Operations {
		if participant, exists := c.participants[participantID]; exists {
			participantsToRollback = append(participantsToRollback, participant)
		}
	}

	c.mu.Unlock()

	type rollbackResult struct {
		participantID string
		err           error
	}

	resultsChan := make(chan rollbackResult, len(participantsToRollback))
	var wg sync.WaitGroup

	for _, participant := range participantsToRollback {
		wg.Add(1)
		go func(p ITransactionParticipant) {
			defer wg.Done()
			err := p.Rollback(txID)
			resultsChan <- rollbackResult{participantID: p.ID(), err: err}
		}(participant)
	}

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(CoordinatorCommitTimeout):
	}

	for i := 0; i < len(participantsToRollback); i++ {
		select {
		case <-resultsChan:
		case <-time.After(time.Second):
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	txState = c.transactions[txID]
	txState.Status = RolledBack
	txState.UpdatedAt = time.Now()
	c.logTransaction(txID, RolledBack)

	return txState, nil
}

func (c *InMemoryCoordinator) GetStatus(txID string) (*TransactionState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	txState, exists := c.transactions[txID]
	if !exists {
		return nil, fmt.Errorf("transaction %s not found", txID)
	}

	return txState, nil
}

func (c *InMemoryCoordinator) GetUnfinishedTransactions() []*TransactionState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	unfinished := make([]*TransactionState, 0)
	for _, tx := range c.transactions {
		if tx.Status != Committed && tx.Status != RolledBack {
			unfinished = append(unfinished, tx)
		}
	}

	return unfinished
}

func (c *InMemoryCoordinator) AddParticipant(participant ITransactionParticipant) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.participants[participant.ID()]; exists {
		return fmt.Errorf("participant %s already exists", participant.ID())
	}

	c.participants[participant.ID()] = participant
	return nil
}

func (c *InMemoryCoordinator) RemoveParticipant(participantID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.participants[participantID]; !exists {
		return fmt.Errorf("participant %s not found", participantID)
	}

	delete(c.participants, participantID)
	return nil
}

func (c *InMemoryCoordinator) GetParticipants() []ITransactionParticipant {
	c.mu.RLock()
	defer c.mu.RUnlock()

	participants := make([]ITransactionParticipant, 0, len(c.participants))
	for _, p := range c.participants {
		participants = append(participants, p)
	}

	return participants
}

func (c *InMemoryCoordinator) Recover() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	fmt.Println("Coordinator recovery started...")
	
	for txID, tx := range c.transactions {
		if tx.Status == Prepared {
			fmt.Printf("Recovering transaction %s (in prepared state)\n", txID)
			return c.recoverPrepared(txID)
		}
		if tx.Status == Pending {
			fmt.Printf("Aborting pending transaction %s\n", txID)
			c.transactions[txID].Status = Aborted
			c.transactions[txID].UpdatedAt = time.Now()
			c.logTransaction(txID, Aborted)
		}
	}

	fmt.Println("Coordinator recovery completed")
	return nil
}

func (c *InMemoryCoordinator) recoverPrepared(txID string) error {
	tx := c.transactions[txID]
	
	participantsToRollback := make([]ITransactionParticipant, 0, len(tx.Operations))
	for participantID := range tx.Operations {
		if participant, exists := c.participants[participantID]; exists {
			participantsToRollback = append(participantsToRollback, participant)
		}
	}

	for _, participant := range participantsToRollback {
		_ = participant.Rollback(txID)
	}

	c.transactions[txID].Status = RolledBack
	c.transactions[txID].UpdatedAt = time.Now()
	c.logTransaction(txID, RolledBack)
	
	return nil
}

func (c *InMemoryCoordinator) abortTransaction(txID string, reason string) (*TransactionState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	txState, exists := c.transactions[txID]
	if !exists {
		return nil, fmt.Errorf("transaction %s not found", txID)
	}

	txState.Status = Aborted
	txState.UpdatedAt = time.Now()
	c.logTransaction(txID, Aborted)

	for participantID := range txState.Operations {
		if participant, exists := c.participants[participantID]; exists {
			go participant.Rollback(txID)
		}
	}

	return txState, nil
}

func (c *InMemoryCoordinator) logTransaction(txID string, status TransactionStatus) {
	entry := TransactionLogEntry{
		TxID:      txID,
		Status:    status,
		Timestamp: time.Now(),
	}
	
	if _, exists := c.transactionLogs[txID]; !exists {
		c.transactionLogs[txID] = []TransactionLogEntry{entry}
	} else {
		c.transactionLogs[txID] = append(c.transactionLogs[txID], entry)
	}
}

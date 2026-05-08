package commitlog

import (
	"fmt"
	"sync"
	"time"
)

type CommitLog struct {
	dataDir    string
	config     Config
	topics     map[string]*Log
	groups     *GroupManager
	cleanupTck *time.Ticker
	stopCh     chan struct{}
	mu         sync.RWMutex
}

func New(dataDir string, config Config) (*CommitLog, error) {
	gm, err := NewGroupManager(dataDir + "/groups")
	if err != nil {
		return nil, err
	}

	cl := &CommitLog{
		dataDir: dataDir,
		config:  config,
		topics:  make(map[string]*Log),
		groups:  gm,
		stopCh:  make(chan struct{}),
	}

	if config.CleanupInterval > 0 {
		cl.cleanupTck = time.NewTicker(config.CleanupInterval)
		go cl.cleanupLoop()
	}

	return cl, nil
}

func (cl *CommitLog) getOrCreateTopic(topic string) (*Log, error) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	if log, exists := cl.topics[topic]; exists {
		return log, nil
	}

	log, err := NewTopicLog(cl.dataDir, topic, cl.config)
	if err != nil {
		return nil, err
	}

	cl.topics[topic] = log
	return log, nil
}

func (cl *CommitLog) Append(topic string, value []byte) (int64, error) {
	log, err := cl.getOrCreateTopic(topic)
	if err != nil {
		return 0, err
	}
	return log.Append(value)
}

func (cl *CommitLog) Fetch(topic string, offset int64, max int) ([]*record, int64, error) {
	cl.mu.RLock()
	log, exists := cl.topics[topic]
	cl.mu.RUnlock()

	if !exists {
		newLog, err := cl.getOrCreateTopic(topic)
		if err != nil {
			return nil, 0, err
		}
		log = newLog
	}

	return log.Fetch(offset, max)
}

func (cl *CommitLog) NextOffset(topic string) int64 {
	cl.mu.RLock()
	log, exists := cl.topics[topic]
	cl.mu.RUnlock()

	if !exists {
		return 0
	}
	return log.NextOffset()
}

func (cl *CommitLog) CreateGroup(groupID, topic string) error {
	_, err := cl.getOrCreateTopic(topic)
	if err != nil {
		return err
	}
	return cl.groups.CreateGroup(groupID, topic)
}

func (cl *CommitLog) JoinGroup(groupID, consumerID string) (int64, error) {
	return cl.groups.JoinGroup(groupID, consumerID, 0)
}

func (cl *CommitLog) CommitOffset(groupID, consumerID string, offset int64) error {
	return cl.groups.CommitOffset(groupID, consumerID, offset)
}

func (cl *CommitLog) GetOffset(groupID, consumerID string) (int64, error) {
	return cl.groups.GetOffset(groupID, consumerID)
}

func (cl *CommitLog) cleanupLoop() {
	for {
		select {
		case <-cl.cleanupTck.C:
			cl.runCleanup()
		case <-cl.stopCh:
			return
		}
	}
}

func (cl *CommitLog) runCleanup() {
	cl.mu.RLock()
	defer cl.mu.RUnlock()

	activeReaders := make(map[int64]struct{})

	for _, log := range cl.topics {
		log.CleanupRetention(activeReaders)
	}
}

func (cl *CommitLog) Close() error {
	select {
	case <-cl.stopCh:
	default:
		close(cl.stopCh)
	}

	if cl.cleanupTck != nil {
		cl.cleanupTck.Stop()
	}

	cl.mu.Lock()
	defer cl.mu.Unlock()

	var firstErr error
	for _, log := range cl.topics {
		if err := log.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if err := cl.groups.Close(); err != nil && firstErr == nil {
		firstErr = err
	}

	return firstErr
}

func RecordToMessage(topic string, rec *record) *record {
	return rec
}

type FetchResult struct {
	Records []*record
	Next    int64
}

func (r *record) Offset() int64    { return r.offset }
func (r *record) Value() []byte     { return r.value }
func (r *record) Timestamp() int64 { return r.timestamp }

func (cl *CommitLog) FetchWithGroup(topic, groupID, consumerID string, max int) (FetchResult, error) {
	offset, err := cl.GetOffset(groupID, consumerID)
	if err != nil {
		return FetchResult{}, fmt.Errorf("get offset: %w", err)
	}

	records, next, err := cl.Fetch(topic, offset, max)
	if err != nil {
		return FetchResult{}, fmt.Errorf("fetch: %w", err)
	}

	return FetchResult{
		Records: records,
		Next:    next,
	}, nil
}

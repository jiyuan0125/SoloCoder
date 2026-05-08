package commitlog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type consumerOffset struct {
	Offset     int64 `json:"offset"`
	LastCommit int64 `json:"last_commit"`
}

type groupState struct {
	GroupID   string                     `json:"group_id"`
	Topic     string                     `json:"topic"`
	Consumers map[string]consumerOffset `json:"consumers"`
	Version   int64                      `json:"version"`
}

type GroupManager struct {
	dir       string
	groups    map[string]*groupState
	mu        sync.RWMutex
	dirty     map[string]struct{}
	flushTick *time.Ticker
	stopCh    chan struct{}
}

func NewGroupManager(dir string) (*GroupManager, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	gm := &GroupManager{
		dir:    dir,
		groups: make(map[string]*groupState),
		dirty:  make(map[string]struct{}),
		stopCh: make(chan struct{}),
	}

	if err := gm.loadGroups(); err != nil {
		return nil, err
	}

	gm.flushTick = time.NewTicker(5 * time.Second)
	go gm.flushLoop()

	return gm, nil
}

func (gm *GroupManager) loadGroups() error {
	files, err := filepath.Glob(filepath.Join(gm.dir, "*.json"))
	if err != nil {
		return err
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var state groupState
		if err := json.Unmarshal(data, &state); err != nil {
			continue
		}

		gm.groups[state.GroupID] = &state
	}

	return nil
}

func (gm *GroupManager) CreateGroup(groupID, topic string) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	if _, exists := gm.groups[groupID]; exists {
		return nil
	}

	gm.groups[groupID] = &groupState{
		GroupID:   groupID,
		Topic:     topic,
		Consumers: make(map[string]consumerOffset),
		Version:   1,
	}

	gm.dirty[groupID] = struct{}{}
	return nil
}

func (gm *GroupManager) JoinGroup(groupID, consumerID string, startOffset int64) (int64, error) {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return 0, fmt.Errorf("group %s not found", groupID)
	}

	if co, exists := group.Consumers[consumerID]; exists {
		return co.Offset, nil
	}

	group.Consumers[consumerID] = consumerOffset{
		Offset:     startOffset,
		LastCommit: time.Now().UnixNano(),
	}
	group.Version++

	gm.dirty[groupID] = struct{}{}
	return startOffset, nil
}

func (gm *GroupManager) CommitOffset(groupID, consumerID string, offset int64) error {
	gm.mu.Lock()
	defer gm.mu.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	co, exists := group.Consumers[consumerID]
	if !exists {
		return fmt.Errorf("consumer %s not found in group %s", consumerID, groupID)
	}

	if offset < co.Offset {
		return fmt.Errorf("offset %d is less than current offset %d", offset, co.Offset)
	}

	co.Offset = offset
	co.LastCommit = time.Now().UnixNano()
	group.Consumers[consumerID] = co
	group.Version++

	gm.dirty[groupID] = struct{}{}
	return nil
}

func (gm *GroupManager) GetOffset(groupID, consumerID string) (int64, error) {
	gm.mu.RLock()
	defer gm.mu.RUnlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return 0, fmt.Errorf("group %s not found", groupID)
	}

	co, exists := group.Consumers[consumerID]
	if !exists {
		return 0, fmt.Errorf("consumer %s not found in group %s", consumerID, groupID)
	}

	return co.Offset, nil
}

func (gm *GroupManager) flushLoop() {
	for {
		select {
		case <-gm.flushTick.C:
			gm.flush()
		case <-gm.stopCh:
			return
		}
	}
}

func (gm *GroupManager) flush() {
	gm.mu.Lock()
	dirty := make(map[string]*groupState)
	for gid := range gm.dirty {
		if group, exists := gm.groups[gid]; exists {
			dirty[gid] = group
		}
	}
	gm.dirty = make(map[string]struct{})
	gm.mu.Unlock()

	for gid, group := range dirty {
		gm.writeGroup(gid, group)
	}
}

func (gm *GroupManager) writeGroup(groupID string, group *groupState) error {
	data, err := json.MarshalIndent(group, "", "  ")
	if err != nil {
		return err
	}

	tempPath := filepath.Join(gm.dir, fmt.Sprintf("%s.json.tmp", groupID))
	finalPath := filepath.Join(gm.dir, fmt.Sprintf("%s.json", groupID))

	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, finalPath)
}

func (gm *GroupManager) Close() error {
	select {
	case <-gm.stopCh:
	default:
		close(gm.stopCh)
	}

	if gm.flushTick != nil {
		gm.flushTick.Stop()
	}

	gm.flush()

	return nil
}

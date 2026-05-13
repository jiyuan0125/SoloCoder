package store

import (
	"bytes"
	"config-center/internal/model"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	MaxValueSize = 64 * 1024

	OperationCreate   = "create"
	OperationUpdate   = "update"
	OperationDelete   = "delete"
	OperationRollback = "rollback"

	WatchStatusActive   = "active"
	WatchStatusFailed   = "failed"
)

var (
	validKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9_.]+$`)
)

type ChangeEvent struct {
	Env      string
	Project  string
	Key      string
	OldValue string
	NewValue string
	Version  int
	Time     time.Time
}

type Store struct {
	mu              sync.RWMutex
	configs         map[string]map[string]map[string]*model.Config
	versionCounter   int
	versions        []*model.VersionSnapshot
	watches         map[string]*model.WatchRegistration
	watchCounter     int
	changeListeners []func(ChangeEvent)
}

func New() *Store {
	return &Store{
		configs:       make(map[string]map[string]map[string]*model.Config),
		versions:        make([]*model.VersionSnapshot, 0),
		versionCounter: 0,
		watches:         make(map[string]*model.WatchRegistration),
		watchCounter:   0,
		changeListeners: make([]func(ChangeEvent), 0),
	}
}

func (s *Store) AddChangeListener(listener func(ChangeEvent)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.changeListeners = append(s.changeListeners, listener)
}

func ValidateKey(key string) bool {
	if key == "" {
		return false
	}
	return validKeyPattern.MatchString(key)
}

func ValidateValue(value string) bool {
	return len(value) <= MaxValueSize
}

func (s *Store) ensureEnvProject(env, project string) {
	if _, ok := s.configs[env]; !ok {
		s.configs[env] = make(map[string]map[string]*model.Config)
	}
	if _, ok := s.configs[env][project]; !ok {
		s.configs[env][project] = make(map[string]*model.Config)
	}
}

func (s *Store) Get(env, project, key string) (*model.Config, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if envMap, ok := s.configs[env]; ok {
		if projMap, ok := envMap[project]; ok {
			if cfg, ok := projMap[key]; ok {
				return cfg, true
			}
		}
	}
	return nil, false
}

func (s *Store) List(env, project string) []*model.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*model.Config

	if env != "" && project != "" {
		if envMap, ok := s.configs[env]; ok {
			if projMap, ok := envMap[project]; ok {
				for _, cfg := range projMap {
					results = append(results, cfg)
				}
			}
		}
	} else if env != "" {
		if envMap, ok := s.configs[env]; ok {
			for _, projMap := range envMap {
				for _, cfg := range projMap {
					results = append(results, cfg)
				}
			}
		}
	} else if project != "" {
		for _, envMap := range s.configs {
			if projMap, ok := envMap[project]; ok {
				for _, cfg := range projMap {
					results = append(results, cfg)
				}
			}
		}
	} else {
		for _, envMap := range s.configs {
			for _, projMap := range envMap {
				for _, cfg := range projMap {
					results = append(results, cfg)
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Key < results[j].Key
	})

	return results
}

func (s *Store) Set(env, project, key, value string) (*model.Config, *ChangeEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureEnvProject(env, project)

	now := time.Now()
	var oldValue string
	var previousVersion int
	var operation string
	var event *ChangeEvent

	if existing, ok := s.configs[env][project][key]; ok {
		if existing.Value == value {
			return existing, nil, nil
		}
		oldValue = existing.Value
		previousVersion = existing.Version
		operation = OperationUpdate
		existing.Value = value
		existing.Version = s.nextVersion()
		existing.UpdatedAt = now
		s.addSnapshot(existing, operation, previousVersion)
		event = &ChangeEvent{
			Env:      env,
			Project:  project,
			Key:      key,
			OldValue: oldValue,
			NewValue: value,
			Version:  existing.Version,
			Time:     now,
		}
		s.notifyListeners(*event)
		return existing, event, nil
	} else {
		operation = OperationCreate
		newVersion := s.nextVersion()
		cfg := &model.Config{
			Env:       env,
			Project:   project,
			Key:       key,
			Value:     value,
			Version:   newVersion,
			CreatedAt: now,
			UpdatedAt: now,
		}
		s.configs[env][project][key] = cfg
		s.addSnapshot(cfg, operation, 0)
		event = &ChangeEvent{
			Env:      env,
			Project:  project,
			Key:      key,
			OldValue: "",
			NewValue: value,
			Version:  newVersion,
			Time:     now,
		}
		s.notifyListeners(*event)
		return cfg, event, nil
	}
}

func (s *Store) Delete(env, project, key string) (*ChangeEvent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if envMap, ok := s.configs[env]; ok {
		if projMap, ok := envMap[project]; ok {
			if existing, ok := projMap[key]; ok {
				oldValue := existing.Value
				previousVersion := existing.Version
				now := time.Now()
				delete(projMap, key)
				if len(projMap) == 0 {
					delete(envMap, project)
					if len(envMap) == 0 {
						delete(s.configs, env)
					}
				}
				s.addSnapshot(&model.Config{
					Env:     env,
					Project: project,
					Key:     key,
					Value:   "",
					Version: s.nextVersion(),
				}, OperationDelete, previousVersion)
				event := &ChangeEvent{
					Env:      env,
					Project:  project,
					Key:      key,
					OldValue: oldValue,
					NewValue: "",
					Version:  s.versionCounter,
					Time:     now,
				}
				s.notifyListeners(*event)
				return event, true
			}
		}
	}
	return nil, false
}

func (s *Store) nextVersion() int {
	s.versionCounter++
	return s.versionCounter
}

func (s *Store) addSnapshot(cfg *model.Config, operation string, previousVersion int) {
	s.versions = append(s.versions, &model.VersionSnapshot{
		Version:    cfg.Version,
		Env:      cfg.Env,
		Project:  cfg.Project,
		Key:      cfg.Key,
		Value:    cfg.Value,
		Operation: operation,
		PreviousVersion: previousVersion,
		CreatedAt:  time.Now(),
	})
}

func (s *Store) GetVersion(version int) (*model.VersionSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if version < 1 || version > len(s.versions) {
		return nil, false
	}
	for _, v := range s.versions {
		if v.Version == version {
			return v, true
		}
	}
	return nil, false
}

func (s *Store) ListVersions(env, project, key string) []*model.VersionSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []*model.VersionSnapshot
	for _, v := range s.versions {
		match := true
		if env != "" && v.Env != env {
			match = false
		}
		if match && project != "" && v.Project != project {
			match = false
		}
		if match && key != "" && v.Key != key {
			match = false
		}
		if match {
			results = append(results, v)
		}
	}
	return results
}

func (s *Store) Rollback(version int) (*model.Config, *ChangeEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var target *model.VersionSnapshot
	for _, v := range s.versions {
		if v.Version == version {
			target = v
			break
		}
	}
	if target == nil {
		return nil, nil, fmt.Errorf("version not found")
	}

	if target.Operation == OperationDelete {
		return nil, nil, fmt.Errorf("cannot rollback delete operation")
	}

	s.ensureEnvProject(target.Env, target.Project)

	var oldValue string
	var previousVersion int

	if existing, ok := s.configs[target.Env][target.Project][target.Key]; ok {
		if existing.Value == target.Value {
			return existing, nil, nil
		}
		oldValue = existing.Value
		previousVersion = existing.Version
		existing.Value = target.Value
		existing.Version = s.nextVersion()
		existing.UpdatedAt = time.Now()
		s.addSnapshot(existing, OperationRollback, previousVersion)
		event := &ChangeEvent{
			Env:      target.Env,
			Project:  target.Project,
			Key:      target.Key,
			OldValue: oldValue,
			NewValue: target.Value,
			Version:  existing.Version,
			Time:     time.Now(),
		}
		s.notifyListeners(*event)
		return existing, event, nil
	} else {
		newVersion := s.nextVersion()
		cfg := &model.Config{
			Env:       target.Env,
			Project:   target.Project,
			Key:       target.Key,
			Value:     target.Value,
			Version:   newVersion,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		s.configs[target.Env][target.Project][target.Key] = cfg
		s.addSnapshot(cfg, OperationRollback, 0)
		event := &ChangeEvent{
			Env:      target.Env,
			Project:  target.Project,
			Key:      target.Key,
			OldValue: "",
			NewValue: target.Value,
			Version:  newVersion,
			Time:     time.Now(),
		}
		s.notifyListeners(*event)
		return cfg, event, nil
	}
}

func (s *Store) Diff(version1, version2 int) []model.DiffItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if version1 > version2 {
		version1, version2 = version2, version1
	}

	v1Map := s.buildStateAtVersion(version1)
	v2Map := s.buildStateAtVersion(version2)

	var diffs []model.DiffItem

	allKeys := make(map[string]bool)
	for k := range v1Map {
		allKeys[k] = true
	}
	for k := range v2Map {
		allKeys[k] = true
	}

	for k := range allKeys {
		v1Val, ok1 := v1Map[k]
		v2Val, ok2 := v2Map[k]

		if !ok1 {
			diffs = append(diffs, model.DiffItem{
				Key:        k,
				OldValue:   "",
				NewValue:   v2Val,
				ChangeType: "added",
			})
		} else if !ok2 {
			diffs = append(diffs, model.DiffItem{
				Key:        k,
				OldValue:   v1Val,
				NewValue:   "",
				ChangeType: "removed",
			})
		} else if v1Val != v2Val {
			diffs = append(diffs, model.DiffItem{
				Key:        k,
				OldValue:   v1Val,
				NewValue:   v2Val,
				ChangeType: "modified",
			})
		}
	}

	sort.Slice(diffs, func(i, j int) bool {
		return diffs[i].Key < diffs[j].Key
	})

	return diffs
}

func (s *Store) buildStateAtVersion(targetVersion int) map[string]string {
	state := make(map[string]string)
	for _, v := range s.versions {
		if v.Version > targetVersion {
			continue
		}
		key := fmt.Sprintf("%s|%s|%s", v.Env, v.Project, v.Key)
		switch v.Operation {
		case OperationCreate, OperationUpdate, OperationRollback:
			state[key] = v.Value
		case OperationDelete:
			delete(state, key)
		}
	}
	return state
}

func (s *Store) notifyListeners(event ChangeEvent) {
	listeners := make([]func(ChangeEvent), len(s.changeListeners))
	copy(listeners, s.changeListeners)
	for _, l := range listeners {
		go l(event)
	}
}

func (s *Store) RegisterWatch(env, project, key, callback string, isKeyWatch bool) *model.WatchRegistration {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.watchCounter++
	id := fmt.Sprintf("w%d", s.watchCounter)

	watch := &model.WatchRegistration{
		ID:        id,
		Env:       env,
		Project:   project,
		Key:       key,
		Callback:  callback,
		IsKeyWatch: isKeyWatch,
		Status:    WatchStatusActive,
		ConsecutiveFailures: 0,
		CreatedAt: time.Now(),
	}
	s.watches[id] = watch
	return watch
}

func (s *Store) UnregisterWatch(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.watches[id]; ok {
		delete(s.watches, id)
		return true
	}
	return false
}

func (s *Store) GetWatch(id string) (*model.WatchRegistration, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	w, ok := s.watches[id]
	return w, ok
}

func (s *Store) ListWatches() []*model.WatchRegistration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*model.WatchRegistration, 0, len(s.watches))
	for _, w := range s.watches {
		result = append(result, w)
	}
	return result
}

func (s *Store) MarkWatchSuccess(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, ok := s.watches[id]; ok {
		w.ConsecutiveFailures = 0
		w.Status = WatchStatusActive
	}
}

func (s *Store) MarkWatchFailure(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if w, ok := s.watches[id]; ok {
		w.ConsecutiveFailures++
		if w.ConsecutiveFailures >= 3 {
			w.Status = WatchStatusFailed
			return true
		}
	}
	return false
}

func (s *Store) GetActiveWatchesForEvent(event ChangeEvent) []*model.WatchRegistration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var matches []*model.WatchRegistration
	for _, w := range s.watches {
		if w.Status != WatchStatusActive {
			continue
		}
		if w.Env != event.Env {
			continue
		}
		if w.Project != event.Project {
			continue
		}
		if w.IsKeyWatch {
			if w.Key == event.Key {
				matches = append(matches, w)
			}
		} else {
			matches = append(matches, w)
		}
	}
	return matches
}

func BuildKey(env, project, key string) string {
	return fmt.Sprintf("%s|%s|%s", env, project, key)
}

func ParseKey(composite string) (env, project, key string) {
	parts := strings.SplitN(composite, "|", 3)
	if len(parts) == 3 {
		return parts[0], parts[1], parts[2]
	}
	return "", "", ""
}

func CompareValues(a, b string) bool {
	return bytes.Equal([]byte(a), []byte(b))
}

package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type QueryHistory struct {
	ID        int       `json:"id"`
	Query     string    `json:"query"`
	Timestamp time.Time `json:"timestamp"`
}

type Shortcut struct {
	Name       string    `json:"name"`
	DBPath     string    `json:"db_path"`
	Table      string    `json:"table"`
	Fields     string    `json:"fields"`
	Conditions string    `json:"conditions"`
	Where      string    `json:"where"`
	GroupBy    string    `json:"group_by"`
	CreatedAt  time.Time `json:"created_at"`
}

type Storage struct {
	historyFile  string
	shortcutFile string
}

func NewStorage() (*Storage, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("无法获取用户目录: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "report-query")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("无法创建配置目录: %w", err)
	}

	return &Storage{
		historyFile:  filepath.Join(configDir, "history.json"),
		shortcutFile: filepath.Join(configDir, "shortcuts.json"),
	}, nil
}

func (s *Storage) loadHistory() ([]QueryHistory, error) {
	if _, err := os.Stat(s.historyFile); os.IsNotExist(err) {
		return []QueryHistory{}, nil
	}

	data, err := os.ReadFile(s.historyFile)
	if err != nil {
		return nil, err
	}

	var history []QueryHistory
	if err := json.Unmarshal(data, &history); err != nil {
		return []QueryHistory{}, nil
	}

	return history, nil
}

func (s *Storage) saveHistory(history []QueryHistory) error {
	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.historyFile, data, 0644)
}

func (s *Storage) AddHistory(query string) error {
	history, err := s.loadHistory()
	if err != nil {
		return err
	}

	newID := 1
	if len(history) > 0 {
		newID = history[len(history)-1].ID + 1
	}

	history = append(history, QueryHistory{
		ID:        newID,
		Query:     query,
		Timestamp: time.Now(),
	})

	if len(history) > 100 {
		history = history[len(history)-100:]
	}

	return s.saveHistory(history)
}

func (s *Storage) ListHistory() ([]QueryHistory, error) {
	return s.loadHistory()
}

func (s *Storage) ClearHistory() error {
	return s.saveHistory([]QueryHistory{})
}

func (s *Storage) loadShortcuts() (map[string]Shortcut, error) {
	if _, err := os.Stat(s.shortcutFile); os.IsNotExist(err) {
		return map[string]Shortcut{}, nil
	}

	data, err := os.ReadFile(s.shortcutFile)
	if err != nil {
		return nil, err
	}

	shortcuts := map[string]Shortcut{}
	if err := json.Unmarshal(data, &shortcuts); err != nil {
		return map[string]Shortcut{}, nil
	}

	return shortcuts, nil
}

func (s *Storage) saveShortcuts(shortcuts map[string]Shortcut) error {
	data, err := json.MarshalIndent(shortcuts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.shortcutFile, data, 0644)
}

func (s *Storage) AddShortcut(sc Shortcut) error {
	shortcuts, err := s.loadShortcuts()
	if err != nil {
		return err
	}

	if _, exists := shortcuts[sc.Name]; exists {
		return errors.New("快捷方式名已存在: " + sc.Name)
	}

	sc.CreatedAt = time.Now()
	shortcuts[sc.Name] = sc

	return s.saveShortcuts(shortcuts)
}

func (s *Storage) GetShortcut(name string) (*Shortcut, error) {
	shortcuts, err := s.loadShortcuts()
	if err != nil {
		return nil, err
	}

	sc, exists := shortcuts[name]
	if !exists {
		return nil, errors.New("快捷方式不存在: " + name)
	}

	return &sc, nil
}

func (s *Storage) ListShortcuts() ([]Shortcut, error) {
	shortcuts, err := s.loadShortcuts()
	if err != nil {
		return nil, err
	}

	list := make([]Shortcut, 0, len(shortcuts))
	for _, sc := range shortcuts {
		list = append(list, sc)
	}

	return list, nil
}

func (s *Storage) DeleteShortcut(name string) error {
	shortcuts, err := s.loadShortcuts()
	if err != nil {
		return err
	}

	if _, exists := shortcuts[name]; !exists {
		return errors.New("快捷方式不存在: " + name)
	}

	delete(shortcuts, name)
	return s.saveShortcuts(shortcuts)
}

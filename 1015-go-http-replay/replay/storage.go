package replay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"http-replay/api"
)

const DefaultSessionsDir = "./sessions"

type Storage struct {
	sessionsDir string
}

func NewStorage(sessionsDir string) *Storage {
	if sessionsDir == "" {
		sessionsDir = DefaultSessionsDir
	}
	return &Storage{sessionsDir: sessionsDir}
}

func (s *Storage) ensureDir() error {
	return os.MkdirAll(s.sessionsDir, 0755)
}

func (s *Storage) generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

func (s *Storage) sessionFilePath(sessionID string) string {
	return filepath.Join(s.sessionsDir, sessionID+".json")
}

func (s *Storage) SaveSession(data *api.RecordedSessionData) (string, error) {
	if err := s.ensureDir(); err != nil {
		return "", fmt.Errorf("create sessions dir: %w", err)
	}

	filePath := s.sessionFilePath(data.Meta.SessionID)

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal session data: %w", err)
	}

	if err := ioutil.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("write session file: %w", err)
	}

	return filePath, nil
}

func (s *Storage) LoadSession(sessionID string) (*api.RecordedSessionData, error) {
	filePath := s.sessionFilePath(sessionID)
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		return nil, fmt.Errorf("read session file: %w", err)
	}

	var session api.RecordedSessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("invalid session data: %w", err)
	}

	if session.Meta.SessionID == "" {
		return nil, fmt.Errorf("session missing meta information")
	}

	return &session, nil
}

func (s *Storage) SessionExists(sessionID string) bool {
	_, err := os.Stat(s.sessionFilePath(sessionID))
	return err == nil
}

func (s *Storage) ListSessions() ([]api.RecordSession, error) {
	if _, err := os.Stat(s.sessionsDir); os.IsNotExist(err) {
		return []api.RecordSession{}, nil
	}

	files, err := filepath.Glob(filepath.Join(s.sessionsDir, "session_*.json"))
	if err != nil {
		return nil, err
	}

	var sessions []api.RecordSession
	for _, f := range files {
		session, err := s.loadSessionFromPath(f)
		if err != nil {
			continue
		}
		sessions = append(sessions, session)
	}
	return sessions, nil
}

func (s *Storage) loadSessionFromPath(path string) (api.RecordSession, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return api.RecordSession{}, err
	}

	var session api.RecordedSessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return api.RecordSession{}, err
	}

	return api.RecordSession{
		SessionID:    session.Meta.SessionID,
		TargetURL:    session.Meta.TargetURL,
		StartTime:    session.Meta.StartTime,
		EndTime:      session.Meta.EndTime,
		RequestCount: session.Meta.RequestCount,
		Active:       false,
		FilePath:     path,
	}, nil
}

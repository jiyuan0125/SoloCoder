package artifact

import (
	"io"
	"os"
	"path/filepath"

	"go-cicd-pipeline/internal/store"
	"go-cicd-pipeline/internal/types"
	"go-cicd-pipeline/internal/utils"
)

type Manager struct {
	artifactDir string
	store       *store.Store
}

func New(artifactDir string, s *store.Store) (*Manager, error) {
	if err := os.MkdirAll(artifactDir, 0755); err != nil {
		return nil, err
	}
	return &Manager{artifactDir: artifactDir, store: s}, nil
}

func (m *Manager) Upload(executionID, taskResultID, name string, reader io.Reader) (*types.Artifact, error) {
	dir := filepath.Join(m.artifactDir, executionID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dir, name)
	f, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	size, err := io.Copy(f, reader)
	if err != nil {
		return nil, err
	}

	artifact := &types.Artifact{
		ID:           utils.GenerateID(),
		ExecutionID:  executionID,
		TaskResultID: taskResultID,
		Name:         name,
		FilePath:     filePath,
		Size:         size,
		CreatedAt:    utils.Now(),
	}

	if err := m.store.CreateArtifact(artifact); err != nil {
		os.Remove(filePath)
		return nil, err
	}

	return artifact, nil
}

func (m *Manager) Download(artifactID string) (io.ReadCloser, *types.Artifact, error) {
	artifact, err := m.store.GetArtifact(artifactID)
	if err != nil {
		return nil, nil, err
	}
	if artifact == nil {
		return nil, nil, nil
	}

	f, err := os.Open(artifact.FilePath)
	if err != nil {
		return nil, nil, err
	}

	return f, artifact, nil
}

func (m *Manager) List(executionID string) ([]*types.Artifact, error) {
	return m.store.ListArtifacts(executionID)
}

func (m *Manager) Get(artifactID string) (*types.Artifact, error) {
	return m.store.GetArtifact(artifactID)
}

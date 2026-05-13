package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"contract-lifecycle/models"
)

const (
	contractsFile = "contracts.json"
	logsFile      = "operations.log"
	resourcesFile  = "resources.json"
)

type Storage struct {
	dataDir string
}

func NewStorage(dataDir string) *Storage {
	return &Storage{dataDir: dataDir}
}

func (s *Storage) ensureDir() error {
	return os.MkdirAll(s.dataDir, 0755)
}

func (s *Storage) contractsPath() string {
	return filepath.Join(s.dataDir, contractsFile)
}

func (s *Storage) logsPath() string {
	return filepath.Join(s.dataDir, logsFile)
}

func (s *Storage) resourcesPath() string {
	return filepath.Join(s.dataDir, resourcesFile)
}

func (s *Storage) LoadContracts() ([]models.Contract, error) {
	if err := s.ensureDir(); err != nil {
		return nil, err
	}

	path := s.contractsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.Contract{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var contracts []models.Contract
	if len(data) == 0 {
		return []models.Contract{}, nil
	}

	if err := json.Unmarshal(data, &contracts); err != nil {
		return nil, err
	}

	return contracts, nil
}

func (s *Storage) SaveContracts(contracts []models.Contract) error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(contracts, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.contractsPath(), data, 0644)
}

func (s *Storage) GetContract(id string) (*models.Contract, error) {
	contracts, err := s.LoadContracts()
	if err != nil {
		return nil, err
	}

	for i := range contracts {
		if contracts[i].ID == id {
			return &contracts[i], nil
		}
	}

	return nil, fmt.Errorf("contract not found: %s", id)
}

func (s *Storage) SaveContract(contract *models.Contract) error {
	contracts, err := s.LoadContracts()
	if err != nil {
		return err
	}

	found := false
	for i := range contracts {
		if contracts[i].ID == contract.ID {
			contracts[i] = *contract
			found = true
			break
		}
	}

	if !found {
		contracts = append(contracts, *contract)
	}

	return s.SaveContracts(contracts)
}

func (s *Storage) AddLog(log models.OperationLog) error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	line, err := json.Marshal(log)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.logsPath(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(string(line) + "\n")
	return err
}

func (s *Storage) LoadLogs() ([]models.OperationLog, error) {
	if err := s.ensureDir(); err != nil {
		return nil, err
	}

	path := s.logsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.OperationLog{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var logs []models.OperationLog
	if len(data) == 0 {
		return logs, nil
	}

	lines := splitLines(data)
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		var log models.OperationLog
		if err := json.Unmarshal([]byte(line), &log); err != nil {
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}

func (s *Storage) LoadResources() ([]models.Resource, error) {
	if err := s.ensureDir(); err != nil {
		return nil, err
	}

	path := s.resourcesPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []models.Resource{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var resources []models.Resource
	if len(data) == 0 {
		return []models.Resource{}, nil
	}

	if err := json.Unmarshal(data, &resources); err != nil {
		return nil, err
	}

	return resources, nil
}

func (s *Storage) SaveResources(resources []models.Resource) error {
	if err := s.ensureDir(); err != nil {
		return err
	}

	data, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.resourcesPath(), data, 0644)
}

func (s *Storage) SaveResource(resource *models.Resource) error {
	resources, err := s.LoadResources()
	if err != nil {
		return err
	}

	found := false
	for i := range resources {
		if resources[i].ID == resource.ID {
			resources[i] = *resource
			found = true
			break
		}
	}

	if !found {
		resources = append(resources, *resource)
	}

	return s.SaveResources(resources)
}

func (s *Storage) GetResource(id string) (*models.Resource, error) {
	resources, err := s.LoadResources()
	if err != nil {
		return nil, err
	}

	for i := range resources {
		if resources[i].ID == id {
			return &resources[i], nil
		}
	}

	return nil, fmt.Errorf("resource not found: %s", id)
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			lines = append(lines, string(data[start:i]))
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}

func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

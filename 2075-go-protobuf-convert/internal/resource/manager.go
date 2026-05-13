package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ResourceType string

const (
	TypeProtoFile     ResourceType = "proto_file"
	TypeProtoBinary   ResourceType = "proto_binary"
	TypeJSONOutput    ResourceType = "json_output"
	TypeJSONInput     ResourceType = "json_input"
)

type RelationType string

const (
	RelationDefines     RelationType = "defines"
	RelationGenerated   RelationType = "generated_from"
	RelationConverted   RelationType = "converted_to"
	RelationPartOf      RelationType = "part_of"
)

type Resource struct {
	ID         string                 `json:"id"`
	Type       ResourceType           `json:"type"`
	Name       string                 `json:"name"`
	Path       string                 `json:"path,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

type Relation struct {
	ID         string       `json:"id"`
	SourceID   string       `json:"source_id"`
	TargetID   string       `json:"target_id"`
	Type       RelationType `json:"type"`
	CreatedAt  time.Time    `json:"created_at"`
}

type Operation struct {
	ID           string      `json:"id"`
	ResourceID   string      `json:"resource_id"`
	Type         string      `json:"type"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	Status       string      `json:"status"`
	ErrorMessage string      `json:"error_message,omitempty"`
}

type ResourceManager struct {
	storagePath string
	mu          sync.RWMutex
}

type storageData struct {
	Resources  map[string]Resource  `json:"resources"`
	Relations  map[string]Relation  `json:"relations"`
	Operations map[string]Operation `json:"operations"`
}

func NewManager(storagePath string) (*ResourceManager, error) {
	if storagePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		storagePath = filepath.Join(home, ".protoconv", "resources.json")
	}

	dir := filepath.Dir(storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	rm := &ResourceManager{
		storagePath: storagePath,
	}

	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		if err := rm.save(&storageData{
			Resources:  make(map[string]Resource),
			Relations:  make(map[string]Relation),
			Operations: make(map[string]Operation),
		}); err != nil {
			return nil, err
		}
	}

	return rm, nil
}

func (rm *ResourceManager) load() (*storageData, error) {
	data, err := os.ReadFile(rm.storagePath)
	if err != nil {
		return nil, err
	}

	var sd storageData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, err
	}

	if sd.Resources == nil {
		sd.Resources = make(map[string]Resource)
	}
	if sd.Relations == nil {
		sd.Relations = make(map[string]Relation)
	}
	if sd.Operations == nil {
		sd.Operations = make(map[string]Operation)
	}

	return &sd, nil
}

func (rm *ResourceManager) save(sd *storageData) error {
	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(rm.storagePath, data, 0644)
}

func (rm *ResourceManager) AddResource(res Resource) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	sd, err := rm.load()
	if err != nil {
		return err
	}

	if res.ID == "" {
		res.ID = generateID()
	}
	res.CreatedAt = time.Now()
	res.UpdatedAt = time.Now()

	sd.Resources[res.ID] = res
	return rm.save(sd)
}

func (rm *ResourceManager) GetResource(id string) (*Resource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	sd, err := rm.load()
	if err != nil {
		return nil, err
	}

	res, ok := sd.Resources[id]
	if !ok {
		return nil, errors.New("resource not found: " + id)
	}

	return &res, nil
}

func (rm *ResourceManager) ListResources(filterType ResourceType) []Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	sd, err := rm.load()
	if err != nil {
		return nil
	}

	result := make([]Resource, 0)
	for _, res := range sd.Resources {
		if filterType == "" || res.Type == filterType {
			result = append(result, res)
		}
	}
	return result
}

func (rm *ResourceManager) AddRelation(rel Relation) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	sd, err := rm.load()
	if err != nil {
		return err
	}

	if _, ok := sd.Resources[rel.SourceID]; !ok {
		return errors.New("source resource not found: " + rel.SourceID)
	}
	if _, ok := sd.Resources[rel.TargetID]; !ok {
		return errors.New("target resource not found: " + rel.TargetID)
	}

	if rel.ID == "" {
		rel.ID = generateID()
	}
	rel.CreatedAt = time.Now()

	sd.Relations[rel.ID] = rel
	return rm.save(sd)
}

func (rm *ResourceManager) GetResourceRelations(resourceID string) []Relation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	sd, err := rm.load()
	if err != nil {
		return nil
	}

	result := make([]Relation, 0)
	for _, rel := range sd.Relations {
		if rel.SourceID == resourceID || rel.TargetID == resourceID {
			result = append(result, rel)
		}
	}
	return result
}

func (rm *ResourceManager) RecordOperation(op Operation) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	sd, err := rm.load()
	if err != nil {
		return err
	}

	if op.ID == "" {
		op.ID = generateID()
	}

	sd.Operations[op.ID] = op
	return rm.save(sd)
}

func (rm *ResourceManager) GetResourceOperations(resourceID string) []Operation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	sd, err := rm.load()
	if err != nil {
		return nil
	}

	result := make([]Operation, 0)
	for _, op := range sd.Operations {
		if op.ResourceID == resourceID {
			result = append(result, op)
		}
	}
	return result
}

type ResourceSummary struct {
	Resource      Resource
	RelationCount int
	Operations    struct {
		Total     int
		Success   int
		Failed    int
		LastOp    *Operation
	}
}

func (rm *ResourceManager) GetSummary(resourceID string) (*ResourceSummary, error) {
	res, err := rm.GetResource(resourceID)
	if err != nil {
		return nil, err
	}

	rm.mu.RLock()
	defer rm.mu.RUnlock()

	sd, err := rm.load()
	if err != nil {
		return nil, err
	}

	summary := &ResourceSummary{
		Resource: *res,
	}

	for _, rel := range sd.Relations {
		if rel.SourceID == resourceID || rel.TargetID == resourceID {
			summary.RelationCount++
		}
	}

	var lastOp *Operation
	for _, op := range sd.Operations {
		if op.ResourceID == resourceID {
			summary.Operations.Total++
			if op.Status == "success" {
				summary.Operations.Success++
			} else if op.Status == "failed" {
				summary.Operations.Failed++
			}
			if lastOp == nil || op.StartTime.After(lastOp.StartTime) {
				opCopy := op
				lastOp = &opCopy
			}
		}
	}
	summary.Operations.LastOp = lastOp

	return summary, nil
}

func (rm *ResourceManager) ListAllSummaries() []ResourceSummary {
	resources := rm.ListResources("")
	summaries := make([]ResourceSummary, 0, len(resources))

	for _, res := range resources {
		summary, err := rm.GetSummary(res.ID)
		if err == nil {
			summaries = append(summaries, *summary)
		}
	}
	return summaries
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

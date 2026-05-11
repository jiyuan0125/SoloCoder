package core

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"energymanagement/common"
)

type Service struct {
	mu          sync.RWMutex
	points      map[string]*common.Point
	data        map[string][]*common.EnergyData
	lastReading map[string]float64
	suggestions map[string][]*common.Suggestion
}

func NewService() *Service {
	return &Service{
		points:      make(map[string]*common.Point),
		data:        make(map[string][]*common.EnergyData),
		lastReading: make(map[string]float64),
		suggestions: make(map[string][]*common.Suggestion),
	}
}

func (s *Service) CreatePoint(req common.CreatePointRequest) (*common.Point, error) {
	if req.Name == "" || req.Area == "" || req.Unit == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID(req.Name + req.Area + string(req.EnergyType))
	if _, exists := s.points[id]; exists {
		return nil, ErrPointExists
	}

	point := &common.Point{
		ID:         id,
		Name:       req.Name,
		Area:       req.Area,
		EnergyType: req.EnergyType,
		Unit:       req.Unit,
		AreaSize:   req.AreaSize,
	}

	s.points[id] = point
	s.data[id] = make([]*common.EnergyData, 0)
	s.lastReading[id] = 0

	return point, nil
}

func (s *Service) ListPoints() []*common.Point {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*common.Point, 0, len(s.points))
	for _, p := range s.points {
		result = append(result, p)
	}
	return result
}

func (s *Service) ReportData(req common.ReportDataRequest) (*common.EnergyData, error) {
	if req.PointID == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	point, exists := s.points[req.PointID]
	if !exists {
		return nil, ErrPointNotFound
	}

	lastReading := s.lastReading[req.PointID]
	var increment float64

	if req.Reading >= lastReading {
		increment = req.Reading - lastReading
	} else {
		increment = 0
	}

	data := &common.EnergyData{
		ID:        generateID(req.PointID + req.Timestamp.String()),
		PointID:   req.PointID,
		Timestamp: req.Timestamp,
		Reading:   req.Reading,
		Increment: increment,
	}

	s.data[req.PointID] = append(s.data[req.PointID], data)
	s.lastReading[req.PointID] = req.Reading

	_ = point
	s.checkAndGenerateSuggestions(req.PointID, req.Timestamp)

	return data, nil
}

func (s *Service) GetPointData(pointID string, start, end time.Time) []*common.EnergyData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pointData, exists := s.data[pointID]
	if !exists {
		return []*common.EnergyData{}
	}

	result := make([]*common.EnergyData, 0)
	for _, d := range pointData {
		if (d.Timestamp.After(start) || d.Timestamp.Equal(start)) &&
			(d.Timestamp.Before(end) || d.Timestamp.Equal(end)) {
			result = append(result, d)
		}
	}
	return result
}

func generateID(input string) string {
	hash := sha256.Sum256([]byte(input + time.Now().String()))
	return hex.EncodeToString(hash[:])[:16]
}

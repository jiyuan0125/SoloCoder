package storage

import (
	"sync"

	"skill-cert/models"
)

type Storage struct {
	mu            sync.RWMutex
	occupations   map[string]*models.Occupation
	batches       map[string]*models.ExamBatch
	certificates  map[string]*models.Certificate
	certificatesByIdCard map[string][]*models.Certificate
	certSerialCounter map[string]int
}

func NewStorage() *Storage {
	return &Storage{
		occupations:   make(map[string]*models.Occupation),
		batches:       make(map[string]*models.ExamBatch),
		certificates:  make(map[string]*models.Certificate),
		certificatesByIdCard: make(map[string][]*models.Certificate),
		certSerialCounter: make(map[string]int),
	}
}

func (s *Storage) SaveOccupation(occ *models.Occupation) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.occupations[occ.ID] = occ
}

func (s *Storage) GetOccupation(id string) *models.Occupation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.occupations[id]
}

func (s *Storage) GetOccupations() []*models.Occupation {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.Occupation, 0, len(s.occupations))
	for _, occ := range s.occupations {
		result = append(result, occ)
	}
	return result
}

func (s *Storage) DeleteOccupation(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.occupations, id)
}

func (s *Storage) SaveBatch(batch *models.ExamBatch) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.batches[batch.ID] = batch
}

func (s *Storage) GetBatch(id string) *models.ExamBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.batches[id]
}

func (s *Storage) GetBatches() []*models.ExamBatch {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.ExamBatch, 0, len(s.batches))
	for _, batch := range s.batches {
		result = append(result, batch)
	}
	return result
}

func (s *Storage) DeleteBatch(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.batches, id)
}

func (s *Storage) SaveCertificate(cert *models.Certificate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.certificates[cert.ID] = cert
	s.certificatesByIdCard[cert.IDCard] = append(s.certificatesByIdCard[cert.IDCard], cert)
}

func (s *Storage) GetCertificate(id string) *models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.certificates[id]
}

func (s *Storage) GetCertificates() []*models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.Certificate, 0, len(s.certificates))
	for _, cert := range s.certificates {
		result = append(result, cert)
	}
	return result
}

func (s *Storage) GetCertificatesByIdCard(idCard string) []*models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.certificatesByIdCard[idCard]
}

func (s *Storage) GetNextSerialNumber(key string) (int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.certSerialCounter[key]
	if current >= 99999 {
		return 0, false
	}
	current++
	s.certSerialCounter[key] = current
	return current, true
}

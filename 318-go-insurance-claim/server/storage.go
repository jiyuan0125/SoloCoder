package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"insurance-claim/common"
)

type Storage struct {
	mu           sync.RWMutex
	writeMu      sync.Mutex
	policies     map[string]*common.Policy
	claims       map[string]*common.Claim
	policyClaims map[string][]*common.Claim
	dataFile     string
	dirty        bool
	writeChan    chan struct{}
	stopChan     chan struct{}
	once         sync.Once
}

func NewStorage(dataFile string) *Storage {
	s := &Storage{
		policies:     make(map[string]*common.Policy),
		claims:       make(map[string]*common.Claim),
		policyClaims: make(map[string][]*common.Claim),
		dataFile:     dataFile,
		writeChan:    make(chan struct{}, 1),
		stopChan:     make(chan struct{}),
	}
	s.loadFromFile()
	go s.writeLoop()
	return s
}

func (s *Storage) Close() {
	s.once.Do(func() {
		close(s.stopChan)
	})
}

func (s *Storage) writeLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.writeChan:
			s.trySave()
		case <-ticker.C:
			s.trySave()
		case <-s.stopChan:
			s.saveToFile()
			return
		}
	}
}

func (s *Storage) trySave() {
	s.mu.RLock()
	if !s.dirty {
		s.mu.RUnlock()
		return
	}
	s.mu.RUnlock()
	s.saveToFile()
}

func (s *Storage) markDirty() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()

	select {
	case s.writeChan <- struct{}{}:
	default:
	}
}

func (s *Storage) loadFromFile() {
	if _, err := os.Stat(s.dataFile); os.IsNotExist(err) {
		return
	}

	file, err := os.Open(s.dataFile)
	if err != nil {
		fmt.Printf("Failed to open data file: %v\n", err)
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Failed to read data file: %v\n", err)
		return
	}

	var storageData struct {
		Policies []*common.Policy `json:"policies"`
		Claims   []*common.Claim  `json:"claims"`
	}

	if err := json.Unmarshal(data, &storageData); err != nil {
		fmt.Printf("Failed to unmarshal data: %v\n", err)
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, p := range storageData.Policies {
		s.policies[p.PolicyNumber] = p
	}

	for _, c := range storageData.Claims {
		s.claims[c.ClaimID] = c
		s.policyClaims[c.PolicyNumber] = append(s.policyClaims[c.PolicyNumber], c)
	}
}

func (s *Storage) saveToFile() {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	s.mu.RLock()
	if !s.dirty {
		s.mu.RUnlock()
		return
	}

	policies := make([]*common.Policy, 0, len(s.policies))
	for _, p := range s.policies {
		policies = append(policies, p)
	}

	claims := make([]*common.Claim, 0, len(s.claims))
	for _, c := range s.claims {
		claims = append(claims, c)
	}
	s.mu.RUnlock()

	storageData := struct {
		Policies []*common.Policy `json:"policies"`
		Claims   []*common.Claim  `json:"claims"`
	}{
		Policies: policies,
		Claims:   claims,
	}

	data, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal data: %v\n", err)
		return
	}

	tempFile := s.dataFile + ".tmp"
	file, err := os.Create(tempFile)
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}

	if _, err := file.Write(data); err != nil {
		fmt.Printf("Failed to write temp file: %v\n", err)
		file.Close()
		os.Remove(tempFile)
		return
	}

	file.Close()

	if err := os.Rename(tempFile, s.dataFile); err != nil {
		fmt.Printf("Failed to rename temp file: %v\n", err)
		os.Remove(tempFile)
		return
	}

	s.mu.Lock()
	s.dirty = false
	s.mu.Unlock()
}

func (s *Storage) CreatePolicy(policy *common.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.PolicyNumber]; exists {
		return fmt.Errorf("policy with number %s already exists", policy.PolicyNumber)
	}

	s.policies[policy.PolicyNumber] = policy
	s.markDirty()
	return nil
}

func (s *Storage) GetPolicy(policyNumber string) (*common.Policy, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, exists := s.policies[policyNumber]
	if !exists {
		return nil, fmt.Errorf("policy with number %s not found", policyNumber)
	}
	return policy, nil
}

func (s *Storage) UpdatePolicy(policy *common.Policy) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.policies[policy.PolicyNumber]; !exists {
		return fmt.Errorf("policy with number %s not found", policy.PolicyNumber)
	}

	s.policies[policy.PolicyNumber] = policy
	s.markDirty()
	return nil
}

func (s *Storage) CreateClaim(claim *common.Claim) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.claims[claim.ClaimID]; exists {
		return fmt.Errorf("claim with ID %s already exists", claim.ClaimID)
	}

	s.claims[claim.ClaimID] = claim
	s.policyClaims[claim.PolicyNumber] = append(s.policyClaims[claim.PolicyNumber], claim)
	s.markDirty()
	return nil
}

func (s *Storage) GetClaim(claimID string) (*common.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claim, exists := s.claims[claimID]
	if !exists {
		return nil, fmt.Errorf("claim with ID %s not found", claimID)
	}
	return claim, nil
}

func (s *Storage) UpdateClaim(claim *common.Claim) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.claims[claim.ClaimID]; !exists {
		return fmt.Errorf("claim with ID %s not found", claim.ClaimID)
	}

	s.claims[claim.ClaimID] = claim
	s.markDirty()
	return nil
}

func (s *Storage) GetPendingClaims() []*common.Claim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pendingClaims []*common.Claim
	for _, claim := range s.claims {
		if claim.Status == common.ClaimStatusPending {
			pendingClaims = append(pendingClaims, claim)
		}
	}
	return pendingClaims
}

func (s *Storage) GetPolicyClaims(policyNumber string) ([]*common.Claim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.policies[policyNumber]; !exists {
		return nil, fmt.Errorf("policy with number %s not found", policyNumber)
	}

	return s.policyClaims[policyNumber], nil
}

func (s *Storage) HasActiveClaim(policyNumber string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, claim := range s.policyClaims[policyNumber] {
		if claim.Status == common.ClaimStatusPending || claim.Status == common.ClaimStatusApproved {
			return true
		}
	}
	return false
}

func (s *Storage) ListAllPolicies() []*common.Policy {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policies := make([]*common.Policy, 0, len(s.policies))
	for _, p := range s.policies {
		policies = append(policies, p)
	}
	return policies
}

func (s *Storage) ListAllClaims() []*common.Claim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	claims := make([]*common.Claim, 0, len(s.claims))
	for _, c := range s.claims {
		claims = append(claims, c)
	}
	return claims
}

func (s *Storage) GetApprovedClaims() []*common.Claim {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var approvedClaims []*common.Claim
	for _, claim := range s.claims {
		if claim.Status == common.ClaimStatusApproved {
			approvedClaims = append(approvedClaims, claim)
		}
	}
	return approvedClaims
}

func generateID() string {
	return fmt.Sprintf("CLM%d", time.Now().UnixNano())
}

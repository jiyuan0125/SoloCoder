package housekeeping

import (
	"fmt"
	"sync"
)

type AuntManager struct {
	aunts map[string]*Aunt
	mu    sync.RWMutex
}

func NewAuntManager() *AuntManager {
	return &AuntManager{
		aunts: make(map[string]*Aunt),
	}
}

func (am *AuntManager) RegisterAunt(aunt *Aunt) error {
	if aunt.ID == "" {
		return fmt.Errorf("阿姨ID不能为空")
	}
	if aunt.Name == "" {
		return fmt.Errorf("阿姨姓名不能为空")
	}
	if aunt.Phone == "" {
		return fmt.Errorf("阿姨手机号不能为空")
	}
	if aunt.ServiceCategory == "" {
		return fmt.Errorf("服务类别不能为空")
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.aunts[aunt.ID]; exists {
		return fmt.Errorf("阿姨ID已存在")
	}

	if aunt.Rating == 0 {
		aunt.Rating = 5.0
	}
	if aunt.Reviews == nil {
		aunt.Reviews = make([]Review, 0)
	}

	am.aunts[aunt.ID] = aunt
	return nil
}

func (am *AuntManager) GetAunt(id string) (*Aunt, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	aunt, exists := am.aunts[id]
	if !exists {
		return nil, fmt.Errorf("阿姨不存在")
	}
	return aunt, nil
}

func (am *AuntManager) GetAllAunts() []*Aunt {
	am.mu.RLock()
	defer am.mu.RUnlock()

	aunts := make([]*Aunt, 0, len(am.aunts))
	for _, aunt := range am.aunts {
		aunts = append(aunts, aunt)
	}
	return aunts
}

func (am *AuntManager) FindAuntsByCategory(category ServiceCategory) []*Aunt {
	am.mu.RLock()
	defer am.mu.RUnlock()

	var aunts []*Aunt
	for _, aunt := range am.aunts {
		if aunt.ServiceCategory == category {
			aunts = append(aunts, aunt)
		}
	}
	return aunts
}

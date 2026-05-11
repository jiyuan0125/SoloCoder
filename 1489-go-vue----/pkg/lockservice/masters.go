package lockservice

import (
	"sync"
	"time"
)

type MasterManager struct {
	mu      sync.RWMutex
	masters map[string]*Master
}

func NewMasterManager() *MasterManager {
	return &MasterManager{
		masters: make(map[string]*Master),
	}
}

func (mm *MasterManager) AddMaster(master *Master) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if master.BookedSlots == nil {
		master.BookedSlots = make(map[string]map[TimeSlot]bool)
	}
	mm.masters[master.ID] = master
}

func (mm *MasterManager) GetMaster(id string) *Master {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	return mm.masters[id]
}

func (mm *MasterManager) ListMasters() []*Master {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	result := make([]*Master, 0, len(mm.masters))
	for _, m := range mm.masters {
		result = append(result, m)
	}
	return result
}

func (mm *MasterManager) UpdateStatus(id string, status MasterStatus) bool {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if m, exists := mm.masters[id]; exists {
		m.Status = status
		return true
	}
	return false
}

func (mm *MasterManager) IsSlotAvailable(masterID string, date time.Time, slot TimeSlot) bool {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	master, exists := mm.masters[masterID]
	if !exists {
		return false
	}
	dateKey := date.Format("2006-01-02")
	daySlots, booked := master.BookedSlots[dateKey]
	if !booked {
		return true
	}
	return !daySlots[slot]
}

func (mm *MasterManager) BookSlot(masterID string, date time.Time, slot TimeSlot) bool {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	master, exists := mm.masters[masterID]
	if !exists {
		return false
	}
	dateKey := date.Format("2006-01-02")
	if master.BookedSlots == nil {
		master.BookedSlots = make(map[string]map[TimeSlot]bool)
	}
	if _, exists := master.BookedSlots[dateKey]; !exists {
		master.BookedSlots[dateKey] = make(map[TimeSlot]bool)
	}
	if master.BookedSlots[dateKey][slot] {
		return false
	}
	master.BookedSlots[dateKey][slot] = true
	return true
}

func (mm *MasterManager) CancelSlot(masterID string, date time.Time, slot TimeSlot) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	master, exists := mm.masters[masterID]
	if !exists {
		return
	}
	dateKey := date.Format("2006-01-02")
	if daySlots, exists := master.BookedSlots[dateKey]; exists {
		delete(daySlots, slot)
	}
}

func (mm *MasterManager) FindQualifiedMasters(lockType LockType) []*Master {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	var result []*Master
	for _, m := range mm.masters {
		for _, lt := range m.SupportedLocks {
			if lt == lockType {
				result = append(result, m)
				break
			}
		}
	}
	return result
}

func (mm *MasterManager) FindAvailableSlots(masterID string, date time.Time) []TimeSlot {
	allSlots := []TimeSlot{TimeSlotMorning, TimeSlotAfternoon, TimeSlotEvening}
	var available []TimeSlot
	for _, slot := range allSlots {
		if mm.IsSlotAvailable(masterID, date, slot) {
			available = append(available, slot)
		}
	}
	return available
}

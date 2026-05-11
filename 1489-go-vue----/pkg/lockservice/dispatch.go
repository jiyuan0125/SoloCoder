package lockservice

import (
	"sync"
)

type Dispatcher struct {
	masterManager *MasterManager
	mu            sync.Mutex
	inUseMasters  map[string]bool
}

func NewDispatcher(mm *MasterManager) *Dispatcher {
	return &Dispatcher{
		masterManager: mm,
		inUseMasters:  make(map[string]bool),
	}
}

func (d *Dispatcher) Dispatch(order *Order) *DispatchResult {
	qualified := d.masterManager.FindQualifiedMasters(order.LockType)
	if len(qualified) == 0 {
		return &DispatchResult{Success: false}
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if order.Urgency == UrgencyUrgent {
		for _, m := range qualified {
			if m.Status == MasterStatusIdle && !d.inUseMasters[m.ID] {
				d.inUseMasters[m.ID] = true
				return &DispatchResult{Success: true, Master: m}
			}
		}
		return &DispatchResult{Success: false}
	}

	var selectedMaster *Master
	availableSlots := map[string][]TimeSlot{}
	alternativeMasters := []*Master{}

	for _, m := range qualified {
		slots := d.masterManager.FindAvailableSlots(m.ID, order.TimeSlotDate)
		if len(slots) > 0 {
			availableSlots[m.ID] = slots
			for _, s := range slots {
				if s == order.TimeSlot {
					if !d.inUseMasters[m.ID] {
						selectedMaster = m
						d.inUseMasters[m.ID] = true
						break
					}
				}
			}
		}
	}

	if selectedMaster != nil {
		return &DispatchResult{Success: true, Master: selectedMaster}
	}

	var otherSlots []TimeSlot
	allSlots := []TimeSlot{TimeSlotMorning, TimeSlotAfternoon, TimeSlotEvening}
	for _, s := range allSlots {
		if s != order.TimeSlot {
			otherSlots = append(otherSlots, s)
		}
	}

	for _, m := range qualified {
		if slots, exists := availableSlots[m.ID]; exists && len(slots) > 0 {
			alternativeMasters = append(alternativeMasters, m)
		}
	}

	return &DispatchResult{
		Success: false,
		Alternative: &AlternativeSlots{
			OtherSlots:   otherSlots,
			OtherMasters: alternativeMasters,
		},
	}
}

func (d *Dispatcher) ReleaseMaster(masterID string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.inUseMasters, masterID)
}

package core

import (
	"errors"
	"piperepair/api"
	"sync"
)

type Dispatcher struct {
	store *Store
	mu    sync.Mutex
}

func NewDispatcher(store *Store) *Dispatcher {
	return &Dispatcher{store: store}
}

func (d *Dispatcher) DispatchOrder(order *api.RepairOrder) (*api.Master, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	requiredSkills := d.getRequiredSkills(order)

	site, ok := d.store.GetSite(order.SiteID)
	if !ok {
		return nil, errors.New("站点不存在")
	}

	master, err := d.findAvailableMaster(site, requiredSkills)
	if err == nil {
		return d.assignMaster(order, master)
	}

	for _, neighborSiteID := range site.NeighborSiteIDs {
		neighborSite, ok := d.store.GetSite(neighborSiteID)
		if !ok {
			continue
		}
		master, err = d.findAvailableMaster(neighborSite, requiredSkills)
		if err == nil {
			return d.assignMaster(order, master)
		}
	}

	return nil, errors.New("暂无可用师傅，请稍后重试")
}

func (d *Dispatcher) getRequiredSkills(order *api.RepairOrder) []api.SkillTag {
	skills := make([]api.SkillTag, 0)

	if order.IsRecurring {
		skills = append(skills, api.SkillInspection)
	}

	switch order.BlockageType {
	case api.BlockageTypeMainPipe:
		skills = append(skills, api.SkillHighPressure)
	default:
		if order.IsRecurring {
			return skills
		}
		skills = append(skills, api.SkillGeneral)
	}

	return skills
}

func (d *Dispatcher) findAvailableMaster(site *api.RepairSite, requiredSkills []api.SkillTag) (*api.Master, error) {
	for _, masterID := range site.MasterIDs {
		master, ok := d.store.GetMaster(masterID)
		if !ok {
			continue
		}

		if master.Status != api.MasterStatusIdle {
			continue
		}

		if master.DailyOrders >= 5 {
			continue
		}

		if !d.hasRequiredSkills(master, requiredSkills) {
			continue
		}

		return master, nil
	}

	return nil, errors.New("站点内无可用师傅")
}

func (d *Dispatcher) hasRequiredSkills(master *api.Master, requiredSkills []api.SkillTag) bool {
	skillSet := make(map[api.SkillTag]bool)
	for _, skill := range master.Skills {
		skillSet[skill] = true
	}

	for _, requiredSkill := range requiredSkills {
		if !skillSet[requiredSkill] {
			return false
		}
	}

	return true
}

func (d *Dispatcher) assignMaster(order *api.RepairOrder, master *api.Master) (*api.Master, error) {
	master.Status = api.MasterStatusBusy
	master.DailyOrders++
	d.store.SaveMaster(master)

	order.AssignedMasterID = master.ID
	order.Status = api.StatusDispatched
	order.DispatchCount++
	d.store.SaveOrder(order)

	d.store.SetMasterCurrentOrder(master.ID, order.ID)

	return master, nil
}

func (d *Dispatcher) ReleaseMaster(masterID string) error {
	master, ok := d.store.GetMaster(masterID)
	if !ok {
		return errors.New("师傅不存在")
	}

	master.Status = api.MasterStatusIdle
	d.store.SaveMaster(master)

	d.store.SetMasterCurrentOrder(masterID, "")

	return nil
}

func (d *Dispatcher) RedispatchOrder(order *api.RepairOrder) (*api.Master, error) {
	if order.DispatchCount >= 3 {
		return nil, errors.New("已达最大重新派单次数")
	}

	if order.AssignedMasterID != "" {
		err := d.ReleaseMaster(order.AssignedMasterID)
		if err != nil {
			return nil, err
		}
		order.AssignedMasterID = ""
	}

	return d.DispatchOrder(order)
}

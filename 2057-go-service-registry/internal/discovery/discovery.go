package discovery

import (
	"math/rand"
	"sync"
	"time"

	"serviceregistry/internal/model"
)

type InstanceProvider interface {
	GetInstances(serviceName string) []*model.ServiceInstance
}

type Discovery struct {
	provider   InstanceProvider
	roundRobin map[string]int
	lock       sync.Mutex
	rng        *rand.Rand
}

func NewDiscovery(provider InstanceProvider) *Discovery {
	return &Discovery{
		provider:   provider,
		roundRobin: make(map[string]int),
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (d *Discovery) ListHealthyInstances(serviceName string, version string, env string) []*model.ServiceInstance {
	instances := d.provider.GetInstances(serviceName)

	filtered := make([]*model.ServiceInstance, 0, len(instances))
	for _, inst := range instances {
		if inst.Status != model.StatusHealthy {
			continue
		}
		if version != "" && inst.Version != version {
			continue
		}
		if env != "" && inst.Environment != env {
			continue
		}
		filtered = append(filtered, inst)
	}

	return filtered
}

func (d *Discovery) SelectInstance(serviceName string, strategy model.LoadBalanceStrategy) *model.ServiceInstance {
	return d.SelectInstanceWithFilter(serviceName, strategy, "", "")
}

func (d *Discovery) SelectInstanceWithFilter(serviceName string, strategy model.LoadBalanceStrategy, version string, env string) *model.ServiceInstance {
	instances := d.ListHealthyInstances(serviceName, version, env)
	if len(instances) == 0 {
		return nil
	}

	switch strategy {
	case model.StrategyRoundRobin:
		return d.roundRobinSelect(serviceName, instances)
	case model.StrategyWeightedRandom:
		return d.weightedRandomSelect(instances)
	default:
		return d.roundRobinSelect(serviceName, instances)
	}
}

func (d *Discovery) roundRobinSelect(serviceName string, instances []*model.ServiceInstance) *model.ServiceInstance {
	d.lock.Lock()
	defer d.lock.Unlock()

	idx := d.roundRobin[serviceName]
	if idx >= len(instances) {
		idx = 0
	}
	inst := instances[idx]
	d.roundRobin[serviceName] = (idx + 1) % len(instances)
	return inst
}

func (d *Discovery) weightedRandomSelect(instances []*model.ServiceInstance) *model.ServiceInstance {
	totalWeight := 0
	for _, inst := range instances {
		totalWeight += inst.Weight
	}

	if totalWeight <= 0 {
		return instances[0]
	}

	d.lock.Lock()
	defer d.lock.Unlock()

	r := d.rng.Intn(totalWeight)
	for _, inst := range instances {
		r -= inst.Weight
		if r < 0 {
			return inst
		}
	}

	return instances[len(instances)-1]
}

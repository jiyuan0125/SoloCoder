package ctxmanager

import (
	"context"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"context-chain/pkg/api"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
)

type trackedContext struct {
	id           string
	parentID     string
	ctx          context.Context
	cancelFunc   context.CancelFunc
	timeout      time.Duration
	deadline     time.Time
	createdAt    time.Time
	cancelled    bool
	cancelledAt  time.Time
	cancelCalled bool
	values       map[string]string
	children     map[string]struct{}
	mu           sync.Mutex
}

type cancelEvent struct {
	triggerID   string
	triggerType string
	cancelledAt time.Time
	children    []string
}

type Manager struct {
	contexts      map[string]*trackedContext
	contextsByReq map[string]string
	events        []*cancelEvent
	eventsMu      sync.Mutex
	mu            sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		contexts:      make(map[string]*trackedContext),
		contextsByReq: make(map[string]string),
		events:        make([]*cancelEvent, 0),
	}
}

func (m *Manager) generateID() string {
	return fmt.Sprintf("%d_%s", time.Now().UnixNano(), randomSuffix())
}

func randomSuffix() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = chars[time.Now().UnixNano()%int64(len(chars))]
		time.Sleep(1 * time.Nanosecond)
	}
	return string(b)
}

func (m *Manager) CreateContext(parentCtx context.Context, timeout time.Duration, values []api.ValueItem, parentID string) (*api.ResponseCreate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := m.generateID()
	requestID := m.generateID()

	var ctx context.Context
	var cancel context.CancelFunc
	var deadline time.Time
	var actualTimeout time.Duration

	if parentID != "" {
		parent, exists := m.contexts[parentID]
		if !exists {
			return nil, fmt.Errorf("parent context not found")
		}
		if parent.cancelled {
			return nil, fmt.Errorf("parent context already cancelled")
		}
		parentCtx = parent.ctx
		parent.mu.Lock()
		if parent.children == nil {
			parent.children = make(map[string]struct{})
		}
		parent.children[id] = struct{}{}
		parent.mu.Unlock()
	} else {
		parentCtx = context.Background()
	}

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parentCtx, timeout)
		deadline, _ = ctx.Deadline()
		actualTimeout = timeout
	} else {
		ctx, cancel = context.WithCancel(parentCtx)
	}

	ctx = context.WithValue(ctx, RequestIDKey, requestID)

	valueMap := make(map[string]string)
	for _, v := range values {
		keyStr := v.Key
		valStr := fmt.Sprintf("%v", v.Value)
		valueMap[keyStr] = valStr
		ctx = context.WithValue(ctx, contextKey(keyStr), v.Value)
	}

	tracked := &trackedContext{
		id:        id,
		parentID:  parentID,
		ctx:       ctx,
		cancelFunc: cancel,
		timeout:   actualTimeout,
		deadline:  deadline,
		createdAt: time.Now(),
		values:    valueMap,
		children:  make(map[string]struct{}),
	}

	m.contexts[id] = tracked
	m.contextsByReq[requestID] = id

	go m.monitorContext(tracked)

	info := api.ContextInfo{
		ID:        id,
		ParentID:  parentID,
		Timeout:   actualTimeout,
		CreatedAt: tracked.createdAt,
		Deadline:  deadline,
		Values:    valueMap,
		Cancelled: false,
	}

	return &api.ResponseCreate{
		RequestID: requestID,
		CancelID:  id,
		Context:   info,
	}, nil
}

func (m *Manager) monitorContext(tc *trackedContext) {
	select {
	case <-tc.ctx.Done():
		tc.mu.Lock()
		if !tc.cancelled {
			tc.cancelled = true
			tc.cancelledAt = time.Now()

			triggerType := "manual"
			if tc.ctx.Err() == context.DeadlineExceeded {
				triggerType = "timeout"
			}

			m.eventsMu.Lock()
			event := &cancelEvent{
				triggerID:   tc.id,
				triggerType: triggerType,
				cancelledAt: tc.cancelledAt,
				children:    make([]string, 0),
			}

			m.collectChildren(tc.id, event)
			m.events = append(m.events, event)
			m.eventsMu.Unlock()
		}
		tc.mu.Unlock()
	}
}

func (m *Manager) collectChildren(id string, event *cancelEvent) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tc, exists := m.contexts[id]
	if !exists {
		return
	}

	for childID := range tc.children {
		event.children = append(event.children, childID)
		m.collectChildren(childID, event)
	}
}

func (m *Manager) CancelContext(cancelID string) (*api.ResponseCancel, error) {
	m.mu.RLock()
	tc, exists := m.contexts[cancelID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("context not found")
	}

	tc.mu.Lock()
	if tc.cancelCalled {
		tc.mu.Unlock()
		return &api.ResponseCancel{
			Success: true,
			Message: "cancel function already called (idempotent, no-op on repeated calls)",
		}, nil
	}
	tc.cancelCalled = true
	tc.mu.Unlock()

	tc.cancelFunc()
	log.Printf("[ctxmanager] context %s cancelled (first call only effective)", cancelID)

	return &api.ResponseCancel{
		Success: true,
		Message: "context cancelled successfully",
	}, nil
}

func (m *Manager) GetCancelChain() *api.ResponseChain {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()

	triggers := make([]api.TriggerInfo, 0, len(m.events))
	totalCancel := 0

	for _, event := range m.events {
		children := make([]api.CanceledCtx, 0, len(event.children))
		for _, childID := range event.children {
			m.mu.RLock()
			childCtx, exists := m.contexts[childID]
			m.mu.RUnlock()
			if exists {
				children = append(children, api.CanceledCtx{
					ID:       childID,
					ParentID: childCtx.parentID,
				})
			}
		}

		triggers = append(triggers, api.TriggerInfo{
			TriggerID:   event.triggerID,
			TriggerType: event.triggerType,
			CanceledAt:  event.cancelledAt,
			Children:    children,
		})

		totalCancel += 1 + len(event.children)
	}

	return &api.ResponseChain{
		Triggers:    triggers,
		TotalCancel: totalCancel,
	}
}

func (m *Manager) TestPrecision(target time.Duration, iterations int) (*api.ResponsePrecision, error) {
	if iterations <= 0 {
		iterations = 10
	}

	results := make([]api.PrecisionResult, 0, iterations)
	durations := make([]time.Duration, 0, iterations)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		ctx, cancel := context.WithTimeout(context.Background(), target)
		<-ctx.Done()
		elapsed := time.Since(start)
		cancel()

		deviation := elapsed - target
		percentError := float64(deviation) / float64(target) * 100

		results = append(results, api.PrecisionResult{
			Index:        i,
			Actual:       elapsed,
			Deviation:    deviation,
			PercentError: percentError,
		})
		durations = append(durations, elapsed)
	}

	stats := calculateStats(durations)

	return &api.ResponsePrecision{
		Iterations: iterations,
		Target:     target,
		Stats:      stats,
		Results:    results,
	}, nil
}

func calculateStats(durations []time.Duration) api.PrecisionStats {
	if len(durations) == 0 {
		return api.PrecisionStats{}
	}

	var sum time.Duration
	min := durations[0]
	max := durations[0]

	for _, d := range durations {
		sum += d
		if d < min {
			min = d
		}
		if d > max {
			max = d
		}
	}

	mean := sum / time.Duration(len(durations))

	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	var median time.Duration
	if len(sorted)%2 == 1 {
		median = sorted[len(sorted)/2]
	} else {
		mid := len(sorted) / 2
		median = (sorted[mid-1] + sorted[mid]) / 2
	}

	var varianceSum float64
	meanFloat := float64(mean)
	for _, d := range durations {
		diff := float64(d) - meanFloat
		varianceSum += diff * diff
	}
	stdDev := time.Duration(math.Sqrt(varianceSum / float64(len(durations))))

	return api.PrecisionStats{
		Mean:   mean,
		Median: median,
		Min:    min,
		Max:    max,
		StdDev: stdDev,
	}
}

func (m *Manager) GetContextByRequestID(requestID string) *trackedContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctxID, exists := m.contextsByReq[requestID]
	if !exists {
		return nil
	}
	return m.contexts[ctxID]
}

func (m *Manager) GetStatus() *api.ResponseStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := len(m.contexts)
	canceled := 0

	for _, tc := range m.contexts {
		tc.mu.Lock()
		if tc.cancelled {
			canceled++
		}
		tc.mu.Unlock()
	}

	return &api.ResponseStatus{
		Contexts: api.ContextSummary{
			Total:    total,
			Active:   total - canceled,
			Canceled: canceled,
		},
	}
}

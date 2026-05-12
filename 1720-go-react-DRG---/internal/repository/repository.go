package repository

import (
	"sync"
	"time"

	"drg-system/internal/model"
)

type Repository interface {
	ListDRGGroups() []model.DRGGroup
	GetDRGGroupByCode(code string) (*model.DRGGroup, error)
	CreateDRGGroup(group model.DRGGroup) (*model.DRGGroup, error)
	UpdateDRGGroup(code string, group model.DRGGroup) (*model.DRGGroup, error)
	DeleteDRGGroup(code string) error

	ListGroupingRules() []model.GroupingRule
	CreateGroupingRule(rule model.GroupingRule) (*model.GroupingRule, error)

	CreateMedicalRecord(record model.MedicalRecord) (*model.MedicalRecord, error)
	GetMedicalRecordByCaseNo(caseNo string) (*model.MedicalRecord, error)
	ListMedicalRecordsBySettlement(settlementID string) []model.MedicalRecord

	GetActivePaymentParam() (*model.PaymentParam, error)
	CreatePaymentParam(param model.PaymentParam) (*model.PaymentParam, error)
	ListPaymentParams() []model.PaymentParam

	CreateSettlement(settlement model.Settlement) (*model.Settlement, error)
	GetSettlement(id string) (*model.Settlement, error)
	ListSettlements() []model.Settlement
	ListSettlementsByHospital(hospitalID string) []model.Settlement
	UpdateSettlement(id string, settlement model.Settlement) (*model.Settlement, error)

	CreateTodo(todo model.Todo) (*model.Todo, error)
	GetTodo(id string) (*model.Todo, error)
	ListTodos() []model.Todo
	ListTodosByHospital(hospitalID string) []model.Todo
	UpdateTodo(id string, todo model.Todo) (*model.Todo, error)
}

type InMemoryRepository struct {
	mu             sync.RWMutex
	drgGroups      map[string]model.DRGGroup
	groupingRules  map[string]model.GroupingRule
	medicalRecords map[string]model.MedicalRecord
	caseNoIndex    map[string]string
	paymentParams  []model.PaymentParam
	settlements    map[string]model.Settlement
	todos          map[string]model.Todo
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		drgGroups:      make(map[string]model.DRGGroup),
		groupingRules:  make(map[string]model.GroupingRule),
		medicalRecords: make(map[string]model.MedicalRecord),
		caseNoIndex:    make(map[string]string),
		settlements:    make(map[string]model.Settlement),
		todos:          make(map[string]model.Todo),
	}
}

func (r *InMemoryRepository) InitPresetData() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	presetGroups := []model.DRGGroup{
		{ID: "g1", Code: "GRG01", Name: "脑外科手术伴CC/MCC", Weight: 2.5678, CreatedAt: now, UpdatedAt: now},
		{ID: "g2", Code: "GRG02", Name: "脑外科手术无CC/MCC", Weight: 1.8901, CreatedAt: now, UpdatedAt: now},
		{ID: "g3", Code: "GRG03", Name: "心脏瓣膜置换术伴CC/MCC", Weight: 3.4567, CreatedAt: now, UpdatedAt: now},
		{ID: "g4", Code: "GRG04", Name: "心脏瓣膜置换术无CC/MCC", Weight: 2.7890, CreatedAt: now, UpdatedAt: now},
		{ID: "g5", Code: "GRG05", Name: "冠脉搭桥术伴CC/MCC", Weight: 3.1234, CreatedAt: now, UpdatedAt: now},
		{ID: "g6", Code: "GRG06", Name: "冠脉搭桥术无CC/MCC", Weight: 2.4567, CreatedAt: now, UpdatedAt: now},
		{ID: "g7", Code: "GRG07", Name: "髋关节置换术伴CC/MCC", Weight: 2.2345, CreatedAt: now, UpdatedAt: now},
		{ID: "g8", Code: "GRG08", Name: "髋关节置换术无CC/MCC", Weight: 1.6789, CreatedAt: now, UpdatedAt: now},
		{ID: "g9", Code: "GRG09", Name: "膝关节置换术伴CC/MCC", Weight: 2.0123, CreatedAt: now, UpdatedAt: now},
		{ID: "g10", Code: "GRG10", Name: "膝关节置换术无CC/MCC", Weight: 1.5678, CreatedAt: now, UpdatedAt: now},
		{ID: "g11", Code: "GRG11", Name: "腹腔镜胆囊切除术伴CC/MCC", Weight: 1.3456, CreatedAt: now, UpdatedAt: now},
		{ID: "g12", Code: "GRG12", Name: "腹腔镜胆囊切除术无CC/MCC", Weight: 1.0234, CreatedAt: now, UpdatedAt: now},
		{ID: "g13", Code: "GRG13", Name: "阑尾切除术伴CC/MCC", Weight: 1.1234, CreatedAt: now, UpdatedAt: now},
		{ID: "g14", Code: "GRG14", Name: "阑尾切除术无CC/MCC", Weight: 0.8765, CreatedAt: now, UpdatedAt: now},
		{ID: "g15", Code: "GRG15", Name: "剖宫产术伴CC/MCC", Weight: 1.5678, CreatedAt: now, UpdatedAt: now},
		{ID: "g16", Code: "GRG16", Name: "剖宫产术无CC/MCC", Weight: 1.1234, CreatedAt: now, UpdatedAt: now},
		{ID: "g17", Code: "GRG17", Name: "顺产伴CC/MCC", Weight: 0.9876, CreatedAt: now, UpdatedAt: now},
		{ID: "g18", Code: "GRG18", Name: "顺产无CC/MCC", Weight: 0.7654, CreatedAt: now, UpdatedAt: now},
		{ID: "g19", Code: "GRG19", Name: "急性心肌梗死伴CC/MCC", Weight: 2.3456, CreatedAt: now, UpdatedAt: now},
		{ID: "g20", Code: "GRG20", Name: "急性心肌梗死无CC/MCC", Weight: 1.8765, CreatedAt: now, UpdatedAt: now},
		{ID: "g21", Code: "GRG21", Name: "心力衰竭伴CC/MCC", Weight: 1.8901, CreatedAt: now, UpdatedAt: now},
		{ID: "g22", Code: "GRG22", Name: "心力衰竭无CC/MCC", Weight: 1.4567, CreatedAt: now, UpdatedAt: now},
		{ID: "g23", Code: "GRG23", Name: "肺炎伴CC/MCC", Weight: 1.6789, CreatedAt: now, UpdatedAt: now},
		{ID: "g24", Code: "GRG24", Name: "肺炎无CC/MCC", Weight: 1.2345, CreatedAt: now, UpdatedAt: now},
		{ID: "g25", Code: "GRG25", Name: "脑梗塞伴CC/MCC", Weight: 1.7890, CreatedAt: now, UpdatedAt: now},
		{ID: "g26", Code: "GRG26", Name: "脑梗塞无CC/MCC", Weight: 1.3456, CreatedAt: now, UpdatedAt: now},
		{ID: "g27", Code: "GRG27", Name: "糖尿病伴并发症", Weight: 1.2345, CreatedAt: now, UpdatedAt: now},
		{ID: "g28", Code: "GRG28", Name: "糖尿病无并发症", Weight: 0.8901, CreatedAt: now, UpdatedAt: now},
		{ID: "g29", Code: "GRG29", Name: "慢性肾功能衰竭伴CC/MCC", Weight: 1.9012, CreatedAt: now, UpdatedAt: now},
		{ID: "g30", Code: "GRG30", Name: "慢性肾功能衰竭无CC/MCC", Weight: 1.4567, CreatedAt: now, UpdatedAt: now},
		{ID: "g99", Code: "未分组", Name: "未分组DRG", Weight: 1.0000, CreatedAt: now, UpdatedAt: now},
	}

	for _, g := range presetGroups {
		r.drgGroups[g.Code] = g
	}

	presetRules := []model.GroupingRule{
		{ID: "r1", DiagnosisPrefix: "I6", ProcedurePrefix: "01", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG01", CreatedAt: now},
		{ID: "r2", DiagnosisPrefix: "I6", ProcedurePrefix: "01", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG01", CreatedAt: now},
		{ID: "r3", DiagnosisPrefix: "I6", ProcedurePrefix: "01", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG02", CreatedAt: now},
		{ID: "r4", DiagnosisPrefix: "I05", ProcedurePrefix: "35", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG03", CreatedAt: now},
		{ID: "r5", DiagnosisPrefix: "I05", ProcedurePrefix: "35", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG03", CreatedAt: now},
		{ID: "r6", DiagnosisPrefix: "I05", ProcedurePrefix: "35", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG04", CreatedAt: now},
		{ID: "r7", DiagnosisPrefix: "I25", ProcedurePrefix: "36", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG05", CreatedAt: now},
		{ID: "r8", DiagnosisPrefix: "I25", ProcedurePrefix: "36", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG05", CreatedAt: now},
		{ID: "r9", DiagnosisPrefix: "I25", ProcedurePrefix: "36", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG06", CreatedAt: now},
		{ID: "r10", DiagnosisPrefix: "M16", ProcedurePrefix: "81", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG07", CreatedAt: now},
		{ID: "r11", DiagnosisPrefix: "M16", ProcedurePrefix: "81", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG07", CreatedAt: now},
		{ID: "r12", DiagnosisPrefix: "M16", ProcedurePrefix: "81", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG08", CreatedAt: now},
		{ID: "r13", DiagnosisPrefix: "M17", ProcedurePrefix: "81", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG09", CreatedAt: now},
		{ID: "r14", DiagnosisPrefix: "M17", ProcedurePrefix: "81", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG09", CreatedAt: now},
		{ID: "r15", DiagnosisPrefix: "M17", ProcedurePrefix: "81", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG10", CreatedAt: now},
		{ID: "r16", DiagnosisPrefix: "K80", ProcedurePrefix: "51", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG11", CreatedAt: now},
		{ID: "r17", DiagnosisPrefix: "K80", ProcedurePrefix: "51", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG11", CreatedAt: now},
		{ID: "r18", DiagnosisPrefix: "K80", ProcedurePrefix: "51", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG12", CreatedAt: now},
		{ID: "r19", DiagnosisPrefix: "K35", ProcedurePrefix: "47", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG13", CreatedAt: now},
		{ID: "r20", DiagnosisPrefix: "K35", ProcedurePrefix: "47", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG13", CreatedAt: now},
		{ID: "r21", DiagnosisPrefix: "K35", ProcedurePrefix: "47", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG14", CreatedAt: now},
		{ID: "r22", DiagnosisPrefix: "O82", ProcedurePrefix: "74", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG15", CreatedAt: now},
		{ID: "r23", DiagnosisPrefix: "O82", ProcedurePrefix: "74", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG15", CreatedAt: now},
		{ID: "r24", DiagnosisPrefix: "O82", ProcedurePrefix: "74", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG16", CreatedAt: now},
		{ID: "r25", DiagnosisPrefix: "O80", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG17", CreatedAt: now},
		{ID: "r26", DiagnosisPrefix: "O80", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG17", CreatedAt: now},
		{ID: "r27", DiagnosisPrefix: "O80", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG18", CreatedAt: now},
		{ID: "r28", DiagnosisPrefix: "I21", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG19", CreatedAt: now},
		{ID: "r29", DiagnosisPrefix: "I21", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG19", CreatedAt: now},
		{ID: "r30", DiagnosisPrefix: "I21", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG20", CreatedAt: now},
		{ID: "r31", DiagnosisPrefix: "I50", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG21", CreatedAt: now},
		{ID: "r32", DiagnosisPrefix: "I50", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG21", CreatedAt: now},
		{ID: "r33", DiagnosisPrefix: "I50", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG22", CreatedAt: now},
		{ID: "r34", DiagnosisPrefix: "J18", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG23", CreatedAt: now},
		{ID: "r35", DiagnosisPrefix: "J18", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG23", CreatedAt: now},
		{ID: "r36", DiagnosisPrefix: "J18", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG24", CreatedAt: now},
		{ID: "r37", DiagnosisPrefix: "I63", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG25", CreatedAt: now},
		{ID: "r38", DiagnosisPrefix: "I63", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG25", CreatedAt: now},
		{ID: "r39", DiagnosisPrefix: "I63", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG26", CreatedAt: now},
		{ID: "r40", DiagnosisPrefix: "E11", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG27", CreatedAt: now},
		{ID: "r41", DiagnosisPrefix: "E11", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG27", CreatedAt: now},
		{ID: "r42", DiagnosisPrefix: "E11", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG28", CreatedAt: now},
		{ID: "r43", DiagnosisPrefix: "N18", ProcedurePrefix: "", CCFlag: model.CCFlagMCC, DRGGroupCode: "GRG29", CreatedAt: now},
		{ID: "r44", DiagnosisPrefix: "N18", ProcedurePrefix: "", CCFlag: model.CCFlagCC, DRGGroupCode: "GRG29", CreatedAt: now},
		{ID: "r45", DiagnosisPrefix: "N18", ProcedurePrefix: "", CCFlag: model.CCFlagNone, DRGGroupCode: "GRG30", CreatedAt: now},
	}

	for _, rule := range presetRules {
		r.groupingRules[rule.ID] = rule
	}

	r.paymentParams = append(r.paymentParams, model.PaymentParam{
		ID:         "p1",
		Rate:       6850.123456,
		Version:    1,
		IsActive:   true,
		EffectiveAt: now,
		CreatedAt:  now,
	})
}

func (r *InMemoryRepository) ListDRGGroups() []model.DRGGroup {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.DRGGroup, 0, len(r.drgGroups))
	for _, g := range r.drgGroups {
		result = append(result, g)
	}
	return result
}

func (r *InMemoryRepository) GetDRGGroupByCode(code string) (*model.DRGGroup, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if g, ok := r.drgGroups[code]; ok {
		return &g, nil
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) CreateDRGGroup(group model.DRGGroup) (*model.DRGGroup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drgGroups[group.Code]; ok {
		return nil, ErrConflict
	}
	now := time.Now()
	group.CreatedAt = now
	group.UpdatedAt = now
	r.drgGroups[group.Code] = group
	return &group, nil
}

func (r *InMemoryRepository) UpdateDRGGroup(code string, group model.DRGGroup) (*model.DRGGroup, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.drgGroups[code]
	if !ok {
		return nil, ErrNotFound
	}
	group.ID = existing.ID
	group.Code = code
	group.CreatedAt = existing.CreatedAt
	group.UpdatedAt = time.Now()
	r.drgGroups[code] = group
	return &group, nil
}

func (r *InMemoryRepository) DeleteDRGGroup(code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.drgGroups[code]; !ok {
		return ErrNotFound
	}
	delete(r.drgGroups, code)
	return nil
}

func (r *InMemoryRepository) ListGroupingRules() []model.GroupingRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.GroupingRule, 0, len(r.groupingRules))
	for _, rule := range r.groupingRules {
		result = append(result, rule)
	}
	return result
}

func (r *InMemoryRepository) CreateGroupingRule(rule model.GroupingRule) (*model.GroupingRule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	rule.CreatedAt = now
	r.groupingRules[rule.ID] = rule
	return &rule, nil
}

func (r *InMemoryRepository) CreateMedicalRecord(record model.MedicalRecord) (*model.MedicalRecord, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.caseNoIndex[record.CaseNo]; ok {
		return nil, ErrConflict
	}
	now := time.Now()
	record.CreatedAt = now
	r.medicalRecords[record.ID] = record
	r.caseNoIndex[record.CaseNo] = record.ID
	return &record, nil
}

func (r *InMemoryRepository) GetMedicalRecordByCaseNo(caseNo string) (*model.MedicalRecord, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if id, ok := r.caseNoIndex[caseNo]; ok {
		if record, ok := r.medicalRecords[id]; ok {
			return &record, nil
		}
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) ListMedicalRecordsBySettlement(settlementID string) []model.MedicalRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.MedicalRecord, 0)
	for _, record := range r.medicalRecords {
		if record.SettlementID == settlementID {
			result = append(result, record)
		}
	}
	return result
}

func (r *InMemoryRepository) GetActivePaymentParam() (*model.PaymentParam, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for i := len(r.paymentParams) - 1; i >= 0; i-- {
		if r.paymentParams[i].IsActive {
			return &r.paymentParams[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) CreatePaymentParam(param model.PaymentParam) (*model.PaymentParam, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.paymentParams {
		r.paymentParams[i].IsActive = false
	}
	now := time.Now()
	param.Version = len(r.paymentParams) + 1
	param.IsActive = true
	param.CreatedAt = now
	r.paymentParams = append(r.paymentParams, param)
	return &param, nil
}

func (r *InMemoryRepository) ListPaymentParams() []model.PaymentParam {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.PaymentParam, len(r.paymentParams))
	copy(result, r.paymentParams)
	return result
}

func (r *InMemoryRepository) CreateSettlement(settlement model.Settlement) (*model.Settlement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	settlement.CreatedAt = now
	settlement.Status = model.SettlementStatusDraft
	r.settlements[settlement.ID] = settlement
	return &settlement, nil
}

func (r *InMemoryRepository) GetSettlement(id string) (*model.Settlement, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if s, ok := r.settlements[id]; ok {
		return &s, nil
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) ListSettlements() []model.Settlement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Settlement, 0, len(r.settlements))
	for _, s := range r.settlements {
		result = append(result, s)
	}
	return result
}

func (r *InMemoryRepository) ListSettlementsByHospital(hospitalID string) []model.Settlement {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Settlement, 0)
	for _, s := range r.settlements {
		if s.HospitalID == hospitalID {
			result = append(result, s)
		}
	}
	return result
}

func (r *InMemoryRepository) UpdateSettlement(id string, settlement model.Settlement) (*model.Settlement, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.settlements[id]
	if !ok {
		return nil, ErrNotFound
	}
	settlement.ID = id
	settlement.CreatedAt = existing.CreatedAt
	r.settlements[id] = settlement
	return &settlement, nil
}

func (r *InMemoryRepository) CreateTodo(todo model.Todo) (*model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	todo.CreatedAt = now
	todo.Status = model.TodoStatusPending
	r.todos[todo.ID] = todo
	return &todo, nil
}

func (r *InMemoryRepository) GetTodo(id string) (*model.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.todos[id]; ok {
		return &t, nil
	}
	return nil, ErrNotFound
}

func (r *InMemoryRepository) ListTodos() []model.Todo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Todo, 0, len(r.todos))
	for _, t := range r.todos {
		result = append(result, t)
	}
	return result
}

func (r *InMemoryRepository) ListTodosByHospital(hospitalID string) []model.Todo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]model.Todo, 0)
	for _, t := range r.todos {
		if t.HospitalID == hospitalID {
			result = append(result, t)
		}
	}
	return result
}

func (r *InMemoryRepository) UpdateTodo(id string, todo model.Todo) (*model.Todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.todos[id]
	if !ok {
		return nil, ErrNotFound
	}
	todo.ID = id
	todo.CreatedAt = existing.CreatedAt
	r.todos[id] = todo
	return &todo, nil
}

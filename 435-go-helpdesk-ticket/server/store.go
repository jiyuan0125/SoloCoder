package main

import (
	"go-helpdesk-ticket/common"
	"sync"
)

type Store struct {
	tickets     map[string]*common.Ticket
	logs        map[string][]*common.OperationLog
	handlers    map[string]*common.Handler
	templates   map[string]*common.QuickReplyTemplate
	mu          sync.RWMutex
}

func NewStore() *Store {
	s := &Store{
		tickets:   make(map[string]*common.Ticket),
		logs:      make(map[string][]*common.OperationLog),
		handlers:  make(map[string]*common.Handler),
		templates: make(map[string]*common.QuickReplyTemplate),
	}
	s.initHandlers()
	s.initTemplates()
	return s
}

func (s *Store) initHandlers() {
	s.handlers = map[string]*common.Handler{
		"h1": {ID: "h1", Name: "张网络", Category: common.CategoryNetwork, Level: 1, IsManager: false},
		"h2": {ID: "h2", Name: "李网络", Category: common.CategoryNetwork, Level: 1, IsManager: false},
		"h3": {ID: "h3", Name: "王硬件", Category: common.CategoryHardware, Level: 1, IsManager: false},
		"h4": {ID: "h4", Name: "赵硬件", Category: common.CategoryHardware, Level: 1, IsManager: false},
		"h5": {ID: "h5", Name: "刘软件", Category: common.CategorySoftware, Level: 1, IsManager: false},
		"h6": {ID: "h6", Name: "陈软件", Category: common.CategorySoftware, Level: 1, IsManager: false},
		"h7": {ID: "h7", Name: "孙账号", Category: common.CategoryAccount, Level: 1, IsManager: false},
		"h8": {ID: "h8", Name: "周账号", Category: common.CategoryAccount, Level: 1, IsManager: false},
		"m1": {ID: "m1", Name: "网络经理", Category: common.CategoryNetwork, Level: 2, IsManager: true},
		"m2": {ID: "m2", Name: "硬件经理", Category: common.CategoryHardware, Level: 2, IsManager: true},
		"m3": {ID: "m3", Name: "软件经理", Category: common.CategorySoftware, Level: 2, IsManager: true},
		"m4": {ID: "m4", Name: "账号经理", Category: common.CategoryAccount, Level: 2, IsManager: true},
		"dm": {ID: "dm", Name: "部门经理", Category: "", Level: 3, IsManager: true},
	}
}

func (s *Store) initTemplates() {
	s.templates = map[string]*common.QuickReplyTemplate{
		"t1": {ID: "t1", Title: "网络连接问题", Content: "您好，您的网络问题我们已经收到。请尝试以下步骤：1. 检查网线连接 2. 重启路由器 3. 如仍有问题请联系我们。", Category: common.CategoryNetwork},
		"t2": {ID: "t2", Title: "密码重置", Content: "您好，您的密码重置请求已处理。请使用临时密码登录后修改。", Category: common.CategoryAccount},
		"t3": {ID: "t3", Title: "软件安装指导", Content: "您好，请按照以下步骤安装：1. 双击安装包 2. 按照向导点击下一步 3. 完成后重启电脑。", Category: common.CategorySoftware},
		"t4": {ID: "t4", Title: "硬件故障报修", Content: "您好，您的硬件故障已登记，维修人员将在2小时内到达现场处理。", Category: common.CategoryHardware},
	}
}

func (s *Store) SaveTicket(ticket *common.Ticket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickets[ticket.ID] = ticket
}

func (s *Store) GetTicket(id string) (*common.Ticket, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickets[id]
	return t, ok
}

func (s *Store) GetAllTickets() []*common.Ticket {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tickets := make([]*common.Ticket, 0, len(s.tickets))
	for _, t := range s.tickets {
		tickets = append(tickets, t)
	}
	return tickets
}

func (s *Store) GetHandler(id string) (*common.Handler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h, ok := s.handlers[id]
	return h, ok
}

func (s *Store) GetHandlersByCategory(category common.TicketCategory) []*common.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var handlers []*common.Handler
	for _, h := range s.handlers {
		if h.Category == category && !h.IsManager {
			handlers = append(handlers, h)
		}
	}
	return handlers
}

func (s *Store) GetManagerByCategory(category common.TicketCategory) (*common.Handler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, h := range s.handlers {
		if h.Category == category && h.IsManager && h.Level == 2 {
			return h, true
		}
	}
	return nil, false
}

func (s *Store) GetDeptManager() *common.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handlers["dm"]
}

func (s *Store) AddLog(log *common.OperationLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs[log.TicketID] = append(s.logs[log.TicketID], log)
}

func (s *Store) GetLogs(ticketID string) []*common.OperationLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.logs[ticketID]
}

func (s *Store) GetAllLogs() map[string][]*common.OperationLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string][]*common.OperationLog)
	for k, v := range s.logs {
		copy[k] = v
	}
	return copy
}

func (s *Store) GetAllHandlers() map[string]*common.Handler {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]*common.Handler)
	for k, v := range s.handlers {
		copy[k] = v
	}
	return copy
}

func (s *Store) GetTemplates() map[string]*common.QuickReplyTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copy := make(map[string]*common.QuickReplyTemplate)
	for k, v := range s.templates {
		copy[k] = v
	}
	return copy
}

func (s *Store) GetTemplatesByCategory(category common.TicketCategory) []*common.QuickReplyTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var templates []*common.QuickReplyTemplate
	for _, t := range s.templates {
		if t.Category == category {
			templates = append(templates, t)
		}
	}
	return templates
}

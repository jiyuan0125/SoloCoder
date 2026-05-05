package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"go-helpdesk-ticket/common"
	"strings"
	"time"
)

type TicketService struct {
	store *Store
}

func NewTicketService(store *Store) *TicketService {
	return &TicketService{store: store}
}

func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "TK" + hex.EncodeToString(b)[:12]
}

func (s *TicketService) CreateTicket(req *common.CreateTicketRequest) (*common.Ticket, error) {
	if strings.TrimSpace(req.Title) == "" {
		return nil, fmt.Errorf("title is required")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, fmt.Errorf("description is required")
	}
	if strings.TrimSpace(req.SubmitterID) == "" {
		return nil, fmt.Errorf("submitter_id is required")
	}
	if strings.TrimSpace(req.SubmitterName) == "" {
		return nil, fmt.Errorf("submitter_name is required")
	}

	category := req.Category
	if req.AutoClassify && category == "" {
		category = s.AutoClassify(req.Title, req.Description)
	}

	if !s.isValidCategory(category) {
		return nil, fmt.Errorf(common.GetErrorMessage(common.ErrCodeInvalidCategory))
	}

	if !s.isValidPriority(req.Priority) {
		return nil, fmt.Errorf(common.GetErrorMessage(common.ErrCodeInvalidPriority))
	}

	now := time.Now()
	ticket := &common.Ticket{
		ID:            generateID(),
		Title:         req.Title,
		Description:   req.Description,
		Category:      category,
		Priority:      req.Priority,
		Status:        common.StatusNew,
		SubmitterID:   req.SubmitterID,
		SubmitterName: req.SubmitterName,
		CreatedAt:     now,
		UpdatedAt:     now,
		EscalationLevel: 0,
		SLAIsMet:      true,
	}

	ticket.SLAResponseDeadline = s.calculateSLADeadline(req.Priority, now)

	assignee := s.findBestAssignee(ticket.Category)
	if assignee != nil {
		ticket.AssigneeID = assignee.ID
		ticket.AssigneeName = assignee.Name
		ticket.Status = common.StatusAssigned
	}

	s.store.SaveTicket(ticket)

	s.addOperationLog(ticket.ID, common.OpCreate, req.SubmitterID, req.SubmitterName, "", "", "工单创建")
	if assignee != nil {
		s.addOperationLog(ticket.ID, common.OpAssign, "system", "System", "", fmt.Sprintf("%s (%s)", assignee.Name, assignee.ID), "自动分配")
	}

	return ticket, nil
}

func (s *TicketService) AutoClassify(title, description string) common.TicketCategory {
	fullText := title + " " + description
	for category, keywords := range common.CategoryKeywords {
		for _, kw := range keywords {
			if strings.Contains(fullText, kw) {
				return category
			}
		}
	}
	return common.CategorySoftware
}

func (s *TicketService) isValidCategory(c common.TicketCategory) bool {
	switch c {
	case common.CategoryNetwork, common.CategoryHardware, common.CategorySoftware, common.CategoryAccount:
		return true
	default:
		return false
	}
}

func (s *TicketService) isValidPriority(p common.TicketPriority) bool {
	switch p {
	case common.PriorityNormal, common.PriorityUrgent, common.PriorityCritical:
		return true
	default:
		return false
	}
}

func (s *TicketService) isValidStatus(status common.TicketStatus) bool {
	switch status {
	case common.StatusNew, common.StatusAssigned, common.StatusInProgress, common.StatusResolved, common.StatusClosed:
		return true
	default:
		return false
	}
}

func (s *TicketService) calculateSLADeadline(priority common.TicketPriority, now time.Time) time.Time {
	var hours int
	switch priority {
	case common.PriorityCritical:
		hours = common.SLAHoursCritical
	case common.PriorityUrgent:
		hours = common.SLAHoursUrgent
	default:
		hours = common.SLAHoursNormal
	}
	return now.Add(time.Duration(hours) * time.Hour)
}

func (s *TicketService) findBestAssignee(category common.TicketCategory) *common.Handler {
	handlers := s.store.GetHandlersByCategory(category)
	if len(handlers) == 0 {
		return nil
	}

	var best *common.Handler
	minCount := -1

	for _, h := range handlers {
		count := s.getActiveTicketCount(h.ID)
		if minCount == -1 || count < minCount {
			minCount = count
			best = h
		}
	}

	return best
}

func (s *TicketService) getActiveTicketCount(handlerID string) int {
	count := 0
	tickets := s.store.GetAllTickets()
	for _, t := range tickets {
		if t.AssigneeID == handlerID && t.Status != common.StatusClosed {
			count++
		}
	}
	return count
}

func (s *TicketService) GetTicket(id string) (*common.Ticket, error) {
	ticket, ok := s.store.GetTicket(id)
	if !ok {
		return nil, fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound))
	}
	return ticket, nil
}

func (s *TicketService) GetAllTickets() []*common.Ticket {
	return s.store.GetAllTickets()
}

func (s *TicketService) ReassignTicket(ticketID string, req *common.ReassignTicketRequest) error {
	if req.Reason == "" {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeReassignReasonRequired))
	}

	ticket, ok := s.store.GetTicket(ticketID)
	if !ok {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound))
	}

	oldAssignee := fmt.Sprintf("%s (%s)", ticket.AssigneeName, ticket.AssigneeID)
	newAssignee := fmt.Sprintf("%s (%s)", req.NewAssigneeName, req.NewAssigneeID)

	ticket.AssigneeID = req.NewAssigneeID
	ticket.AssigneeName = req.NewAssigneeName
	ticket.UpdatedAt = time.Now()
	s.store.SaveTicket(ticket)

	s.addOperationLog(ticketID, common.OpReassign, req.NewAssigneeID, req.NewAssigneeName, oldAssignee, newAssignee, req.Reason)
	return nil
}

func (s *TicketService) UpdateStatus(ticketID string, req *common.UpdateStatusRequest) error {
	if !s.isValidStatus(req.Status) {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeInvalidStatus))
	}

	ticket, ok := s.store.GetTicket(ticketID)
	if !ok {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound))
	}

	if ticket.Status == common.StatusClosed {
		return fmt.Errorf("cannot update closed ticket")
	}

	oldStatus := ticket.Status
	ticket.Status = req.Status
	ticket.UpdatedAt = time.Now()

	if ticket.FirstResponse.IsZero() && req.Status != common.StatusNew {
		ticket.FirstResponse = time.Now()
		if ticket.FirstResponse.After(ticket.SLAResponseDeadline) {
			ticket.SLAIsMet = false
		}
	}

	if req.Status == common.StatusClosed {
		ticket.ClosedAt = time.Now()
		if err := s.checkAndCloseParent(ticket.ParentID); err != nil {
			return err
		}
	}

	s.store.SaveTicket(ticket)

	handlerName := req.OperatorID
	if handler, ok := s.store.GetHandler(req.OperatorID); ok {
		handlerName = handler.Name
	}
	s.addOperationLog(ticketID, common.OpStatusChange, req.OperatorID, handlerName, string(oldStatus), string(req.Status), req.Comment)
	return nil
}

func (s *TicketService) checkAndCloseParent(parentID string) error {
	if parentID == "" {
		return nil
	}

	parent, ok := s.store.GetTicket(parentID)
	if !ok {
		return nil
	}

	allChildrenClosed := true
	for _, childID := range parent.ChildIDs {
		child, ok := s.store.GetTicket(childID)
		if !ok {
			continue
		}
		if child.Status != common.StatusClosed {
			allChildrenClosed = false
			break
		}
	}

	if allChildrenClosed && parent.Status != common.StatusClosed {
		parent.Status = common.StatusClosed
		parent.ClosedAt = time.Now()
		s.store.SaveTicket(parent)
		s.addOperationLog(parentID, common.OpStatusChange, "system", "System", string(common.StatusAssigned), string(common.StatusClosed), "所有子工单已关闭，自动关闭主工单")
	}

	return nil
}

func (s *TicketService) RateTicket(ticketID string, req *common.RateTicketRequest) error {
	if req.Rating < 1 || req.Rating > 5 {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeInvalidRating))
	}

	ticket, ok := s.store.GetTicket(ticketID)
	if !ok {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound))
	}

	if ticket.Status != common.StatusClosed {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotClosed))
	}

	if ticket.SubmitterID != req.UserID {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodePermissionDenied))
	}

	ticket.Rating = req.Rating
	ticket.RatingComment = req.Comment
	ticket.UpdatedAt = time.Now()

	s.store.SaveTicket(ticket)
	s.addOperationLog(ticketID, common.OpRate, req.UserID, req.UserName, "", fmt.Sprintf("评分: %d星", req.Rating), req.Comment)
	return nil
}

func (s *TicketService) LinkTickets(req *common.LinkTicketsRequest) error {
	if req.ParentID == req.ChildID {
		return fmt.Errorf("cannot link ticket to itself")
	}

	parent, ok := s.store.GetTicket(req.ParentID)
	if !ok {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound) + ": parent")
	}

	child, ok := s.store.GetTicket(req.ChildID)
	if !ok {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound) + ": child")
	}

	if child.ParentID != "" {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeParentTicketExists))
	}

	if s.checkCircularDependency(req.ParentID, req.ChildID) {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeCircularDependency))
	}

	child.ParentID = req.ParentID
	child.UpdatedAt = time.Now()

	parent.ChildIDs = append(parent.ChildIDs, req.ChildID)
	parent.UpdatedAt = time.Now()

	s.store.SaveTicket(child)
	s.store.SaveTicket(parent)

	s.addOperationLog(req.ChildID, common.OpLink, "system", "System", "", req.ParentID, "关联到主工单")
	s.addOperationLog(req.ParentID, common.OpLink, "system", "System", "", req.ChildID, "添加子工单")

	return nil
}

func (s *TicketService) checkCircularDependency(parentID, childID string) bool {
	current := parentID
	visited := make(map[string]bool)

	for current != "" {
		if visited[current] {
			return true
		}
		if current == childID {
			return true
		}
		visited[current] = true

		t, ok := s.store.GetTicket(current)
		if !ok {
			break
		}
		current = t.ParentID
	}

	return false
}

func (s *TicketService) TransferTicket(ticketID string, req *common.TransferTicketRequest) error {
	ticket, ok := s.store.GetTicket(ticketID)
	if !ok {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeTicketNotFound))
	}

	if req.Reason == "" {
		return fmt.Errorf(common.GetErrorMessage(common.ErrCodeReassignReasonRequired))
	}

	newTicket := &common.Ticket{
		ID:            generateID(),
		Title:         ticket.Title,
		Description:   ticket.Description + " [转自工单: " + ticketID + "]",
		Category:      ticket.Category,
		Priority:      ticket.Priority,
		Status:        common.StatusNew,
		SubmitterID:   ticket.SubmitterID,
		SubmitterName: ticket.SubmitterName,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LinkedTicketID: ticketID,
		EscalationLevel: 0,
		SLAIsMet:      true,
	}

	newTicket.SLAResponseDeadline = s.calculateSLADeadline(newTicket.Priority, newTicket.CreatedAt)

	targetCategory := s.mapTeamToCategory(req.TargetTeam)
	if targetCategory != "" {
		newTicket.Category = targetCategory
	}

	assignee := s.findBestAssignee(newTicket.Category)
	if assignee != nil {
		newTicket.AssigneeID = assignee.ID
		newTicket.AssigneeName = assignee.Name
		newTicket.Status = common.StatusAssigned
	}

	s.store.SaveTicket(newTicket)
	ticket.LinkedTicketID = newTicket.ID
	ticket.UpdatedAt = time.Now()
	s.store.SaveTicket(ticket)

	s.addOperationLog(newTicket.ID, common.OpTransfer, req.OperatorID, req.OperatorName, "", "", "从工单 "+ticketID+" 转移而来，原因: "+req.Reason)
	s.addOperationLog(ticketID, common.OpTransfer, req.OperatorID, req.OperatorName, "", newTicket.ID, "转移到新工单，原因: "+req.Reason)

	return nil
}

func (s *TicketService) mapTeamToCategory(team string) common.TicketCategory {
	switch strings.ToLower(team) {
	case "network", "网络":
		return common.CategoryNetwork
	case "hardware", "硬件":
		return common.CategoryHardware
	case "software", "软件":
		return common.CategorySoftware
	case "account", "账号":
		return common.CategoryAccount
	default:
		return ""
	}
}

func (s *TicketService) CheckSLAAndEscalate() {
	now := time.Now()
	tickets := s.store.GetAllTickets()

	for _, t := range tickets {
		if t.Status == common.StatusClosed || t.Status == common.StatusResolved {
			continue
		}

		elapsedHours := now.Sub(t.CreatedAt).Hours()

		if t.EscalationLevel == 0 && now.After(t.SLAResponseDeadline) && t.FirstResponse.IsZero() {
			s.escalateToManager(t)
			continue
		}

		if t.EscalationLevel >= 1 && elapsedHours >= common.SLAHoursEscalateToManager {
			s.escalateToDeptManager(t)
		}
	}
}

func (s *TicketService) escalateToManager(ticket *common.Ticket) {
	if ticket.EscalationLevel >= 1 {
		return
	}

	manager, ok := s.store.GetManagerByCategory(ticket.Category)
	if !ok {
		return
	}

	oldAssignee := fmt.Sprintf("%s (%s)", ticket.AssigneeName, ticket.AssigneeID)
	newAssignee := fmt.Sprintf("%s (%s)", manager.Name, manager.ID)

	ticket.AssigneeID = manager.ID
	ticket.AssigneeName = manager.Name
	ticket.EscalationLevel = 1
	ticket.UpdatedAt = time.Now()
	ticket.SLAIsMet = false

	s.store.SaveTicket(ticket)
	s.addOperationLog(ticket.ID, common.OpEscalate, "system", "System", oldAssignee, newAssignee, "SLA超时，升级到部门经理")
}

func (s *TicketService) escalateToDeptManager(ticket *common.Ticket) {
	if ticket.EscalationLevel >= 2 {
		return
	}

	manager := s.store.GetDeptManager()
	if manager == nil {
		return
	}

	oldAssignee := fmt.Sprintf("%s (%s)", ticket.AssigneeName, ticket.AssigneeID)
	newAssignee := fmt.Sprintf("%s (%s)", manager.Name, manager.ID)

	ticket.AssigneeID = manager.ID
	ticket.AssigneeName = manager.Name
	ticket.EscalationLevel = 2
	ticket.UpdatedAt = time.Now()

	s.store.SaveTicket(ticket)
	s.addOperationLog(ticket.ID, common.OpEscalate, "system", "System", oldAssignee, newAssignee, "7天未关闭，升级到部门总监")
}

func (s *TicketService) GetOperationLogs(ticketID string) []*common.OperationLog {
	return s.store.GetLogs(ticketID)
}

func (s *TicketService) GetStatistics() *common.StatisticsReport {
	tickets := s.store.GetAllTickets()
	report := &common.StatisticsReport{
		ByCategory: make(map[common.TicketCategory]common.CategoryStats),
		TotalTickets: len(tickets),
	}

	categoryTickets := make(map[common.TicketCategory][]*common.Ticket)
	var totalHandleTime float64
	var closedCount int
	var slaMetCount int
	var slaTotalCount int

	for _, t := range tickets {
		categoryTickets[t.Category] = append(categoryTickets[t.Category], t)

		if t.Status == common.StatusClosed && !t.ClosedAt.IsZero() {
			handleHours := t.ClosedAt.Sub(t.CreatedAt).Hours()
			totalHandleTime += handleHours
			closedCount++
		}

		if t.FirstResponse.IsZero() && t.Status != common.StatusNew && t.Status != common.StatusClosed {
			continue
		}
		if !t.FirstResponse.IsZero() {
			slaTotalCount++
			if t.SLAIsMet {
				slaMetCount++
			}
		}
	}

	if closedCount > 0 {
		report.AverageHandleTime = totalHandleTime / float64(closedCount)
	}

	if slaTotalCount > 0 {
		report.OverallSLARate = float64(slaMetCount) / float64(slaTotalCount) * 100
	}

	for cat, tickets := range categoryTickets {
		stats := common.CategoryStats{Count: len(tickets)}
		var catHandleTime float64
		var catClosedCount int
		var catSLAMetCount int
		var catSLATotalCount int

		for _, t := range tickets {
			if t.Status == common.StatusClosed && !t.ClosedAt.IsZero() {
				handleHours := t.ClosedAt.Sub(t.CreatedAt).Hours()
				catHandleTime += handleHours
				catClosedCount++
			}
			if !t.FirstResponse.IsZero() {
				catSLATotalCount++
				if t.SLAIsMet {
					catSLAMetCount++
				}
			}
		}

		if catClosedCount > 0 {
			stats.HandleTimeHours = catHandleTime / float64(catClosedCount)
		}
		if catSLATotalCount > 0 {
			stats.SLARate = float64(catSLAMetCount) / float64(catSLATotalCount) * 100
		}

		report.ByCategory[cat] = stats
	}

	return report
}

func (s *TicketService) GetHandlers() map[string]*common.Handler {
	return s.store.GetAllHandlers()
}

func (s *TicketService) GetTemplates() map[string]*common.QuickReplyTemplate {
	return s.store.GetTemplates()
}

func (s *TicketService) addOperationLog(ticketID string, opType common.OperationType, operatorID, operatorName, oldValue, newValue, comment string) {
	log := &common.OperationLog{
		ID:            generateID(),
		TicketID:      ticketID,
		OperationType: opType,
		OperatorID:    operatorID,
		OperatorName:  operatorName,
		OldValue:      oldValue,
		NewValue:      newValue,
		Comment:       comment,
		CreatedAt:     time.Now(),
	}
	s.store.AddLog(log)
}

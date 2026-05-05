package main

import (
	"announcement-board/common"
	"sync"
	"time"
)

type Store struct {
	mu sync.RWMutex

	users         map[string]*common.User
	departments   map[string]*common.Department
	announcements map[string]*common.Announcement
	approvals     map[string]*common.ApprovalRecord
}

var globalStore *Store

func initStore() {
	globalStore = &Store{
		users:         make(map[string]*common.User),
		departments:   make(map[string]*common.Department),
		announcements: make(map[string]*common.Announcement),
		approvals:     make(map[string]*common.ApprovalRecord),
	}
	initMockData()
}

func initMockData() {
	globalStore.departments["dept_1"] = &common.Department{ID: "dept_1", Name: "技术部"}
	globalStore.departments["dept_2"] = &common.Department{ID: "dept_2", Name: "产品部"}
	globalStore.departments["dept_3"] = &common.Department{ID: "dept_3", Name: "市场部"}
	globalStore.departments["dept_4"] = &common.Department{ID: "dept_4", Name: "人事部"}

	globalStore.users["user_admin"] = &common.User{
		ID:       "user_admin",
		Name:     "系统管理员",
		DeptID:   "dept_1",
		DeptName: "技术部",
		Role:     common.RoleAdmin,
	}
	globalStore.users["user_manager_1"] = &common.User{
		ID:       "user_manager_1",
		Name:     "张经理",
		DeptID:   "dept_1",
		DeptName: "技术部",
		Role:     common.RoleManager,
	}
	globalStore.users["user_emp_1"] = &common.User{
		ID:       "user_emp_1",
		Name:     "李四",
		DeptID:   "dept_1",
		DeptName: "技术部",
		Role:     common.RoleEmployee,
	}
	globalStore.users["user_emp_2"] = &common.User{
		ID:       "user_emp_2",
		Name:     "王五",
		DeptID:   "dept_2",
		DeptName: "产品部",
		Role:     common.RoleEmployee,
	}
}

func (s *Store) GetUser(userID string) (*common.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[userID]
	return u, ok
}

func (s *Store) GetAllUsers() []*common.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*common.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func (s *Store) GetDepartment(deptID string) (*common.Department, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.departments[deptID]
	return d, ok
}

func (s *Store) GetAllDepartments() []*common.Department {
	s.mu.RLock()
	defer s.mu.RUnlock()
	depts := make([]*common.Department, 0, len(s.departments))
	for _, d := range s.departments {
		depts = append(depts, d)
	}
	return depts
}

func (s *Store) GetUsersByDept(deptID string) []*common.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var users []*common.User
	for _, u := range s.users {
		if u.DeptID == deptID {
			users = append(users, u)
		}
	}
	return users
}

func (s *Store) CreateAnnouncement(ann *common.Announcement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.announcements[ann.ID] = ann
}

func (s *Store) GetAnnouncement(id string) (*common.Announcement, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.announcements[id]
	return a, ok
}

func (s *Store) UpdateAnnouncement(ann *common.Announcement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.announcements[ann.ID] = ann
}

func (s *Store) DeleteAnnouncement(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.announcements, id)
}

func (s *Store) GetPinnedAnnouncements() []*common.Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var pinned []*common.Announcement
	now := time.Now()
	for _, a := range s.announcements {
		if a.IsPinned && a.Status == common.StatusPublished &&
			!a.EffectiveStart.After(now) && !a.EffectiveEnd.Before(now) {
			pinned = append(pinned, a)
		}
	}
	return pinned
}

func (s *Store) GetAllAnnouncements() []*common.Announcement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	anns := make([]*common.Announcement, 0, len(s.announcements))
	for _, a := range s.announcements {
		anns = append(anns, a)
	}
	return anns
}

func (s *Store) CreateApprovalRecord(ar *common.ApprovalRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[ar.ID] = ar
}

func (s *Store) GetApprovalRecord(id string) (*common.ApprovalRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ar, ok := s.approvals[id]
	return ar, ok
}

func (s *Store) GetApprovalRecordByAnnouncementID(annID string) (*common.ApprovalRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, ar := range s.approvals {
		if ar.AnnouncementID == annID {
			return ar, true
		}
	}
	return nil, false
}

func (s *Store) UpdateApprovalRecord(ar *common.ApprovalRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[ar.ID] = ar
}

func (s *Store) GetPendingApprovalsByApprover(approverDeptID string) []*common.ApprovalRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var pending []*common.ApprovalRecord
	for _, ar := range s.approvals {
		if ar.Status == common.ApprovalPending {
			ann, ok := s.announcements[ar.AnnouncementID]
			if !ok {
				continue
			}
			if ann.Scope == common.ScopeAll {
				continue
			}
			for _, deptID := range ann.TargetDeptIDs {
				if deptID == approverDeptID {
					pending = append(pending, ar)
					break
				}
			}
		}
	}
	return pending
}

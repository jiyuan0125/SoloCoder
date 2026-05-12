package storage

import (
	"research-collaboration/src/models"
	"sync"
)

type Storage struct {
	mu           sync.RWMutex
	projects     map[string]*models.Project
	tasks        map[string]*models.Task
	sharedData   map[string]*models.SharedData
	downloadLogs map[string]*models.DownloadLog
	achievements map[string]*models.Achievement
	approvals    map[string]*models.Approval
	todos        map[string]*models.TodoItem
	users        map[string]*models.User
	departments  map[string]*models.Department
	nextID       int
}

func NewStorage() *Storage {
	s := &Storage{
		projects:     make(map[string]*models.Project),
		tasks:        make(map[string]*models.Task),
		sharedData:   make(map[string]*models.SharedData),
		downloadLogs: make(map[string]*models.DownloadLog),
		achievements: make(map[string]*models.Achievement),
		approvals:    make(map[string]*models.Approval),
		todos:        make(map[string]*models.TodoItem),
		users:        make(map[string]*models.User),
		departments:  make(map[string]*models.Department),
		nextID:       1,
	}
	s.initSeedData()
	return s
}

func (s *Storage) initSeedData() {
	s.departments["dept-1"] = &models.Department{ID: "dept-1", Name: "计算机学院"}
	s.departments["dept-2"] = &models.Department{ID: "dept-2", Name: "电子工程学院"}
	s.departments["dept-3"] = &models.Department{ID: "dept-3", Name: "数学学院"}

	s.users["user-1"] = &models.User{ID: "user-1", Name: "张教授", DepartmentID: "dept-1"}
	s.users["user-2"] = &models.User{ID: "user-2", Name: "李博士", DepartmentID: "dept-1"}
	s.users["user-3"] = &models.User{ID: "user-3", Name: "王研究员", DepartmentID: "dept-2"}
	s.users["user-4"] = &models.User{ID: "user-4", Name: "赵老师", DepartmentID: "dept-3"}
}

func (s *Storage) generateID(prefix string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := s.nextID
	s.nextID++
	return prefix + "-" + string(rune('0'+id))
}

func (s *Storage) GetUsers() map[string]*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make(map[string]*models.User)
	for k, v := range s.users {
		users[k] = v
	}
	return users
}

func (s *Storage) GetDepartments() map[string]*models.Department {
	s.mu.RLock()
	defer s.mu.RUnlock()
	depts := make(map[string]*models.Department)
	for k, v := range s.departments {
		depts[k] = v
	}
	return depts
}

func (s *Storage) GetUser(id string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *Storage) CreateProject(p *models.Project) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		s.nextID++
		p.ID = "proj-" + string(rune('0'+s.nextID-1))
	}
	s.projects[p.ID] = p
}

func (s *Storage) GetProject(id string) (*models.Project, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.projects[id]
	return p, ok
}

func (s *Storage) GetAllProjects() []*models.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.Project, 0, len(s.projects))
	for _, p := range s.projects {
		result = append(result, p)
	}
	return result
}

func (s *Storage) UpdateProject(p *models.Project) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[p.ID] = p
}

func (s *Storage) DeleteProject(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.projects, id)
}

func (s *Storage) CreateTask(t *models.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		s.nextID++
		t.ID = "task-" + string(rune('0'+s.nextID-1))
	}
	s.tasks[t.ID] = t
}

func (s *Storage) GetTask(id string) (*models.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *Storage) GetTasksByProject(projectID string) []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*models.Task{}
	for _, t := range s.tasks {
		if t.ProjectID == projectID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Storage) GetAllTasks() []*models.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		result = append(result, t)
	}
	return result
}

func (s *Storage) UpdateTask(t *models.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
}

func (s *Storage) DeleteTasksByProject(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, t := range s.tasks {
		if t.ProjectID == projectID {
			delete(s.tasks, id)
		}
	}
}

func (s *Storage) CreateSharedData(d *models.SharedData) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if d.ID == "" {
		s.nextID++
		d.ID = "data-" + string(rune('0'+s.nextID-1))
	}
	s.sharedData[d.ID] = d
}

func (s *Storage) GetSharedData(id string) (*models.SharedData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	d, ok := s.sharedData[id]
	return d, ok
}

func (s *Storage) GetSharedDataByProject(projectID string) []*models.SharedData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*models.SharedData{}
	for _, d := range s.sharedData {
		if d.ProjectID == projectID {
			result = append(result, d)
		}
	}
	return result
}

func (s *Storage) GetAllSharedData() []*models.SharedData {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.SharedData, 0, len(s.sharedData))
	for _, d := range s.sharedData {
		result = append(result, d)
	}
	return result
}

func (s *Storage) CreateDownloadLog(log *models.DownloadLog) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if log.ID == "" {
		s.nextID++
		log.ID = "log-" + string(rune('0'+s.nextID-1))
	}
	s.downloadLogs[log.ID] = log
}

func (s *Storage) GetDownloadLogsByData(dataID string) []*models.DownloadLog {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*models.DownloadLog{}
	for _, l := range s.downloadLogs {
		if l.DataID == dataID {
			result = append(result, l)
		}
	}
	return result
}

func (s *Storage) CreateAchievement(a *models.Achievement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		s.nextID++
		a.ID = "ach-" + string(rune('0'+s.nextID-1))
	}
	s.achievements[a.ID] = a
}

func (s *Storage) GetAchievement(id string) (*models.Achievement, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.achievements[id]
	return a, ok
}

func (s *Storage) GetAllAchievements() []*models.Achievement {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.Achievement, 0, len(s.achievements))
	for _, a := range s.achievements {
		result = append(result, a)
	}
	return result
}

func (s *Storage) UpdateAchievement(a *models.Achievement) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.achievements[a.ID] = a
}

func (s *Storage) DeleteAchievementContribution(projectID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.achievements {
		newContribs := []models.ProjectContribution{}
		for _, c := range a.Contributions {
			if c.ProjectID != projectID {
				newContribs = append(newContribs, c)
			}
		}
		a.Contributions = newContribs
	}
}

func (s *Storage) CreateApproval(a *models.Approval) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if a.ID == "" {
		s.nextID++
		a.ID = "app-" + string(rune('0'+s.nextID-1))
	}
	s.approvals[a.ID] = a
}

func (s *Storage) GetApproval(id string) (*models.Approval, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.approvals[id]
	return a, ok
}

func (s *Storage) UpdateApproval(a *models.Approval) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.approvals[a.ID] = a
}

func (s *Storage) CreateTodo(t *models.TodoItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if t.ID == "" {
		s.nextID++
		t.ID = "todo-" + string(rune('0'+s.nextID-1))
	}
	s.todos[t.ID] = t
}

func (s *Storage) GetTodosByUser(userID string) []*models.TodoItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*models.TodoItem{}
	for _, t := range s.todos {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result
}

func (s *Storage) IsProjectMember(userID string, projectID string) bool {
	project, ok := s.GetProject(projectID)
	if !ok {
		return false
	}
	user, ok := s.GetUser(userID)
	if !ok {
		return false
	}
	if user.DepartmentID == project.LeadDepartment {
		return true
	}
	for _, dept := range project.ParticipatingDepts {
		if user.DepartmentID == dept {
			return true
		}
	}
	return false
}

func (s *Storage) IsInProjectDepartment(userID string, projectID string) bool {
	return s.IsProjectMember(userID, projectID)
}

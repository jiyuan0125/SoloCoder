package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"paper-management-platform/models"
)

type PaperStore struct {
	mu     sync.RWMutex
	papers map[string]*models.Paper
	users  map[string]*models.User
}

var store *PaperStore

func init() {
	store = &PaperStore{
		papers: make(map[string]*models.Paper),
		users:  make(map[string]*models.User),
	}
	store.seedUsers()
}

func (s *PaperStore) seedUsers() {
	s.users["author1"] = &models.User{ID: "author1", Name: "张三 (作者)", Role: models.RoleAuthor}
	s.users["author2"] = &models.User{ID: "author2", Name: "李四 (作者)", Role: models.RoleAuthor}
	s.users["editor1"] = &models.User{ID: "editor1", Name: "王编辑", Role: models.RoleEditor}
	s.users["reviewer1"] = &models.User{ID: "reviewer1", Name: "赵专家", Role: models.RoleReviewer}
	s.users["reviewer2"] = &models.User{ID: "reviewer2", Name: "钱教授", Role: models.RoleReviewer}
	s.users["reviewer3"] = &models.User{ID: "reviewer3", Name: "孙研究员", Role: models.RoleReviewer}
	s.users["admin1"] = &models.User{ID: "admin1", Name: "系统管理员", Role: models.RoleAdmin}
}

func GetStore() *PaperStore {
	return store
}

func (s *PaperStore) GetUser(id string) (*models.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, exists := s.users[id]
	return user, exists
}

func (s *PaperStore) ListUsers() []*models.User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	users := make([]*models.User, 0, len(s.users))
	for _, u := range s.users {
		users = append(users, u)
	}
	return users
}

func (s *PaperStore) CreatePaper(paper *models.Paper) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.checkTitleUnique(paper.Title, ""); err != nil {
		return err
	}

	s.papers[paper.ID] = paper
	return nil
}

func (s *PaperStore) GetPaper(id string) (*models.Paper, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	paper, exists := s.papers[id]
	if !exists {
		return nil, false
	}
	return s.clonePaper(paper), true
}

func (s *PaperStore) ListPapers() []*models.Paper {
	s.mu.RLock()
	defer s.mu.RUnlock()
	papers := make([]*models.Paper, 0, len(s.papers))
	for _, p := range s.papers {
		papers = append(papers, s.clonePaper(p))
	}
	return papers
}

func (s *PaperStore) UpdatePaper(id string, updater func(*models.Paper) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	paper, exists := s.papers[id]
	if !exists {
		return errors.New("paper not found")
	}

	if err := updater(paper); err != nil {
		return err
	}
	return nil
}

func (s *PaperStore) checkTitleUnique(title, excludeID string) error {
	for id, p := range s.papers {
		if id == excludeID {
			continue
		}
		if strings.EqualFold(p.Title, title) {
			return errors.New("title already exists")
		}
	}
	return nil
}

func (s *PaperStore) CheckTitleUnique(title, excludeID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.checkTitleUnique(title, excludeID)
}

func (s *PaperStore) DeletePaper(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.papers[id]; !exists {
		return false
	}
	delete(s.papers, id)
	return true
}

func (s *PaperStore) clonePaper(p *models.Paper) *models.Paper {
	clone := *p
	clone.Keywords = append([]string(nil), p.Keywords...)
	clone.Authors = append([]models.Author(nil), p.Authors...)
	clone.AssignedReviewers = append([]models.Reviewer(nil), p.AssignedReviewers...)
	clone.Reviews = append([]models.Review(nil), p.Reviews...)
	clone.VersionHistory = append([]models.VersionHistory(nil), p.VersionHistory...)
	if p.PublicationInfo != nil {
		pubInfo := *p.PublicationInfo
		clone.PublicationInfo = &pubInfo
	}
	return &clone
}

func GenerateID() string {
	return fmt.Sprintf("p%d", time.Now().UnixNano())
}

func GenerateReviewID() string {
	return fmt.Sprintf("r%d", time.Now().UnixNano())
}

func GenerateHistoryID() string {
	return fmt.Sprintf("h%d", time.Now().UnixNano())
}

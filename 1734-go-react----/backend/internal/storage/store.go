package storage

import (
	"fmt"
	"school-connect/internal/models"
	"sync"
	"time"
)

type MemoryStore struct {
	mu sync.RWMutex

	Users         map[string]*models.User
	Parents       map[string]*models.Parent
	Students      map[string]*models.Student
	Teachers      map[string]*models.Teacher

	Announcements    map[string]*models.Announcement
	AnnouncementReads map[string][]*models.AnnouncementRead

	Scores         map[string]*models.ExamScore

	Messages       map[string]*models.Message
	Conversations  map[string]*models.Conversation
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		Users:         make(map[string]*models.User),
		Parents:       make(map[string]*models.Parent),
		Students:      make(map[string]*models.Student),
		Teachers:      make(map[string]*models.Teacher),
		Announcements:    make(map[string]*models.Announcement),
		AnnouncementReads: make(map[string][]*models.AnnouncementRead),
		Scores:         make(map[string]*models.ExamScore),
		Messages:       make(map[string]*models.Message),
		Conversations:  make(map[string]*models.Conversation),
	}
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

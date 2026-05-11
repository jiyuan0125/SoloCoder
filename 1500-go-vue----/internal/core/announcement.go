package core

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AnnouncementService struct {
	store *Store
}

func NewAnnouncementService(store *Store) *AnnouncementService {
	return &AnnouncementService{store: store}
}

func (s *AnnouncementService) CreateAnnouncement(title, content string, days int) (*Announcement, error) {
	if title == "" || content == "" {
		return nil, errors.New("title and content are required")
	}
	if days <= 0 {
		return nil, errors.New("validity days must be positive")
	}

	now := time.Now()
	announcement := &Announcement{
		ID:          uuid.NewString(),
		Title:       title,
		Content:     content,
		PublishedAt: now,
		ExpireAt:    now.Add(time.Duration(days) * 24 * time.Hour),
		Status:      AnnouncementStatusActive,
	}
	s.store.AddAnnouncement(announcement)
	return announcement, nil
}

func (s *AnnouncementService) GetAnnouncement(id string) (*Announcement, error) {
	announcement := s.store.GetAnnouncement(id)
	if announcement == nil {
		return nil, errors.New("announcement not found")
	}
	return announcement, nil
}

func (s *AnnouncementService) ListAnnouncements(includeArchived bool) []*Announcement {
	s.checkExpired()
	all := s.store.GetAnnouncements()
	if includeArchived {
		return all
	}
	result := []*Announcement{}
	for _, a := range all {
		if a.Status == AnnouncementStatusActive {
			result = append(result, a)
		}
	}
	return result
}

func (s *AnnouncementService) checkExpired() {
	now := time.Now()
	for _, a := range s.store.GetAnnouncements() {
		if a.Status == AnnouncementStatusActive && now.After(a.ExpireAt) {
			a.Status = AnnouncementStatusArchived
			s.store.UpdateAnnouncement(a)
		}
	}
}

func (s *AnnouncementService) ArchiveAnnouncement(id string) (*Announcement, error) {
	announcement, err := s.GetAnnouncement(id)
	if err != nil {
		return nil, err
	}
	if announcement.Status != AnnouncementStatusActive {
		return nil, errors.New("announcement is not active")
	}
	announcement.Status = AnnouncementStatusArchived
	s.store.UpdateAnnouncement(announcement)
	return announcement, nil
}

package announcement

import (
	"errors"
	"property-management/pkg/core"
	"property-management/pkg/core/models"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store *core.Store
}

func NewService(store *core.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(title string, content string, scope models.AnnouncementScope, targetBuilding string, effectiveTime, expiryTime time.Time, createdBy string) (*models.Announcement, error) {
	s.store.Lock()
	defer s.store.Unlock()

	if title == "" || content == "" {
		return nil, errors.New("title and content are required")
	}

	if effectiveTime.After(expiryTime) {
		return nil, errors.New("effective time must be before expiry time")
	}

	if scope == models.AnnouncementScopeBuilding && targetBuilding == "" {
		return nil, errors.New("target building is required for building scope")
	}

	user, ok := s.store.Users()[createdBy]
	if !ok {
		return nil, errors.New("creator not found")
	}
	if user.Role != models.RoleAdmin && user.Role != models.RoleSupervisor {
		return nil, errors.New("only admin or supervisor can create announcements")
	}

	now := time.Now()
	announcement := &models.Announcement{
		ID:             uuid.New().String(),
		Title:          title,
		Content:        content,
		Scope:          scope,
		TargetBuilding: targetBuilding,
		EffectiveTime:  effectiveTime,
		ExpiryTime:     expiryTime,
		CreatedBy:      createdBy,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	s.store.SetAnnouncement(announcement.ID, announcement)
	return announcement, nil
}

func (s *Service) Update(id string, title string, content string, scope models.AnnouncementScope, targetBuilding string, effectiveTime, expiryTime time.Time) (*models.Announcement, error) {
	s.store.Lock()
	defer s.store.Unlock()

	announcement, ok := s.store.Announcements()[id]
	if !ok {
		return nil, errors.New("announcement not found")
	}

	if effectiveTime.After(expiryTime) {
		return nil, errors.New("effective time must be before expiry time")
	}

	if scope == models.AnnouncementScopeBuilding && targetBuilding == "" {
		return nil, errors.New("target building is required for building scope")
	}

	announcement.Title = title
	announcement.Content = content
	announcement.Scope = scope
	announcement.TargetBuilding = targetBuilding
	announcement.EffectiveTime = effectiveTime
	announcement.ExpiryTime = expiryTime
	announcement.UpdatedAt = time.Now()

	s.store.SetAnnouncement(announcement.ID, announcement)
	return announcement, nil
}

func (s *Service) Delete(id string) error {
	s.store.Lock()
	defer s.store.Unlock()

	if _, ok := s.store.Announcements()[id]; !ok {
		return errors.New("announcement not found")
	}

	delete(s.store.Announcements(), id)
	return nil
}

func (s *Service) GetByID(id string) (*models.Announcement, error) {
	s.store.RLock()
	defer s.store.RUnlock()

	announcement, ok := s.store.Announcements()[id]
	if !ok {
		return nil, errors.New("announcement not found")
	}
	return announcement, nil
}

func (s *Service) ListValid(building string) []*models.Announcement {
	s.store.RLock()
	defer s.store.RUnlock()

	now := time.Now()
	var result []*models.Announcement

	for _, announcement := range s.store.Announcements() {
		if now.Before(announcement.EffectiveTime) || now.After(announcement.ExpiryTime) {
			continue
		}

		if announcement.Scope == models.AnnouncementScopeAll {
			result = append(result, announcement)
		} else if announcement.Scope == models.AnnouncementScopeBuilding && announcement.TargetBuilding == building {
			result = append(result, announcement)
		}
	}

	return result
}

func (s *Service) ListAll() []*models.Announcement {
	s.store.RLock()
	defer s.store.RUnlock()

	var result []*models.Announcement
	for _, announcement := range s.store.Announcements() {
		result = append(result, announcement)
	}
	return result
}

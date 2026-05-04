package store

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"notification-hub/pkg/models"
)

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrDeliveryNotFound     = errors.New("delivery not found")
	ErrInvalidChannel       = errors.New("invalid channel")
)

type Store struct {
	notifications map[string]*models.Notification
	deliveries    map[string]*models.UserDelivery
	users         map[string]struct{}
	deliveriesByUser map[string][]*models.UserDelivery
	deliveriesByNotification map[string][]*models.UserDelivery
	mu            sync.RWMutex
	dataFile      string
}

func NewStore(dataFile string) *Store {
	s := &Store{
		notifications:            make(map[string]*models.Notification),
		deliveries:               make(map[string]*models.UserDelivery),
		users:                    make(map[string]struct{}),
		deliveriesByUser:         make(map[string][]*models.UserDelivery),
		deliveriesByNotification: make(map[string][]*models.UserDelivery),
		dataFile:                 dataFile,
	}
	
	s.initDefaultUsers()
	s.load()
	
	return s
}

func (s *Store) initDefaultUsers() {
	defaultUsers := []string{"user1", "user2", "user3", "user4", "user5"}
	for _, u := range defaultUsers {
		s.users[u] = struct{}{}
	}
}

type storeData struct {
	Notifications map[string]*models.Notification `json:"notifications"`
	Deliveries    map[string]*models.UserDelivery `json:"deliveries"`
	Users         map[string]struct{}              `json:"users"`
}

func (s *Store) load() {
	data, err := os.ReadFile(s.dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}

	var sd storeData
	if err := json.Unmarshal(data, &sd); err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.notifications = sd.Notifications
	s.deliveries = sd.Deliveries
	
	for id := range sd.Users {
		s.users[id] = struct{}{}
	}

	s.rebuildIndexes()
}

func (s *Store) rebuildIndexes() {
	s.deliveriesByUser = make(map[string][]*models.UserDelivery)
	s.deliveriesByNotification = make(map[string][]*models.UserDelivery)

	for _, d := range s.deliveries {
		s.deliveriesByUser[d.UserID] = append(s.deliveriesByUser[d.UserID], d)
		s.deliveriesByNotification[d.NotificationID] = append(s.deliveriesByNotification[d.NotificationID], d)
	}
}

func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sd := storeData{
		Notifications: s.notifications,
		Deliveries:    s.deliveries,
		Users:         s.users,
	}

	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.dataFile, data, 0644)
}

func (s *Store) CreateNotification(n *models.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.notifications[n.ID] = n
	return s.Save()
}

func (s *Store) GetNotification(id string) (*models.Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n, ok := s.notifications[id]
	if !ok {
		return nil, ErrNotificationNotFound
	}
	return n, nil
}

func (s *Store) GetAllNotifications() []*models.Notification {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Notification, 0, len(s.notifications))
	for _, n := range s.notifications {
		result = append(result, n)
	}
	return result
}

func (s *Store) CreateDelivery(d *models.UserDelivery) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.deliveries[d.ID] = d
	s.deliveriesByUser[d.UserID] = append(s.deliveriesByUser[d.UserID], d)
	s.deliveriesByNotification[d.NotificationID] = append(s.deliveriesByNotification[d.NotificationID], d)
	return s.Save()
}

func (s *Store) GetDelivery(id string) (*models.UserDelivery, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.deliveries[id]
	if !ok {
		return nil, ErrDeliveryNotFound
	}
	return d, nil
}

func (s *Store) UpdateDeliveryStatus(id string, status models.DeliveryStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.deliveries[id]
	if !ok {
		return ErrDeliveryNotFound
	}

	d.Status = status
	d.UpdatedAt = time.Now()
	return s.Save()
}

func (s *Store) MarkAsRead(deliveryID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, ok := s.deliveries[deliveryID]
	if !ok {
		return ErrDeliveryNotFound
	}

	if d.UserID != userID {
		return errors.New("unauthorized")
	}

	if d.Channel != models.ChannelInSite {
		return errors.New("only insite notifications can be marked as read")
	}

	d.IsRead = true
	d.UpdatedAt = time.Now()
	return s.Save()
}

func (s *Store) GetUserDeliveries(userID string, isRead *bool) []*models.UserDelivery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	deliveries := s.deliveriesByUser[userID]
	if isRead == nil {
		return deliveries
	}

	result := make([]*models.UserDelivery, 0)
	for _, d := range deliveries {
		if d.IsRead == *isRead {
			result = append(result, d)
		}
	}
	return result
}

func (s *Store) GetDeliveriesByNotification(notificationID string) []*models.UserDelivery {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.deliveriesByNotification[notificationID]
}

func (s *Store) GetAllUsers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]string, 0, len(s.users))
	for u := range s.users {
		result = append(result, u)
	}
	return result
}

func (s *Store) UserExists(userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.users[userID]
	return ok
}

func (s *Store) GetStatistics() []models.Statistics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Statistics, 0, len(s.notifications))
	for _, n := range s.notifications {
		stats := models.Statistics{
			NotificationID: n.ID,
			Title:          n.Title,
			TotalUsers:     n.TotalUsers,
			ChannelStats:   make(map[models.Channel]models.ChannelStat),
		}

		for _, ch := range n.Channels {
			stats.ChannelStats[ch] = models.ChannelStat{
				Channel: ch,
			}
		}

		deliveries := s.deliveriesByNotification[n.ID]
		for _, d := range deliveries {
			chStat := stats.ChannelStats[d.Channel]
			chStat.Total++

			switch d.Status {
			case models.StatusDelivered:
				chStat.Delivered++
			case models.StatusFailed:
				chStat.Failed++
			default:
				chStat.Pending++
			}

			stats.ChannelStats[d.Channel] = chStat
		}

		result = append(result, stats)
	}

	return result
}

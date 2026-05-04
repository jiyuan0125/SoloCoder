package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"notification-hub/internal/server/store"
	"notification-hub/pkg/models"
)

var (
	ErrEmptyTitle        = errors.New("title cannot be empty")
	ErrTitleTooLong      = errors.New("title cannot exceed 50 characters")
	ErrNoChannels        = errors.New("at least one channel must be specified")
	ErrNoTargetUsers     = errors.New("no target users specified")
	ErrInvalidChannel    = errors.New("invalid channel")
)

type NotificationService struct {
	store      *store.Store
	dispatcher *Dispatcher
}

func NewNotificationService(s *store.Store) *NotificationService {
	return &NotificationService{
		store:      s,
		dispatcher: NewDispatcher(s),
	}
}

func (s *NotificationService) CreateNotification(req *models.CreateNotificationRequest) (*models.CreateNotificationResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	targetUsers := s.getTargetUsers(req)
	if len(targetUsers) == 0 {
		return nil, ErrNoTargetUsers
	}

	notification := &models.Notification{
		ID:          generateID(),
		Title:       strings.TrimSpace(req.Title),
		Content:     req.Content,
		Channels:    req.Channels,
		TargetAll:   req.TargetAll,
		TargetUsers: targetUsers,
		TotalUsers:  len(targetUsers),
		CreatedAt:   time.Now(),
	}

	if err := s.store.CreateNotification(notification); err != nil {
		return nil, err
	}

	s.dispatcher.Dispatch(notification)

	return &models.CreateNotificationResponse{
		NotificationID: notification.ID,
		TotalUsers:     notification.TotalUsers,
	}, nil
}

func (s *NotificationService) validateRequest(req *models.CreateNotificationRequest) error {
	trimmedTitle := strings.TrimSpace(req.Title)
	if trimmedTitle == "" {
		return ErrEmptyTitle
	}
	if len(trimmedTitle) > 50 {
		return ErrTitleTooLong
	}

	if len(req.Channels) == 0 {
		return ErrNoChannels
	}

	validChannels := map[models.Channel]bool{
		models.ChannelInSite: true,
		models.ChannelEmail:  true,
		models.ChannelSMS:    true,
	}

	for _, ch := range req.Channels {
		if !validChannels[ch] {
			return ErrInvalidChannel
		}
	}

	return nil
}

func (s *NotificationService) getTargetUsers(req *models.CreateNotificationRequest) []string {
	if req.TargetAll {
		return s.store.GetAllUsers()
	}

	uniqueUsers := make(map[string]struct{})
	for _, u := range req.TargetUsers {
		trimmed := strings.TrimSpace(u)
		if trimmed != "" && s.store.UserExists(trimmed) {
			uniqueUsers[trimmed] = struct{}{}
		}
	}

	result := make([]string, 0, len(uniqueUsers))
	for u := range uniqueUsers {
		result = append(result, u)
	}

	return result
}

func (s *NotificationService) GetStatistics() []models.Statistics {
	return s.store.GetStatistics()
}

func (s *NotificationService) GetUserNotifications(userID string, isRead *bool) ([]models.UserNotification, error) {
	if !s.store.UserExists(userID) {
		return nil, errors.New("user not found")
	}

	deliveries := s.store.GetUserDeliveries(userID, isRead)
	result := make([]models.UserNotification, 0, len(deliveries))

	for _, d := range deliveries {
		n, err := s.store.GetNotification(d.NotificationID)
		if err != nil {
			continue
		}

		result = append(result, models.UserNotification{
			ID:             d.ID,
			NotificationID: d.NotificationID,
			Title:          n.Title,
			Content:        n.Content,
			Channel:        d.Channel,
			Status:         d.Status,
			IsRead:         d.IsRead,
			CreatedAt:      d.CreatedAt,
		})
	}

	return result, nil
}

func (s *NotificationService) MarkAsRead(userID, deliveryID string) error {
	return s.store.MarkAsRead(deliveryID, userID)
}

type Dispatcher struct {
	store      *store.Store
	channelSenders map[models.Channel]Sender
}

func NewDispatcher(s *store.Store) *Dispatcher {
	d := &Dispatcher{
		store:           s,
		channelSenders: make(map[models.Channel]Sender),
	}

	d.channelSenders[models.ChannelInSite] = &InSiteSender{}
	d.channelSenders[models.ChannelEmail] = &EmailSender{}
	d.channelSenders[models.ChannelSMS] = &SMSSender{}

	return d
}

func (d *Dispatcher) Dispatch(notification *models.Notification) {
	var wg sync.WaitGroup

	for _, channel := range notification.Channels {
		wg.Add(1)
		go func(ch models.Channel, n *models.Notification) {
			defer wg.Done()
			d.sendToChannel(ch, n)
		}(channel, notification)
	}

	go func() {
		wg.Wait()
		log.Printf("Notification %s dispatch completed for all channels", notification.ID)
	}()
}

func (d *Dispatcher) sendToChannel(channel models.Channel, notification *models.Notification) {
	sender, ok := d.channelSenders[channel]
	if !ok {
		log.Printf("Unknown channel: %s", channel)
		return
	}

	for _, userID := range notification.TargetUsers {
		delivery := &models.UserDelivery{
			ID:             generateID(),
			NotificationID: notification.ID,
			UserID:         userID,
			Channel:        channel,
			Status:         models.StatusPending,
			IsRead:         false,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := d.store.CreateDelivery(delivery); err != nil {
			log.Printf("Failed to create delivery for user %s, channel %s: %v", userID, channel, err)
			continue
		}

		go func(deliv *models.UserDelivery, s Sender) {
			if err := s.Send(deliv, notification); err != nil {
				log.Printf("Failed to send to user %s via %s: %v", deliv.UserID, deliv.Channel, err)
				d.store.UpdateDeliveryStatus(deliv.ID, models.StatusFailed)
			}
		}(delivery, sender)
	}
}

type Sender interface {
	Send(delivery *models.UserDelivery, notification *models.Notification) error
}

type InSiteSender struct{}

func (s *InSiteSender) Send(delivery *models.UserDelivery, notification *models.Notification) error {
	log.Printf("[InSite] Sending notification to user %s: %s", delivery.UserID, notification.Title)
	delivery.Status = models.StatusDelivered
	delivery.UpdatedAt = time.Now()
	return nil
}

type EmailSender struct{}

func (s *EmailSender) Send(delivery *models.UserDelivery, notification *models.Notification) error {
	log.Printf("[Email] Sending email to user %s: %s", delivery.UserID, notification.Title)
	delivery.Status = models.StatusSending
	delivery.UpdatedAt = time.Now()
	
	go func() {
		time.Sleep(2 * time.Second)
		log.Printf("[Email] Email sent to user %s: %s (simulated success)", delivery.UserID, notification.Title)
	}()
	
	return nil
}

type SMSSender struct{}

func (s *SMSSender) Send(delivery *models.UserDelivery, notification *models.Notification) error {
	log.Printf("[SMS] Sending SMS to user %s: %s", delivery.UserID, notification.Title)
	delivery.Status = models.StatusSending
	delivery.UpdatedAt = time.Now()
	
	go func() {
		time.Sleep(1 * time.Second)
		log.Printf("[SMS] SMS sent to user %s: %s (simulated success)", delivery.UserID, notification.Title)
	}()
	
	return nil
}

func generateID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return time.Now().Format("20060102150405")
	}
	return hex.EncodeToString(bytes)
}

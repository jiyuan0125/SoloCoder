package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// 存储接口
type Storage interface {
	// 通知相关
	SaveNotification(n *Notification) error
	GetNotification(id string) (*Notification, error)
	
	// 送达记录相关
	SaveDeliveryRecord(r *DeliveryRecord) error
	GetDeliveryRecordsByNotificationID(notificationID string) ([]*DeliveryRecord, error)
	GetDeliveryRecordsByUserID(userID string) ([]*DeliveryRecord, error)
	GetDeliveryRecordsByUserIDAndChannel(userID string, channel Channel) ([]*DeliveryRecord, error)
	UpdateDeliveryRecordStatus(id string, status DeliveryStatus) error
	MarkAsRead(id string) error
	
	// 持久化
	Load() error
	Save() error
}

// 内存存储实现，带文件持久化
type FileStorage struct {
	notifications   map[string]*Notification
	deliveryRecords map[string]*DeliveryRecord
	filePath        string
	mu              sync.RWMutex
}

// 创建新的存储实例
func NewStorage() (*FileStorage, error) {
	// 获取当前目录下的 data 目录
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	
	filePath := filepath.Join(dataDir, "notifications.json")
	
	fs := &FileStorage{
		notifications:   make(map[string]*Notification),
		deliveryRecords: make(map[string]*DeliveryRecord),
		filePath:        filePath,
	}
	
	// 尝试从文件加载数据
	if err := fs.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	
	return fs, nil
}

// 从文件加载数据
func (fs *FileStorage) Load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	data, err := os.ReadFile(fs.filePath)
	if err != nil {
		return err
	}
	
	// 定义一个临时结构来存储加载的数据
	var storageData struct {
		Notifications   []*Notification   `json:"notifications"`
		DeliveryRecords []*DeliveryRecord `json:"delivery_records"`
	}
	
	if err := json.Unmarshal(data, &storageData); err != nil {
		return err
	}
	
	// 重建映射
	fs.notifications = make(map[string]*Notification)
	for _, n := range storageData.Notifications {
		fs.notifications[n.ID] = n
	}
	
	fs.deliveryRecords = make(map[string]*DeliveryRecord)
	for _, r := range storageData.DeliveryRecords {
		fs.deliveryRecords[r.ID] = r
	}
	
	return nil
}

// 保存数据到文件
func (fs *FileStorage) Save() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	// 构建要保存的数据结构
	storageData := struct {
		Notifications   []*Notification   `json:"notifications"`
		DeliveryRecords []*DeliveryRecord `json:"delivery_records"`
	}{}
	
	// 添加通知
	for _, n := range fs.notifications {
		storageData.Notifications = append(storageData.Notifications, n)
	}
	
	// 添加送达记录
	for _, r := range fs.deliveryRecords {
		storageData.DeliveryRecords = append(storageData.DeliveryRecords, r)
	}
	
	// 序列化
	data, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		return err
	}
	
	// 写入文件
	return os.WriteFile(fs.filePath, data, 0644)
}

// 保存通知
func (fs *FileStorage) SaveNotification(n *Notification) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if n.ID == "" {
		n.ID = generateID()
	}
	n.CreatedAt = time.Now()
	
	fs.notifications[n.ID] = n
	
	// 立即持久化
	go fs.Save()
	
	return nil
}

// 获取通知
func (fs *FileStorage) GetNotification(id string) (*Notification, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	n, ok := fs.notifications[id]
	if !ok {
		return nil, errors.New("notification not found")
	}
	
	return n, nil
}

// 保存送达记录
func (fs *FileStorage) SaveDeliveryRecord(r *DeliveryRecord) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	if r.ID == "" {
		r.ID = generateID()
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	
	fs.deliveryRecords[r.ID] = r
	
	// 立即持久化
	go fs.Save()
	
	return nil
}

// 根据通知ID获取送达记录
func (fs *FileStorage) GetDeliveryRecordsByNotificationID(notificationID string) ([]*DeliveryRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	var records []*DeliveryRecord
	for _, r := range fs.deliveryRecords {
		if r.NotificationID == notificationID {
			records = append(records, r)
		}
	}
	
	return records, nil
}

// 根据用户ID获取送达记录
func (fs *FileStorage) GetDeliveryRecordsByUserID(userID string) ([]*DeliveryRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	var records []*DeliveryRecord
	for _, r := range fs.deliveryRecords {
		if r.UserID == userID {
			records = append(records, r)
		}
	}
	
	return records, nil
}

// 根据用户ID和渠道获取送达记录
func (fs *FileStorage) GetDeliveryRecordsByUserIDAndChannel(userID string, channel Channel) ([]*DeliveryRecord, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	
	var records []*DeliveryRecord
	for _, r := range fs.deliveryRecords {
		if r.UserID == userID && r.Channel == channel {
			records = append(records, r)
		}
	}
	
	return records, nil
}

// 更新送达记录状态
func (fs *FileStorage) UpdateDeliveryRecordStatus(id string, status DeliveryStatus) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	r, ok := fs.deliveryRecords[id]
	if !ok {
		return errors.New("delivery record not found")
	}
	
	r.Status = status
	r.UpdatedAt = time.Now()
	
	// 立即持久化
	go fs.Save()
	
	return nil
}

// 标记为已读
func (fs *FileStorage) MarkAsRead(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	
	r, ok := fs.deliveryRecords[id]
	if !ok {
		return errors.New("delivery record not found")
	}
	
	// 只有站内信可以标记已读
	if r.Channel != ChannelInbox {
		return errors.New("only inbox notifications can be marked as read")
	}
	
	r.IsRead = true
	r.UpdatedAt = time.Now()
	
	// 立即持久化
	go fs.Save()
	
	return nil
}

// 生成唯一ID
func generateID() string {
	return uuid.New().String()
}

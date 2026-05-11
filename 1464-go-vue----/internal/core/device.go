package core

import (
	"firemanagement/pkg/api"
	"time"
)

type DeviceService struct {
	store *Store
}

func NewDeviceService(store *Store) *DeviceService {
	return &DeviceService{store: store}
}

func (s *DeviceService) CreateDevice(req api.CreateDeviceRequest) (string, error) {
	if !req.Type.Valid() {
		return "", ErrInvalidDeviceType
	}

	device := api.Device{
		Code:          req.Code,
		Type:          req.Type,
		Location:      req.Location,
		InstallDate:   req.InstallDate,
		ExpiryDate:    req.ExpiryDate,
		LastCheckDate: req.LastCheckDate,
		Status:        api.DeviceStatusNormal,
	}

	return s.store.AddDevice(device)
}

func (s *DeviceService) UpdateDevice(id string, req api.UpdateDeviceRequest) error {
	return s.store.UpdateDevice(id, func(d *api.Device) {
		if req.Code != nil {
			d.Code = *req.Code
		}
		if req.Location != nil {
			d.Location = *req.Location
		}
		if req.InstallDate != nil {
			d.InstallDate = *req.InstallDate
		}
		if req.ExpiryDate != nil {
			d.ExpiryDate = *req.ExpiryDate
		}
		if req.LastCheckDate != nil {
			d.LastCheckDate = *req.LastCheckDate
		}
		if req.Status != nil {
			d.Status = *req.Status
		}
	})
}

func (s *DeviceService) GetDevice(id string) (*api.Device, error) {
	device, exists := s.store.GetDevice(id)
	if !exists {
		return nil, ErrNotFound
	}
	return device, nil
}

func (s *DeviceService) ListDevices(req api.ListDevicesRequest) []api.Device {
	return s.store.ListDevices(func(d api.Device) bool {
		if req.Type != nil && d.Type != *req.Type {
			return false
		}
		if req.Location != nil && d.Location != *req.Location {
			return false
		}
		if req.Status != nil && d.Status != *req.Status {
			return false
		}
		return true
	})
}

func (s *DeviceService) CheckExpiringDevices() []api.Device {
	now := time.Now()
	threshold := now.AddDate(0, 0, 30)

	var expiring []api.Device
	for _, device := range s.store.GetAllDevices() {
		if device.Status == api.DeviceStatusNormal &&
			device.ExpiryDate.After(now) &&
			!device.ExpiryDate.After(threshold) {
			expiring = append(expiring, device)
		}
	}
	return expiring
}

func (s *DeviceService) CreateExpiryReminder(deviceID string) (string, error) {
	if s.store.ReminderExists(deviceID, api.ReminderTypeDeviceExpiry) {
		return "", nil
	}

	device, exists := s.store.GetDevice(deviceID)
	if !exists {
		return "", ErrNotFound
	}

	reminder := api.Reminder{
		Type:        api.ReminderTypeDeviceExpiry,
		Title:       "设备即将过期提醒",
		Content:     "设备 " + device.Code + "(" + device.Type.String() + ") 即将于 " + device.ExpiryDate.Format("2006-01-02") + " 过期，请及时处理",
		ReferenceID: deviceID,
		CreatedAt:   time.Now(),
		Read:        false,
	}

	return s.store.AddReminder(reminder), nil
}



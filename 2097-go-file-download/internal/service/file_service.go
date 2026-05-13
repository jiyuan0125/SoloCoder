package service

import (
	"time"

	"filedownload/internal/model"
	"filedownload/internal/store"
)

type FileService struct {
	store   *store.SQLiteStore
	storage *StorageService
}

func NewFileService(s *store.SQLiteStore, st *StorageService) *FileService {
	return &FileService{store: s, storage: st}
}

func (fs *FileService) Store() *store.SQLiteStore {
	return fs.store
}

func (fs *FileService) Storage() *StorageService {
	return fs.storage
}

func (fs *FileService) CreateFile(userID, filename string, expireHours int, maxDownloads int) (*model.File, error) {
	if expireHours <= 0 {
		return nil, ErrInvalidExpireTime
	}
	if maxDownloads < 0 {
		return nil, ErrInvalidDownloadLimit
	}

	now := time.Now()
	file := &model.File{
		ID:              generateID(),
		UserID:          userID,
		Filename:        filename,
		UploadTime:      now,
		ExpireTime:      now.Add(time.Duration(expireHours) * time.Hour),
		MaxDownloads:    maxDownloads,
		CurrentDownload: 0,
	}

	return file, nil
}

func (fs *FileService) SaveFileToDB(file *model.File) error {
	return fs.store.SaveFile(file)
}

func (fs *FileService) GetFileByID(fileID string) (*model.File, error) {
	return fs.store.GetFileByID(fileID)
}

func (fs *FileService) GetFileByPath(userID, filename string) (*model.File, error) {
	return fs.store.GetFileByUserAndPath(userID, filename)
}

func (fs *FileService) CanDownload(file *model.File) bool {
	return !fs.store.IsLinkExpired(file)
}

func (fs *FileService) RecordDownload(file *model.File, ip, userAgent string) error {
	if err := fs.store.UpdateDownloadCount(file.ID); err != nil {
		return err
	}

	record := &model.DownloadRecord{
		FileID:    file.ID,
		UserID:    file.UserID,
		Filename:  file.Filename,
		IP:        ip,
		UserAgent: userAgent,
		Time:      time.Now(),
	}
	return fs.store.SaveDownloadRecord(record)
}

func (fs *FileService) ListUserFiles(userID string) ([]*model.File, error) {
	return fs.store.ListUserFiles(userID)
}

func (fs *FileService) SearchFiles(userID, keyword string) ([]*model.File, error) {
	return fs.store.SearchFiles(userID, keyword)
}

func (fs *FileService) ListAllFiles() ([]*model.File, error) {
	return fs.store.ListAllFiles()
}

func (fs *FileService) ListAllDownloadRecords() ([]*model.DownloadRecord, error) {
	return fs.store.ListAllDownloadRecords()
}

var (
	ErrInvalidExpireTime    = &ServiceError{Code: 400, Message: "expire time must be greater than 0"}
	ErrInvalidDownloadLimit = &ServiceError{Code: 400, Message: "download limit must be non-negative"}
)

type ServiceError struct {
	Code    int
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

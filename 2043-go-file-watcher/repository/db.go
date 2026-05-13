package repository

import (
	"file-watcher/models"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dir := "./data"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	dbPath := filepath.Join(dir, "watcher.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.WatchDirectory{}, &models.FileEvent{}); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	DB = db
	log.Printf("Database initialized at: %s", dbPath)
	return nil
}

func AddWatchDirectory(path string, recursive bool, callback string) (*models.WatchDirectory, error) {
	dir := models.WatchDirectory{
		Path:      path,
		Recursive: recursive,
		Callback:  callback,
		Status:    "active",
	}

	if err := DB.Create(&dir).Error; err != nil {
		return nil, err
	}
	return &dir, nil
}

func GetWatchDirectoryByPath(path string) (*models.WatchDirectory, error) {
	var dir models.WatchDirectory
	err := DB.Where("path = ?", path).First(&dir).Error
	if err != nil {
		return nil, err
	}
	return &dir, nil
}

func GetWatchDirectoryByID(id uint) (*models.WatchDirectory, error) {
	var dir models.WatchDirectory
	err := DB.Where("id = ?", id).First(&dir).Error
	if err != nil {
		return nil, err
	}
	return &dir, nil
}

func GetAllActiveWatchDirectories() ([]models.WatchDirectory, error) {
	var dirs []models.WatchDirectory
	err := DB.Where("status = ?", "active").Find(&dirs).Error
	return dirs, err
}

func UpdateWatchDirectoryStatus(id uint, status string) error {
	return DB.Model(&models.WatchDirectory{}).Where("id = ?", id).Update("status", status).Error
}

func DeleteWatchDirectory(id uint) error {
	return DB.Delete(&models.WatchDirectory{}, id).Error
}

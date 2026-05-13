package handlers

import (
	"file-watcher/models"
	"file-watcher/repository"
	"file-watcher/watcher"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AddWatchRequest struct {
	Path      string `json:"path" binding:"required"`
	Recursive bool   `json:"recursive"`
	Callback  string `json:"callback" binding:"required"`
}

func AddWatch(c *gin.Context) {
	var req AddWatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	if _, err := repository.GetWatchDirectoryByPath(req.Path); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Directory already being watched"})
		return
	}

	mgr := watcher.GetManager()
	dw, err := mgr.AddWatch(req.Path, req.Recursive, req.Callback)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":        dw.ID,
		"path":      dw.Path,
		"recursive": dw.Recursive,
		"callback":  dw.Callback,
		"status":    dw.Status,
	})
}

func RemoveWatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid watch ID"})
		return
	}

	if _, err := repository.GetWatchDirectoryByID(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Watch not found"})
		return
	}

	mgr := watcher.GetManager()
	if err := mgr.RemoveWatch(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Watch removed successfully"})
}

func ListWatches(c *gin.Context) {
	dirs, err := repository.DB.Find(&[]models.WatchDirectory{}).Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer dirs.Close()

	var watches []models.WatchDirectory
	for dirs.Next() {
		var watch models.WatchDirectory
		if err := repository.DB.ScanRows(dirs, &watch); err != nil {
			continue
		}
		watches = append(watches, watch)
	}

	c.JSON(http.StatusOK, gin.H{"watches": watches})
}

func GetWatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid watch ID"})
		return
	}

	watch, err := repository.GetWatchDirectoryByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Watch not found"})
		return
	}

	c.JSON(http.StatusOK, watch)
}

package config

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type fileWatcher struct {
	watcher      *fsnotify.Watcher
	filePath     string
	stableTime   time.Duration
	onChange     func()
	ctx          context.Context
	cancel       context.CancelFunc
	lastSize     int64
	lastModTime  time.Time
}

func newFileWatcher(filePath string, stableTime time.Duration, onChange func()) (*fileWatcher, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, err
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	fw := &fileWatcher{
		watcher:    w,
		filePath:   absPath,
		stableTime: stableTime,
		onChange:   onChange,
		ctx:        ctx,
		cancel:     cancel,
	}

	return fw, nil
}

func (fw *fileWatcher) start() error {
	dir := filepath.Dir(fw.filePath)
	if err := fw.watcher.Add(dir); err != nil {
		return err
	}

	go fw.eventLoop()
	return nil
}

func (fw *fileWatcher) stop() {
	fw.cancel()
	fw.watcher.Close()
}

func (fw *fileWatcher) eventLoop() {
	var pendingTimer *time.Timer
	pendingTimer = time.NewTimer(0)
	<-pendingTimer.C

	for {
		select {
		case <-fw.ctx.Done():
			if pendingTimer != nil && !pendingTimer.Stop() {
				select {
				case <-pendingTimer.C:
				default:
				}
			}
			return
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}

			if event.Name != fw.filePath {
				continue
			}

			if event.Op&fsnotify.Remove == fsnotify.Remove {
				continue
			}

			if pendingTimer != nil {
				pendingTimer.Stop()
			}
			pendingTimer = time.AfterFunc(fw.stableTime, func() {
				if fw.isFileStable() {
					fw.onChange()
				}
			})
		case _, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}

func (fw *fileWatcher) isFileStable() bool {
	info, err := os.Stat(fw.filePath)
	if err != nil {
		return false
	}

	if info.Size() == 0 {
		return false
	}

	if fw.lastSize == info.Size() && !info.ModTime().After(fw.lastModTime) {
		return true
	}

	fw.lastSize = info.Size()
	fw.lastModTime = info.ModTime()
	return false
}

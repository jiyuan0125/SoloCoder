package watcher

import "errors"

var (
	ErrPathRequired      = errors.New("path is required")
	ErrCallbackRequired  = errors.New("callback is required")
	ErrPathNotDirectory  = errors.New("path is not a directory")
	ErrPathNotExist      = errors.New("path does not exist")
	ErrWatcherNotActive    = errors.New("watcher is not active")
)

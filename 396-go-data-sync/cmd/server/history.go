package main

import (
	"sync"

	"go-data-sync/pkg/protocol"
)

var (
	historyMutex sync.RWMutex
	historyList  []protocol.SyncHistory
	maxHistory   = 100
)

func addHistory(h protocol.SyncHistory) {
	historyMutex.Lock()
	defer historyMutex.Unlock()
	
	historyList = append(historyList, h)
	
	if len(historyList) > maxHistory {
		historyList = historyList[len(historyList)-maxHistory:]
	}
}

func getHistory() []protocol.SyncHistory {
	historyMutex.RLock()
	defer historyMutex.RUnlock()
	
	result := make([]protocol.SyncHistory, len(historyList))
	copy(result, historyList)
	
	return result
}

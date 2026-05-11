package main

import (
	"container/list"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"fifo-ipc/internal/common"
	"fifo-ipc/pkg/fifo"
)

type messageStore struct {
	mu            sync.RWMutex
	messages      *list.List
	maxRecent     int
	totalCount    int64
	truncateCount int64
}

func newMessageStore(maxRecent int) *messageStore {
	return &messageStore{
		messages:  list.New(),
		maxRecent: maxRecent,
	}
}

func (s *messageStore) add(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalCount++

	info := common.MessageInfo{
		ID:        s.totalCount,
		Content:   msg,
		Timestamp: time.Now(),
		Size:      len(msg),
	}

	s.messages.PushBack(info)
	if s.messages.Len() > s.maxRecent {
		s.messages.Remove(s.messages.Front())
	}
}

func (s *messageStore) addTruncation() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.truncateCount++
}

func (s *messageStore) getRecent(n int) []common.MessageInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if n <= 0 || n > s.messages.Len() {
		n = s.messages.Len()
	}

	result := make([]common.MessageInfo, 0, n)
	i := 0
	e := s.messages.Back()
	for e != nil && i < n {
		result = append([]common.MessageInfo{e.Value.(common.MessageInfo)}, result...)
		e = e.Prev()
		i++
	}

	return result
}

func (s *messageStore) stats() (int64, int64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalCount, s.truncateCount
}

type server struct {
	store     *messageStore
	fifoPath  string
	startTime time.Time
}

func (s *server) statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	total, truncates := s.store.stats()
	recent := s.store.getRecent(100)

	resp := common.StatsResponse{
		TotalMessages:   total,
		RecentMessages:  recent,
		HasTruncation:   truncates > 0,
		TruncationCount: truncates,
		FifoPath:        s.fifoPath,
		ServerUptime:    time.Since(s.startTime).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func startFifoReader(fifoPath string, store *messageStore, done <-chan struct{}) {
	var reader *fifo.Reader
	var err error

	for {
		reader, err = fifo.NewReader(fifoPath)
		if err == nil {
			break
		}
		log.Printf("failed to open fifo for reading: %v, retrying...", err)
		select {
		case <-time.After(500 * time.Millisecond):
		case <-done:
			return
		}
	}

	defer reader.Close()
	log.Printf("started reading from fifo: %s", fifoPath)

	buf := make([]byte, 0)
	for {
		select {
		case <-done:
			log.Println("stopping fifo reader")
			return
		default:
		}

		msg, truncated, err := reader.ReadMessage()
		if err != nil {
			if err == fifo.ErrTruncated {
				store.addTruncation()
				log.Printf("message truncated: %v", err)
			} else if err != fifo.ErrPipeClosed {
				log.Printf("read error: %v", err)
			}
			time.Sleep(100 * time.Millisecond)
			continue
		}

		if truncated {
			store.addTruncation()
			log.Println("detected truncated message")
			continue
		}

		if msg == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		_ = buf
		store.add(string(msg.Data))
	}
}

func main() {
	fifoPath := flag.String("fifo", "/tmp/fifo-ipc.pipe", "path to fifo pipe")
	addr := flag.String("addr", ":8080", "http server address")
	maxRecent := flag.Int("max-recent", 1000, "maximum number of recent messages to keep")
	flag.Parse()

	if err := fifo.Create(*fifoPath); err != nil {
		log.Fatalf("failed to create fifo: %v", err)
	}

	store := newMessageStore(*maxRecent)
	srv := &server{
		store:     store,
		fifoPath:  *fifoPath,
		startTime: time.Now(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stats", srv.statsHandler)
	mux.HandleFunc("/health", srv.healthHandler)

	httpServer := &http.Server{
		Addr:    *addr,
		Handler: mux,
	}

	done := make(chan struct{})
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("shutdown signal received")
		close(done)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("http server shutdown error: %v", err)
		}
	}()

	go startFifoReader(*fifoPath, store, done)

	log.Printf("server starting on %s, fifo: %s", *addr, *fifoPath)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server error: %v", err)
	}

	fmt.Println("server stopped")
}
